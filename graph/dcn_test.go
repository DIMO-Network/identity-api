package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	test "github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func TestDCNQuery(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	logger := zerolog.New(os.Stdout)
	settings := config.Settings{
		DCNRegistryAddr:     "0xE9F4dfE02f895DC17E2e146e578873c9095bA293", // For realism.
		DIMORegistryChainID: 137,
		DCNResolverAddr:     "0x60627326F55054Ea448e0a7BC750785bD65EF757",
	}

	repo := base.NewRepository(pdb, settings, &logger)
	resolver := NewResolver(repo)

	_, wallet, err := test.GenerateWallet()
	assert.NoError(err)
	veh := models.Vehicle{
		ID:            1,
		OwnerAddress:  wallet.Bytes(),
		Make:          null.StringFrom("Toyota"),
		Model:         null.StringFrom("Corolla"),
		Year:          null.IntFrom(2000),
		DefinitionURI: null.StringFrom("mockUri"),
	}
	err = veh.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(err)

	contractEventConsumer := services.NewContractsEventsConsumer(pdb, &logger, &settings)

	err = contractEventConsumer.Process(ctx, &shared.CloudEvent[json.RawMessage]{
		Source: "chain/137",
		Type:   "zone.dimo.contract.event",
		Data: json.RawMessage(`
		{
			"contract": "0xE9F4dfE02f895DC17E2e146e578873c9095bA293",
			"eventName": "NewNode",
			"arguments": {
				"node": "ZmUlXZ4s/E7W0wZChcTSDIZK+B3A0myUxTgPZ/ndV+0=",
				"owner": "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"
			}
		}
	`)})

	require.NoError(err)
	cfg := Config{Resolvers: resolver}
	cfg.Directives.OneOf = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		return next(ctx)
	}
	c := client.New(loader.Middleware(pdb, handler.NewDefaultServer(NewExecutableSchema(cfg)), settings))

	type response struct {
		DCN struct {
			Node      string
			Owner     string
			ExpiresAt *string
			Name      *string
			Vehicle   *model.Vehicle
		}
	}

	var dcnr response
	req := map[string]string{}
	req["node"] = "0x6665255d9e2cfc4ed6d3064285c4d20c864af81dc0d26c94c5380f67f9dd57ed"
	c.MustPost(`
		query DCN($input:DCNBy!) {
			dcn(by: $input) {
				node
				owner
				expiresAt
				name
			}
		}
	`, &dcnr, client.Var("input", req))

	assert.Equal("0x6665255d9e2cfc4ed6d3064285c4d20c864af81dc0d26c94c5380f67f9dd57ed", dcnr.DCN.Node)
	assert.Equal("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045", dcnr.DCN.Owner)
	assert.Nil(dcnr.DCN.Name)
	assert.Nil(dcnr.DCN.ExpiresAt)

	currTime := time.Now().UTC().Truncate(time.Second)
	err = contractEventConsumer.Process(ctx, &shared.CloudEvent[json.RawMessage]{
		Source: "chain/137",
		Type:   "zone.dimo.contract.event",
		Data: json.RawMessage(fmt.Sprintf(`
		{
			"contract": "0xE9F4dfE02f895DC17E2e146e578873c9095bA293",
			"eventName": "NewExpiration",
			"arguments": {
				"node": "ZmUlXZ4s/E7W0wZChcTSDIZK+B3A0myUxTgPZ/ndV+0=",
				"expiration": %d
			}
		}
	`, int(currTime.Unix())))})
	require.NoError(err)

	c.MustPost(`
		query DCN($input:DCNBy!) {
			dcn(by: $input) {
				node
				owner
				expiresAt
				name
			}
		}
	`, &dcnr, client.Var("input", req))

	expected, err := time.Parse(time.RFC3339, *dcnr.DCN.ExpiresAt)
	assert.NoError(err)
	assert.Equal(expected, currTime)

	// NameChanged
	mockName := "SomeMockName"
	err = contractEventConsumer.Process(ctx, &shared.CloudEvent[json.RawMessage]{
		Source: "chain/137",
		Type:   "zone.dimo.contract.event",
		Data: json.RawMessage(fmt.Sprintf(`
	   		{
	   			"contract": "0x60627326F55054Ea448e0a7BC750785bD65EF757",
	   			"eventName": "NameChanged",
	   			"arguments": {
	   				"node": "ZmUlXZ4s/E7W0wZChcTSDIZK+B3A0myUxTgPZ/ndV+0=",
	   				"name_": "%s"
	   			}
	   		}

	   `, mockName))})
	require.NoError(err)

	c.MustPost(`
	   	query DCN($input:DCNBy!) {
	   		dcn(by: $input) {
	   			node
	   			owner
	   			expiresAt
	   			name
	   		}
	   	}

	   `, &dcnr, client.Var("input", req))

	if assert.NotNil(dcnr.DCN.Name) {
		assert.Equal(mockName, *dcnr.DCN.Name)
	}
}

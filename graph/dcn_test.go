package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDCNQuery(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	repo := repositories.New(pdb)
	resolver := NewResolver(repo)

	logger := zerolog.New(os.Stdout)
	settings := config.Settings{
		DCNRegistryAddr:     "0xE9F4dfE02f895DC17E2e146e578873c9095bA293", // For realism.
		DIMORegistryChainID: 137,
	}

	contractEventConsumer := services.NewContractsEventsConsumer(pdb, &logger, &settings)

	err := contractEventConsumer.Process(ctx, &shared.CloudEvent[json.RawMessage]{
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

	c := client.New(loader.Middleware(pdb, handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver}))))

	type response struct {
		DCN struct {
			Node      string
			Owner     string
			ExpiresAt *string
			Name      *string
		}
	}

	var dcnr response

	c.MustPost(`
		query DCN($node: Bytes!) {
			dcn(node: $node) {
				node
				owner
				expiresAt
			}
		}
	`, &dcnr, client.Var("node", "0x6665255d9e2cfc4ed6d3064285c4d20c864af81dc0d26c94c5380f67f9dd57ed"))

	assert.Equal("0x6665255d9e2cfc4ed6d3064285c4d20c864af81dc0d26c94c5380f67f9dd57ed", dcnr.DCN.Node)
	assert.Equal("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045", dcnr.DCN.Owner)
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
		query DCN($node: Bytes!) {
			dcn(node: $node) {
				node
				owner
				expiresAt
			}
		}
	`, &dcnr, client.Var("node", "0x6665255d9e2cfc4ed6d3064285c4d20c864af81dc0d26c94c5380f67f9dd57ed"))

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
			"contract": "0xE9F4dfE02f895DC17E2e146e578873c9095bA293",
			"eventName": "NameChanged",
			"arguments": {
				"node": "ZmUlXZ4s/E7W0wZChcTSDIZK+B3A0myUxTgPZ/ndV+0=",
				"_name": "%s"
			}
		}
	`, mockName))})
	require.NoError(err)

	c.MustPost(`
		query DCN($node: Bytes!) {
			dcn(node: $node) {
				node
				owner
				expiresAt
				name
			}
		}
	`, &dcnr, client.Var("node", "0x6665255d9e2cfc4ed6d3064285c4d20c864af81dc0d26c94c5380f67f9dd57ed"))

	assert.NoError(err)
	assert.Equal(*dcnr.DCN.Name, mockName)
}

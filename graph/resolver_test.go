package graph

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/ast"
)

var aftermarketDevice = models.AftermarketDevice{
	ID:             1,
	ManufacturerID: 137,
	Address:        common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf5").Bytes(),
	Owner:          common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes(),
	Serial:         null.StringFrom("aftermarketDeviceSerial-1"),
	Imei:           null.StringFrom("aftermarketDeviceIMEI-1"),
	MintedAt:       time.Now(),
	VehicleID:      null.IntFrom(11),
	Beneficiary:    common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
}

var ad2 = models.AftermarketDevice{
	ID:             100,
	ManufacturerID: 137,
	Address:        common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf5").Bytes(),
	Owner:          common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes(),
	Serial:         null.StringFrom("aftermarketDeviceSerial-1"),
	Imei:           null.StringFrom("aftermarketDeviceIMEI-1"),
	MintedAt:       time.Now(),
	Beneficiary:    common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
}

var fordMfr = models.Manufacturer{
	ID:       41,
	Owner:    common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	Name:     "Ford",
	MintedAt: time.Now(),
	Slug:     "ford",
}

var apMfr = models.Manufacturer{
	ID:       137,
	Owner:    common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDff"),
	Name:     "AutoPi",
	MintedAt: time.Now(),
	Slug:     "autopi",
}

var testVehicle = models.Vehicle{
	ID:             11,
	ManufacturerID: 41,
	OwnerAddress:   common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	Make:           null.StringFrom("Ford"),
	Model:          null.StringFrom("Bronco"),
	Year:           null.IntFrom(2022),
	MintedAt:       time.Now(),
}

var syntheticDevice = models.SyntheticDevice{
	ID:            1,
	IntegrationID: 2,
	VehicleID:     11,
	DeviceAddress: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	MintedAt:      time.Now(),
}

const migrationsDir = "../migrations"

func TestResolver(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	err := fordMfr.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	err = apMfr.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	err = testVehicle.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	err = aftermarketDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	err = ad2.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	err = syntheticDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)
	vehicleAddress := common.HexToAddress("0x123")
	settings := config.Settings{
		VehicleNFTAddr: vehicleAddress.String(),
	}
	logger := zerolog.Nop()
	repo := base.NewRepository(pdb, settings, &logger)
	resolver := NewResolver(repo)
	c := client.New(loader.Middleware(pdb, NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver})), settings))

	t.Run("ownedAftermarketDevices, return only one response", func(t *testing.T) {
		var resp interface{}
		c.MustPost(`{aftermarketDevices(filterBy: {owner: "46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"}, first: 1) {edges {node {tokenId owner}}}}`, &resp)
		b, _ := json.Marshal(resp)
		fmt.Println(string(b))
		assert.JSONEq(`{"aftermarketDevices":{"edges":[{"node":{"tokenId":100,"owner":"0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"}}]}}`, string(b))
	})

	t.Run("ownedAftermarketDevices, search after", func(t *testing.T) {
		var resp interface{}
		c.MustPost(`{aftermarketDevices(filterBy: {owner: "46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"}, after: "MQ==", first: 5) {edges {node {id owner}}}}`, &resp)
		b, _ := json.Marshal(resp)
		fmt.Println(string(b))
		assert.Equal(`{"aftermarketDevices":{"edges":[]}}`, string(b))
	})

	t.Run("ownedAftermarketDevices and linked vehicle", func(t *testing.T) {
		var resp interface{}
		c.MustPost(`{aftermarketDevices(filterBy: {owner: "46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"}, first: 5) {edges {node {tokenId owner vehicle{tokenId owner}}}}}`, &resp)
		b, _ := json.Marshal(resp)
		fmt.Println(string(b))
		assert.JSONEq(
			`{"aftermarketDevices":{"edges":[{"node":{"tokenId":100,"owner":"0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4","vehicle":null}},{"node":{"tokenId":1,"owner":"0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4","vehicle":{"tokenId":11,"owner":"0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"}}}]}}`,
			string(b))
	})

	t.Run("accessibleVehicles", func(t *testing.T) {
		var resp interface{}
		c.MustPost(`{vehicles(filterBy: {privileged: "46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"}, first: 2) {edges {node {tokenId owner}}}}`, &resp)
		b, _ := json.Marshal(resp)
		fmt.Println(string(b))
		assert.JSONEq(
			`{"vehicles":{"edges":[{"node":{"tokenId":11,"owner":"0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"}}]}}`,
			string(b))
	})

	t.Run("accessibleVehicles and syntheticDevices", func(t *testing.T) {
		var resp interface{}
		c.MustPost(`{
			vehicles(first: 5, filterBy: {privileged: "46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"}) {edges {node {tokenId owner syntheticDevice {tokenId}}}}}`, &resp)
		b, _ := json.Marshal(resp)
		fmt.Println(string(b))
		assert.JSONEq(
			`{"vehicles":{"edges":[{"node":{"tokenId":11,"owner":"0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4","syntheticDevice":{"tokenId":1}}}]}}`,
			string(b))
	})

	t.Run("vehicle query with tokenId", func(t *testing.T) {
		var resp interface{}
		c.MustPost(`{vehicle(tokenId: 11) {tokenId owner}}`, &resp)
		b, _ := json.Marshal(resp)
		fmt.Println(string(b))
		assert.JSONEq(
			`{"vehicle":{"tokenId":11,"owner":"0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"}}`,
			string(b))
	})

	t.Run("vehicle query with tokenDID", func(t *testing.T) {
		var resp interface{}
		c.MustPost(`{vehicle(tokenDID: "did:erc721:1:`+vehicleAddress.String()+`:11") {tokenId owner}}`, &resp)
		b, _ := json.Marshal(resp)
		fmt.Println(string(b))
		assert.JSONEq(
			`{"vehicle":{"tokenId":11,"owner":"0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"}}`,
			string(b))
	})

	t.Run("vehicle query with both tokenId and tokenDID should fail", func(t *testing.T) {
		var resp interface{}
		err := c.Post(`{vehicle(tokenId: 11, tokenDID: "did:erc721:1:`+vehicleAddress.String()+`:11") {tokenId owner}}`, &resp)
		assert.NotNil(err)
	})

	t.Run("vehicle query with no parameters should fail", func(t *testing.T) {
		var resp interface{}
		err := c.Post(`{vehicle {tokenId owner}}`, &resp)
		assert.NotNil(err)
	})
}

func NewDefaultServer(es graphql.ExecutableSchema) *handler.Server {
	srv := handler.New(es)

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	return srv
}

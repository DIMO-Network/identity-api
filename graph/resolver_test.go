package graph

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var aftermarketDevice = models.AftermarketDevice{
	ID:          1,
	Address:     common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf5").Bytes(),
	Owner:       common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes(),
	Serial:      null.StringFrom("aftermarketDeviceSerial-1"),
	Imei:        null.StringFrom("aftermarketDeviceIMEI-1"),
	MintedAt:    time.Now(),
	VehicleID:   null.IntFrom(11),
	Beneficiary: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
}

var ad2 = models.AftermarketDevice{
	ID:          100,
	Address:     common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf5").Bytes(),
	Owner:       common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes(),
	Serial:      null.StringFrom("aftermarketDeviceSerial-1"),
	Imei:        null.StringFrom("aftermarketDeviceIMEI-1"),
	MintedAt:    time.Now(),
	Beneficiary: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
}

var testVehicle = models.Vehicle{
	ID:           11,
	OwnerAddress: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	Make:         null.StringFrom("Ford"),
	Model:        null.StringFrom("Bronco"),
	Year:         null.IntFrom(2022),
	MintedAt:     time.Now(),
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

	err := testVehicle.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	err = aftermarketDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	err = ad2.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	err = syntheticDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	settings := config.Settings{}

	repo := repositories.New(pdb, settings)
	resolver := NewResolver(repo)
	c := client.New(loader.Middleware(pdb, handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver})), settings))

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
}

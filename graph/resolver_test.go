package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var cloudEvent = shared.CloudEvent[json.RawMessage]{
	ID:          "2SiTVhP3WBhfQQnnnpeBdMR7BSY",
	Source:      "chain/80001",
	SpecVersion: "1.0",
	Subject:     "0x4de1bcf2b7e851e31216fc07989caa902a604784",
	Time:        time.Now(),
	Type:        "zone.dimo.contract.event",
}

var contractEventData = services.ContractEventData{
	ChainID:         80001,
	EventName:       "AftermarketDeviceNodeMinted",
	Contract:        common.HexToAddress("0x4de1bcf2b7e851e31216fc07989caa902a604784"),
	TransactionHash: common.HexToHash("0x811a85e24d0129a2018c9a6668652db63d73bc6d1c76f21b07da2162c6bfea7d"),
	EventSignature:  common.HexToHash("0xd624fd4c3311e1803d230d97ce71fd60c4f658c30a31fbe08edcb211fd90f63f"),
}

var aftermarketDeviceNodeMintedArgs = services.AftermarketDeviceNodeMintedData{
	AftermarketDeviceAddress: common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	ManufacturerID:           big.NewInt(137),
	Owner:                    common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	TokenID:                  big.NewInt(42),
}

var aftermarketDevice = models.AftermarketDevice{
	ID:        1,
	Address:   null.BytesFrom(common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf5").Bytes()),
	Owner:     null.BytesFrom(common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes()),
	Serial:    null.StringFrom("aftermarketDeviceSerial-1"),
	Imei:      null.StringFrom("aftermarketDeviceIMEI-1"),
	MintedAt:  null.TimeFrom(time.Now()),
	VehicleID: null.IntFrom(11),
}

var ad2 = models.AftermarketDevice{
	ID:       100,
	Address:  null.BytesFrom(common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf5").Bytes()),
	Owner:    null.BytesFrom(common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes()),
	Serial:   null.StringFrom("aftermarketDeviceSerial-1"),
	Imei:     null.StringFrom("aftermarketDeviceIMEI-1"),
	MintedAt: null.TimeFrom(time.Now()),
}

var vehicle = models.Vehicle{
	ID:           11,
	OwnerAddress: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	Make:         null.StringFrom("Ford"),
	Model:        null.StringFrom("Bronco"),
	Year:         null.IntFrom(2022),
	MintedAt:     time.Now(),
}

func TestNew(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	pdb, _ := helpers.StartContainerDatabase(ctx, t, helpers.MigrationsDirRelPath)

	err := vehicle.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	err = aftermarketDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	err = ad2.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	repo := repositories.NewRepository(pdb, 0)
	resolver := NewResolver(repo)
	c := client.New(loader.Middleware(pdb, handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: &resolver}))))

	t.Run("ownedAftermarketDevices, return only one response", func(t *testing.T) {
		var resp interface{}
		c.MustPost(`{ownedAftermarketDevices(address: "46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4", first: 1) {edges {node {id owner}}}}`, &resp)
		b, _ := json.Marshal(resp)
		fmt.Println(string(b))
		assert.Equal(`{"ownedAftermarketDevices":{"edges":[{"node":{"id":"100","owner":"0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"}}]}}`, string(b))
	})

	t.Run("ownedAftermarketDevices, search after", func(t *testing.T) {
		var resp interface{}
		c.MustPost(`{ownedAftermarketDevices(address: "46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4" after: "MQ==") {edges {node {id owner}}}}`, &resp)
		b, _ := json.Marshal(resp)
		fmt.Println(string(b))
		assert.Equal(`{"ownedAftermarketDevices":{"edges":[]}}`, string(b))
	})

	t.Run("ownedAftermarketDevices and linked vehicle", func(t *testing.T) {
		var resp interface{}
		c.MustPost(`{ownedAftermarketDevices(address: "46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4") {edges {node {id owner vehicle{id owner}}}}}`, &resp)
		b, _ := json.Marshal(resp)
		fmt.Println(string(b))
		assert.Equal(
			`{"ownedAftermarketDevices":{"edges":[{"node":{"id":"100","owner":"0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4","vehicle":null}},{"node":{"id":"1","owner":"0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4","vehicle":{"id":"11","owner":"0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"}}}]}}`,
			string(b))
	})

	t.Run("ownedVehicles", func(t *testing.T) {
		var resp interface{}
		c.MustPost(`{ownedVehicles(address: "46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4") {edges {node {id owner}}}}`, &resp)
		b, _ := json.Marshal(resp)
		fmt.Println(string(b))
		assert.Equal(
			`{"ownedVehicles":{"edges":[{"node":{"id":"11","owner":"0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"}}]}}`,
			string(b))
	})
}

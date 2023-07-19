package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/internal/test"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
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

func TestAftermarketDeviceNodeMintSingleResponse(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", test.DBSettings.Name).Logger()

	settings, err := shared.LoadConfig[config.Settings](test.SettingsPath)
	settings.DB = test.DBSettings
	assert.NoError(t, err)

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(t, config)

	pdb, _ := test.StartContainerDatabase(ctx, t, test.MigrationsDirRelPath)
	contractEventConsumer := services.NewContractsEventsConsumer(pdb, &logger, &settings)

	argBytes, err := json.Marshal(aftermarketDeviceNodeMintedArgs)
	assert.NoError(t, err)

	contractEventData.Arguments = argBytes
	ctEventDataBytes, err := json.Marshal(contractEventData)
	assert.NoError(t, err)

	cloudEvent.Data = ctEventDataBytes
	expectedBytes, err := json.Marshal(cloudEvent)
	assert.NoError(t, err)

	consumer.ExpectConsumePartition(settings.ContractsEventTopic, 0, 0).YieldMessage(&sarama.ConsumerMessage{Value: expectedBytes})

	outputTest, err := consumer.ConsumePartition(settings.ContractsEventTopic, 0, 0)
	assert.NoError(t, err)

	m := <-outputTest.Messages()
	var e shared.CloudEvent[json.RawMessage]
	err = json.Unmarshal(m.Value, &e)
	assert.NoError(t, err)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	ad, err := models.AftermarketDevices(models.AftermarketDeviceWhere.ID.EQ(int(aftermarketDeviceNodeMintedArgs.TokenID.Int64()))).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)
	assert.Equal(t, ad.Address.Bytes, aftermarketDeviceNodeMintedArgs.Owner.Bytes())

	adController := NewVehiclesRepo(pdb)
	res, err := adController.GetOwnedAftermarketDevices(ctx, aftermarketDeviceNodeMintedArgs.Owner, nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, *res.Edges[0].Node.Address, aftermarketDeviceNodeMintedArgs.Owner)
}

func TestAftermarketDeviceNodeMintMultiResponse(t *testing.T) {
	ctx := context.Background()
	settings, err := shared.LoadConfig[config.Settings](test.SettingsPath)
	settings.DB = test.DBSettings
	assert.NoError(t, err)
	pdb, _ := test.StartContainerDatabase(ctx, t, test.MigrationsDirRelPath)

	for i := 1; i < 6; i++ {
		ad := models.AftermarketDevice{
			ID:    i,
			Owner: null.BytesFrom(aftermarketDeviceNodeMintedArgs.Owner.Bytes()),
		}

		err = ad.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	adController := NewVehiclesRepo(pdb)
	first := 2
	after := "4"
	res, err := adController.GetOwnedAftermarketDevices(ctx, aftermarketDeviceNodeMintedArgs.Owner, &first, &after)
	assert.NoError(t, err)

	a, err := strconv.Atoi(after)
	assert.NoError(t, err)

	for i := 0; i < first; i++ {
		a--
		assert.Equal(t, res.Edges[i].Node.ID, fmt.Sprintf("%d", a))
	}

}

package services

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const migrationsDirRelPath = "../../migrations"
const aftermarketDeviceAddr = "0xcf9af64522162da85164a714c23a7705e6e466b3"

var mintedAt = time.Now()

var cloudEvent = shared.CloudEvent[json.RawMessage]{
	ID:          "2SiTVhP3WBhfQQnnnpeBdMR7BSY",
	Source:      "chain/80001",
	SpecVersion: "1.0",
	Subject:     "0x4de1bcf2b7e851e31216fc07989caa902a604784",
	Time:        mintedAt,
	Type:        "zone.dimo.contract.event",
}

var contractEventData = ContractEventData{
	ChainID:         80001,
	Contract:        common.HexToAddress("0x4de1bcf2b7e851e31216fc07989caa902a604784"),
	TransactionHash: common.HexToHash("0x811a85e24d0129a2018c9a6668652db63d73bc6d1c76f21b07da2162c6bfea7d"),
	EventSignature:  common.HexToHash("0xd624fd4c3311e1803d230d97ce71fd60c4f658c30a31fbe08edcb211fd90f63f"),
	Block: Block{
		Time: mintedAt,
	},
}

func eventBytes(args interface{}, contractEventData ContractEventData, t *testing.T) []byte {

	argBytes, err := json.Marshal(args)
	assert.NoError(t, err)

	contractEventData.Arguments = argBytes
	ctEventDataBytes, err := json.Marshal(contractEventData)
	assert.NoError(t, err)

	cloudEvent.Data = ctEventDataBytes
	expectedBytes, err := json.Marshal(cloudEvent)
	assert.NoError(t, err)

	return expectedBytes
}

func TestHandleAftermarketDeviceAttributeSetEvent(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "AftermarketDeviceAttributeSet"

	var aftermarketDeviceAttributesSerial = AftermarketDeviceAttributeSetData{
		Attribute: "Serial",
		Info:      "randomgarbagevalue",
		TokenID:   big.NewInt(43),
	}

	settings := config.Settings{
		DIMORegistryAddr:    contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
	}

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(t, config)

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	expectedBytes := eventBytes(aftermarketDeviceAttributesSerial, contractEventData, t)

	d := models.AftermarketDevice{
		ID:          int(aftermarketDeviceAttributesSerial.TokenID.Int64()),
		Owner:       common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Beneficiary: common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Imei:        null.StringFrom("garbage-imei-value"),
	}
	err := d.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
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

	ad, err := models.AftermarketDevices(models.AftermarketDeviceWhere.ID.EQ(int(aftermarketDeviceAttributesSerial.TokenID.Int64()))).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Equal(t, aftermarketDeviceAttributesSerial.TokenID.Int64(), int64(ad.ID))
	assert.Equal(t, aftermarketDeviceAttributesSerial.Info, ad.Serial.String)
	assert.Equal(t, d.Imei.String, ad.Imei.String)
	assert.Equal(t, null.Bytes{Bytes: []uint8{}}, ad.Address)
	assert.Equal(t, common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"), ad.Owner)
	assert.Equal(t, time.Time{}, ad.MintedAt)
}

func TestHandleAftermarketDevicePairedEvent(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "AftermarketDevicePaired"

	var aftermarketDevicePairData = AftermarketDevicePairData{
		Owner:                 common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		AftermarketDeviceNode: big.NewInt(1),
		VehicleNode:           big.NewInt(11),
	}

	settings := config.Settings{
		DIMORegistryAddr:    contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
	}

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(t, config)

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	expectedBytes := eventBytes(aftermarketDevicePairData, contractEventData, t)

	v := models.Vehicle{
		ID:           11,
		OwnerAddress: common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Make:         null.StringFrom("Tesla"),
		Model:        null.StringFrom("Model-3"),
		Year:         null.IntFrom(2023),
		MintedAt:     time.Now(),
	}
	err := v.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	d := models.AftermarketDevice{
		ID:          1,
		Owner:       common.FromHex("0x12b3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Beneficiary: common.FromHex("0x12b3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	}
	err = d.Insert(ctx, pdb.DBS().Writer, boil.Infer())
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

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.ID.EQ(int(aftermarketDevicePairData.AftermarketDeviceNode.Int64())),
		qm.Load(models.AftermarketDeviceRels.Vehicle),
	).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Equal(t, aftermarketDevicePairData.AftermarketDeviceNode.Int64(), int64(ad.ID))
	assert.Equal(t, aftermarketDevicePairData.VehicleNode.Int64(), int64(ad.R.Vehicle.ID))
}

func TestHandleAftermarketDeviceUnPairedEvent(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "AftermarketDeviceUnpaired"

	var aftermarketDevicePairData = AftermarketDevicePairData{
		Owner:                 common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		AftermarketDeviceNode: big.NewInt(1),
		VehicleNode:           big.NewInt(11),
	}

	settings := config.Settings{
		DIMORegistryAddr:    contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
	}

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(t, config)

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	expectedBytes := eventBytes(aftermarketDevicePairData, contractEventData, t)

	v := models.Vehicle{
		ID:           11,
		OwnerAddress: common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Make:         null.StringFrom("Tesla"),
		Model:        null.StringFrom("Model-3"),
		Year:         null.IntFrom(2023),
		MintedAt:     time.Now(),
	}
	err := v.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	d := models.AftermarketDevice{
		ID:          1,
		Owner:       common.FromHex("0x12b3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Beneficiary: common.FromHex("0x12b3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	}
	err = d.Insert(ctx, pdb.DBS().Writer, boil.Infer())
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

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.ID.EQ(int(aftermarketDevicePairData.AftermarketDeviceNode.Int64())),
		qm.Load(models.AftermarketDeviceRels.Vehicle),
	).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Equal(t, aftermarketDevicePairData.AftermarketDeviceNode.Int64(), int64(ad.ID))
	assert.Equal(t, null.Bytes(null.Bytes{Bytes: []uint8(nil)}), null.Bytes{})

	if ad.R.Vehicle != nil {
		assert.Fail(t, "failed to unlink vehicle and aftermarket device while unpairing")
	}
}

func TestHandleAftermarketDeviceTransferredEventNewTokenID(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "Transfer"
	contractEventData.Contract = common.HexToAddress(aftermarketDeviceAddr)

	var aftermarketDeviceTransferredData = TransferEventData{
		From:    common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		To:      common.HexToAddress("0x55a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		TokenID: big.NewInt(100),
	}

	settings := config.Settings{
		AftermarketDeviceAddr: contractEventData.Contract.String(),
		DIMORegistryChainID:   contractEventData.ChainID,
	}

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(t, config)

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	expectedBytes := eventBytes(aftermarketDeviceTransferredData, contractEventData, t)

	consumer.ExpectConsumePartition(settings.ContractsEventTopic, 0, 0).YieldMessage(&sarama.ConsumerMessage{Value: expectedBytes})

	outputTest, err := consumer.ConsumePartition(settings.ContractsEventTopic, 0, 0)
	assert.NoError(t, err)

	m := <-outputTest.Messages()
	var e shared.CloudEvent[json.RawMessage]
	err = json.Unmarshal(m.Value, &e)
	assert.NoError(t, err)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.ID.EQ(int(aftermarketDeviceTransferredData.TokenID.Int64())),
	).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Equal(t, aftermarketDeviceTransferredData.TokenID.Int64(), int64(ad.ID))
	assert.Equal(t, aftermarketDeviceTransferredData.To.Bytes(), ad.Owner)
	assert.Equal(t, aftermarketDeviceTransferredData.To.Bytes(), ad.Beneficiary)
	assert.Equal(t, null.Int{}, ad.VehicleID)

}

func TestHandleAftermarketDeviceTransferredEventExistingTokenID(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "Transfer"
	contractEventData.Contract = common.HexToAddress(aftermarketDeviceAddr)

	var aftermarketDeviceTransferredData = TransferEventData{
		From:    common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		To:      common.HexToAddress("0x55a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		TokenID: big.NewInt(100),
	}

	settings := config.Settings{
		AftermarketDeviceAddr: contractEventData.Contract.String(),
		DIMORegistryChainID:   contractEventData.ChainID,
	}

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(t, config)

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	expectedBytes := eventBytes(aftermarketDeviceTransferredData, contractEventData, t)

	v := models.Vehicle{
		ID:           1,
		OwnerAddress: common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	}
	err := v.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	d := models.AftermarketDevice{
		ID:          100,
		Owner:       common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Beneficiary: common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		VehicleID:   null.IntFrom(v.ID),
	}
	err = d.Insert(ctx, pdb.DBS().Writer, boil.Infer())
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

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.ID.EQ(int(aftermarketDeviceTransferredData.TokenID.Int64())),
	).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Equal(t, aftermarketDeviceTransferredData.TokenID.Int64(), int64(ad.ID))
	assert.Equal(t, aftermarketDeviceTransferredData.To.Bytes(), ad.Owner)
	assert.Equal(t, aftermarketDeviceTransferredData.To.Bytes(), ad.Beneficiary)
	assert.Equal(t, v.ID, ad.VehicleID.Int)

}

func TestHandleBeneficiarySetEvent(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "BeneficiarySet"
	contractEventData.Contract = common.HexToAddress(aftermarketDeviceAddr)

	var beneficiarySetData = BeneficiarySetEventData{
		IdProxyAddress: common.HexToAddress(aftermarketDeviceAddr),
		Beneficiary:    common.HexToAddress("0x55b6D41bd932244Dd08186e4c19F1a7E48cbcDg3"),
		NodeId:         big.NewInt(100),
	}

	settings := config.Settings{
		DIMORegistryAddr:      contractEventData.Contract.String(),
		DIMORegistryChainID:   contractEventData.ChainID,
		AftermarketDeviceAddr: aftermarketDeviceAddr,
	}

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(t, config)

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	expectedBytes := eventBytes(beneficiarySetData, contractEventData, t)

	d := models.AftermarketDevice{
		ID:          100,
		Owner:       common.HexToAddress("0x22a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes(),
		Beneficiary: common.HexToAddress("0x22a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes(),
	}

	err := d.Insert(ctx, pdb.DBS().Writer, boil.Infer())
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

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.ID.EQ(int(beneficiarySetData.NodeId.Int64())),
	).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Equal(t, beneficiarySetData.NodeId.Int64(), int64(ad.ID))
	assert.Equal(t, beneficiarySetData.Beneficiary.Bytes(), ad.Beneficiary)

}

func TestHandleClearBeneficiaryEvent(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "BeneficiarySet"

	var beneficiarySetData = BeneficiarySetEventData{
		IdProxyAddress: common.HexToAddress(aftermarketDeviceAddr),
		Beneficiary:    zeroAddress,
		NodeId:         big.NewInt(100),
	}

	settings := config.Settings{
		DIMORegistryAddr:      contractEventData.Contract.String(),
		DIMORegistryChainID:   contractEventData.ChainID,
		AftermarketDeviceAddr: aftermarketDeviceAddr,
	}

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(t, config)

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	expectedBytes := eventBytes(beneficiarySetData, contractEventData, t)

	d := models.AftermarketDevice{
		ID:          100,
		Owner:       common.FromHex("0x22a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Beneficiary: common.FromHex("0x22a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	}
	err := d.Insert(ctx, pdb.DBS().Writer, boil.Infer())
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

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.ID.EQ(int(beneficiarySetData.NodeId.Int64())),
	).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Equal(t, beneficiarySetData.NodeId.Int64(), int64(ad.ID))
	assert.Equal(t, ad.Owner, ad.Beneficiary)
}

func TestHandle_SyntheticDeviceNodeMintedEvent_Success(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = string(SyntheticDeviceNodeMinted)

	mintedTime := mintedAt.UTC().Truncate(time.Second)
	contractEventData.Block.Time = mintedTime
	_, wallet, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	_, wallet2, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	eventData := SyntheticDeviceNodeMintedData{
		IntegrationNode:        big.NewInt(1),
		SyntheticDeviceNode:    big.NewInt(2),
		VehicleNode:            big.NewInt(1),
		SyntheticDeviceAddress: *wallet,
		Owner:                  *wallet2,
	}

	settings := config.Settings{
		DIMORegistryAddr:    contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
	}

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(t, config)

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	expectedBytes := eventBytes(eventData, contractEventData, t)

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:           1,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Camry"),
			Year:         null.IntFrom(2020),
			MintedAt:     currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			assert.NoError(t, err)
		}
	}

	consumer.ExpectConsumePartition(settings.ContractsEventTopic, 0, 0).YieldMessage(&sarama.ConsumerMessage{Value: expectedBytes})

	outputTest, err := consumer.ConsumePartition(settings.ContractsEventTopic, 0, 0)
	assert.NoError(t, err)

	m := <-outputTest.Messages()
	var e shared.CloudEvent[json.RawMessage]
	err = json.Unmarshal(m.Value, &e)
	assert.NoError(t, err)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	sd, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.VehicleID.EQ(1),
	).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Exactly(t, &models.SyntheticDevice{
		ID:            int(eventData.SyntheticDeviceNode.Int64()),
		IntegrationID: int(eventData.IntegrationNode.Int64()),
		VehicleID:     int(eventData.VehicleNode.Int64()),
		DeviceAddress: eventData.SyntheticDeviceAddress.Bytes(),
		MintedAt:      mintedTime,
	}, sd)
}

func TestHandle_SyntheticDeviceNodeBurnedEvent_Success(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = string(SyntheticDeviceNodeBurned)
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)

	_, wallet, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	_, wallet2, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:           1,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Camry"),
			Year:         null.IntFrom(2020),
			MintedAt:     currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			assert.NoError(t, err)
		}
	}

	sd := models.SyntheticDevice{
		ID:            1,
		IntegrationID: 2,
		VehicleID:     1,
		DeviceAddress: wallet2.Bytes(),
		MintedAt:      currTime,
	}

	err = sd.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	eventData := SyntheticDeviceNodeBurnedData{
		SyntheticDeviceNode: big.NewInt(2),
		VehicleNode:         big.NewInt(1),
		Owner:               *wallet2,
	}
	settings := config.Settings{
		DIMORegistryAddr:    contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
	}

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(t, config)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	expectedBytes := eventBytes(eventData, contractEventData, t)

	consumer.ExpectConsumePartition(settings.ContractsEventTopic, 0, 0).YieldMessage(&sarama.ConsumerMessage{Value: expectedBytes})
	outputTest, err := consumer.ConsumePartition(settings.ContractsEventTopic, 0, 0)
	assert.NoError(t, err)

	m := <-outputTest.Messages()
	var e shared.CloudEvent[json.RawMessage]
	err = json.Unmarshal(m.Value, &e)
	assert.NoError(t, err)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	sds, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.VehicleID.EQ(1),
		models.SyntheticDeviceWhere.ID.EQ(2),
	).All(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Len(t, sds, 0)
}

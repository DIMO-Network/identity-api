package services

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	cmodels "github.com/DIMO-Network/identity-api/internal/services/models"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	migrationsDirRelPath  = "../../migrations"
	aftermarketDeviceAddr = "0xcf9af64522162da85164a714c23a7705e6e466b3"
	syntheticDeviceAddr   = "0x85226A67FF1b3Ec6cb033162f7df5038a6C3bAB2"
	rewardContractAddr    = "0x375885164266d48C48abbbb439Be98864Ae62bBE"
	templateAddr          = "0xf369532e2144034E34Be08DBaA3A589C87dBbE3A"
)

var (
	zeroDecimal = types.NewDecimal(decimal.New(0, 0))
	mintedAt    = time.Now()
	cloudEvent  = cloudevent.RawEvent{
		CloudEventHeader: cloudevent.CloudEventHeader{
			ID:          "2SiTVhP3WBhfQQnnnpeBdMR7BSY",
			Source:      "chain/80001",
			SpecVersion: "1.0",
			Subject:     "0x4de1bcf2b7e851e31216fc07989caa902a604784",
			Time:        mintedAt,
			Type:        "zone.dimo.contract.event",
		},
	}
	contractEventData = cmodels.ContractEventData{
		ChainID:         80001,
		Contract:        common.HexToAddress("0x4de1bcf2b7e851e31216fc07989caa902a604784"),
		TransactionHash: common.HexToHash("0x811a85e24d0129a2018c9a6668652db63d73bc6d1c76f21b07da2162c6bfea7d"),
		EventSignature:  common.HexToHash("0xd624fd4c3311e1803d230d97ce71fd60c4f658c30a31fbe08edcb211fd90f63f"),
		Block: cmodels.Block{
			Time: mintedAt,
		},
	}
)

// prepareEvent turns ContractEventData (the block time, number, etc) and the event arguments (from, to, tokenId, etc)
// into a cloudevent.RawEvent like the processor expects.
//
// Note that this relies on the global variable cloudEvent to fill in the top level object.
func prepareEvent(t *testing.T, contractEventData cmodels.ContractEventData, args any) cloudevent.RawEvent {
	// Copy, just in case.
	ce := cloudEvent
	ced := contractEventData

	argBytes, err := json.Marshal(args)
	require.NoError(t, err)

	ced.Arguments = argBytes
	cedBytes, err := json.Marshal(ced)
	require.NoError(t, err)

	ce.Data = cedBytes

	return ce
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

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, aftermarketDeviceAttributesSerial)

	mfr := models.Manufacturer{
		ID:    137,
		Name:  "AutoPi",
		Owner: common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:  "autopi",
	}

	err := mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	d := models.AftermarketDevice{
		ManufacturerID: 137,
		ID:             int(aftermarketDeviceAttributesSerial.TokenID.Int64()),
		Address:        common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Owner:          common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Beneficiary:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Imei:           null.StringFrom("garbage-imei-value"),
		DevEui:         null.StringFrom("garbage-deveui-value"),
	}
	err = d.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	ad, err := models.AftermarketDevices(models.AftermarketDeviceWhere.ID.EQ(int(aftermarketDeviceAttributesSerial.TokenID.Int64()))).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Equal(t, aftermarketDeviceAttributesSerial.TokenID.Int64(), int64(ad.ID))
	assert.Equal(t, aftermarketDeviceAttributesSerial.Info, ad.Serial.String)
	assert.Equal(t, d.Imei.String, ad.Imei.String)
	assert.Equal(t, d.DevEui.String, ad.DevEui.String)
	assert.Equal(t, common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"), ad.Address)
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

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, aftermarketDevicePairData)

	m := models.Manufacturer{
		ID:       130,
		Name:     "Tesla",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "tesla",
	}

	err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	v := models.Vehicle{
		ID:             11,
		ManufacturerID: 130,
		OwnerAddress:   common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Make:           null.StringFrom("Tesla"),
		Model:          null.StringFrom("Model-3"),
		Year:           null.IntFrom(2023),
		MintedAt:       time.Now(),
	}
	err = v.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	mfr := models.Manufacturer{
		ID:    137,
		Name:  "AutoPi",
		Owner: common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:  "autopi",
	}

	err = mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	d := models.AftermarketDevice{
		ID:             1,
		ManufacturerID: 137,
		Address:        common.FromHex("0xabb3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Owner:          common.FromHex("0x12b3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Beneficiary:    common.FromHex("0x12b3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	}
	err = d.Insert(ctx, pdb.DBS().Writer, boil.Infer())
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

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, aftermarketDevicePairData)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

	v := models.Vehicle{
		ID:             11,
		ManufacturerID: 131,
		OwnerAddress:   common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Make:           null.StringFrom("Tesla"),
		Model:          null.StringFrom("Model-3"),
		Year:           null.IntFrom(2023),
		MintedAt:       time.Now(),
	}
	err := v.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	mfr := models.Manufacturer{
		ID:    137,
		Name:  "AutoPi",
		Owner: common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:  "autopi",
	}

	err = mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	d := models.AftermarketDevice{
		ID:             1,
		ManufacturerID: 137,
		Address:        common.FromHex("0xabb3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Owner:          common.FromHex("0x12b3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Beneficiary:    common.FromHex("0x12b3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	}
	err = d.Insert(ctx, pdb.DBS().Writer, boil.Infer())
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
	contractEventData.EventName = "AftermarketDeviceNodeMinted"
	contractEventData.Contract = common.HexToAddress(aftermarketDeviceAddr)

	_, wallet, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	manufacturer := models.Manufacturer{
		ID:    7,
		Owner: wallet.Bytes(),
	}
	err = manufacturer.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	var aftermarketDeviceTransferredData = AftermarketDeviceNodeMintedData{
		ManufacturerID:           big.NewInt(7),
		AftermarketDeviceAddress: common.HexToAddress("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Owner:                    common.HexToAddress("0x55a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		TokenID:                  big.NewInt(100),
	}

	settings := config.Settings{
		DIMORegistryAddr:    contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
	}

	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, aftermarketDeviceTransferredData)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.ID.EQ(int(aftermarketDeviceTransferredData.TokenID.Int64())),
	).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Equal(t, aftermarketDeviceTransferredData.TokenID.Int64(), int64(ad.ID))
	assert.Equal(t, aftermarketDeviceTransferredData.Owner.Bytes(), ad.Owner)
	assert.Equal(t, aftermarketDeviceTransferredData.Owner.Bytes(), ad.Beneficiary)
	assert.Equal(t, null.Int{}, ad.VehicleID)

}

func TestHandleAftermarketDeviceTransferredEventExistingTokenID(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "Transfer"
	contractEventData.Contract = common.HexToAddress(aftermarketDeviceAddr)

	var aftermarketDeviceTransferredData = TransferData{
		From:    common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		To:      common.HexToAddress("0x55a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		TokenID: big.NewInt(100),
	}

	settings := config.Settings{
		AftermarketDeviceAddr: contractEventData.Contract.String(),
		DIMORegistryChainID:   contractEventData.ChainID,
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, aftermarketDeviceTransferredData)

	m := models.Manufacturer{
		ID:       130,
		Name:     "Tesla",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "tesla",
	}

	err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	v := models.Vehicle{
		ID:             1,
		ManufacturerID: 130,
		OwnerAddress:   common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	}
	err = v.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	mfr := models.Manufacturer{
		ID:    137,
		Name:  "AutoPi",
		Owner: common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:  "autopi",
	}

	err = mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	d := models.AftermarketDevice{
		ID:             100,
		ManufacturerID: 137,
		Owner:          common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Beneficiary:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Address:        common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		VehicleID:      null.IntFrom(v.ID),
	}
	err = d.Insert(ctx, pdb.DBS().Writer, boil.Infer())
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

	var beneficiarySetData = BeneficiarySetData{
		IdProxyAddress: common.HexToAddress(aftermarketDeviceAddr),
		Beneficiary:    common.HexToAddress("0x55b6D41bd932244Dd08186e4c19F1a7E48cbcDg3"),
		NodeId:         big.NewInt(100),
	}

	settings := config.Settings{
		DIMORegistryAddr:      contractEventData.Contract.String(),
		DIMORegistryChainID:   contractEventData.ChainID,
		AftermarketDeviceAddr: aftermarketDeviceAddr,
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, beneficiarySetData)

	var mfr2 = models.Manufacturer{
		ID:       137,
		Owner:    common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDff"),
		Name:     "AutoPi",
		MintedAt: time.Now(),
		Slug:     "autopi",
	}
	err := mfr2.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	d := models.AftermarketDevice{
		ID:             100,
		ManufacturerID: 137,
		Owner:          common.HexToAddress("0x22a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes(),
		Beneficiary:    common.HexToAddress("0x22a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes(),
		Address:        common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	}

	err = d.Insert(ctx, pdb.DBS().Writer, boil.Infer())
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

	var beneficiarySetData = BeneficiarySetData{
		IdProxyAddress: common.HexToAddress(aftermarketDeviceAddr),
		Beneficiary:    zeroAddress,
		NodeId:         big.NewInt(100),
	}

	settings := config.Settings{
		DIMORegistryAddr:      contractEventData.Contract.String(),
		DIMORegistryChainID:   contractEventData.ChainID,
		AftermarketDeviceAddr: aftermarketDeviceAddr,
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, beneficiarySetData)

	var mfr2 = models.Manufacturer{
		ID:       137,
		Owner:    common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDff"),
		Name:     "AutoPi",
		MintedAt: time.Now(),
		Slug:     "autopi",
	}
	err := mfr2.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	d := models.AftermarketDevice{
		ID:             100,
		ManufacturerID: 137,
		Owner:          common.FromHex("0x22a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Beneficiary:    common.FromHex("0x22a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Address:        common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	}
	err = d.Insert(ctx, pdb.DBS().Writer, boil.Infer())
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

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, eventData)

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

	v := models.Vehicle{
		ID:             1,
		ManufacturerID: 131,
		OwnerAddress:   wallet.Bytes(),
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Camry"),
		Year:           null.IntFrom(2020),
		MintedAt:       currTime,
	}

	if err := v.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

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
		ConnectionID:  null.NewBytes([]byte{}, false),
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

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

	v := models.Vehicle{
		ID:             1,
		ManufacturerID: 131,
		OwnerAddress:   wallet.Bytes(),
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Camry"),
		Year:           null.IntFrom(2020),
		MintedAt:       currTime,
	}

	if err := v.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
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

	rw := models.Reward{
		IssuanceWeek:     1,
		VehicleID:        1,
		ConnectionStreak: null.IntFrom(12),
		SyntheticTokenID: null.IntFrom(1),
	}
	err = rw.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	eventData := SyntheticDeviceNodeBurnedData{
		SyntheticDeviceNode: big.NewInt(1),
		VehicleNode:         big.NewInt(1),
		Owner:               *wallet2,
	}
	settings := config.Settings{
		DIMORegistryAddr:    contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
	}

	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, eventData)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	sds, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.VehicleID.EQ(1),
		models.SyntheticDeviceWhere.ID.EQ(2),
	).All(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Empty(t, sds)
}

func Test_HandleVehicle_Transferred_Event(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "Transfer"

	tkID := 100
	var vehicleTransferredData = TransferData{
		From:    common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		To:      common.HexToAddress("0x55a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		TokenID: big.NewInt(int64(tkID)),
	}

	currTime := time.Now().UTC().Truncate(time.Second)

	settings := config.Settings{
		VehicleNFTAddr:      contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, vehicleTransferredData)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

	vehicle := models.Vehicle{
		ID:             tkID,
		ManufacturerID: 131,
		OwnerAddress:   vehicleTransferredData.From[:],
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Camry"),
		Year:           null.IntFrom(2020),
		MintedAt:       currTime,
	}

	if err := vehicle.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

	err := contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	veh, err := models.Vehicles(
		models.VehicleWhere.ID.EQ(tkID),
	).All(ctx, pdb.DBS().Reader.DB)
	assert.NoError(t, err)

	assert.Len(t, veh, 1)

	assert.Equal(t, tkID, veh[0].ID)
	assert.Equal(t, vehicleTransferredData.To.Bytes(), veh[0].OwnerAddress)
}

func Test_HandleVehicle_Transferred_To_Zero_Event_ShouldDelete(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "Transfer"
	var zeroAddress common.Address

	_, wallet, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	tkID := 100
	var vehicleTransferredData = TransferData{
		From:    *wallet,
		To:      zeroAddress,
		TokenID: big.NewInt(int64(tkID)),
	}

	settings := config.Settings{
		VehicleNFTAddr:      contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, vehicleTransferredData)

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

	vehicle := models.Vehicle{
		ID:             tkID,
		ManufacturerID: 131,
		OwnerAddress:   wallet.Bytes(),
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Camry"),
		Year:           null.IntFrom(2020),
		MintedAt:       currTime,
	}

	if err := vehicle.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

	privilege := models.Privilege{
		TokenID:     tkID,
		PrivilegeID: 1,
		UserAddress: wallet.Bytes(),
		SetAt:       currTime,
		ExpiresAt:   currTime.Add(time.Hour),
	}

	err = privilege.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	reward := models.Reward{
		IssuanceWeek: 1,
		VehicleID:    tkID,
		EarnedAt:     time.Now(),
	}

	err = reward.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	dcn := models.DCN{
		Node:         common.Hash{}.Bytes(),
		OwnerAddress: wallet.Bytes(),
		MintedAt:     time.Now(),
	}

	err = dcn.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	exists, err := models.VehicleExists(ctx, pdb.DBS().Reader, tkID)
	assert.NoError(t, err)
	assert.False(t, exists)

	numPrivs, err := models.Privileges().Count(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)
	assert.Zero(t, numPrivs)

	numRewards, err := models.Rewards().Count(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)
	assert.Zero(t, numRewards)

	err = dcn.Reload(ctx, pdb.DBS().Reader.DB)
	assert.NoError(t, err)
	assert.False(t, dcn.VehicleID.Valid)
}

func Test_HandleVehicle_Transferred_To_Zero_Event_NoDelete_SyntheticDevice(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "Transfer"
	var zeroAddress common.Address

	_, wallet, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	_, wallet2, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	tkID := 100
	var vehicleTransferredData = TransferData{
		From:    *wallet,
		To:      zeroAddress,
		TokenID: big.NewInt(int64(tkID)),
	}

	settings := config.Settings{
		VehicleNFTAddr:      contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, vehicleTransferredData)

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

	vehicle := models.Vehicle{
		ID:             tkID,
		ManufacturerID: 131,
		OwnerAddress:   wallet.Bytes(),
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Camry"),
		Year:           null.IntFrom(2020),
		MintedAt:       currTime,
	}

	if err := vehicle.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

	sd := models.SyntheticDevice{
		ID:            1,
		IntegrationID: 2,
		VehicleID:     tkID,
		DeviceAddress: wallet2.Bytes(),
		MintedAt:      currTime,
	}

	err = sd.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = contractEventConsumer.Process(ctx, &e)
	assert.Error(t, err)

	veh, err := models.Vehicles(
		models.VehicleWhere.ID.EQ(tkID),
	).All(ctx, pdb.DBS().Reader.DB)
	assert.NoError(t, err)

	assert.Len(t, veh, 1)
	assert.Equal(t, tkID, veh[0].ID)
	assert.Equal(t, vehicle.OwnerAddress, veh[0].OwnerAddress)
}

func Test_HandleVehicle_Transferred_To_Zero_Event_NoDelete_AfterMarketDevice(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "Transfer"
	var zeroAddress common.Address

	_, wallet, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	tkID := 100
	var vehicleTransferredData = TransferData{
		From:    *wallet,
		To:      zeroAddress,
		TokenID: big.NewInt(int64(tkID)),
	}

	settings := config.Settings{
		VehicleNFTAddr:      contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, vehicleTransferredData)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

	currTime := time.Now().UTC().Truncate(time.Second)

	vehicle := models.Vehicle{
		ID:             tkID,
		ManufacturerID: 131,
		OwnerAddress:   wallet.Bytes(),
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Camry"),
		Year:           null.IntFrom(2020),
		MintedAt:       currTime,
	}

	if err := vehicle.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

	m2 := models.Manufacturer{
		ID:       137,
		Name:     "AutoPi",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDff"),
		MintedAt: time.Now(),
		Slug:     "autopi",
	}

	if err := m2.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

	d := models.AftermarketDevice{
		ID:             200,
		ManufacturerID: 137,
		VehicleID:      null.IntFrom(tkID),
		Owner:          common.FromHex("0x22a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Beneficiary:    common.FromHex("0x22a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Address:        common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	}
	err = d.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	err = contractEventConsumer.Process(ctx, &e)
	assert.Error(t, err)

	veh, err := models.Vehicles(
		models.VehicleWhere.ID.EQ(tkID),
	).All(ctx, pdb.DBS().Reader.DB)
	assert.NoError(t, err)

	assert.Len(t, veh, 1)
	assert.Equal(t, tkID, veh[0].ID)
	assert.Equal(t, vehicle.OwnerAddress, veh[0].OwnerAddress)
}

func getCommonEntities(_ context.Context, vehicleID, aftermarketDeviceID, syntheticDeviceID int, owner, beneficiary common.Address) (models.Manufacturer, models.Manufacturer, models.Vehicle, models.AftermarketDevice, models.SyntheticDevice) {
	mfr := models.Manufacturer{
		ID:       130,
		Name:     "Tesla",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "tesla",
	}

	adMfr := models.Manufacturer{
		ID:       137,
		Name:     "AutoPi",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDff"),
		MintedAt: time.Now(),
		Slug:     "autopi",
	}

	veh := models.Vehicle{
		ID:             vehicleID,
		ManufacturerID: 130,
		OwnerAddress:   owner.Bytes(),
		Make:           null.StringFrom("Tesla"),
		Model:          null.StringFrom("Model-3"),
		Year:           null.IntFrom(2023),
		MintedAt:       time.Now(),
	}

	aftDevice := models.AftermarketDevice{
		ID:             aftermarketDeviceID,
		ManufacturerID: 137,
		Address:        common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Beneficiary:    beneficiary.Bytes(),
		Owner:          owner.Bytes(),
	}

	syntDevice := models.SyntheticDevice{
		ID:            syntheticDeviceID,
		IntegrationID: 11,
		VehicleID:     vehicleID,
		DeviceAddress: common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf5"),
		MintedAt:      time.Now(),
	}

	return mfr, adMfr, veh, aftDevice, syntDevice
}

func Test_Handle_TokensTransferred_ForDevice_AftermarketDevice_Event(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "TokensTransferredForDevice"
	mintedTime := mintedAt.UTC().Truncate(time.Second)
	contractEventData.Block.Time = mintedTime
	contractEventData.Contract = common.HexToAddress(rewardContractAddr)

	_, user, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	_, beneficiary, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	amt := big.NewInt(100)
	vID := big.NewInt(1)
	deviceNode := big.NewInt(2)
	aftAddr := common.HexToAddress(aftermarketDeviceAddr)

	settings := config.Settings{
		DIMORegistryChainID:   contractEventData.ChainID,
		AftermarketDeviceAddr: aftAddr.Hex(),
		RewardsContractAddr:   contractEventData.Contract.String(),
	}

	var tokensTransferredForDeviceData = TokensTransferredForDeviceData{
		User:           *user,
		Amount:         amt,
		VehicleNodeID:  vID,
		DeviceNode:     deviceNode,
		DeviceNftProxy: aftAddr,
		Week:           big.NewInt(1),
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	afterMktID := 2
	mfr, adMfr, vehicle, afterMarketDevice, _ := getCommonEntities(ctx, int(vID.Int64()), afterMktID, 0, *user, *beneficiary)

	err = mfr.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = adMfr.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = vehicle.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = afterMarketDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, tokensTransferredForDeviceData)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	reward, err := models.Rewards().All(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Len(t, reward, 1)

	if len(reward) > 0 {
		assert.Equal(t, reward[0], &models.Reward{
			IssuanceWeek:        1,
			VehicleID:           int(vID.Int64()),
			ReceivedByAddress:   null.BytesFrom(user.Bytes()),
			EarnedAt:            mintedTime,
			AftermarketTokenID:  null.IntFrom(afterMktID),
			AftermarketEarnings: types.NewDecimal(decimal.New(amt.Int64(), 0)),
			ConnectionStreak:    null.Int{},
			SyntheticTokenID:    null.Int{},
			SyntheticEarnings:   zeroDecimal,
			StreakEarnings:      zeroDecimal,
		})
	}
}

func Test_Handle_TokensTransferred_ForDevice_SyntheticDevice_Event(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "TokensTransferredForDevice"
	mintedTime := mintedAt.UTC().Truncate(time.Second)
	contractEventData.Block.Time = mintedTime
	contractEventData.Contract = common.HexToAddress(rewardContractAddr)

	_, user, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	_, beneficiary, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	amt := big.NewInt(100)
	vID := big.NewInt(1)
	synthID := 3
	deviceNode := big.NewInt(int64(synthID))
	synthAddr := common.HexToAddress(syntheticDeviceAddr)

	settings := config.Settings{
		RewardsContractAddr: contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
		SyntheticDeviceAddr: synthAddr.Hex(),
	}

	var tokensTransferredForDeviceData = TokensTransferredForDeviceData{
		User:           *user,
		Amount:         amt,
		VehicleNodeID:  vID,
		DeviceNode:     deviceNode,
		DeviceNftProxy: synthAddr,
		Week:           big.NewInt(1),
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	afterMktID := 2

	mfr, adMfr, vehicle, afterMarketDevice, syntheticDevice := getCommonEntities(ctx, int(vID.Int64()), afterMktID, synthID, *user, *beneficiary)

	err = mfr.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = adMfr.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = vehicle.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = afterMarketDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	err = syntheticDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, tokensTransferredForDeviceData)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	reward, err := models.Rewards().All(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Len(t, reward, 1)

	if len(reward) > 0 {
		assert.Equal(t, reward[0], &models.Reward{
			IssuanceWeek:        1,
			VehicleID:           int(vID.Int64()),
			ReceivedByAddress:   null.BytesFrom(user.Bytes()),
			EarnedAt:            mintedTime,
			AftermarketTokenID:  null.Int{},
			AftermarketEarnings: zeroDecimal,
			ConnectionStreak:    null.Int{},
			StreakEarnings:      zeroDecimal,
			SyntheticTokenID:    null.IntFrom(synthID),
			SyntheticEarnings:   types.NewDecimal(decimal.New(amt.Int64(), 0)),
		})
	}
}

func Test_Handle_TokensTransferred_ForDevice_UpdateSynthetic_WhenAftermarketExists(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "TokensTransferredForDevice"
	mintedTime := mintedAt.UTC().Truncate(time.Second)
	contractEventData.Block.Time = mintedTime
	contractEventData.Contract = common.HexToAddress(rewardContractAddr)

	_, user, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	_, beneficiary, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	vID := big.NewInt(1)

	aftAddr := common.HexToAddress(aftermarketDeviceAddr)
	synthAddr := common.HexToAddress(syntheticDeviceAddr)

	settings := config.Settings{
		RewardsContractAddr:   contractEventData.Contract.String(),
		DIMORegistryChainID:   contractEventData.ChainID,
		AftermarketDeviceAddr: aftAddr.Hex(),
		SyntheticDeviceAddr:   synthAddr.Hex(),
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	afterMktID := 2
	synthID := 3
	mfr, adMfr, vehicle, afterMarketDevice, syntheticDevice := getCommonEntities(ctx, int(vID.Int64()), afterMktID, synthID, *user, *beneficiary)

	err = mfr.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = adMfr.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = vehicle.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = afterMarketDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	err = syntheticDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)

	payloads := []struct {
		node      *big.Int
		proxyAddr common.Address
		amount    *big.Int
	}{
		{
			node:      big.NewInt(int64(afterMktID)),
			proxyAddr: aftAddr,
			amount:    big.NewInt(100),
		},
		{
			node:      big.NewInt(int64(synthID)),
			proxyAddr: synthAddr,
			amount:    big.NewInt(200),
		},
	}
	var tokensTransferredForDeviceData = TokensTransferredForDeviceData{
		User:          *user,
		VehicleNodeID: vID,
		Week:          big.NewInt(1),
	}
	for _, event := range payloads {
		tokensTransferredForDeviceData.Amount = event.amount
		tokensTransferredForDeviceData.DeviceNode = event.node
		tokensTransferredForDeviceData.DeviceNftProxy = event.proxyAddr

		e := prepareEvent(t, contractEventData, tokensTransferredForDeviceData)

		err = contractEventConsumer.Process(ctx, &e)
		assert.NoError(t, err)
	}

	reward, err := models.Rewards().All(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Len(t, reward, 1)

	if len(reward) > 0 {
		assert.Equal(t, reward[0], &models.Reward{
			IssuanceWeek:        1,
			VehicleID:           int(vID.Int64()),
			ReceivedByAddress:   null.BytesFrom(user.Bytes()),
			EarnedAt:            mintedTime,
			AftermarketTokenID:  null.IntFrom(afterMktID),
			AftermarketEarnings: types.NewDecimal(decimal.New(payloads[0].amount.Int64(), 0)),
			SyntheticTokenID:    null.IntFrom(synthID),
			SyntheticEarnings:   types.NewDecimal(decimal.New(payloads[1].amount.Int64(), 0)),
			ConnectionStreak:    null.Int{},
			StreakEarnings:      zeroDecimal,
		})
	}
}

func Test_Handle_TokensTransferred_ForDevice_UpdateAftermarket_WhenSyntheticExists(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "TokensTransferredForDevice"
	mintedTime := mintedAt.UTC().Truncate(time.Second)
	contractEventData.Block.Time = mintedTime
	contractEventData.Contract = common.HexToAddress(rewardContractAddr)

	_, user, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	_, beneficiary, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	vID := big.NewInt(1)

	aftAddr := common.HexToAddress(aftermarketDeviceAddr)
	synthAddr := common.HexToAddress(syntheticDeviceAddr)

	settings := config.Settings{
		RewardsContractAddr:   contractEventData.Contract.String(),
		DIMORegistryChainID:   contractEventData.ChainID,
		AftermarketDeviceAddr: aftAddr.Hex(),
		SyntheticDeviceAddr:   synthAddr.Hex(),
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	afterMktID := 1
	synthID := 11
	mfr, adMfr, vehicle, afterMarketDevice, syntheticDevice := getCommonEntities(ctx, int(vID.Int64()), afterMktID, synthID, *user, *beneficiary)

	err = mfr.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = adMfr.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = vehicle.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = afterMarketDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	err = syntheticDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)

	payloads := []struct {
		node      *big.Int
		proxyAddr common.Address
		amount    *big.Int
	}{
		{
			node:      big.NewInt(int64(synthID)),
			proxyAddr: synthAddr,
			amount:    big.NewInt(200),
		},
		{
			node:      big.NewInt(int64(afterMktID)),
			proxyAddr: aftAddr,
			amount:    big.NewInt(100),
		},
	}
	var tokensTransferredForDeviceData = TokensTransferredForDeviceData{
		User:          *user,
		VehicleNodeID: vID,
		Week:          big.NewInt(1),
	}
	for _, event := range payloads {
		tokensTransferredForDeviceData.Amount = event.amount
		tokensTransferredForDeviceData.DeviceNode = event.node
		tokensTransferredForDeviceData.DeviceNftProxy = event.proxyAddr

		e := prepareEvent(t, contractEventData, tokensTransferredForDeviceData)

		err = contractEventConsumer.Process(ctx, &e)
		assert.NoError(t, err)
	}

	reward, err := models.Rewards().All(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Len(t, reward, 1)

	if len(reward) > 0 {
		assert.Equal(t, &models.Reward{
			IssuanceWeek:        1,
			VehicleID:           int(vID.Int64()),
			ReceivedByAddress:   null.BytesFrom(user.Bytes()),
			EarnedAt:            mintedTime,
			AftermarketTokenID:  null.IntFrom(afterMktID),
			AftermarketEarnings: types.NewDecimal(decimal.New(payloads[1].amount.Int64(), 0)),
			SyntheticTokenID:    null.IntFrom(synthID),
			SyntheticEarnings:   types.NewDecimal(decimal.New(payloads[0].amount.Int64(), 0)),
			ConnectionStreak:    null.Int{},
			StreakEarnings:      zeroDecimal,
		},
			reward[0])
	}
}

func Test_Handle_TokensTransferredForConnectionStreak_Event(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	mintedTime := mintedAt.UTC().Truncate(time.Second)
	contractEventData.Block.Time = mintedTime
	contractEventData.Contract = common.HexToAddress(rewardContractAddr)

	_, user, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	_, beneficiary, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	vID := big.NewInt(1)

	aftAddr := common.HexToAddress(aftermarketDeviceAddr)
	synthAddr := common.HexToAddress(syntheticDeviceAddr)

	settings := config.Settings{
		RewardsContractAddr:   contractEventData.Contract.String(),
		DIMORegistryChainID:   contractEventData.ChainID,
		AftermarketDeviceAddr: aftAddr.Hex(),
		SyntheticDeviceAddr:   synthAddr.Hex(),
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	afterMktID := 2
	synthID := 3
	mfr, adMfr, vehicle, afterMarketDevice, syntheticDevice := getCommonEntities(ctx, int(vID.Int64()), afterMktID, synthID, *user, *beneficiary)

	err = mfr.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = adMfr.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = vehicle.Insert(ctx, pdb.DBS().Writer.DB, boil.Infer())
	assert.NoError(t, err)

	err = afterMarketDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	err = syntheticDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)

	payloads := []struct {
		node             *big.Int
		proxyAddr        *common.Address
		amount           *big.Int
		connectionStreak *big.Int
		eventName        string
	}{
		{
			node:      big.NewInt(int64(synthID)),
			proxyAddr: &synthAddr,
			amount:    big.NewInt(200),
			eventName: "TokensTransferredForDevice",
		},
		{
			node:      big.NewInt(int64(afterMktID)),
			proxyAddr: &aftAddr,
			amount:    big.NewInt(100),
			eventName: "TokensTransferredForDevice",
		},
		{
			amount:           big.NewInt(50),
			connectionStreak: big.NewInt(11),
			eventName:        "TokensTransferredForConnectionStreak",
		},
	}
	var tokensTransferredForDeviceData = TokensTransferredForDeviceData{
		User:          *user,
		VehicleNodeID: vID,
		Week:          big.NewInt(1),
	}
	var tokensTransferredForConnectionStreakData = TokensTransferredForConnectionStreakData{
		User:          *user,
		VehicleNodeID: vID,
		Week:          big.NewInt(1),
	}
	for _, event := range payloads {
		contractEventData.EventName = event.eventName

		var e cloudevent.RawEvent
		if event.proxyAddr != nil {
			tokensTransferredForDeviceData.Amount = event.amount
			tokensTransferredForDeviceData.DeviceNode = event.node
			tokensTransferredForDeviceData.DeviceNftProxy = *event.proxyAddr
			e = prepareEvent(t, contractEventData, tokensTransferredForDeviceData)
		}

		if event.connectionStreak != nil {
			tokensTransferredForConnectionStreakData.Amount = event.amount
			tokensTransferredForConnectionStreakData.ConnectionStreak = event.connectionStreak
			e = prepareEvent(t, contractEventData, tokensTransferredForConnectionStreakData)
		}

		err = contractEventConsumer.Process(ctx, &e)
		assert.NoError(t, err)
	}

	reward, err := models.Rewards().All(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Len(t, reward, 1)

	if len(reward) > 0 {
		assert.Equal(t, &models.Reward{
			IssuanceWeek:        1,
			VehicleID:           int(vID.Int64()),
			ReceivedByAddress:   null.BytesFrom(user.Bytes()),
			EarnedAt:            mintedTime,
			AftermarketTokenID:  null.IntFrom(afterMktID),
			AftermarketEarnings: types.NewDecimal(decimal.New(payloads[1].amount.Int64(), 0)),
			SyntheticTokenID:    null.IntFrom(synthID),
			SyntheticEarnings:   types.NewDecimal(decimal.New(payloads[0].amount.Int64(), 0)),
			ConnectionStreak:    null.IntFrom(int(payloads[2].connectionStreak.Int64())),
			StreakEarnings:      types.NewDecimal(decimal.New(payloads[2].amount.Int64(), 0)),
		}, reward[0])
	}
}

func TestHandle_SyntheticDevice_NodeBurnet_RewardsNulled_Success(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = string(SyntheticDeviceNodeBurned)
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)

	_, wallet, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	_, wallet2, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
	}

	v := models.Vehicle{
		ID:             1,
		ManufacturerID: 131,
		OwnerAddress:   wallet.Bytes(),
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Camry"),
		Year:           null.IntFrom(2020),
		MintedAt:       currTime,
	}

	if err := v.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(t, err)
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

	rw := models.Reward{
		IssuanceWeek:     1,
		VehicleID:        1,
		ConnectionStreak: null.IntFrom(12),
		SyntheticTokenID: null.IntFrom(1),
	}
	err = rw.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	eventData := SyntheticDeviceNodeBurnedData{
		SyntheticDeviceNode: big.NewInt(1),
		VehicleNode:         big.NewInt(1),
		Owner:               *wallet2,
	}
	settings := config.Settings{
		DIMORegistryAddr:    contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
	}

	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, eventData)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	sds, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.VehicleID.EQ(1),
		models.SyntheticDeviceWhere.ID.EQ(2),
	).All(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	rws, err := models.Rewards(models.RewardWhere.IssuanceWeek.EQ(1)).One(ctx, pdb.DBS().Reader.DB)
	assert.NoError(t, err)

	assert.Equal(t, rws.SyntheticTokenID, null.Int{})

	assert.Empty(t, sds)
}

func TestHandleAftermarketDeviceAddressResetEvent(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = string(AftermarketDeviceAddressReset)
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)

	settings := config.Settings{
		DIMORegistryAddr:    contractEventData.Contract.String(),
		DIMORegistryChainID: contractEventData.ChainID,
	}

	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	eventData := AftermarketDeviceAddressResetData{
		TokenId:                  big.NewInt(1),
		ManufacturerId:           big.NewInt(2),
		AftermarketDeviceAddress: common.HexToAddress("0x19995Cee27AbBe71b85A09B73D24EA26Fa9325a0"),
	}
	e := prepareEvent(t, contractEventData, eventData)

	mfr := models.Manufacturer{
		ID:    137,
		Name:  "AutoPi",
		Owner: common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:  "autopi",
	}
	err := mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	amd := models.AftermarketDevice{
		ID:             1,
		ManufacturerID: 137,
		Address:        common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Owner:          common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Beneficiary:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Imei:           null.StringFrom("garbage-imei-value"),
	}
	err = amd.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	updatedAmd, err := models.AftermarketDevices(models.AftermarketDeviceWhere.ID.EQ(amd.ID)).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Equal(t, eventData.AftermarketDeviceAddress, common.BytesToAddress(updatedAmd.Address))
}

func TestHandleTemplateCreatedEvent(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	contractEventData.EventName = "TemplateCreated"
	contractEventData.Contract = common.HexToAddress(templateAddr)

	_, wallet, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	// uint256(keccak256(bytes("QmYA2fn8cMbVWo4v95RwcwJVyQsNtnEwHerfWR8UNtEwoE")))
	templateIdStr := "39432737238797479986393736422353506685818102154178527542856781838379111614015"
	templateId := new(big.Int)
	templateId.SetString(templateIdStr, 10)

	var templateCreatedData = TemplateCreatedData{
		TemplateId:  templateId,
		Creator:     *wallet,
		Asset:       common.HexToAddress("0xc6e7DF5E7b4f2A278906862b61205850344D4e7d"),
		Permissions: big.NewInt(3888), // 11 11 00 11 00 00
		Cid:         "QmYA2fn8cMbVWo4v95RwcwJVyQsNtnEwHerfWR8UNtEwoE",
	}

	settings := config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784", // Different from template address
		DIMORegistryChainID: contractEventData.ChainID,
		TemplateAddr:        templateAddr,
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, templateCreatedData)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	expectedTemplateID, err := helpers.ConvertTokenIDToID(templateCreatedData.TemplateId)
	assert.NoError(t, err)

	template, err := models.Templates(
		models.TemplateWhere.ID.EQ(expectedTemplateID),
	).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Equal(t, expectedTemplateID, template.ID)
	assert.Equal(t, templateCreatedData.Creator.Bytes(), template.Creator)
	assert.Equal(t, templateCreatedData.Asset.Bytes(), template.Asset)
	assert.Equal(t, fmt.Sprintf("%016b", templateCreatedData.Permissions.Uint64()), template.Permissions)
	assert.Equal(t, templateCreatedData.Cid, template.Cid)
	assert.Equal(t, contractEventData.Block.Time.UTC(), template.CreatedAt)
}

func TestHandlePermissionsSetEventLegacy(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()

	// Setup test data
	sacdAddr := "0x5555555555555555555555555555555555555555"
	contractEventData.EventName = "PermissionsSet"
	contractEventData.Contract = common.HexToAddress(sacdAddr) // SACD contract address

	var permissionsSetData = PermissionsSetData{
		Asset:       common.HexToAddress("0xc6e7DF5E7b4f2A278906862b61205850344D4e7d"),
		TokenId:     big.NewInt(123),
		Permissions: big.NewInt(3888), // 11 11 00 11 00 00
		Grantee:     common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Expiration:  big.NewInt(time.Now().Add(time.Hour * 24 * 30).Unix()), // 30 days from now
		Source:      "test-source",
	}

	settings := config.Settings{
		VehicleNFTAddr:      "0xc6e7DF5E7b4f2A278906862b61205850344D4e7d",
		SACDAddress:         sacdAddr, // Add SACD address
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: contractEventData.ChainID,
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)

	// Create a test vehicle first
	vehicle := models.Vehicle{
		ID:           int(permissionsSetData.TokenId.Int64()),
		OwnerAddress: permissionsSetData.Grantee.Bytes(),
		MintedAt:     time.Now(),
		ManufacturerID: 1,
	}
	
	// Create a test manufacturer first
	manufacturer := models.Manufacturer{
		ID:       1,
		Name:     "Test Manufacturer",
		Owner:    permissionsSetData.Grantee.Bytes(),
		MintedAt: time.Now(),
		Slug:     "test-manufacturer",
	}
	err := manufacturer.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	err = vehicle.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, permissionsSetData)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	// Verify the SACD was created correctly (without template_id)
	sacd, err := models.VehicleSacds(
		models.VehicleSacdWhere.VehicleID.EQ(int(permissionsSetData.TokenId.Int64())),
		models.VehicleSacdWhere.Grantee.EQ(permissionsSetData.Grantee.Bytes()),
	).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Equal(t, int(permissionsSetData.TokenId.Int64()), sacd.VehicleID)
	assert.Equal(t, permissionsSetData.Grantee.Bytes(), sacd.Grantee)
	assert.Equal(t, permissionsSetData.Permissions.Text(2), sacd.Permissions)
	assert.Equal(t, permissionsSetData.Source, sacd.Source)
	assert.Equal(t, contractEventData.Block.Time.UTC().Truncate(time.Microsecond), sacd.CreatedAt.UTC().Truncate(time.Microsecond))
	assert.Equal(t, time.Unix(permissionsSetData.Expiration.Int64(), 0).UTC().Truncate(time.Microsecond), sacd.ExpiresAt.UTC().Truncate(time.Microsecond))
	
	// Template ID should be null for legacy events
	assert.False(t, sacd.TemplateID.Valid)
}

func TestHandlePermissionsSetEventWithTemplate(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()

	// Setup test data
	sacdAddr := "0x5555555555555555555555555555555555555555"
	contractEventData.EventName = "PermissionsSet"
	contractEventData.Contract = common.HexToAddress(sacdAddr) // SACD contract address

	templateId := big.NewInt(456)
	var permissionsSetWithTemplateData = PermissionsSetWithTemplateData{
		Asset:       common.HexToAddress("0xc6e7DF5E7b4f2A278906862b61205850344D4e7d"),
		TokenId:     big.NewInt(124),
		Permissions: big.NewInt(3888), // 11 11 00 11 00 00
		Grantee:     common.HexToAddress("0x1234567890123456789012345678901234567891"),
		Expiration:  big.NewInt(time.Now().Add(time.Hour * 24 * 30).Unix()), // 30 days from now
		TemplateId:  templateId,
		Source:      "test-source-with-template",
	}

	settings := config.Settings{
		VehicleNFTAddr:      "0xc6e7DF5E7b4f2A278906862b61205850344D4e7d",
		SACDAddress:         sacdAddr, // Add SACD address
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: contractEventData.ChainID,
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)

	// Create a test manufacturer first
	manufacturer := models.Manufacturer{
		ID:       1,
		Name:     "Test Manufacturer",
		Owner:    permissionsSetWithTemplateData.Grantee.Bytes(),
		MintedAt: time.Now(),
		Slug:     "test-manufacturer",
	}
	err := manufacturer.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	// Create a test vehicle first
	vehicle := models.Vehicle{
		ID:           int(permissionsSetWithTemplateData.TokenId.Int64()),
		OwnerAddress: permissionsSetWithTemplateData.Grantee.Bytes(),
		MintedAt:     time.Now(),
		ManufacturerID: 1,
	}
	err = vehicle.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	// Create a test template first
	templateIDBytes, err := helpers.ConvertTokenIDToID(templateId)
	require.NoError(t, err)
	
	template := models.Template{
		ID:          templateIDBytes,
		Creator:     permissionsSetWithTemplateData.Grantee.Bytes(),
		Asset:       permissionsSetWithTemplateData.Asset.Bytes(),
		Permissions: fmt.Sprintf("%016b", permissionsSetWithTemplateData.Permissions.Uint64()),
		Cid:         "QmTestTemplateForPermissions",
		CreatedAt:   time.Now(),
	}
	err = template.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	contractEventConsumer := NewContractsEventsConsumer(pdb, &logger, &settings)
	e := prepareEvent(t, contractEventData, permissionsSetWithTemplateData)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	// Verify the SACD was created correctly (with template_id)
	sacd, err := models.VehicleSacds(
		models.VehicleSacdWhere.VehicleID.EQ(int(permissionsSetWithTemplateData.TokenId.Int64())),
		models.VehicleSacdWhere.Grantee.EQ(permissionsSetWithTemplateData.Grantee.Bytes()),
	).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)

	assert.Equal(t, int(permissionsSetWithTemplateData.TokenId.Int64()), sacd.VehicleID)
	assert.Equal(t, permissionsSetWithTemplateData.Grantee.Bytes(), sacd.Grantee)
	assert.Equal(t, permissionsSetWithTemplateData.Permissions.Text(2), sacd.Permissions)
	assert.Equal(t, permissionsSetWithTemplateData.Source, sacd.Source)
	assert.Equal(t, contractEventData.Block.Time.UTC().Truncate(time.Microsecond), sacd.CreatedAt.UTC().Truncate(time.Microsecond))
	assert.Equal(t, time.Unix(permissionsSetWithTemplateData.Expiration.Int64(), 0).UTC().Truncate(time.Microsecond), sacd.ExpiresAt.UTC().Truncate(time.Microsecond))
	
	// Template ID should be set for new events
	assert.True(t, sacd.TemplateID.Valid)
	assert.Equal(t, templateIDBytes, sacd.TemplateID.Bytes)

	// Verify we can query the related template directly
	relatedTemplate, err := models.Templates(
		models.TemplateWhere.ID.EQ(sacd.TemplateID.Bytes),
	).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)
	assert.NotNil(t, relatedTemplate)
	assert.Equal(t, template.ID, relatedTemplate.ID)
	assert.Equal(t, template.Cid, relatedTemplate.Cid)
}

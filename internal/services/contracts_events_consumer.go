package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

type ContractsEventsConsumer struct {
	dbs      db.Store
	log      *zerolog.Logger
	settings *config.Settings
}

type EventName string

const (
	VehicleNodeMinted                  EventName = "VehicleNodeMinted"
	VehicleAttributeSet                EventName = "VehicleAttributeSet"
	AftermarketDeviceNodeMintedEvent   EventName = "AftermarketDeviceNodeMinted"
	AftermarketDeviceAttributeSetEvent EventName = "AftermarketDeviceAttributeSet"
	AftermarketDevicePairedEvent       EventName = "AftermarketDevicePaired"
	AftermarketDeviceUnpairedEvent     EventName = "AftermarketDeviceUnpaired"
	AftermarketDeviceTransferredEvent  EventName = "AftermarketDeviceTransferred"
	BeneficiarySetEvent                EventName = "BeneficiarySet"
	Transfer                           EventName = "Transfer"
)

func (r EventName) String() string {
	return string(r)
}

const contractEventCEType = "zone.dimo.contract.event"

type ContractEventData struct {
	ChainID         int64           `json:"chainId"`
	EventName       string          `json:"eventName"`
	Block           Block           `json:"block,omitempty"`
	Contract        common.Address  `json:"contract"`
	TransactionHash common.Hash     `json:"transactionHash"`
	EventSignature  common.Hash     `json:"eventSignature"`
	Arguments       json.RawMessage `json:"arguments"`
}

type Block struct {
	Number *big.Int    `json:"number,omitempty"`
	Hash   common.Hash `json:"hash,omitempty"`
	Time   time.Time   `json:"time,omitempty"`
}

type VehicleAttributeSetData struct {
	TokenID   *big.Int
	Attribute string
	Info      string
}

type AftermarketDeviceNodeMintedData struct {
	ManufacturerID           *big.Int
	TokenID                  *big.Int
	AftermarketDeviceAddress common.Address
	Owner                    common.Address
}

type AftermarketDeviceAttributeSetData struct {
	TokenID   *big.Int
	Attribute string
	Info      string
}

type AftermarketDevicePairData struct {
	AftermarketDeviceNode *big.Int
	VehicleNode           *big.Int
	Owner                 common.Address
}

type TransferEventData struct {
	From    common.Address
	To      common.Address
	TokenID *big.Int
}

type AftermarketDeviceTransferredEventData struct {
	OldOwner              common.Address
	NewOwner              common.Address
	AftermarketDeviceNode *big.Int
}

type BeneficiarySetEventData struct {
	IdProxyAddress common.Address
	NodeId         *big.Int
	Beneficiary    common.Address
}

func NewContractsEventsConsumer(dbs db.Store, log *zerolog.Logger, settings *config.Settings) *ContractsEventsConsumer {
	return &ContractsEventsConsumer{
		dbs:      dbs,
		log:      log,
		settings: settings,
	}
}

func (c *ContractsEventsConsumer) Process(ctx context.Context, event *shared.CloudEvent[json.RawMessage]) error {
	if event.Type != contractEventCEType {
		return nil
	}

	registryAddr := common.HexToAddress(c.settings.DIMORegistryAddr)
	vehicleNFTAddr := common.HexToAddress(c.settings.VehicleNFTAddr)

	var data ContractEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	if event.Source != fmt.Sprintf("chain/%d", c.settings.DIMORegistryChainID) {
		c.log.Debug().Str("event", data.EventName).Interface("event data", event).Msg("Handler not provided for event ===.")
		return nil
	}
	eventName := EventName(data.EventName)
	switch data.Contract {
	case registryAddr:
		switch eventName {
		case VehicleAttributeSet:
			return c.handleVehicleAttributeSetEvent(ctx, &data)
		case AftermarketDeviceNodeMintedEvent:
			return c.handleAftermarketDeviceNodeMintedEvent(ctx, &data)
		case AftermarketDeviceAttributeSetEvent:
			return c.handleAftermarketDeviceAttributeSetEvent(ctx, &data)
		case AftermarketDevicePairedEvent:
			return c.handleAftermarketDevicePairedEvent(ctx, &data)
		case AftermarketDeviceUnpairedEvent:
			return c.handleAftermarketDeviceUnpairedEvent(ctx, &data)
		case AftermarketDeviceTransferredEvent:
			return c.handleAftermarketDeviceTransferredEvent(ctx, &data)
		case BeneficiarySetEvent:
			return c.handleBeneficiarySetEvent(ctx, &data)
		}
	case vehicleNFTAddr:
		if eventName == Transfer {
			return c.handleVehicleTransferEvent(ctx, &data)
		}
	}

	c.log.Debug().Str("event", data.EventName).Msg("Handler not provided for event.")

	return nil
}

func (c *ContractsEventsConsumer) handleVehicleAttributeSetEvent(ctx context.Context, e *ContractEventData) error {
	var args VehicleAttributeSetData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	veh, err := models.Vehicles(
		models.VehicleWhere.ID.EQ(int(args.TokenID.Int64())),
	).One(ctx, c.dbs.DBS().Writer)
	if err != nil {
		return err
	}

	switch args.Attribute {
	case "Make":
		veh.Make = null.StringFrom(args.Info)
	case "Model":
		veh.Model = null.StringFrom(args.Info)
	case "Year":
		year, err := strconv.Atoi(args.Info)
		if err != nil {
			return err
		}
		veh.Year = null.IntFrom(year)
	default:
		return nil
	}

	colToLower := strings.ToLower(args.Attribute)

	if _, err := veh.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(colToLower)); err != nil {
		return err
	}

	return nil
}

func (c *ContractsEventsConsumer) handleVehicleTransferEvent(ctx context.Context, e *ContractEventData) error {
	logger := c.log.With().Str("EventName", Transfer.String()).Logger()

	var args TransferEventData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	vehicle := models.Vehicle{
		ID:           int(args.TokenID.Int64()),
		OwnerAddress: args.To.Bytes(),
		MintedAt:     e.Block.Time,
	}

	if err := vehicle.Upsert(ctx,
		c.dbs.DBS().Writer, true,
		[]string{models.VehicleColumns.ID},
		boil.Whitelist(models.VehicleColumns.OwnerAddress),
		boil.Whitelist(models.VehicleColumns.ID, models.VehicleColumns.OwnerAddress, models.VehicleColumns.MintedAt)); err != nil {
		return err
	}

	logger.Info().Str("TokenID", args.TokenID.String()).Msg("Event processed successfuly")

	return nil
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceNodeMintedEvent(ctx context.Context, e *ContractEventData) error {
	var args AftermarketDeviceNodeMintedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID:          int(args.TokenID.Int64()),
		Address:     null.BytesFrom(args.AftermarketDeviceAddress.Bytes()),
		Owner:       args.Owner.Bytes(),
		Beneficiary: null.BytesFrom(args.Owner.Bytes()),
		MintedAt:    e.Block.Time,
	}

	if err := ad.Upsert(ctx, c.dbs.DBS().Writer,
		true,
		[]string{models.AftermarketDeviceColumns.ID},
		boil.Whitelist(models.AftermarketDeviceColumns.ID, models.AftermarketDeviceColumns.Address, models.AftermarketDeviceColumns.Owner, models.AftermarketDeviceColumns.MintedAt, models.AftermarketDeviceColumns.Beneficiary),
		boil.Whitelist(models.AftermarketDeviceColumns.ID, models.AftermarketDeviceColumns.Address, models.AftermarketDeviceColumns.Owner, models.AftermarketDeviceColumns.MintedAt, models.AftermarketDeviceColumns.Beneficiary),
	); err != nil {
		return err
	}

	return nil
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceAttributeSetEvent(ctx context.Context, e *ContractEventData) error {
	var args AftermarketDeviceAttributeSetData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID: int(args.TokenID.Int64()),
	}
	switch args.Attribute {
	case "Serial":
		ad.Serial = null.StringFrom(args.Info)
		if _, err := ad.Update(
			ctx,
			c.dbs.DBS().Writer,
			boil.Whitelist(models.AftermarketDeviceColumns.Serial)); err != nil {
			return err
		}
	case "IMEI":
		ad.Imei = null.StringFrom(args.Info)
		if _, err := ad.Update(
			ctx,
			c.dbs.DBS().Writer,
			boil.Whitelist(models.AftermarketDeviceColumns.Imei)); err != nil {
			return err
		}
	}

	return nil
}

func (c *ContractsEventsConsumer) handleAftermarketDevicePairedEvent(ctx context.Context, e *ContractEventData) error {
	var args AftermarketDevicePairData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID:          int(args.AftermarketDeviceNode.Int64()),
		VehicleID:   null.IntFrom(int(args.VehicleNode.Int64())),
		Owner:       args.Owner.Bytes(),
		Beneficiary: null.BytesFrom(args.Owner.Bytes()),
	}

	if err := ad.Upsert(
		ctx,
		c.dbs.DBS().Writer,
		true,
		[]string{models.AftermarketDeviceColumns.ID},
		boil.Whitelist(models.AftermarketDeviceColumns.VehicleID, models.AftermarketDeviceColumns.Owner),
		boil.Infer(),
	); err != nil {
		return err
	}

	return nil
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceUnpairedEvent(ctx context.Context, e *ContractEventData) error {
	var args AftermarketDevicePairData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID:        int(args.AftermarketDeviceNode.Int64()),
		VehicleID: null.Int{},
		Owner:     args.Owner.Bytes(),
	}

	if err := ad.Upsert(
		ctx,
		c.dbs.DBS().Writer,
		true,
		[]string{models.AftermarketDeviceColumns.ID},
		boil.Whitelist(models.AftermarketDeviceColumns.VehicleID, models.AftermarketDeviceColumns.Owner),
		boil.Infer(),
	); err != nil {
		return err
	}

	return nil
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceTransferredEvent(ctx context.Context, e *ContractEventData) error {
	var args AftermarketDeviceTransferredEventData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID:          int(args.AftermarketDeviceNode.Int64()),
		Owner:       args.NewOwner.Bytes(),
		Beneficiary: null.BytesFrom(args.NewOwner.Bytes()),
		VehicleID:   null.Int{},
	}

	_, err := ad.Update(
		ctx,
		c.dbs.DBS().Writer,
		boil.Whitelist(models.AftermarketDeviceColumns.Owner, models.AftermarketDeviceColumns.VehicleID, models.AftermarketDeviceColumns.Beneficiary))
	return err
}

func (c *ContractsEventsConsumer) handleBeneficiarySetEvent(ctx context.Context, e *ContractEventData) error {
	var args BeneficiarySetEventData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID:          int(args.NodeId.Int64()),
		Beneficiary: null.BytesFrom(args.Beneficiary.Bytes()),
	}

	_, err := ad.Update(
		ctx,
		c.dbs.DBS().Writer,
		boil.Whitelist(models.AftermarketDeviceColumns.Beneficiary),
	)
	return err
}

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/ksuid"
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

var zeroAddress common.Address

const (
	Transfer                           EventName = "Transfer"
	VehicleAttributeSet                EventName = "VehicleAttributeSet"
	AftermarketDeviceNodeMintedEvent   EventName = "AftermarketDeviceNodeMinted"
	AftermarketDeviceAttributeSetEvent EventName = "AftermarketDeviceAttributeSet"
	PrivilegeSet                       EventName = "PrivilegeSet"
	AftermarketDevicePairedEvent       EventName = "AftermarketDevicePaired"
	AftermarketDeviceUnpairedEvent     EventName = "AftermarketDeviceUnpaired"
	BeneficiarySetEvent                EventName = "BeneficiarySet"
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

type PrivilegeSetData struct {
	TokenId *big.Int
	PrivId  *big.Int
	User    common.Address
	Expires *big.Int
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
	// Filter out end-of-block events.
	if event.Type != contractEventCEType {
		return nil
	}

	if event.Source != fmt.Sprintf("chain/%d", c.settings.DIMORegistryChainID) {
		return nil
	}

	registryAddr := common.HexToAddress(c.settings.DIMORegistryAddr)
	vehicleNFTAddr := common.HexToAddress(c.settings.VehicleNFTAddr)
	aftermarketDeviceAddr := common.HexToAddress(c.settings.AftermarketDeviceAddr)

	var data ContractEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	eventName := EventName(data.EventName)

	c.log.Info().Str("Event", string(eventName)).Str("Contract", data.Contract.Hex()).Msg("Event Received")

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
		case BeneficiarySetEvent:
			return c.handleBeneficiarySetEvent(ctx, &data)
		}
	case vehicleNFTAddr:
		switch eventName {
		case Transfer:
			return c.handleVehicleTransferEvent(ctx, &data)
		case PrivilegeSet:
			return c.handlePrivilegeSetEvent(ctx, &data)
		}
	case aftermarketDeviceAddr:
		switch eventName {
		case Transfer:
			return c.handleAftermarketDeviceTransferredEvent(ctx, &data)
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

	veh, err := models.FindVehicle(ctx, c.dbs.DBS().Reader, int(args.TokenID.Int64()))
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
		return fmt.Errorf("unrecognized vehicle attribute %q", args.Attribute)
	}

	colToLower := strings.ToLower(args.Attribute)

	_, err = veh.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(colToLower))
	return err
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

	// Insert is the mint case.
	if err := vehicle.Upsert(
		ctx,
		c.dbs.DBS().Writer,
		true,
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
		ID:       int(args.TokenID.Int64()),
		Address:  null.BytesFrom(args.AftermarketDeviceAddress.Bytes()),
		Owner:    null.BytesFrom(args.Owner.Bytes()),
		MintedAt: null.TimeFrom(e.Block.Time),
	}

	if _, err := ad.Update(
		ctx,
		c.dbs.DBS().Writer,
		boil.Whitelist(models.AftermarketDeviceColumns.Address, models.AftermarketDeviceColumns.Owner, models.AftermarketDeviceColumns.MintedAt),
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

func (c *ContractsEventsConsumer) handlePrivilegeSetEvent(ctx context.Context, e *ContractEventData) error {
	logger := c.log.With().Str("EventName", Transfer.String()).Logger()

	var args PrivilegeSetData

	privilege := models.Privilege{
		ID:          ksuid.New().String(),
		TokenID:     int(args.TokenId.Int64()),
		PrivilegeID: int(args.PrivId.Int64()),
		UserAddress: args.User.Bytes(),
		SetAt:       e.Block.Time,
		ExpiresAt:   time.Unix(args.Expires.Int64(), 0),
	}

	if err := privilege.Insert(ctx, c.dbs.DBS().Writer, boil.Infer()); err != nil {
		return err
	}

	logger.Info().
		Str("PrivilegeID", args.PrivId.String()).
		Str("TokenID", args.TokenId.String()).
		Msg("Event processed successfuly")

	return nil
}

func (c *ContractsEventsConsumer) handleAftermarketDevicePairedEvent(ctx context.Context, e *ContractEventData) error {
	var args AftermarketDevicePairData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID:        int(args.AftermarketDeviceNode.Int64()),
		VehicleID: null.IntFrom(int(args.VehicleNode.Int64())),
	}

	_, err := ad.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.AftermarketDeviceColumns.VehicleID))
	return err
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceUnpairedEvent(ctx context.Context, e *ContractEventData) error {
	var args AftermarketDevicePairData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{ID: int(args.AftermarketDeviceNode.Int64())}

	_, err := ad.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.AftermarketDeviceColumns.VehicleID))
	return err
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceTransferredEvent(ctx context.Context, e *ContractEventData) error {
	var args TransferEventData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID:          int(args.TokenID.Int64()),
		Owner:       null.BytesFrom(args.To.Bytes()),
		MintedAt:    null.TimeFrom(e.Block.Time),
		Beneficiary: args.To.Bytes(),
	}

	return ad.Upsert(
		ctx,
		c.dbs.DBS().Writer,
		true,
		[]string{models.AftermarketDeviceColumns.ID},
		boil.Whitelist(models.AftermarketDeviceColumns.ID, models.AftermarketDeviceColumns.Owner, models.AftermarketDeviceColumns.MintedAt, models.AftermarketDeviceColumns.Beneficiary),
		boil.Whitelist(models.AftermarketDeviceColumns.ID, models.AftermarketDeviceColumns.Owner, models.AftermarketDeviceColumns.MintedAt, models.AftermarketDeviceColumns.Beneficiary),
	)
}

func (c *ContractsEventsConsumer) handleBeneficiarySetEvent(ctx context.Context, e *ContractEventData) error {
	var args BeneficiarySetEventData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	if args.IdProxyAddress != common.HexToAddress(c.settings.AftermarketDeviceAddr) {
		c.log.Warn().Msgf("beneficiary set on an unexpected contract: %s", args.IdProxyAddress.Hex())
		return nil
	}

	var err error
	ad := new(models.AftermarketDevice)
	ad.ID = int(args.NodeId.Int64())
	ad.Beneficiary = args.Beneficiary.Bytes()

	if args.Beneficiary == zeroAddress {
		ad, err = models.AftermarketDevices(
			models.AftermarketDeviceWhere.ID.EQ(int(args.NodeId.Int64())),
		).One(ctx, c.dbs.DBS().Reader)
		if err != nil {
			return err
		}
		ad.Beneficiary = ad.Owner.Bytes
	}

	if _, err = ad.Update(
		ctx,
		c.dbs.DBS().Writer,
		boil.Whitelist(models.AftermarketDeviceColumns.Beneficiary),
	); err != nil {
		return err
	}

	return nil
}

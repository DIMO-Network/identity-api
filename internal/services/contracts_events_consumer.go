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
	AftermarketDeviceNodeMinted        EventName = "AftermarketDeviceNodeMinted"
	AftermarketDeviceAttributeSetEvent EventName = "AftermarketDeviceAttributeSet"
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

type VehicleNodeMintedData struct {
	TokenId *big.Int
	Owner   common.Address
}

type VehicleAttributeSetData struct {
	TokenId   *big.Int
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

func NewContractsEventsConsumer(dbs db.Store, log *zerolog.Logger, settings *config.Settings) *ContractsEventsConsumer {
	return &ContractsEventsConsumer{
		dbs:      dbs,
		log:      log,
		settings: settings,
	}
}

func (c *ContractsEventsConsumer) Process(event *shared.CloudEvent[json.RawMessage]) error {
	if event.Type != contractEventCEType {
		return nil
	}

	registryAddr := common.HexToAddress(c.settings.DIMORegistryAddr)
	var data ContractEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	if event.Source != fmt.Sprintf("chain/%d", c.settings.DIMORegistryChainID) {
		c.log.Debug().Str("event", data.EventName).Interface("event data", event).Msg("Handler not provided for event ===.")
		return nil
	}

	if data.Contract == registryAddr {
		switch EventName(data.EventName) {
		case VehicleNodeMinted:
			return c.handleVehicleNodeMintedEvent(&data)
		case VehicleAttributeSet:
			return c.handleVehicleAttributeSetEvent(&data)
		case AftermarketDeviceNodeMinted:
			return c.handleAftermarketDeviceNodeMintedEvent(&data)
		case AftermarketDeviceAttributeSetEvent:
			return c.handleAftermarketDeviceAttributeSetEvent(&data)
		}
	}

	c.log.Debug().Str("event", data.EventName).Msg("Handler not provided for event.")

	return nil
}

func (c *ContractsEventsConsumer) handleVehicleNodeMintedEvent(e *ContractEventData) error {
	var args VehicleNodeMintedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	dm := models.Vehicle{
		OwnerAddress: null.BytesFrom(args.Owner.Bytes()),
		MintedAt:     null.TimeFrom(e.Block.Time),
		ID:           int(args.TokenId.Int64()),
	}

	if err := dm.Upsert(context.TODO(), c.dbs.DBS().Writer, true, []string{models.VehicleColumns.ID},
		boil.Whitelist(models.VehicleColumns.OwnerAddress, models.VehicleColumns.MintedAt),
		boil.Whitelist(models.VehicleColumns.ID, models.VehicleColumns.OwnerAddress, models.VehicleColumns.MintedAt)); err != nil {
		return err
	}

	return nil
}

func (c *ContractsEventsConsumer) handleVehicleAttributeSetEvent(e *ContractEventData) error {
	var args VehicleAttributeSetData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	veh := models.Vehicle{ID: int(args.TokenId.Int64())}

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

	if err := veh.Upsert(context.TODO(), c.dbs.DBS().Writer, true, []string{models.VehicleColumns.ID}, boil.Whitelist(colToLower), boil.Whitelist(models.VehicleColumns.ID, colToLower)); err != nil {
		return err
	}

	return nil
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceNodeMintedEvent(e *ContractEventData) error {
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

	if err := ad.Upsert(context.Background(), c.dbs.DBS().Writer,
		true,
		[]string{models.AftermarketDeviceColumns.ID},
		boil.Whitelist(models.AftermarketDeviceColumns.Owner, models.AftermarketDeviceColumns.MintedAt, models.AftermarketDeviceColumns.Address),
		boil.Infer()); err != nil {
		return err
	}

	return nil
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceAttributeSetEvent(e *ContractEventData) error {
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
		if err := ad.Upsert(
			context.Background(),
			c.dbs.DBS().Writer,
			true,
			[]string{models.AftermarketDeviceColumns.ID},
			boil.Whitelist(models.AftermarketDeviceColumns.Serial),
			boil.Infer()); err != nil {
			return err
		}
	case "IMEI":
		ad.Imei = null.StringFrom(args.Info)
		if err := ad.Upsert(
			context.Background(),
			c.dbs.DBS().Writer,
			true,
			[]string{models.AftermarketDeviceColumns.ID},
			boil.Whitelist(models.AftermarketDeviceColumns.Imei),
			boil.Infer()); err != nil {
			return err
		}
	}

	return nil
}

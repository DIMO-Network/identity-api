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
	VehicleNodeMinted   EventName = "VehicleNodeMinted"
	VehicleAttributeSet EventName = "VehicleAttributeSet"
	Transfer            EventName = "Transfer"
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
	TokenID   *big.Int
	Attribute string
	Info      string
}

type TransferEventData struct {
	From    common.Address
	To      common.Address
	TokenID *big.Int
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

	switch data.Contract {
	case registryAddr:
		if EventName(data.EventName) == VehicleNodeMinted {
			return c.handleVehicleNodeMintedEvent(ctx, &data)
		}

		if EventName(data.EventName) == VehicleAttributeSet {
			return c.handleVehicleAttributeSetEvent(ctx, &data)
		}
	case vehicleNFTAddr:
		if EventName(data.EventName) == Transfer {
			return c.handleVehicleTransferEvent(ctx, &data)
		}
	}

	c.log.Debug().Str("event", data.EventName).Msg("Handler not provided for event.")

	return nil
}

func (c *ContractsEventsConsumer) handleVehicleNodeMintedEvent(ctx context.Context, e *ContractEventData) error {
	var args VehicleNodeMintedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	veh, err := models.Vehicles(
		models.VehicleWhere.ID.EQ(int(args.TokenId.Int64())),
		models.VehicleWhere.OwnerAddress.EQ(null.BytesFrom(args.Owner.Bytes())),
	).One(ctx, c.dbs.DBS().Writer)
	if err != nil {
		return err
	}

	veh.MintedAt = null.TimeFrom(e.Block.Time)

	if _, err := veh.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.VehicleColumns.MintedAt)); err != nil {
		return err
	}

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

	vehicle := models.Vehicle{ID: int(args.TokenID.Int64()), OwnerAddress: null.BytesFrom(args.To.Bytes())}
	if err := vehicle.Upsert(ctx, c.dbs.DBS().Writer, true, []string{models.VehicleColumns.ID}, boil.Whitelist(models.VehicleColumns.ID), boil.Whitelist(models.VehicleColumns.ID, models.VehicleColumns.OwnerAddress)); err != nil {
		return err
	}

	logger.Info().Str("TokenID", args.TokenID.String()).Msg("Event processed successfuly")

	return nil
}

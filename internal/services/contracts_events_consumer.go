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
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type ContractsEventsConsumer struct {
	ctx      context.Context
	db       db.Store
	log      *zerolog.Logger
	settings *config.Settings
}

type EventName string

const (
	VehicleNodeMinted   EventName = "VehicleNodeMinted"
	VehicleAttributeSet EventName = "VehicleAttributeSet"
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

func NewContractsEventsConsumer(ctx context.Context, pdb db.Store, log *zerolog.Logger, settings *config.Settings) *ContractsEventsConsumer {
	return &ContractsEventsConsumer{
		ctx:      ctx,
		db:       pdb,
		log:      log,
		settings: settings,
	}
}

func (c *ContractsEventsConsumer) ProcessContractsEventsMessages(messages <-chan *message.Message) {
	for msg := range messages {
		err := c.processMessage(msg)
		if err != nil {
			c.log.Err(err).Msg("error processing contract events messages")
		}
	}
}

func (c *ContractsEventsConsumer) processMessage(msg *message.Message) error {
	// Keep the pipeline moving no matter what.
	defer func() { msg.Ack() }()

	// Deletion messages. We're the only actor that produces these, so ignore them.
	if msg.Payload == nil {
		return nil
	}

	event := new(shared.CloudEvent[json.RawMessage])
	if err := json.Unmarshal(msg.Payload, event); err != nil {
		return errors.Wrap(err, "error parsing device event payload")
	}

	return c.processEvent(event)
}

func (c *ContractsEventsConsumer) processEvent(event *shared.CloudEvent[json.RawMessage]) error {
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
	switch data.EventName {
	case VehicleNodeMinted.String():
		if data.Contract == registryAddr {
			c.log.Info().Str("event", data.EventName).Msg("Event received")
			return c.handleVehicleNodeMintedEvent(&data)
		}
	case VehicleAttributeSet.String():
		if data.Contract == registryAddr {
			c.log.Info().Str("event", data.EventName).Msg("Event received")
			return c.handleVehicleAttributeSetEvent(&data)
		}
	default:
		c.log.Debug().Str("event", data.EventName).Msg("Handler not provided for event.")
	}

	return nil
}

func (c *ContractsEventsConsumer) handleVehicleNodeMintedEvent(e *ContractEventData) error {
	var args VehicleNodeMintedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	dm := models.Vehicle{
		OwnerAddress: null.BytesFrom(args.Owner.Bytes()),
		MintTime:     null.TimeFrom(e.Block.Time),
		ID:           int(args.TokenId.Int64()),
	}

	if err := dm.Upsert(c.ctx, c.db.DBS().Writer, true, []string{models.VehicleColumns.ID},
		boil.Whitelist(models.VehicleColumns.OwnerAddress, models.VehicleColumns.MintTime),
		boil.Whitelist(models.VehicleColumns.ID, models.VehicleColumns.OwnerAddress, models.VehicleColumns.MintTime)); err != nil {
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

	if err := veh.Upsert(c.ctx, c.db.DBS().Writer, true, []string{models.VehicleColumns.ID}, boil.Whitelist(colToLower), boil.Whitelist(models.VehicleColumns.ID, colToLower)); err != nil {
		return err
	}

	return nil
}

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	dgrpc "github.com/DIMO-Network/devices-api/pkg/grpc"
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
	dSvc     dgrpc.UserDeviceServiceClient
	ddFSvc   ddgrpc.DeviceDefinitionServiceClient
}

type EventName string

const (
	VehicleNodeMinted EventName = "VehicleNodeMinted"
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

func NewContractsEventsConsumer(ctx context.Context, pdb db.Store, log *zerolog.Logger, settings *config.Settings, dSvc dgrpc.UserDeviceServiceClient, ddFSvc ddgrpc.DeviceDefinitionServiceClient) *ContractsEventsConsumer {
	return &ContractsEventsConsumer{
		ctx:      ctx,
		db:       pdb,
		log:      log,
		settings: settings,
		dSvc:     dSvc,
		ddFSvc:   ddFSvc,
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
		c.log.Info().Str("event", data.EventName).Msg("Event received")
		return c.handleVehicleNodeMintedEvent(&data)
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

	device, err := c.dSvc.GetUserDeviceByTokenId(c.ctx, &dgrpc.GetUserDeviceByTokenIdRequest{
		TokenId: args.TokenId.Int64(),
	})
	if err != nil {
		return err
	}

	if device == nil {
		return fmt.Errorf("could not find device with tokenID %d", args.TokenId)
	}

	deviceDef, err := c.ddFSvc.GetDeviceDefinitionByID(c.ctx, &ddgrpc.GetDeviceDefinitionRequest{
		Ids: []string{device.DeviceDefinitionId},
	})
	if err != nil {
		return err
	}

	if len(deviceDef.DeviceDefinitions) == 0 {
		return fmt.Errorf("could not find device with tokenID %d", args.TokenId)
	}

	ddf := deviceDef.DeviceDefinitions[0]

	dm := models.Vehicle{
		OwnerAddress: null.BytesFrom(args.Owner.Bytes()),
		Make:         ddf.Type.Make,
		Model:        ddf.Type.Model,
		Year:         int(ddf.Type.Year),
		ID:           int(args.TokenId.Int64()),
	}

	if err := dm.Insert(c.ctx, c.db.DBS().Writer, boil.Infer()); err != nil {
		return err
	}

	return nil
}

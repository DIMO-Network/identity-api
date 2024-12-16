package services

import (
	"context"
	"time"

	cmodels "github.com/DIMO-Network/identity-api/internal/services/models"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/goccy/go-json"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func (c *ContractsEventsConsumer) handleNewDcnNode(ctx context.Context, e *cmodels.ContractEventData) error {
	logger := c.log.With().Str("EventName", NewNode.String()).Logger()

	var args NewDCNNodeData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	dcn := models.DCN{
		Node:         args.Node,
		OwnerAddress: args.Owner.Bytes(),
		MintedAt:     e.Block.Time,
	}

	err := dcn.Insert(ctx, c.dbs.DBS().Writer, boil.Infer())
	if err != nil {
		return err
	}

	logger.Info().Str("Node", hexutil.Encode(args.Node)).Msg(NewNode.String() + " Event processed successfuly")

	return nil
}

func (c *ContractsEventsConsumer) handleNewDCNExpiration(ctx context.Context, e *cmodels.ContractEventData) error {
	logger := c.log.With().Str("EventName", NewExpiration.String()).Logger()

	var args NewDCNExpirationData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	dcn := models.DCN{
		Node:       args.Node,
		Expiration: null.TimeFrom(time.Unix(int64(args.Expiration), 0)),
	}

	_, err := dcn.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.DCNColumns.Expiration))
	if err != nil {
		return err
	}

	logger.Info().Str("Node", hexutil.Encode(args.Node)).Msg(NewExpiration.String() + " Event processed successfuly")

	return nil
}

func (c *ContractsEventsConsumer) handleNameChanged(ctx context.Context, e *cmodels.ContractEventData) error {
	eventName := NameChanged.String()
	logger := c.log.With().Str("EventName", eventName).Logger()

	var args DCNNameChangedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	dcn := models.DCN{
		Node: args.Node,
		Name: null.StringFrom(args.Name),
	}

	_, err := dcn.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.DCNColumns.Name))
	if err != nil {
		return err
	}

	logger.Info().Str("Node", hexutil.Encode(args.Node)).Msg(eventName + " Event processed successfuly")

	return nil
}

func (c *ContractsEventsConsumer) handleVehicleIdChanged(ctx context.Context, e *cmodels.ContractEventData) error {
	eventName := VehicleIdChanged.String()
	logger := c.log.With().Str("EventName", eventName).Logger()

	var args DCNVehicleIdChangedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	dcn := models.DCN{
		Node:      args.Node,
		VehicleID: null.IntFrom(int(args.VehicleID.Int64())),
	}

	_, err := dcn.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.DCNColumns.VehicleID))
	if err != nil {
		return err
	}

	logger.Info().Str("Node", hexutil.Encode(args.Node)).Msg(eventName + " Event processed successfuly")

	return nil
}

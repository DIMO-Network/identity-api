package services

import (
	"context"
	"time"

	"github.com/DIMO-Network/identity-api/models"
	"github.com/goccy/go-json"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func (c *ContractsEventsConsumer) handleNewDcnNode(ctx context.Context, e *ContractEventData) error {
	logger := c.log.With().Str("EventName", NewNode.String()).Logger()

	var args NewDCNNodeEventData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	dcn := models.DCN{
		Node:         args.Node,
		OwnerAddress: args.Owner.Bytes(),
	}
	err := dcn.Insert(ctx, c.dbs.DBS().Writer, boil.Infer())
	if err != nil {
		return nil
	}

	logger.Info().Str("Node", string(args.Node)).Msg(NewNode.String() + " Event processed successfuly")

	return nil
}

func (c *ContractsEventsConsumer) handleNewDcnResolver(ctx context.Context, e *ContractEventData) error {
	logger := c.log.With().Str("EventName", NewExpiration.String()).Logger()

	var args NewDCNResolverEventData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	dcn, err := models.DCNS(
		models.DCNWhere.Node.EQ(args.Node),
	).One(ctx, c.dbs.DBS().Reader)
	if err != nil {
		return err
	}

	dcn.ResolverAddress = null.BytesFrom(args.Resolver.Bytes())
	_, err = dcn.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.DCNColumns.ResolverAddress))
	if err != nil {
		return err
	}

	logger.Info().Str("Node", string(args.Node)).Msg(NewExpiration.String() + " Event processed successfuly")

	return nil
}

func (c *ContractsEventsConsumer) handleNewDCNExpiration(ctx context.Context, e *ContractEventData) error {
	logger := c.log.With().Str("EventName", NewExpiration.String()).Logger()

	var args NewDCNExpirationEventData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	dcn, err := models.DCNS(
		models.DCNWhere.Node.EQ(args.Node),
	).One(ctx, c.dbs.DBS().Reader)
	if err != nil {
		return err
	}

	dcn.Expiration = null.TimeFrom(
		time.Unix(int64(args.Expiration), 0),
	)
	_, err = dcn.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.DCNColumns.Expiration))
	if err != nil {
		return err
	}

	logger.Info().Str("Node", string(args.Node)).Msg(NewExpiration.String() + " Event processed successfuly")

	return nil
}

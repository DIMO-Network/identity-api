package staking

import (
	"context"
	"time"

	"github.com/DIMO-Network/identity-api/internal/dbtypes"
	cmodels "github.com/DIMO-Network/identity-api/internal/services/models"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

type Handler struct {
	DBS db.Store
}

func (h *Handler) HandleStaked(ctx context.Context, event *cmodels.ContractEventData, args *Staked) error {
	stake := models.Stake{
		ID:       int(args.StakeId.Int64()),
		Owner:    args.User.Bytes(),
		Level:    int(args.Level),
		Points:   int(args.Points.Int64()),
		Amount:   dbtypes.IntToDecimal(args.Amount),
		StakedAt: event.Block.Time,
		EndsAt:   time.Unix(args.LockEndTime.Int64(), 0),
	}

	cols := models.StakeColumns

	return stake.Upsert(ctx, h.DBS.DBS().Writer, true, []string{models.StakeColumns.ID},
		boil.Whitelist(cols.Level, cols.Points, cols.Amount, cols.StakedAt, cols.EndsAt), // Updating. This happens when the stake is upgraded. Definitely don't want to touch the vehicle attachment.
		boil.Infer(), // Inserting. Could omit the vehicle here.
	)
}

func (h *Handler) HandleWithdrawn(ctx context.Context, event *cmodels.ContractEventData, args *Withdrawn) error {
	stake := models.Stake{
		ID:          int(args.StakeId.Int64()),
		WithdrawnAt: null.TimeFrom(event.Block.Time),
	}

	_, err := stake.Update(ctx, h.DBS.DBS().Writer, boil.Whitelist(models.StakeColumns.WithdrawnAt))
	return err
}

func (h *Handler) HandleStakingExtended(ctx context.Context, event *cmodels.ContractEventData, args *StakingExtended) error {
	stake := models.Stake{
		ID:     int(args.StakeId.Int64()),
		EndsAt: time.Unix(args.NewLockEndTime.Int64(), 0),
	}

	_, err := stake.Update(ctx, h.DBS.DBS().Writer, boil.Whitelist(models.StakeColumns.EndsAt))
	return err
}

func (h *Handler) HandleVehicleAttached(ctx context.Context, event *cmodels.ContractEventData, args *VehicleAttached) error {
	stake := models.Stake{
		ID:        int(args.StakeId.Int64()),
		VehicleID: null.IntFrom(int(args.VehicleId.Int64())),
	}

	_, err := stake.Update(ctx, h.DBS.DBS().Writer, boil.Whitelist(models.StakeColumns.VehicleID))
	return err
}

func (h *Handler) HandleVehicleDetached(ctx context.Context, event *cmodels.ContractEventData, args *VehicleDetached) error {
	stake := models.Stake{
		ID: int(args.StakeId.Int64()),
	}

	_, err := stake.Update(ctx, h.DBS.DBS().Writer, boil.Whitelist(models.StakeColumns.VehicleID))
	return err
}

func (h *Handler) HandleTransfer(ctx context.Context, event *cmodels.ContractEventData, args *Transfer) error {
	stake := models.Stake{
		ID:    int(args.TokenId.Int64()),
		Owner: args.To.Bytes(),
	}

	_, err := stake.Update(ctx, h.DBS.DBS().Writer, boil.Whitelist(models.StakeColumns.Owner))
	return err
}

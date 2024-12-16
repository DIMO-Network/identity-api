package staking

import (
	"context"
	"time"

	cmodels "github.com/DIMO-Network/identity-api/internal/services/models"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/dbtypes"
)

type Handler struct {
	dbs db.Store
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
	return nil
}

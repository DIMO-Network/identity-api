package staking

import (
	"context"

	"github.com/DIMO-Network/identity-api/internal/services/models"
	"github.com/DIMO-Network/shared/db"
)

type Handler struct {
	dbs db.Store
}

func (h *Handler) HandleStaked(ctx context.Context, event *models.ContractEventData, args *Staked) error {
	return nil
}

package staking

import (
	"context"

	"github.com/DIMO-Network/identity-api/internal/services/models"
)

type Handler struct{}

func (h *Handler) HandleStaked(ctx context.Context, event *models.ContractEventData, args *Staked) error {
	return nil
}

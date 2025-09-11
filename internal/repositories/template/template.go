package template

import (
	"context"
	"math/big"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
)

type Repository struct {
	*base.Repository
	chainID         uint64
	contractAddress common.Address
}

// New creates a new template repository.
func New(db *base.Repository) *Repository {
	return &Repository{
		Repository:      db,
		chainID:         uint64(db.Settings.DIMORegistryChainID),
		contractAddress: common.HexToAddress(db.Settings.ConnectionAddr),
	}
}

func (r *Repository) ToAPI(template *models.Template) *model.Template {
	tokenID := new(big.Int).SetBytes(template.ID)

	return &model.Template{
		TokenID:     tokenID,
		Creator:     common.BytesToAddress(template.Creator),
		Asset:       common.BytesToAddress(template.Asset),
		Permissions: template.Permissions,
		Cid:         template.Cid,
		CreatedAt:   template.CreatedAt,
	}
}

// TODO maybe add a filterby
func (r *Repository) GetTemplates(ctx context.Context, first *int, after *string, last *int, before *string) (*model.TemplateConnection, error) {
	// TODO implement GetTemplates
	return nil, nil
}

func (r *Repository) GetTemplate(ctx context.Context, by model.ConnectionBy) (*model.Template, error) {
	// TODO implement GetTemplate
	return nil, nil
}

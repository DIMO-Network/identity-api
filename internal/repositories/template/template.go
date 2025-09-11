package template

import (
	"context"
	"database/sql"
	"errors"
	"math/big"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
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

func (r *Repository) GetTemplate(ctx context.Context, by model.TemplateBy) (*model.Template, error) {
	if base.CountTrue(by.TokenID != nil, by.Cid != nil) != 1 {
		return nil, gqlerror.Errorf("must specify exactly one of `TokenID` or `Cid`")
	}

	var mod qm.QueryMod

	switch {
	case by.TokenID != nil:
		id, err := helpers.ConvertTokenIDToID(by.TokenID)
		if err != nil {
			return nil, err
		}

		mod = models.TemplateWhere.ID.EQ(id)
	case by.Cid != nil:
		mod = models.TemplateWhere.Cid.EQ(*by.Cid)
	}

	dl, err := models.Templates(mod).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}

	return r.ToAPI(dl), nil
}

package template

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"time"

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

type Cursor struct {
	CreatedAt time.Time
	TokenID   []byte
}

var pageHelper = helpers.PaginationHelper[Cursor]{}

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

func (r *Repository) GetTemplates(ctx context.Context, first *int, after *string, last *int, before *string) (*model.TemplateConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	// None now, but making room.
	var queryMods []qm.QueryMod

	totalCount, err := models.Templates().Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if after != nil {
		afterCur, err := pageHelper.DecodeCursor(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, qm.Expr(
			models.TemplateWhere.CreatedAt.EQ(afterCur.CreatedAt),
			models.TemplateWhere.ID.GT(afterCur.TokenID),
			qm.Or2(models.TemplateWhere.CreatedAt.LT(afterCur.CreatedAt)),
		))
	}

	if before != nil {
		beforeCur, err := pageHelper.DecodeCursor(*before)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, qm.Expr(
			models.TemplateWhere.CreatedAt.EQ(beforeCur.CreatedAt),
			models.TemplateWhere.ID.LT(beforeCur.TokenID),
			qm.Or2(models.TemplateWhere.CreatedAt.GT(beforeCur.CreatedAt)),
		))
	}

	orderBy := fmt.Sprintf("%s DESC, %s ASC", models.TemplateColumns.CreatedAt, models.TemplateColumns.ID)
	if last != nil {
		orderBy = fmt.Sprintf("%s ASC, %s DESC", models.TemplateColumns.CreatedAt, models.TemplateColumns.ID)
	}

	queryMods = append(queryMods,
		// Use limit + 1 here to check if there's another page.
		qm.Limit(limit+1),
		qm.OrderBy(orderBy),
	)

	all, err := models.Templates(queryMods...).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	// We assume that cursors come from real elements.
	hasNext := before != nil
	hasPrevious := after != nil

	if first != nil && len(all) == limit+1 {
		hasNext = true
		all = all[:limit]
	} else if last != nil && len(all) == limit+1 {
		hasPrevious = true
		all = all[:limit]
	}

	if last != nil {
		slices.Reverse(all)
	}

	nodes := make([]*model.Template, len(all))
	edges := make([]*model.TemplateEdge, len(all))

	for i, dv := range all {
		dlv := r.ToAPI(dv)

		nodes[i] = dlv

		crsr, err := pageHelper.EncodeCursor(Cursor{
			CreatedAt: dv.CreatedAt,
			TokenID:   dv.ID,
		})
		if err != nil {
			return nil, err
		}

		edges[i] = &model.TemplateEdge{
			Node:   dlv,
			Cursor: crsr,
		}
	}

	var endCur, startCur *string

	if len(edges) != 0 {
		startCur = &edges[0].Cursor
		endCur = &edges[len(edges)-1].Cursor
	}

	res := &model.TemplateConnection{
		Edges: edges,
		Nodes: nodes,
		PageInfo: &model.PageInfo{
			EndCursor:       endCur,
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCur,
		},
		TotalCount: int(totalCount),
	}

	return res, nil
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

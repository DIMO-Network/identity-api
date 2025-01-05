package stake

import (
	"context"
	"slices"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type Repository struct {
	*base.Repository
}

var weiPerEther = decimal.New(params.Ether, 0)

func weiToToken(wei types.Decimal) *decimal.Big {
	return new(decimal.Big).Quo(wei.Big, weiPerEther)
}

func ToAPI(v *models.Stake) *gmodel.Stake {
	return &gmodel.Stake{
		TokenID:     v.ID,
		Owner:       common.BytesToAddress(v.Owner),
		Level:       v.Level + 2, // TODO(elffjs): Is this what we want to do? https://docs.dimo.org/governance/amendments/dip2a4
		Points:      v.Points,
		Amount:      weiToToken(v.Amount),
		StakedAt:    v.StakedAt,
		EndsAt:      v.EndsAt,
		WithdrawnAt: v.WithdrawnAt.Ptr(),
		VehicleID:   v.VehicleID.Ptr(),
	}
}

func (r *Repository) GetStakes(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *gmodel.StakeFilterBy) (*gmodel.StakeConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	// None now, but making room.
	var queryMods []qm.QueryMod

	if filterBy != nil {
		if filterBy.Owner != nil {
			queryMods = append(queryMods, models.StakeWhere.Owner.EQ(filterBy.Owner.Bytes()))
		}
		if filterBy.Attachable != nil {
			if *filterBy.Attachable {
				// Shouldn't need to worry about withdrawn ones.
				queryMods = append(queryMods, models.StakeWhere.VehicleID.IsNull(), models.StakeWhere.EndsAt.GT(time.Now()))
			} else {
				queryMods = append(queryMods, qm.Expr(
					models.StakeWhere.VehicleID.IsNotNull(),
					qm.Or2(models.StakeWhere.EndsAt.LTE(time.Now())),
				))
			}
		}
	}

	totalCount, err := models.Stakes(queryMods...).Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if after != nil {
		afterID, err := helpers.CursorToID(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.StakeWhere.ID.LT(afterID))
	}

	if before != nil {
		beforeID, err := helpers.CursorToID(*before)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.StakeWhere.ID.GT(beforeID))
	}

	orderBy := "DESC"
	if last != nil {
		orderBy = "ASC"
	}

	queryMods = append(queryMods,
		// Use limit + 1 here to check if there's another page.
		qm.Limit(limit+1),
		qm.OrderBy(models.StakeColumns.ID+" "+orderBy),
	)

	all, err := models.Stakes(queryMods...).All(ctx, r.PDB.DBS().Reader)
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

	var endCur, startCur *string
	if len(all) != 0 {
		ec := helpers.IDToCursor(all[len(all)-1].ID)
		endCur = &ec

		sc := helpers.IDToCursor(all[0].ID)
		startCur = &sc
	}

	edges := make([]*gmodel.StakeEdge, len(all))
	nodes := make([]*gmodel.Stake, len(all))

	for i, dv := range all {
		dlv := ToAPI(dv)

		edges[i] = &gmodel.StakeEdge{
			Node:   dlv,
			Cursor: helpers.IDToCursor(dv.ID),
		}

		nodes[i] = dlv
	}

	res := &gmodel.StakeConnection{
		Edges: edges,
		Nodes: nodes,
		PageInfo: &gmodel.PageInfo{
			EndCursor:       endCur,
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCur,
		},
		TotalCount: int(totalCount),
	}

	return res, nil
}

package connection

import (
	"context"
	"fmt"
	"slices"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Repository struct {
	*base.Repository
}

type Cursor struct {
	MintedAt time.Time
	Address  []byte
}

var pageHelper = helpers.PaginationHelper[Cursor]{}

func ToAPI(v *models.Connection) *gmodel.Connection {
	return &gmodel.Connection{
		Name:     v.Name,
		Address:  common.BytesToAddress(v.Address),
		Owner:    common.BytesToAddress(v.Owner),
		MintedAt: v.MintedAt,
	}
}

func (r *Repository) GetConnections(ctx context.Context, first *int, after *string, last *int, before *string) (*gmodel.ConnectionConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	// None now, but making room.
	var queryMods []qm.QueryMod

	totalCount, err := models.Connections().Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if after != nil {
		afterCur, err := pageHelper.DecodeCursor(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, qm.Expr(
			models.ConnectionWhere.MintedAt.EQ(afterCur.MintedAt),
			models.ConnectionWhere.Address.GT(afterCur.Address),
			qm.Or2(models.ConnectionWhere.MintedAt.LT(afterCur.MintedAt)),
		))
	}

	if before != nil {
		beforeCur, err := pageHelper.DecodeCursor(*before)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, qm.Expr(
			models.ConnectionWhere.MintedAt.EQ(beforeCur.MintedAt),
			models.ConnectionWhere.Address.LT(beforeCur.Address),
			qm.Or2(models.ConnectionWhere.MintedAt.GT(beforeCur.MintedAt)),
		))
	}

	orderBy := fmt.Sprintf("%s DESC, %s ASC", models.ConnectionColumns.MintedAt, models.ConnectionColumns.Address)
	if last != nil {
		orderBy = fmt.Sprintf("%s ASC, %s DESC", models.ConnectionColumns.MintedAt, models.ConnectionColumns.Address)
	}

	queryMods = append(queryMods,
		// Use limit + 1 here to check if there's another page.
		qm.Limit(limit+1),
		qm.OrderBy(orderBy),
	)

	all, err := models.Connections(queryMods...).All(ctx, r.PDB.DBS().Reader)
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
		ec, err := pageHelper.EncodeCursor(Cursor{
			MintedAt: all[len(all)-1].MintedAt,
			Address:  all[len(all)-1].Address,
		})
		if err != nil {
			return nil, err
		}
		endCur = &ec

		sc, err := pageHelper.EncodeCursor(Cursor{
			MintedAt: all[0].MintedAt,
			Address:  all[0].Address,
		})
		if err != nil {
			return nil, err
		}
		startCur = &sc
	}

	edges := make([]*gmodel.ConnectionEdge, len(all))
	nodes := make([]*gmodel.Connection, len(all))

	for i, dv := range all {
		dlv := ToAPI(dv)

		crsr, err := pageHelper.EncodeCursor(Cursor{
			MintedAt: dv.MintedAt,
			Address:  dv.Address,
		})
		if err != nil {
			return nil, err
		}

		edges[i] = &gmodel.ConnectionEdge{
			Node:   dlv,
			Cursor: crsr,
		}

		nodes[i] = dlv
	}

	res := &gmodel.ConnectionConnection{
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

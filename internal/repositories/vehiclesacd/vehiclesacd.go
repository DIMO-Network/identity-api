package vehiclesacd

import (
	"context"
	"fmt"
	"math/big"
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

type SacdCursor struct {
	CreatedAt time.Time
	Grantee   []byte
}

func sacdToAPIResponse(pr *models.VehicleSacd) (*gmodel.Sacd, error) {
	b, ok := new(big.Int).SetString(pr.Permissions, 2)
	if !ok {
		return nil, fmt.Errorf("couldn't parse permission string %q as binary", pr.Permissions)
	}

	return &gmodel.Sacd{
		Grantee:     common.BytesToAddress(pr.Grantee),
		Permissions: "0x" + b.Text(16),
		Source:      pr.Source,
		CreatedAt:   pr.CreatedAt,
		ExpiresAt:   pr.ExpiresAt,
	}, nil
}

func (p *Repository) createSacdResponse(sacds models.VehicleSacdSlice, totalCount int64, hasNext, hasPrevious bool, pHelper helpers.PaginationHelper[SacdCursor]) (*gmodel.SacdConnection, error) {
	var endCur, startCur *string

	if len(sacds) != 0 {
		ec, err := pHelper.EncodeCursor(SacdCursor{
			CreatedAt: sacds[len(sacds)-1].CreatedAt,
			Grantee:   sacds[len(sacds)-1].Grantee,
		})
		if err != nil {
			return nil, err
		}
		endCur = &ec

		sc, err := pHelper.EncodeCursor(SacdCursor{
			CreatedAt: sacds[0].CreatedAt,
			Grantee:   sacds[0].Grantee,
		})
		if err != nil {
			return nil, err
		}

		startCur = &sc
	}

	edges := make([]*gmodel.SacdEdge, len(sacds))
	nodes := make([]*gmodel.Sacd, len(sacds))

	for i, dp := range sacds {
		gp, err := sacdToAPIResponse(dp)
		if err != nil {
			return nil, err
		}

		crsr, err := pHelper.EncodeCursor(SacdCursor{
			CreatedAt: dp.CreatedAt,
			Grantee:   dp.Grantee,
		})
		if err != nil {
			return nil, err
		}
		edges[i] = &gmodel.SacdEdge{
			Node:   gp,
			Cursor: crsr,
		}

		nodes[i] = gp
	}

	res := &gmodel.SacdConnection{
		TotalCount: int(totalCount),
		Edges:      edges,
		Nodes:      nodes,
		PageInfo: &gmodel.PageInfo{
			StartCursor:     startCur,
			EndCursor:       endCur,
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
		},
	}

	return res, nil
}

func (p *Repository) GetSacdsForVehicle(ctx context.Context, tokenID int, first *int, after *string, last *int, before *string, grantee *common.Address) (*gmodel.SacdConnection, error) {
	pHelp := helpers.PaginationHelper[SacdCursor]{}

	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{
		models.VehicleSacdWhere.VehicleID.EQ(tokenID),
		models.VehicleSacdWhere.ExpiresAt.GT(time.Now()),
	}

	totalCount, err := models.VehicleSacds(queryMods...).Count(ctx, p.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if totalCount == 0 {
		return &gmodel.SacdConnection{
			TotalCount: int(totalCount),
			Edges:      []*gmodel.SacdEdge{},
			Nodes:      []*gmodel.Sacd{},
			PageInfo:   &gmodel.PageInfo{},
		}, nil
	}

	if after != nil {
		afterCursor, err := pHelp.DecodeCursor(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(
			queryMods,
			qm.Expr(
				models.VehicleSacdWhere.CreatedAt.EQ(afterCursor.CreatedAt),
				models.VehicleSacdWhere.Grantee.GT(afterCursor.Grantee),
				qm.Or2(models.VehicleSacdWhere.CreatedAt.LT(afterCursor.CreatedAt)),
			),
		)
	}

	if before != nil {
		beforeCursor, err := pHelp.DecodeCursor(*before)
		if err != nil {
			return nil, err
		}

		queryMods = append(
			queryMods,
			qm.Expr(
				models.VehicleSacdWhere.CreatedAt.EQ(beforeCursor.CreatedAt),
				models.VehicleSacdWhere.Grantee.LT(beforeCursor.Grantee),
				qm.Or2(models.VehicleSacdWhere.CreatedAt.GT(beforeCursor.CreatedAt)),
			),
		)
	}

	orderBy := fmt.Sprintf("%s DESC, %s ASC", models.VehicleSacdColumns.CreatedAt, models.VehicleSacdColumns.Grantee)
	if last != nil {
		orderBy = fmt.Sprintf("%s ASC, %s DESC", models.VehicleSacdColumns.CreatedAt, models.VehicleSacdColumns.Grantee)
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(orderBy),
	)

	page, err := models.VehicleSacds(queryMods...).All(ctx, p.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if len(page) == 0 {
		return &gmodel.SacdConnection{
			TotalCount: int(totalCount),
			Edges:      []*gmodel.SacdEdge{},
		}, nil
	}

	hasNext := before != nil
	hasPrevious := after != nil

	if first != nil && len(page) == limit+1 {
		hasNext = true
		page = page[:limit]
	} else if last != nil && len(page) == limit+1 {
		hasPrevious = true
		page = page[:limit]
	}

	if last != nil {
		slices.Reverse(page)
	}

	return p.createSacdResponse(page, totalCount, hasNext, hasPrevious, pHelp)
}

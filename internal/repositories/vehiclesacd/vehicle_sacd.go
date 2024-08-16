package vehiclesacd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const defaultPageSize = 20

type Repository struct {
	*base.Repository
}

type SacdCursor struct {
	CreatedAt   time.Time
	Permissions string
	Grantee     []byte
}

func sacdToAPIResponse(pr *models.Sacd) (*gmodel.Sacd, error) {
	ui, err := strconv.ParseUint(pr.Permissions, 2, 64)
	if err != nil {
		return nil, err
	}

	return &gmodel.Sacd{
		Grantee:     common.BytesToAddress(pr.Grantee),
		Permissions: fmt.Sprintf("0x%x", ui),
		Source:      pr.Source.String,
		CreatedAt:   pr.CreatedAt,
		ExpiresAt:   pr.ExpiresAt,
	}, err
}

func (p *Repository) createSacdResponse(sacds models.SacdSlice, totalCount int64, hasNext bool, pHelper helpers.PaginationHelper[SacdCursor]) (*gmodel.SacdsConnection, error) {
	lastPriv := sacds[len(sacds)-1]
	endCursr, err := pHelper.EncodeCursor(SacdCursor{
		CreatedAt:   lastPriv.CreatedAt,
		Permissions: lastPriv.Permissions,
		Grantee:     lastPriv.Grantee,
	})

	if err != nil {
		return nil, err
	}

	edges := make([]*gmodel.SacdEdge, len(sacds))
	nodes := make([]*gmodel.Sacd, len(sacds))

	for i, dp := range sacds {
		gp, err := sacdToAPIResponse(dp)
		if err != nil {
			return nil, err
		}

		crsr, err := pHelper.EncodeCursor(SacdCursor{
			CreatedAt:   dp.CreatedAt,
			Permissions: dp.Permissions,
			Grantee:     dp.Grantee,
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

	res := &gmodel.SacdsConnection{
		TotalCount: int(totalCount),
		Edges:      edges,
		Nodes:      nodes,
		PageInfo: &gmodel.PageInfo{
			EndCursor:   &endCursr,
			HasNextPage: hasNext,
		},
	}

	return res, nil
}

func (p *Repository) GetSacdsForVehicle(ctx context.Context, tokenID int, first *int, after *string, last *int, before *string, grantee *common.Address) (*gmodel.SacdsConnection, error) {
	pHelp := helpers.PaginationHelper[SacdCursor]{}

	limit := defaultPageSize
	if first != nil {
		limit = *first
		if limit <= 0 {
			return nil, errors.New("invalid value provided for number of sacd permissions to retrieve")
		}
	}

	queryMods := []qm.QueryMod{
		models.SacdWhere.TokenID.EQ(tokenID),
		models.SacdWhere.ExpiresAt.GTE(time.Now()),
	}

	if grantee != nil {
		queryMods = append(queryMods, models.SacdWhere.Grantee.EQ(grantee.Bytes()))
	}

	totalCount, err := models.Sacds(queryMods...).Count(ctx, p.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if totalCount == 0 {
		return &gmodel.SacdsConnection{
			TotalCount: int(totalCount),
			Edges:      []*gmodel.SacdEdge{},
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
				models.SacdWhere.CreatedAt.EQ(afterCursor.CreatedAt),
				qm.And(
					fmt.Sprintf("(%s, %s) > (?, ?)", models.SacdColumns.Permissions, models.SacdColumns.Grantee),
					afterCursor.Permissions, afterCursor.Grantee,
				),
				qm.Or2(models.SacdWhere.CreatedAt.LT(afterCursor.CreatedAt)),
			),
		)
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(
			fmt.Sprintf("%s DESC, (%s, %s)", models.SacdColumns.CreatedAt, models.SacdColumns.Permissions, models.SacdColumns.Grantee),
		),
	)

	page, err := models.Sacds(queryMods...).All(ctx, p.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if len(page) == 0 {
		return &gmodel.SacdsConnection{
			TotalCount: int(totalCount),
			Edges:      []*gmodel.SacdEdge{},
		}, nil
	}

	hasNext := len(page) > limit
	if hasNext {
		page = page[:limit]
	}

	return p.createSacdResponse(page, totalCount, hasNext, pHelp)
}

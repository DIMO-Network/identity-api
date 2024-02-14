package vehicleprivilege

import (
	"context"
	"errors"
	"fmt"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const defaultPageSize = 20

type Repository struct {
	*repositories.Repository
}

type PrivilegeCursor struct {
	SetAt       time.Time
	PrivilegeID int
	User        []byte
}

func privilegeToAPIResponse(pr *models.Privilege) *gmodel.Privilege {
	return &gmodel.Privilege{
		ID:        pr.PrivilegeID,
		User:      common.Address(pr.UserAddress),
		SetAt:     pr.SetAt,
		ExpiresAt: pr.ExpiresAt,
	}
}

func (p *Repository) createPrivilegeResponse(privs models.PrivilegeSlice, totalCount int64, hasNext bool, pHelper helpers.PaginationHelper[PrivilegeCursor]) (*gmodel.PrivilegesConnection, error) {
	lastPriv := privs[len(privs)-1]
	endCursr, err := pHelper.EncodeCursor(PrivilegeCursor{
		SetAt:       lastPriv.SetAt,
		PrivilegeID: lastPriv.PrivilegeID,
		User:        lastPriv.UserAddress,
	})

	if err != nil {
		return nil, err
	}

	edges := make([]*gmodel.PrivilegeEdge, len(privs))
	nodes := make([]*gmodel.Privilege, len(privs))

	for i, dp := range privs {
		gp := privilegeToAPIResponse(dp)

		crsr, err := pHelper.EncodeCursor(PrivilegeCursor{
			SetAt:       dp.SetAt,
			PrivilegeID: dp.PrivilegeID,
			User:        dp.UserAddress,
		})
		if err != nil {
			return nil, err
		}
		edges[i] = &gmodel.PrivilegeEdge{
			Node:   gp,
			Cursor: crsr,
		}

		nodes[i] = gp
	}

	res := &gmodel.PrivilegesConnection{
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

func (p *Repository) GetPrivilegesForVehicle(ctx context.Context, tokenID int, first *int, after *string, last *int, before *string, filterBy *gmodel.PrivilegeFilterBy) (*gmodel.PrivilegesConnection, error) {
	pHelp := helpers.PaginationHelper[PrivilegeCursor]{}

	limit := defaultPageSize
	if first != nil {
		limit = *first
		if limit <= 0 {
			return nil, errors.New("invalid value provided for number of privileges to retrieve")
		}
	}

	queryMods := []qm.QueryMod{
		models.PrivilegeWhere.TokenID.EQ(tokenID),
		models.PrivilegeWhere.ExpiresAt.GTE(time.Now()),
	}

	if filterBy != nil && filterBy.User != nil {
		queryMods = append(queryMods, models.PrivilegeWhere.UserAddress.EQ(filterBy.User.Bytes()))
	}

	totalCount, err := models.Privileges(queryMods...).Count(ctx, p.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if totalCount == 0 {
		return &gmodel.PrivilegesConnection{
			TotalCount: int(totalCount),
			Edges:      []*gmodel.PrivilegeEdge{},
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
				models.PrivilegeWhere.SetAt.EQ(afterCursor.SetAt),
				qm.And(
					fmt.Sprintf("(%s, %s) > (?, ?)", models.PrivilegeColumns.PrivilegeID, models.PrivilegeColumns.UserAddress),
					afterCursor.PrivilegeID, afterCursor.User,
				),
				qm.Or2(models.PrivilegeWhere.SetAt.LT(afterCursor.SetAt)),
			),
		)
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(
			fmt.Sprintf("%s DESC, (%s, %s)", models.PrivilegeColumns.SetAt, models.PrivilegeColumns.PrivilegeID, models.PrivilegeColumns.UserAddress),
		),
	)

	page, err := models.Privileges(queryMods...).All(ctx, p.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if len(page) == 0 {
		return &gmodel.PrivilegesConnection{
			TotalCount: int(totalCount),
			Edges:      []*gmodel.PrivilegeEdge{},
		}, nil
	}

	hasNext := len(page) > limit
	if hasNext {
		page = page[:limit]
	}

	return p.createPrivilegeResponse(page, totalCount, hasNext, pHelp)
}

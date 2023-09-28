package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type PrivilegeCursor struct {
	SetAt       time.Time
	PrivilegeID int
	User        []byte
}

func (p *Repository) privilegeToAPIResponse(pr *models.Privilege) *gmodel.Privilege {
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

	var pEdges []*gmodel.PrivilegeEdge
	for _, pr := range privs {
		crsr, err := pHelper.EncodeCursor(PrivilegeCursor{
			SetAt:       pr.SetAt,
			PrivilegeID: pr.PrivilegeID,
			User:        pr.UserAddress,
		})
		if err != nil {
			return nil, err
		}
		edge := &gmodel.PrivilegeEdge{
			Node:   p.privilegeToAPIResponse(pr),
			Cursor: crsr,
		}

		pEdges = append(pEdges, edge)
	}

	res := &gmodel.PrivilegesConnection{
		TotalCount: int(totalCount),
		Edges:      pEdges,
		PageInfo: &gmodel.PageInfo{
			EndCursor:   &endCursr,
			HasNextPage: hasNext,
		},
	}

	return res, nil
}

func (p *Repository) GetPrivilegesForVehicle(ctx context.Context, tokenID int, filterBy gmodel.PrivilegeFilterBy) (*gmodel.PrivilegesConnection, error) {
	pHelp := helpers.PaginationHelper[PrivilegeCursor]{}

	limit := defaultPageSize
	if filterBy.First != nil {
		limit = *filterBy.First
		if limit <= 0 {
			return nil, errors.New("invalid value provided for number of privileges to retrieve")
		}
	}

	queryMods := []qm.QueryMod{
		models.PrivilegeWhere.TokenID.EQ(tokenID),
		models.PrivilegeWhere.ExpiresAt.GTE(time.Now()),
	}

	if filterBy.User != nil {
		queryMods = append(queryMods, models.PrivilegeWhere.UserAddress.EQ(filterBy.User.Bytes()))
	}

	totalCount, err := models.Privileges(queryMods...).Count(ctx, p.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if totalCount == 0 {
		return &gmodel.PrivilegesConnection{
			TotalCount: int(totalCount),
			Edges:      []*gmodel.PrivilegeEdge{},
		}, nil
	}

	if filterBy.After != nil {
		afterCursor, err := pHelp.DecodeCursor(*filterBy.After)
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

	page, err := models.Privileges(queryMods...).All(ctx, p.pdb.DBS().Reader)
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

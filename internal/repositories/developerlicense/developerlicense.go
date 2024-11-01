package developerlicense

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
)

type Repository struct {
	*base.Repository
}

func ToAPI(v *models.DeveloperLicense) *gmodel.DeveloperLicense {
	return &gmodel.DeveloperLicense{
		TokenID:  v.ID,
		Owner:    common.BytesToAddress(v.Owner),
		ClientID: common.BytesToAddress(v.ClientID),
		Alias:    v.Alias.Ptr(),
		MintedAt: v.MintedAt,
	}
}

func SignerToAPI(v *models.Signer) *gmodel.Signer {
	return &gmodel.Signer{
		Address:   common.BytesToAddress(v.Signer),
		EnabledAt: v.EnabledAt,
	}
}

func (r *Repository) GetDeveloperLicenses(ctx context.Context, first *int, after *string, last *int, before *string) (*gmodel.DeveloperLicenseConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	totalCount, err := models.DeveloperLicenses().Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	var queryMods []qm.QueryMod

	if after != nil {
		afterID, err := helpers.CursorToID(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.DeveloperLicenseWhere.ID.LT(afterID))
	}

	if before != nil {
		beforeID, err := helpers.CursorToID(*before)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.DeveloperLicenseWhere.ID.GT(beforeID))
	}

	orderBy := "DESC"
	if last != nil {
		orderBy = "ASC"
	}

	queryMods = append(queryMods,
		// Use limit + 1 here to check if there's another page.
		qm.Limit(limit+1),
		qm.OrderBy(models.DeveloperLicenseColumns.ID+" "+orderBy),
	)

	all, err := models.DeveloperLicenses(queryMods...).All(ctx, r.PDB.DBS().Reader)
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

	edges := make([]*gmodel.DeveloperLicenseEdge, len(all))
	nodes := make([]*gmodel.DeveloperLicense, len(all))

	for i, dv := range all {
		dlv := ToAPI(dv)

		edges[i] = &gmodel.DeveloperLicenseEdge{
			Node:   dlv,
			Cursor: helpers.IDToCursor(dv.ID),
		}

		nodes[i] = dlv
	}

	res := &gmodel.DeveloperLicenseConnection{
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

type SignerCursor struct {
	EnabledAt time.Time
	Signer    [20]byte
}

func (r *Repository) GetSignersForLicense(ctx context.Context, obj *gmodel.DeveloperLicense, first *int, after *string, last *int, before *string) (*gmodel.SignerConnection, error) {
	pHelp := helpers.PaginationHelper[SignerCursor]{}

	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	totalCount, err := models.Signers().Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{
		models.SignerWhere.DeveloperLicenseID.EQ(obj.TokenID),
	}

	if after != nil {
		afterCursor, err := pHelp.DecodeCursor(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods,
			qm.Expr(
				models.SignerWhere.EnabledAt.EQ(afterCursor.EnabledAt),
				models.SignerWhere.Signer.GT(afterCursor.Signer[:]),
				qm.Or2(models.SignerWhere.EnabledAt.LT(afterCursor.EnabledAt)),
			),
		)
	}

	if before != nil {
		beforeCursor, err := pHelp.DecodeCursor(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods,
			qm.Expr(
				models.SignerWhere.EnabledAt.EQ(beforeCursor.EnabledAt),
				models.SignerWhere.Signer.LT(beforeCursor.Signer[:]),
				qm.Or2(models.SignerWhere.EnabledAt.GT(beforeCursor.EnabledAt)),
			),
		)
	}

	orderBy := fmt.Sprintf("%s DESC, %s ASC", models.SignerColumns.EnabledAt, models.SignerColumns.Signer)
	if last != nil {
		orderBy = fmt.Sprintf("%s ASC, %s DESC", models.SignerColumns.EnabledAt, models.SignerColumns.Signer)
	}

	queryMods = append(queryMods,
		// Use limit + 1 here to check if there's another page.
		qm.Limit(limit+1),
		qm.OrderBy(orderBy),
	)

	all, err := models.Signers(queryMods...).All(ctx, r.PDB.DBS().Reader)
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
		ec, err := pHelp.EncodeCursor(SignerCursor{all[len(all)-1].EnabledAt, common.BytesToAddress(all[len(all)-1].Signer)})
		if err != nil {
			return nil, err
		}
		endCur = &ec

		sc, err := pHelp.EncodeCursor(SignerCursor{all[0].EnabledAt, common.BytesToAddress(all[0].Signer)})
		if err != nil {
			return nil, err
		}
		startCur = &sc
	}

	edges := make([]*gmodel.SignerEdge, len(all))
	nodes := make([]*gmodel.Signer, len(all))

	for i, dv := range all {
		dlv := SignerToAPI(dv)

		crs, err := pHelp.EncodeCursor(SignerCursor{dv.EnabledAt, common.BytesToAddress(dv.Signer)})
		if err != nil {
			return nil, err
		}

		edges[i] = &gmodel.SignerEdge{
			Node:   dlv,
			Cursor: crs,
		}

		nodes[i] = dlv
	}

	res := &gmodel.SignerConnection{
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

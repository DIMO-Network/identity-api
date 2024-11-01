package devicelicense

import (
	"context"
	"slices"

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

		queryMods = append(queryMods, models.VehicleWhere.ID.LT(afterID))
	}

	if before != nil {
		beforeID, err := helpers.CursorToID(*before)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.VehicleWhere.ID.GT(beforeID))
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

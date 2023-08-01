package repositories

import (
	"context"
	"errors"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	defaultPageSize = 20
)

type Repository struct {
	pdb db.Store
}

func New(pdb db.Store) *Repository {
	return &Repository{
		pdb: pdb,
	}
}

func (v *Repository) createVehiclesResponse(totalCount int64, vehicles models.VehicleSlice, hasNext bool) *gmodel.VehicleConnection {
	endCursr := helpers.IDToCursor(vehicles[len(vehicles)-1].ID)

	var vEdges []*gmodel.VehicleEdge
	for _, v := range vehicles {
		edge := &gmodel.VehicleEdge{
			Node: &gmodel.Vehicle{
				ID:       v.ID,
				Owner:    common.BytesToAddress(v.OwnerAddress),
				Make:     v.Make.Ptr(),
				Model:    v.Model.Ptr(),
				Year:     v.Year.Ptr(),
				MintedAt: v.MintedAt,
			},
			Cursor: helpers.IDToCursor(v.ID),
		}

		vEdges = append(vEdges, edge)
	}

	res := &gmodel.VehicleConnection{
		TotalCount: int(totalCount),
		PageInfo: &gmodel.PageInfo{
			HasNextPage: hasNext,
			EndCursor:   &endCursr,
		},
		Edges: vEdges,
	}

	return res
}

func (v *Repository) GetAccessibleVehicles(ctx context.Context, addr common.Address, first *int, after *string) (*gmodel.VehicleConnection, error) {
	limit := defaultPageSize
	if first != nil {
		limit = *first
		if limit <= 0 {
			return nil, errors.New("invalid value provided for number of vehicles to retrieve")
		}
	}

	queryMods := []qm.QueryMod{
		qm.Select("DISTINCT ON (" + models.VehicleTableColumns.ID + ") " + helpers.WithSchema(models.TableNames.Vehicles) + ".*"),
		qm.LeftOuterJoin(
			helpers.WithSchema(models.TableNames.Privileges) + " ON " + models.VehicleTableColumns.ID + " = " + models.PrivilegeTableColumns.TokenID,
		),
		models.VehicleWhere.OwnerAddress.EQ(addr.Bytes()),
		qm.Or2(models.PrivilegeWhere.UserAddress.EQ(addr.Bytes())),
		// Use limit + 1 here to check if there's a next page.
	}

	totalCount, err := models.Vehicles(queryMods...).Count(ctx, v.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if totalCount == 0 {
		return &gmodel.VehicleConnection{
			TotalCount: 0,
			Edges:      []*gmodel.VehicleEdge{},
			PageInfo:   &gmodel.PageInfo{},
		}, nil
	}

	if after != nil {
		afterID, err := helpers.CursorToID(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.VehicleWhere.ID.LT(afterID))
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(models.VehicleColumns.ID+" DESC"),
	)

	all, err := models.Vehicles(queryMods...).All(ctx, v.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if len(all) == 0 {
		return &gmodel.VehicleConnection{
			TotalCount: int(totalCount),
			Edges:      []*gmodel.VehicleEdge{},
		}, nil
	}

	hasNext := len(all) > limit
	if hasNext {
		all = all[:limit]
	}

	return v.createVehiclesResponse(totalCount, all, hasNext), nil
}

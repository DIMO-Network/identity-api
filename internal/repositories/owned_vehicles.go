package repositories

import (
	"context"
	"errors"
	"fmt"

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
	PDB      db.Store
	PageSize int
}

func NewRepository(pdb db.Store, pgSize int) *Repository {
	if pgSize == 0 {
		pgSize = defaultPageSize
	}
	return &Repository{
		PDB:      pdb,
		PageSize: pgSize,
	}
}

func (r *Repository) GetOwnedVehicles(ctx context.Context, addr common.Address, first *int, after *string) (*gmodel.VehicleConnection, error) {
	totalCount, err := models.Vehicles(
		models.VehicleWhere.OwnerAddress.EQ(addr.Bytes()),
	).Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if totalCount == 0 {
		return &gmodel.VehicleConnection{
			TotalCount: 0,
			Edges:      []*gmodel.VehicleEdge{},
		}, nil
	}

	limit := r.PageSize
	if first != nil {
		limit = *first
		if limit <= 0 {
			return nil, errors.New("invalid value provided for number of vehicles to retrieve")
		}
	}

	queryMods := []qm.QueryMod{
		models.VehicleWhere.OwnerAddress.EQ(addr.Bytes()),
		// Use limit + 1 here to check if there's a next page.
		qm.Limit(limit + 1),
		qm.OrderBy(models.VehicleColumns.ID + " DESC"),
		// qm.Load(models.VehicleRels.TokenPrivileges, models.PrivilegeWhere.ExpiresAt.GTE(time.Now())),
	}

	if after != nil {
		afterID, err := helpers.CursorToID(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor %q", *after)
		}

		queryMods = append(queryMods, models.VehicleWhere.ID.LT(afterID))
	}

	vehicles, err := models.Vehicles(queryMods...).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if len(vehicles) == 0 {
		return &gmodel.VehicleConnection{
			TotalCount: int(totalCount),
			Edges:      []*gmodel.VehicleEdge{},
		}, nil
	}

	hasNextPage := len(vehicles) > limit
	if hasNextPage {
		vehicles = vehicles[:limit]
	}

	lastItmID := vehicles[len(vehicles)-1].ID
	endCursr := helpers.IDToCursor(lastItmID)

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
			HasNextPage: hasNextPage,
			EndCursor:   &endCursr,
		},
		Edges: vEdges,
	}

	return res, nil
}

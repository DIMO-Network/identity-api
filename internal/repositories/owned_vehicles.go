package repositories

import (
	"context"
	"errors"
	"strconv"

	"encoding/base64"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
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
	}

	if after != nil {
		searchAfter, err := strconv.Atoi(string([]byte(*after)))
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.VehicleWhere.ID.LT(searchAfter))
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
	endCursr := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(lastItmID)))

	var vEdges []*gmodel.VehicleEdge
	for _, v := range vehicles {
		edge := &gmodel.VehicleEdge{
			Node: &gmodel.Vehicle{
				ID:       strconv.Itoa(v.ID),
				Owner:    common.BytesToAddress(v.OwnerAddress),
				Make:     v.Make.Ptr(),
				Model:    v.Model.Ptr(),
				Year:     v.Year.Ptr(),
				MintedAt: v.MintedAt,
			},
			Cursor: base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(v.ID))),
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

func (r *Repository) GetLinkedVehicleByID(ctx context.Context, aftermarketDevID string) (*gmodel.Vehicle, error) {
	adID, err := strconv.Atoi(aftermarketDevID)
	if err != nil {
		return nil, err
	}

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.ID.EQ(adID),
		qm.Load(models.AftermarketDeviceRels.Vehicle),
	).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if ad.R.Vehicle == nil {
		return nil, nil
	}

	res := &gmodel.Vehicle{
		ID:       strconv.Itoa(ad.R.Vehicle.ID),
		Owner:    common.BytesToAddress(ad.R.Vehicle.OwnerAddress),
		Make:     ad.R.Vehicle.Make.Ptr(),
		Model:    ad.R.Vehicle.Model.Ptr(),
		Year:     ad.R.Vehicle.Year.Ptr(),
		MintedAt: ad.R.Vehicle.MintedAt,
	}

	return res, nil
}

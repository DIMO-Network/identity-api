package controllers

import (
	"context"
	"errors"
	"strconv"

	"encoding/base64"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	defaultPageSize = 20
)

type VehiclesRepo struct {
	ctx context.Context
	pdb db.Store
}

func NewVehiclesRepo(ctx context.Context, pdb db.Store) VehiclesRepo {
	return VehiclesRepo{
		ctx: ctx,
		pdb: pdb,
	}
}

func (v *VehiclesRepo) createVehiclesResponse(totalCount int64, vehicles []models.Vehicle, hasNext bool) *gmodel.VehicleConnection {
	lastItmID := vehicles[len(vehicles)-1].ID
	endCursr := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(lastItmID)))

	var vEdges []*gmodel.VehicleEdge
	for _, v := range vehicles {
		crs := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(v.ID)))
		cursor := crs

		owner := common.BytesToAddress(v.OwnerAddress.Bytes)

		edge := &gmodel.VehicleEdge{
			Node: &gmodel.Vehicle{
				ID:       strconv.Itoa(v.ID),
				Owner:    &owner,
				Make:     v.Make.Ptr(),
				Model:    v.Model.Ptr(),
				Year:     v.Year.Ptr(),
				MintedAt: v.MintedAt.Ptr(),
			},
			Cursor: cursor,
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

func (v *VehiclesRepo) GetOwnedVehicles(addr common.Address, first *int, after *string) (*gmodel.VehicleConnection, error) {
	totalCount, err := models.Vehicles(
		models.VehicleWhere.OwnerAddress.EQ(null.BytesFrom(addr.Bytes())),
	).Count(v.ctx, v.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if totalCount == 0 {
		return &gmodel.VehicleConnection{
			TotalCount: 0,
			Edges:      []*gmodel.VehicleEdge{},
		}, nil
	}

	limit := defaultPageSize
	if first != nil {
		limit = *first
		if limit <= 0 {
			return nil, errors.New("invalid value provided for number of vehicles to retrieve")
		}
	}

	var queryMods []qm.QueryMod

	queryMods = append(queryMods, models.VehicleWhere.OwnerAddress.EQ(null.BytesFrom(addr.Bytes())))
	// Used to determine if there are more values, if returned is not up to limit + 1 then we don't have anymore records
	queryMods = append(queryMods, qm.Limit(limit+1))

	if after != nil {
		lastCursor, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return nil, err
		}

		lastCursorVal, err := strconv.Atoi(string(lastCursor))
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods, models.VehicleWhere.ID.LT(lastCursorVal))
	}
	queryMods = append(queryMods, qm.OrderBy(models.VehicleColumns.ID+" DESC"))

	all, err := models.Vehicles(queryMods...).All(v.ctx, v.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if len(all) == 0 {
		return &gmodel.VehicleConnection{
			TotalCount: int(totalCount),
			Edges:      []*gmodel.VehicleEdge{},
		}, nil
	}

	vIterate := all
	if len(all) > limit {
		vIterate = all[:limit]
	}

	vehicles := []models.Vehicle{}
	for _, v := range vIterate {
		vehicles = append(vehicles, models.Vehicle{
			ID:           v.ID,
			OwnerAddress: v.OwnerAddress,
			Make:         v.Make,
			Model:        v.Model,
			Year:         v.Year,
			MintedAt:     v.MintedAt,
		})
	}

	return v.createVehiclesResponse(totalCount, vehicles, len(all) > limit), nil
}

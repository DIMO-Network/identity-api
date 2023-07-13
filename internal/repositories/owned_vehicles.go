package controllers

import (
	"context"
	"fmt"
	"strconv"

	b64 "encoding/base64"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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

func (v *VehiclesRepo) createVehiclesResponse(totalCount int64, vhs []models.Vehicle, hasNext bool) *gmodel.Vehicles {
	lastItmID := vhs[len(vhs)-1].ID
	endCursr := b64.StdEncoding.EncodeToString([]byte(strconv.Itoa(lastItmID)))

	var vEdges []*gmodel.VehicleEdge
	for _, v := range vhs {
		crs := b64.StdEncoding.EncodeToString([]byte(strconv.Itoa(v.ID)))
		cursor := &crs
		edge := &gmodel.VehicleEdge{
			Node: &gmodel.Vehicle{
				ID:       v.ID,
				Owner:    common.BytesToAddress(v.OwnerAddress.Bytes),
				Make:     v.Make.String,
				Model:    v.Model.String,
				Year:     v.Year.Int,
				MintTime: v.MintTime.Time,
			},
			Cursor: cursor,
		}

		vEdges = append(vEdges, edge)
	}

	res := &gmodel.Vehicles{
		TotalCount: &[]int{int(totalCount)}[0],
		PageInfo: &gmodel.PageInfo{
			HasNextPage: &hasNext,
			EndCursor:   endCursr,
		},
		Edges: vEdges,
	}

	return res
}

func (v *VehiclesRepo) GetOwnedVehicles(addr common.Address, first *int, after *string) (*gmodel.Vehicles, error) {
	tCount, err := models.Vehicles(
		models.VehicleWhere.OwnerAddress.EQ(null.BytesFrom(addr.Bytes())),
	).Count(v.ctx, v.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if tCount == 0 {
		return &gmodel.Vehicles{
			TotalCount: &[]int{int(0)}[0],
			Edges:      []*gmodel.VehicleEdge{},
		}, nil
	}

	var limit int

	if first == nil {
		limit = 20
	} else {
		limit = *first
	}
	var queryMods []qm.QueryMod

	queryMods = append(queryMods, models.VehicleWhere.OwnerAddress.EQ(null.BytesFrom(addr.Bytes())))
	// Used to determine if there are more values, if returned is not up to limit + 1 then we don't have anymore records
	queryMods = append(queryMods, qm.Limit(limit+1))

	if after != nil {
		lastCursor, err := b64.StdEncoding.DecodeString(*after)
		if err != nil {
			return nil, err
		}

		lastCursorVal, err := strconv.Atoi(string(lastCursor))
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods, models.VehicleWhere.ID.GT(lastCursorVal))
	}
	queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s ASC", models.VehicleColumns.ID)))

	all, err := models.Vehicles(queryMods...).All(v.ctx, v.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if len(all) == 0 {
		return &gmodel.Vehicles{
			TotalCount: &[]int{int(0)}[0],
			Edges:      []*gmodel.VehicleEdge{},
		}, nil

	}

	vhs := []models.Vehicle{}
	for _, v := range all[:limit] {
		vhs = append(vhs, models.Vehicle{
			ID:           v.ID,
			OwnerAddress: v.OwnerAddress,
			Make:         v.Make,
			Model:        v.Model,
			Year:         v.Year,
			MintTime:     v.MintTime,
		})
	}

	return v.createVehiclesResponse(tCount, vhs, len(all) > limit), nil
}

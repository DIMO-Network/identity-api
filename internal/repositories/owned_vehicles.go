package controllers

import (
	"context"
	"fmt"
	"log"

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

/* func (v *VehiclesRepo) createVehiclesResponse(totalCount int) {


} */

func (v *VehiclesRepo) GetOwnedVehicles(addr common.Address, first *int, after *string) ([]*gmodel.Vehicles, error) {
	limit := *first

	if first == nil {
		limit = 20
	}
	var queryMods []qm.QueryMod

	queryMods = append(queryMods, models.VehicleWhere.OwnerAddress.EQ(null.BytesFrom(addr.Bytes())))
	queryMods = append(queryMods, qm.Limit(limit))

	/* if after != nil {
		lastCursor, err := b64.StdEncoding.DecodeString(*after)
		if err != nil {
			return nil, err
		}

		lastCursorVal, err := strconv.Atoi(string(lastCursor))
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods, models.VehicleWhere.ID.GT(lastCursorVal))
	} */
	queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s DESC", models.VehicleColumns.ID)))

	// all, err := models.Vehicles(queryMods...).All(v.ctx, v.pdb.DBS().Reader)

	/* tCount, err := models.Vehicles(
		models.VehicleWhere.OwnerAddress.EQ(null.BytesFrom(addr.Bytes())),
	).Count(v.ctx, v.pdb.DBS().Reader) */
	// TODO - return from here if tCount == 0
	mv, err := models.Vehicles(
		models.VehicleWhere.OwnerAddress.EQ(null.BytesFrom(addr.Bytes())),
		// models.VehicleWhere.ID.GT(),
		qm.OrderBy(fmt.Sprintf("%s DESC", models.VehicleColumns.ID)),
	).All(v.ctx, v.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}
	log.Println(mv)
	res := []*gmodel.Vehicles{}
	/* for _, m := range mv {
		res = append(res, &gmodel.Vehicle{
			ID:       m.ID.String(),
			Owner:    addr,
			Make:     m.Make,
			Model:    m.Model,
			Year:     int(m.Year),
			MintTime: m.MintTime,
		})
	} */

	return res, nil
}

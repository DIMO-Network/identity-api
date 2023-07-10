package controllers

import (
	"context"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
)

type VehiclesCtrl struct {
	ctx context.Context
	pdb db.Store
}

func NewVehiclesRepo(ctx context.Context, pdb db.Store) VehiclesCtrl {
	return VehiclesCtrl{
		ctx: ctx,
		pdb: pdb,
	}
}

func (v *VehiclesCtrl) GetOwnedVehicles(addr common.Address) ([]*gmodel.Vehicle, error) {
	mv, err := models.Vehicles(
		models.VehicleWhere.OwnerAddress.EQ(null.BytesFrom(addr.Bytes())),
	).All(v.ctx, v.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	res := []*gmodel.Vehicle{}
	for _, m := range mv {
		res = append(res, &gmodel.Vehicle{
			ID:       m.ID.String(),
			Owner:    addr,
			Make:     m.Make,
			Model:    m.Model,
			Year:     int(m.Year),
			MintTime: m.MintTime,
		})
	}

	return res, nil
}

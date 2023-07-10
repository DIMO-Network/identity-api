package controllers

import (
	"context"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
)

type AftermarketDevicesCtrl struct {
	ctx context.Context
	pdb db.Store
}

func NewADRepo(ctx context.Context, pdb db.Store) AftermarketDevicesCtrl {
	return AftermarketDevicesCtrl{
		ctx: ctx,
		pdb: pdb,
	}
}

func (ad *AftermarketDevicesCtrl) GetAftermarketDevices(addr common.Address) ([]*gmodel.AftermarketDevice, error) {
	ads, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.OwnerAddress.EQ(null.BytesFrom(addr.Bytes())),
	).All(ad.ctx, ad.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	res := []*gmodel.AftermarketDevice{}
	for _, d := range ads {
		res = append(res, &gmodel.AftermarketDevice{
			ID:                 d.ID.String(),
			Owner:              addr,
			BeneficiaryAddress: common.BytesToAddress(d.BeneficiaryAddress.Bytes),
			VehicleID:          d.VehicleID.String(),
			MintTime:           d.MintTime,
		})
	}

	return res, nil
}

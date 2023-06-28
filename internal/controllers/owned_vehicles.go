package controllers

import (
	"context"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/segmentio/ksuid"
)

type VehiclesCtrl struct{}

func NewVehiclesCtrl() VehiclesCtrl {
	return VehiclesCtrl{}
}

func (v *VehiclesCtrl) GetOwnedVehicles(ctx context.Context, addr string) ([]*gmodel.Vehicle, error) {
	res := []*gmodel.Vehicle{
		{
			ID:       ksuid.New().String(),
			Owner:    common.HexToAddress(addr).Bytes(),
			Make:     "someMake",
			Model:    "someModel",
			Year:     2022,
			MintTime: time.Now(),
		},
	}

	return res, nil
}

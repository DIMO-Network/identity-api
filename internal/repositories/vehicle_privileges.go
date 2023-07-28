package repositories

import (
	"context"
	"strconv"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
)

func (v *VehiclesRepo) GetPrivilegesForVehicles(ctx context.Context, vehicle *gmodel.Vehicle) ([]*gmodel.Privilege, error) {
	tkID, err := strconv.Atoi(vehicle.ID)
	if err != nil {
		return nil, err
	}
	privileges, err := models.Privileges(
		models.PrivilegeWhere.TokenID.EQ(tkID),
		models.PrivilegeWhere.ExpiresAt.GTE(time.Now()),
	).All(v.ctx, v.pdb.DBS().Reader)

	if err != nil {
		return nil, err
	}

	privs := []*gmodel.Privilege{}
	for _, p := range privileges {
		privs = append(privs, &gmodel.Privilege{
			User:      common.BytesToAddress(p.UserAddress),
			SetAt:     p.SetAt,
			ExpiresAt: p.ExpiresAt,
			ID:        p.PrivilegeID,
		})
	}

	return privs, nil
}

package repositories

import (
	"context"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
)

func (v *VehiclesRepo) GetPrivilegesForVehicle(ctx context.Context, vehicleID int) ([]*gmodel.Privilege, error) {
	privileges, err := models.Privileges(
		models.PrivilegeWhere.TokenID.EQ(vehicleID),
		models.PrivilegeWhere.ExpiresAt.GTE(time.Now()),
	).All(ctx, v.pdb.DBS().Reader)

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

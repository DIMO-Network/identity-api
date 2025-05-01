package loader

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/stake"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/graph-gophers/dataloader/v7"
)

type StakeLoader struct {
	db db.Store
}

func GetStakeByVehicleID(ctx context.Context, vehicleID int) (*model.Stake, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	// invoke and get thunk
	thunk := loaders.StakeByVehicleID.Load(ctx, vehicleID)
	// read value from thunk
	return thunk()
}

func (ad *StakeLoader) BatchGetLinkedStakesByVehicleID(ctx context.Context, vehicleIDs []int) []*dataloader.Result[*model.Stake] {
	results := make([]*dataloader.Result[*model.Stake], len(vehicleIDs))

	stakes, err := models.Stakes(models.StakeWhere.VehicleID.IN(vehicleIDs)).All(ctx, ad.db.DBS().Reader)
	if err != nil {
		for i := range vehicleIDs {
			results[i] = &dataloader.Result[*model.Stake]{Data: nil, Error: err}
		}
		return results
	}

	stakeByVehicleID := map[int]*models.Stake{}

	for _, d := range stakes {
		stakeByVehicleID[d.VehicleID.Int] = d
	}

	for i, vID := range vehicleIDs {
		if am, ok := stakeByVehicleID[vID]; ok {
			var retErr error
			obj := stake.ToAPI(am)
			results[i] = &dataloader.Result[*model.Stake]{
				Data:  obj,
				Error: retErr,
			}
		} else {
			results[i] = &dataloader.Result[*model.Stake]{}
		}
	}

	return results
}

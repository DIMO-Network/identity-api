package loader

import (
	"context"
	"errors"
	"fmt"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/repositories/vehicle"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/graph-gophers/dataloader/v7"
)

type VehicleLoader struct {
	db       db.Store
	settings config.Settings
}

func GetVehicleByID(ctx context.Context, vehicleID int) (*model.Vehicle, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	// invoke and get thunk
	thunk := loaders.VehicleByID.Load(ctx, vehicleID)
	// read value from thunk
	return thunk()
}

// BatchGetVehicleByID implements the dataloader for finding vehicles by their ids.
func (v *VehicleLoader) BatchGetVehicleByID(ctx context.Context, vehicleIDs []int) []*dataloader.Result[*model.Vehicle] {
	results := make([]*dataloader.Result[*model.Vehicle], len(vehicleIDs))

	vehicles, err := models.Vehicles(models.VehicleWhere.ID.IN(vehicleIDs)).All(ctx, v.db.DBS().Reader)
	if err != nil {
		for i := range results {
			results[i] = &dataloader.Result[*model.Vehicle]{Error: err}
		}
		return results
	}

	vehicleByID := map[int]*models.Vehicle{}

	for _, v := range vehicles {
		vehicleByID[v.ID] = v
	}

	for i, k := range vehicleIDs {
		if veh, ok := vehicleByID[k]; ok {
			var retErr error
			imageURL, err := vehicle.DefaultImageURI(v.settings.BaseImageURL, veh.ID)
			if err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("error getting vehicle image url: %w", err))
			}
			dataURI, err := vehicle.GetVehicleDataURI(v.settings.BaseVehicleDataURI, veh.ID)
			if err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("error getting vehicle data uri: %w", err))
			}
			obj, err := vehicle.ToAPI(veh, imageURL, dataURI)
			if err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("error converting vehicle to API: %w", err))
			}
			results[i] = &dataloader.Result[*model.Vehicle]{
				Data:  obj,
				Error: retErr,
			}
		} else {
			results[i] = &dataloader.Result[*model.Vehicle]{Error: fmt.Errorf("no vehicle with id %d", k)}
		}
	}

	return results
}

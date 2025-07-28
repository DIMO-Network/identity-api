package loader

import (
	"context"
	"errors"
	"fmt"
	"github.com/DIMO-Network/identity-api/internal/repositories/devicedefinition"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/vehicle"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/graph-gophers/dataloader/v7"
)

type VehicleLoader struct {
	repo            *vehicle.Repository
	definitionsRepo *devicedefinition.Repository
}

func NewVehicleLoader(repo *vehicle.Repository, definitionsRepo *devicedefinition.Repository) *VehicleLoader {
	return &VehicleLoader{repo: repo, definitionsRepo: definitionsRepo}
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

	vehicles, err := models.Vehicles(models.VehicleWhere.ID.IN(vehicleIDs)).All(ctx, v.repo.PDB.DBS().Reader)
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

			var imageURI string

			if veh.ImageURI.Valid {
				imageURI = veh.ImageURI.String
			} else {
				var err error
				imageURI, err = vehicle.DefaultImageURI(v.repo.Settings.BaseImageURL, veh.ID)
				if err != nil {
					retErr = errors.Join(retErr, fmt.Errorf("error getting vehicle image url: %w", err))
				}
			}

			dataURI, err := vehicle.GetVehicleDataURI(v.repo.Settings.BaseVehicleDataURI, veh.ID)
			if err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("error getting vehicle data uri: %w", err))
			}
			definition, err := v.definitionsRepo.GetDeviceDefinition(ctx, model.DeviceDefinitionBy{ID: veh.DeviceDefinitionID.String})
			if err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("error getting device definition: %w", err))
			}

			obj, err := v.repo.ToAPI(veh, imageURI, dataURI, definition)
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

package loader

import (
	"context"
	"errors"
	"fmt"

	"github.com/DIMO-Network/identity-api/internal/repositories/devicedefinition"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
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

func GetVehicleByID(ctx context.Context, vehicleID int) (*gmodel.Vehicle, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	// invoke and get thunk
	thunk := loaders.VehicleByID.Load(ctx, vehicleID)
	// read value from thunk
	return thunk()
}

// BatchGetVehicleByID implements the dataloader for finding vehicles by their ids.
func (v *VehicleLoader) BatchGetVehicleByID(ctx context.Context, vehicleIDs []int) []*dataloader.Result[*gmodel.Vehicle] {
	results := make([]*dataloader.Result[*gmodel.Vehicle], len(vehicleIDs))

	vehicles, err := models.Vehicles(models.VehicleWhere.ID.IN(vehicleIDs)).All(ctx, v.repo.PDB.DBS().Reader)
	if err != nil {
		for i := range results {
			results[i] = &dataloader.Result[*gmodel.Vehicle]{Error: err}
		}
		return results
	}

	vehicleByID := map[int]*models.Vehicle{}

	for _, v := range vehicles {
		vehicleByID[v.ID] = v
	}

	// populate a map of device definitions by id for fast lookup later
	definitions := make(map[string]*gmodel.DeviceDefinition)
	for _, veh := range vehicles {
		definitions[veh.DeviceDefinitionID.String] = &gmodel.DeviceDefinition{}
	}
	ids := make([]string, 0, len(definitions))
	for id := range definitions {
		ids = append(ids, id)
	}
	dds, err := v.definitionsRepo.GetDeviceDefinitionsByIDs(ctx, ids)
	if err != nil {
		for i := range results {
			results[i] = &dataloader.Result[*gmodel.Vehicle]{Error: err}
		}
		return results
	}
	for _, dd := range dds {
		definitions[dd.DeviceDefinitionID] = dd
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
			definition, ok := definitions[veh.DeviceDefinitionID.String]
			if !ok {
				retErr = errors.Join(retErr, fmt.Errorf("error getting device definition: not found %s", veh.DeviceDefinitionID.String))
			}

			obj, err := v.repo.ToAPI(veh, imageURI, dataURI, definition)
			if err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("error converting vehicle to API: %w", err))
			}
			results[i] = &dataloader.Result[*gmodel.Vehicle]{
				Data:  obj,
				Error: retErr,
			}
		} else {
			results[i] = &dataloader.Result[*gmodel.Vehicle]{Error: fmt.Errorf("no vehicle with id %d", k)}
		}
	}

	return results
}

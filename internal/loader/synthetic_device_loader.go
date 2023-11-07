package loader

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/graph-gophers/dataloader/v7"
)

type SyntheticDeviceLoader struct {
	db db.Store
}

func GetSyntheticDeviceByVehicleID(ctx context.Context, vehicleID int) (*gmodel.SyntheticDevice, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	// invoke and get thunk
	thunk := loaders.SyntheticDeviceByVehicleID.Load(ctx, vehicleID)
	// read value from thunk
	return thunk()
}

func GetSyntheticDeviceByID(ctx context.Context, vehicleID int) (*gmodel.SyntheticDevice, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	// invoke and get thunk
	thunk := loaders.SyntheticDeviceByID.Load(ctx, vehicleID)
	// read value from thunk
	return thunk()
}

// BatchGetSyntheticDeviceByVehicleID implements the dataloader for finding synthetic devices linked to vehicles and returns
// them in the order requested
func (sd *SyntheticDeviceLoader) BatchGetSyntheticDeviceByVehicleID(ctx context.Context, vehicleIDs []int) []*dataloader.Result[*gmodel.SyntheticDevice] {
	results := make([]*dataloader.Result[*gmodel.SyntheticDevice], len(vehicleIDs))

	devices, err := models.SyntheticDevices(models.SyntheticDeviceWhere.VehicleID.IN(vehicleIDs)).All(ctx, sd.db.DBS().Reader)
	if err != nil {
		for i := range vehicleIDs {
			results[i] = &dataloader.Result[*gmodel.SyntheticDevice]{Data: nil, Error: err}
		}
		return results
	}

	sdByVehicleID := map[int]*models.SyntheticDevice{}

	for _, d := range devices {
		sdByVehicleID[d.VehicleID] = d
	}

	for idx, vid := range vehicleIDs {
		if sdv, ok := sdByVehicleID[vid]; ok {
			results[idx] = &dataloader.Result[*gmodel.SyntheticDevice]{
				Data: repositories.SyntheticDeviceToAPI(sdv),
			}
		} else {
			results[idx] = &dataloader.Result[*gmodel.SyntheticDevice]{}
		}

	}

	return results
}

// BatchGetSyntheticDeviceByVehicleID implements the dataloader for finding synthetic devices by their ids and returns
// them in the order requested
func (sd *SyntheticDeviceLoader) BatchGetSyntheticDeviceByID(ctx context.Context, syntheticDeviceIDs []int) []*dataloader.Result[*gmodel.SyntheticDevice] {
	results := make([]*dataloader.Result[*gmodel.SyntheticDevice], len(syntheticDeviceIDs))

	devices, err := models.SyntheticDevices(models.SyntheticDeviceWhere.ID.IN(syntheticDeviceIDs)).All(ctx, sd.db.DBS().Reader)
	if err != nil {
		for i := range results {
			results[i] = &dataloader.Result[*model.SyntheticDevice]{Error: err}
		}
		return results
	}

	for i, sdv := range devices {
		results[i] = &dataloader.Result[*gmodel.SyntheticDevice]{Data: repositories.SyntheticDeviceToAPI(sdv)}
	}

	return results
}

package loader

import (
	"context"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/synthetic"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/graph-gophers/dataloader/v7"
)

type SyntheticDeviceLoader struct {
	repo *synthetic.Repository
}

func NewSyntheticDeviceLoader(repo *synthetic.Repository) *SyntheticDeviceLoader {
	return &SyntheticDeviceLoader{repo: repo}
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
func (s *SyntheticDeviceLoader) BatchGetSyntheticDeviceByVehicleID(ctx context.Context, vehicleIDs []int) []*dataloader.Result[*gmodel.SyntheticDevice] {
	results := make([]*dataloader.Result[*gmodel.SyntheticDevice], len(vehicleIDs))

	devices, err := models.SyntheticDevices(models.SyntheticDeviceWhere.VehicleID.IN(vehicleIDs)).All(ctx, s.repo.PDB.DBS().Reader)
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
			synthAPI, err := s.repo.ToAPI(sdv)
			results[idx] = &dataloader.Result[*gmodel.SyntheticDevice]{Data: synthAPI, Error: err}
		} else {
			results[idx] = &dataloader.Result[*gmodel.SyntheticDevice]{}
		}
	}

	return results
}

// BatchGetSyntheticDeviceByID implements the dataloader for finding synthetic devices by their ids and returns
// them in the order requested
func (s *SyntheticDeviceLoader) BatchGetSyntheticDeviceByID(ctx context.Context, syntheticDeviceIDs []int) []*dataloader.Result[*gmodel.SyntheticDevice] {
	results := make([]*dataloader.Result[*gmodel.SyntheticDevice], len(syntheticDeviceIDs))

	sds, err := models.SyntheticDevices(models.SyntheticDeviceWhere.ID.IN(syntheticDeviceIDs)).All(ctx, s.repo.PDB.DBS().Reader)
	if err != nil {
		for i := range results {
			results[i] = &dataloader.Result[*gmodel.SyntheticDevice]{Error: err}
		}
		return results
	}

	sdByID := make(map[int]*models.SyntheticDevice)
	for _, sdv := range sds {
		sdByID[sdv.ID] = sdv
	}

	for i, sdID := range syntheticDeviceIDs {
		if sdv, ok := sdByID[sdID]; ok {
			synthAPI, err := s.repo.ToAPI(sdv)
			results[i] = &dataloader.Result[*gmodel.SyntheticDevice]{Data: synthAPI, Error: err}
		} else {
			results[i] = &dataloader.Result[*gmodel.SyntheticDevice]{}
		}
	}

	return results
}

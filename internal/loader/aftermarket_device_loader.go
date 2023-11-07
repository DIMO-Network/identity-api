package loader

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/graph-gophers/dataloader/v7"
)

type AftermarketDeviceLoader struct {
	db db.Store
}

func GetAftermarketDeviceByVehicleID(ctx context.Context, vehicleID int) (*model.AftermarketDevice, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	// invoke and get thunk
	thunk := loaders.AftermarketDeviceByVehicleID.Load(ctx, vehicleID)
	// read value from thunk
	return thunk()
}

func GetAftermarketDeviceByID(ctx context.Context, id int) (*model.AftermarketDevice, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	// invoke and get thunk
	thunk := loaders.AftermarketDeviceByID.Load(ctx, id)
	// read value from thunk
	return thunk()
}

// BatchGetLinkedAftermarketDeviceByVehicleID implements the dataloader for finding aftermarket devices linked to vehicles and returns
// them in the order requested
func (ad *AftermarketDeviceLoader) BatchGetLinkedAftermarketDeviceByVehicleID(ctx context.Context, vehicleIDs []int) []*dataloader.Result[*model.AftermarketDevice] {
	results := make([]*dataloader.Result[*model.AftermarketDevice], len(vehicleIDs))

	devices, err := models.AftermarketDevices(models.AftermarketDeviceWhere.VehicleID.IN(vehicleIDs)).All(ctx, ad.db.DBS().Reader)
	if err != nil {
		for i := range vehicleIDs {
			results[i] = &dataloader.Result[*model.AftermarketDevice]{Data: nil, Error: err}
		}
		return results
	}

	amByVehicleID := map[int]*models.AftermarketDevice{}

	for _, d := range devices {
		amByVehicleID[d.VehicleID.Int] = d
	}

	for i, vID := range vehicleIDs {
		if am, ok := amByVehicleID[vID]; ok {
			results[i] = &dataloader.Result[*model.AftermarketDevice]{
				Data: repositories.AftermarketDeviceToAPI(am),
			}
		} else {
			results[i] = &dataloader.Result[*model.AftermarketDevice]{}
		}
	}

	return results
}

// BatchGetLinkedAftermarketDeviceByID implements the dataloader for finding aftermarket devices by their ids and returns
// them in the order requested
func (ad *AftermarketDeviceLoader) BatchGetLinkedAftermarketDeviceByID(ctx context.Context, ids []int) []*dataloader.Result[*model.AftermarketDevice] {
	results := make([]*dataloader.Result[*model.AftermarketDevice], len(ids))

	devices, err := models.AftermarketDevices(models.AftermarketDeviceWhere.ID.IN(ids)).All(ctx, ad.db.DBS().Reader)
	if err != nil {
		for i := range ids {
			results[i] = &dataloader.Result[*model.AftermarketDevice]{Error: err}
		}
		return results
	}

	for i, adv := range devices {
		results[i] = &dataloader.Result[*model.AftermarketDevice]{Data: repositories.AftermarketDeviceToAPI(adv)}
	}

	return results
}

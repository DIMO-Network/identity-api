package loader

import (
	"context"
	"errors"
	"fmt"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/aftermarket"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/graph-gophers/dataloader/v7"
)

type AftermarketDeviceLoader struct {
	repo *aftermarket.Repository
}

func NewAftermarketDeviceLoader(repo *aftermarket.Repository) *AftermarketDeviceLoader {
	return &AftermarketDeviceLoader{repo: repo}
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
func (a *AftermarketDeviceLoader) BatchGetLinkedAftermarketDeviceByVehicleID(ctx context.Context, vehicleIDs []int) []*dataloader.Result[*model.AftermarketDevice] {
	results := make([]*dataloader.Result[*model.AftermarketDevice], len(vehicleIDs))

	devices, err := models.AftermarketDevices(models.AftermarketDeviceWhere.VehicleID.IN(vehicleIDs)).All(ctx, a.repo.PDB.DBS().Reader)
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
			var retErr error
			imageURL, err := aftermarket.GetAftermarketDeviceImageURL(a.repo.Settings.BaseImageURL, am.ID)
			if err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("failed getting image url: %w", err))
			}
			obj, err := a.repo.ToAPI(am, imageURL)
			if err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("failed converting to API: %w", err))
			}
			results[i] = &dataloader.Result[*model.AftermarketDevice]{
				Data:  obj,
				Error: retErr,
			}
		} else {
			results[i] = &dataloader.Result[*model.AftermarketDevice]{}
		}
	}

	return results
}

// BatchGetAftermarketDeviceByID implements the dataloader for finding aftermarket devices by their ids and returns
// them in the order requested
func (a *AftermarketDeviceLoader) BatchGetAftermarketDeviceByID(ctx context.Context, aftermarketDeviceIDs []int) []*dataloader.Result[*model.AftermarketDevice] {
	results := make([]*dataloader.Result[*model.AftermarketDevice], len(aftermarketDeviceIDs))

	ads, err := models.AftermarketDevices(models.AftermarketDeviceWhere.ID.IN(aftermarketDeviceIDs)).All(ctx, a.repo.PDB.DBS().Reader)
	if err != nil {
		for i := range aftermarketDeviceIDs {
			results[i] = &dataloader.Result[*model.AftermarketDevice]{Error: err}
		}
		return results
	}

	adByID := make(map[int]*models.AftermarketDevice)
	for _, ad := range ads {
		adByID[ad.ID] = ad
	}

	for i, adID := range aftermarketDeviceIDs {
		var retErr error
		if ads, ok := adByID[adID]; ok {
			imageURL, err := aftermarket.GetAftermarketDeviceImageURL(a.repo.Settings.BaseImageURL, ads.ID)
			if err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("failed getting image url: %w", err))
			}
			obj, err := a.repo.ToAPI(ads, imageURL)
			if err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("failed converting to API: %w", err))
			}
			results[i] = &dataloader.Result[*model.AftermarketDevice]{
				Data:  obj,
				Error: retErr,
			}
		} else {
			results[i] = &dataloader.Result[*model.AftermarketDevice]{}
		}
	}

	return results
}

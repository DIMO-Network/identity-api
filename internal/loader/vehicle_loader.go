package loader

import (
	"context"
	"strconv"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type VehicleLoader struct {
	db db.Store
}

func GetLinkedVehicleByID(ctx context.Context, vehicleID string) (*model.Vehicle, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	// invoke and get thunk
	thunk := loaders.VehicleByID.Load(ctx, vehicleID)
	// read value from thunk
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// BatchGetLinkedVehicleByAftermarketID implements the dataloader for finding vehicles linked to aftermarket devices and returns
// them in the order requested
func (v *VehicleLoader) BatchGetLinkedVehicleByAftermarketID(ctx context.Context, aftermarketDeviceIDs []string) []*dataloader.Result[*model.Vehicle] {
	keyOrder := make(map[int]int)
	results := make([]*dataloader.Result[*model.Vehicle], len(aftermarketDeviceIDs))
	var adIDs []int

	for ix, key := range aftermarketDeviceIDs {
		k, err := strconv.Atoi(key)
		if err != nil {
			results[ix] = &dataloader.Result[*model.Vehicle]{Data: nil, Error: err}
		}
		keyOrder[k] = ix
		adIDs = append(adIDs, k)
	}

	adVehicleLink, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.ID.IN(adIDs),
		qm.Load(models.AftermarketDeviceRels.Vehicle),
	).All(ctx, v.db.DBS().Reader)
	if err != nil {
		for ix := range adIDs {
			results[ix] = &dataloader.Result[*model.Vehicle]{Data: nil, Error: err}
		}
		return results
	}

	for _, device := range adVehicleLink {
		if device.R.Vehicle == nil {
			results[keyOrder[device.ID]] = &dataloader.Result[*model.Vehicle]{Data: nil, Error: nil}
			continue
		}

		v := &model.Vehicle{
			ID:       strconv.Itoa(device.R.Vehicle.ID),
			Owner:    *repositories.BytesToAddr(device.R.Vehicle.OwnerAddress),
			Make:     device.R.Vehicle.Make.Ptr(),
			Model:    device.R.Vehicle.Model.Ptr(),
			Year:     device.R.Vehicle.Year.Ptr(),
			MintedAt: device.R.Vehicle.MintedAt.Time,
		}
		results[keyOrder[device.ID]] = &dataloader.Result[*model.Vehicle]{Data: v, Error: nil}
		delete(keyOrder, device.ID)

	}

	return results
}

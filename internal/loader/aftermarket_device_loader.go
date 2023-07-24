package loader

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/graph-gophers/dataloader/v7"
)

type AftermarketDeviceLoader struct {
	db db.Store
}

func GetAftermarketDeviceByVehicleID(ctx context.Context, vehicleID string) (*model.AftermarketDevice, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	// invoke and get thunk
	thunk := loaders.AftermarketDeviceByVehicleID.Load(ctx, vehicleID)
	// read value from thunk
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// BatchGetLinkedAftermarketDeviceByVehicleID implements the dataloader for finding aftermarket devices linked to vehicles and returns
// them in the order requested
func (ad *AftermarketDeviceLoader) BatchGetLinkedAftermarketDeviceByVehicleID(ctx context.Context, vehicleIDs []string) []*dataloader.Result[*model.AftermarketDevice] {
	keyOrder := make(map[int]int)
	results := make([]*dataloader.Result[*model.AftermarketDevice], len(vehicleIDs))
	var vIDs []int

	for ix, key := range vehicleIDs {
		k, err := strconv.Atoi(key)
		if err != nil {
			results[ix] = &dataloader.Result[*model.AftermarketDevice]{Data: nil, Error: err}
		}
		keyOrder[k] = ix
		vIDs = append(vIDs, k)
	}

	devices, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.VehicleID.IN(vIDs),
	).All(ctx, ad.db.DBS().Reader)
	if err != nil {
		for ix := range vIDs {
			results[ix] = &dataloader.Result[*model.AftermarketDevice]{Data: nil, Error: err}
		}
		return results
	}

	for _, device := range devices {
		v := &model.AftermarketDevice{
			ID:       strconv.Itoa(device.ID),
			Address:  repositories.BytesToAddr(device.Address),
			Owner:    repositories.BytesToAddr(device.Owner),
			Serial:   device.Serial.Ptr(),
			Imei:     device.Imei.Ptr(),
			MintedAt: device.MintedAt.Ptr(),
		}
		results[keyOrder[device.VehicleID.Int]] = &dataloader.Result[*model.AftermarketDevice]{Data: v, Error: nil}
		delete(keyOrder, device.VehicleID.Int)
	}

	for k, v := range keyOrder {
		results[v] = &dataloader.Result[*model.AftermarketDevice]{Data: nil, Error: fmt.Errorf("aftermarket device associated with vehicle id %d not found", k)}
	}

	return results
}

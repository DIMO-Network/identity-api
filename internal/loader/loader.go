package loader

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type loaderKey struct{}
type loadersString string

const (
	dataLoadersKey loadersString = "dataLoadersKey"
)

type Loaders struct {
	VehicleByID dataloader.Interface[string, *model.Vehicle]
}

type VehicleLoader struct {
	db db.Store
}

// NewDataLoader returns the instantiated Loaders struct for use in a request
func NewDataLoader(dbs db.Store) *Loaders {
	// instantiate the user dataloader
	vehicle := &VehicleLoader{db: dbs}
	// return the DataLoader
	return &Loaders{
		VehicleByID: dataloader.NewBatchedLoader(
			vehicle.BatchGetLinkedVehicleByAftermarketID,
			dataloader.WithClearCacheOnBatch[string, *model.Vehicle](),
		),
	}
}

// Middleware injects a DataLoader into the request context so it can be
// used later in the schema resolvers
func Middleware(db db.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loader := NewDataLoader(db)
		nextCtx := context.WithValue(r.Context(), dataLoadersKey, loader)
		r = r.WithContext(nextCtx)
		next.ServeHTTP(w, r)
	})
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

	vehicles, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.ID.IN(adIDs),
		qm.Load(models.AftermarketDeviceRels.Vehicle),
	).All(ctx, v.db.DBS().Reader)
	if err != nil {
		for ix := range adIDs {
			results[ix] = &dataloader.Result[*model.Vehicle]{Data: nil, Error: err}
		}
		return results
	}

	for _, vehicle := range vehicles {
		if vehicle.R.Vehicle != nil {
			v := &model.Vehicle{
				ID:       strconv.Itoa(vehicle.R.Vehicle.ID),
				Owner:    repositories.BytesToAddr(vehicle.R.Vehicle.OwnerAddress),
				Make:     vehicle.R.Vehicle.Make.Ptr(),
				Model:    vehicle.R.Vehicle.Model.Ptr(),
				Year:     vehicle.R.Vehicle.Year.Ptr(),
				MintedAt: vehicle.R.Vehicle.MintedAt.Ptr(),
			}
			results[keyOrder[vehicle.ID]] = &dataloader.Result[*model.Vehicle]{Data: v, Error: nil}
			delete(keyOrder, vehicle.ID)
		}
	}

	for k, v := range keyOrder {
		results[v] = &dataloader.Result[*model.Vehicle]{Data: nil, Error: fmt.Errorf("vehicle associated with aftermarket id %d not found", k)}
	}

	return results
}

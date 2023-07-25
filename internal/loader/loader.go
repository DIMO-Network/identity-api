package loader

import (
	"context"
	"net/http"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/shared/db"
	"github.com/graph-gophers/dataloader/v7"
)

type loaderKey struct{}
type loadersString string

const (
	dataLoadersKey loadersString = "dataLoadersKey"
)

type Loaders struct {
	VehicleByID                  dataloader.Interface[string, *model.Vehicle]
	AftermarketDeviceByVehicleID dataloader.Interface[string, *model.AftermarketDevice]
}

// NewDataLoader returns the instantiated Loaders struct for use in a request
func NewDataLoader(dbs db.Store) *Loaders {
	// instantiate the user dataloader
	vehicle := &VehicleLoader{db: dbs}
	aftermarketDevice := &AftermarketDeviceLoader{db: dbs}
	// return the DataLoader
	return &Loaders{
		VehicleByID: dataloader.NewBatchedLoader(
			vehicle.BatchGetLinkedVehicleByAftermarketID,
			dataloader.WithClearCacheOnBatch[string, *model.Vehicle](),
		),
		AftermarketDeviceByVehicleID: dataloader.NewBatchedLoader(
			aftermarketDevice.BatchGetLinkedAftermarketDeviceByVehicleID,
			dataloader.WithClearCacheOnBatch[string, *model.AftermarketDevice](),
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

package loader

import (
	"context"
	"net/http"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/repositories/aftermarket"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/repositories/connection"
	"github.com/DIMO-Network/identity-api/internal/repositories/dcn"
	"github.com/DIMO-Network/identity-api/internal/repositories/devicedefinition"
	"github.com/DIMO-Network/identity-api/internal/repositories/manufacturer"
	"github.com/DIMO-Network/identity-api/internal/repositories/stake"
	"github.com/DIMO-Network/identity-api/internal/repositories/storagenode"
	"github.com/DIMO-Network/identity-api/internal/repositories/synthetic"
	"github.com/DIMO-Network/identity-api/internal/repositories/vehicle"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/rs/zerolog"
)

type loadersString string

const (
	dataLoadersKey loadersString = "dataLoadersKey"
)

type Loaders struct {
	VehicleByID                  dataloader.Interface[int, *model.Vehicle]
	AftermarketDeviceByVehicleID dataloader.Interface[int, *model.AftermarketDevice]
	SyntheticDeviceByVehicleID   dataloader.Interface[int, *model.SyntheticDevice]
	DCNByVehicleID               dataloader.Interface[int, *model.Dcn]
	ManufacturerByID             dataloader.Interface[int, *model.Manufacturer]
	AftermarketDeviceByID        dataloader.Interface[int, *model.AftermarketDevice]
	SyntheticDeviceByID          dataloader.Interface[int, *model.SyntheticDevice]
	StakeByVehicleID             dataloader.Interface[int, *model.Stake]
	ConnectionByID               dataloader.Interface[ConnectionQueryKey, *model.Connection]
	StorageNodeByID              dataloader.Interface[[32]byte, *model.StorageNode]
}

// NewDataLoader returns the instantiated Loaders struct for use in a request
func NewDataLoader(dbs db.Store, settings config.Settings, lg *zerolog.Logger) *Loaders {
	// instantiate the user dataloader
	baseRepo := &base.Repository{PDB: dbs, Settings: settings}

	// Create device definition repository first
	tablelandAPI := services.NewTablelandApiService(lg, &settings)
	deviceDefRepo := devicedefinition.New(baseRepo, tablelandAPI)

	vehicle := NewVehicleLoader(vehicle.New(baseRepo, deviceDefRepo), deviceDefRepo)
	aftermarketDevice := NewAftermarketDeviceLoader(aftermarket.New(baseRepo))
	syntheticDevice := NewSyntheticDeviceLoader(synthetic.New(baseRepo))
	dcn := NewDCNLoader(dcn.New(baseRepo))
	manufacturer := NewManufacturerLoader(manufacturer.New(baseRepo))
	stake := NewStakeLoader(stake.New(baseRepo))
	connection := ConnectionLoader{repo: connection.New(baseRepo)}
	storageNode := NewStorageNodeLoader(storagenode.New(baseRepo))

	// return the DataLoader
	return &Loaders{
		VehicleByID: dataloader.NewBatchedLoader(
			vehicle.BatchGetVehicleByID,
			dataloader.WithClearCacheOnBatch[int, *model.Vehicle](),
		),
		AftermarketDeviceByVehicleID: dataloader.NewBatchedLoader(
			aftermarketDevice.BatchGetLinkedAftermarketDeviceByVehicleID,
			dataloader.WithClearCacheOnBatch[int, *model.AftermarketDevice](),
		),
		SyntheticDeviceByVehicleID: dataloader.NewBatchedLoader(
			syntheticDevice.BatchGetSyntheticDeviceByVehicleID,
			dataloader.WithClearCacheOnBatch[int, *model.SyntheticDevice](),
		),
		DCNByVehicleID: dataloader.NewBatchedLoader(
			dcn.BatchGetDCNByVehicleID,
			dataloader.WithClearCacheOnBatch[int, *model.Dcn](),
		),
		ManufacturerByID: dataloader.NewBatchedLoader(
			manufacturer.BatchGetManufacturerByID,
			dataloader.WithClearCacheOnBatch[int, *model.Manufacturer](),
		),
		AftermarketDeviceByID: dataloader.NewBatchedLoader(
			aftermarketDevice.BatchGetAftermarketDeviceByID,
			dataloader.WithClearCacheOnBatch[int, *model.AftermarketDevice](),
		),
		SyntheticDeviceByID: dataloader.NewBatchedLoader(
			syntheticDevice.BatchGetSyntheticDeviceByID,
			dataloader.WithClearCacheOnBatch[int, *model.SyntheticDevice](),
		),
		StakeByVehicleID: dataloader.NewBatchedLoader(
			stake.BatchGetLinkedStakesByVehicleID,
			dataloader.WithClearCacheOnBatch[int, *model.Stake](),
		),
		ConnectionByID: dataloader.NewBatchedLoader(
			connection.BatchGetConnectionsByIDs,
			dataloader.WithClearCacheOnBatch[ConnectionQueryKey, *model.Connection](),
		),
		StorageNodeByID: dataloader.NewBatchedLoader(
			storageNode.BatchGetStorageNodesByIDs,
			dataloader.WithClearCacheOnBatch[[32]byte, *model.StorageNode](),
		),
	}
}

// Middleware injects a DataLoader into the request context so it can be
// used later in the schema resolvers
func Middleware(db db.Store, next http.Handler, settings config.Settings, lg *zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loader := NewDataLoader(db, settings, lg)
		nextCtx := context.WithValue(r.Context(), dataLoadersKey, loader)
		r = r.WithContext(nextCtx)
		next.ServeHTTP(w, r)
	})
}

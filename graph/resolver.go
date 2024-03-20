package graph

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/aftermarket"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/repositories/dcn"
	"github.com/DIMO-Network/identity-api/internal/repositories/manufacturer"
	"github.com/DIMO-Network/identity-api/internal/repositories/reward"
	"github.com/DIMO-Network/identity-api/internal/repositories/synthetic"
	"github.com/DIMO-Network/identity-api/internal/repositories/vehicle"
	"github.com/DIMO-Network/identity-api/internal/repositories/vehicleprivilege"
)

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// AftermarketDeviceRepository interface for mocking aftermarket.Repository.
//
//go:generate mockgen -destination=./mock_aftermarket_test.go -package=graph github.com/DIMO-Network/identity-api/graph AftermarketDeviceRepository
type AftermarketDeviceRepository interface {
	GetAftermarketDevices(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *model.AftermarketDevicesFilter) (*model.AftermarketDeviceConnection, error)
	GetAftermarketDevice(ctx context.Context, by model.AftermarketDeviceBy) (*model.AftermarketDevice, error)
	GetAftermarketDevicesForManufacturer(ctx context.Context, obj *model.Manufacturer, first *int, after *string, last *int, before *string, filterBy *model.AftermarketDevicesFilter) (*model.AftermarketDeviceConnection, error)
}

// DCNRepository interface for mocking dcn.Repository.
//
//go:generate mockgen -destination=./mock_dcn_test.go -package=graph github.com/DIMO-Network/identity-api/graph DCNRepository
type DCNRepository interface {
	GetDCN(ctx context.Context, by model.DCNBy) (*model.Dcn, error)
	GetDCNByNode(ctx context.Context, node []byte) (*model.Dcn, error)
	GetDCNByName(ctx context.Context, name string) (*model.Dcn, error)
	GetDCNs(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *model.DCNFilter) (*model.DCNConnection, error)
}

// ManufacturerRepository interface for mocking manufacturer.Repository.
//
//go:generate mockgen -destination=./mock_manufacturer_test.go -package=graph github.com/DIMO-Network/identity-api/graph ManufacturerRepository
type ManufacturerRepository interface {
	GetManufacturer(ctx context.Context, by model.ManufacturerBy) (*model.Manufacturer, error)
}

// SyntheticRepository interface for mocking synthetic.Repository.
//
//go:generate mockgen -destination=./mock_synthetic_test.go -package=graph github.com/DIMO-Network/identity-api/graph SyntheticRepository
type SyntheticRepository interface {
	GetSyntheticDevice(ctx context.Context, by model.SyntheticDeviceBy) (*model.SyntheticDevice, error)
	GetSyntheticDevices(ctx context.Context, first *int, last *int, after *string, before *string, filterBy *model.SyntheticDevicesFilter) (*model.SyntheticDeviceConnection, error)
}

// VehicleRepository interface for mocking vehicle.Repository.
//
//go:generate mockgen -destination=./mock_vehicle_test.go -package=graph github.com/DIMO-Network/identity-api/graph VehicleRepository
type VehicleRepository interface {
	GetVehicles(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *model.VehiclesFilter) (*model.VehicleConnection, error)
	GetVehicle(ctx context.Context, id int) (*model.Vehicle, error)
}

// Resolver holds the repositories for the graph resolvers.
type Resolver struct {
	aftermarket      AftermarketDeviceRepository
	dcn              DCNRepository
	manufacturer     ManufacturerRepository
	reward           reward.Repository
	synthetic        SyntheticRepository
	vehicle          VehicleRepository
	vehicleprivilege vehicleprivilege.Repository
}

// NewResolver creates a new Resolver with allocated repositories.
func NewResolver(baseRepo *base.Repository) *Resolver {
	return &Resolver{
		aftermarket:      &aftermarket.Repository{Repository: baseRepo},
		dcn:              &dcn.Repository{Repository: baseRepo},
		manufacturer:     &manufacturer.Repository{Repository: baseRepo},
		reward:           reward.Repository{Repository: baseRepo},
		synthetic:        &synthetic.Repository{Repository: baseRepo},
		vehicle:          &vehicle.Repository{Repository: baseRepo},
		vehicleprivilege: vehicleprivilege.Repository{Repository: baseRepo},
	}
}

package graph

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/aftermarket"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/repositories/dcn"
	"github.com/DIMO-Network/identity-api/internal/repositories/manufacturer"
	"github.com/DIMO-Network/identity-api/internal/repositories/reward"
	"github.com/DIMO-Network/identity-api/internal/repositories/vehicle"
	"github.com/DIMO-Network/identity-api/internal/repositories/vehicleprivilege"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver holds the repositories for the graph resolvers.
type Resolver struct {
	aftermarket      AftermarketDeviceRepository
	dcn              DCNRepository
	manufacturer     ManufacturerRepository
	reward           RewardRepository
	vehicle          VehicleRepository
	vehicleprivilege VehiclePrivilegeRepository
}

// NewResolver creates a new Resolver with allocated repositories.
func NewResolver(repo *base.Repository) *Resolver {
	return &Resolver{
		aftermarket:      &aftermarket.Repository{Repository: repo},
		dcn:              &dcn.Repository{Repository: repo},
		manufacturer:     &manufacturer.Repository{Repository: repo},
		reward:           &reward.Repository{Repository: repo},
		vehicle:          &vehicle.Repository{Repository: repo},
		vehicleprivilege: &vehicleprivilege.Repository{Repository: repo},
	}
}

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

// VehicleRepository interface for mocking vehicle.Repository.
//
//go:generate mockgen -destination=./mock_vehicle_test.go -package=graph github.com/DIMO-Network/identity-api/graph VehicleRepository
type VehicleRepository interface {
	GetVehicles(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *model.VehiclesFilter) (*model.VehicleConnection, error)
	GetVehicle(ctx context.Context, id int) (*model.Vehicle, error)
}

// RewardRepositoryInterface interface for mocking reward.Repository.
type RewardRepository interface {
	GetEarningsSummary(ctx context.Context, conditions []qm.QueryMod) (*reward.EarningsSummary, error)
	GetEarningsByVehicleID(ctx context.Context, tokenID int) (*model.VehicleEarnings, error)
	PaginateVehicleEarningsByID(ctx context.Context, vehicleEarnings *model.VehicleEarnings, first *int, after *string, last *int, before *string) (*model.EarningsConnection, error)
	GetEarningsByAfterMarketDeviceID(ctx context.Context, tokenID int) (*model.AftermarketDeviceEarnings, error)
	PaginateAftermarketDeviceEarningsByID(ctx context.Context, afterMarketDeviceEarnings *model.AftermarketDeviceEarnings, first *int, after *string, last *int, before *string) (*model.EarningsConnection, error)
	GetEarningsByUserAddress(ctx context.Context, user common.Address) (*model.UserRewards, error)
	PaginateGetEarningsByUsersDevices(ctx context.Context, userDeviceEarnings *model.UserRewards, first *int, after *string, last *int, before *string) (*model.EarningsConnection, error)
}

// VehiclePrivilegeRepository interface for mocking vehicleprivilege.Repository.
type VehiclePrivilegeRepository interface {
	GetPrivilegesForVehicle(ctx context.Context, tokenID int, first *int, after *string, last *int, before *string, filterBy *model.PrivilegeFilterBy) (*model.PrivilegesConnection, error)
}

package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.34

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/ethereum/go-ethereum/common"
)

// Vehicle is the resolver for the vehicle field.
func (r *aftermarketDeviceResolver) Vehicle(ctx context.Context, obj *model.AftermarketDevice) (*model.Vehicle, error) {
	if obj.VehicleID == nil {
		return nil, nil
	}
	return loader.GetVehicleByID(ctx, *obj.VehicleID)
}

// AccessibleVehicles is the resolver for the accessibleVehicles field.
func (r *queryResolver) AccessibleVehicles(ctx context.Context, address common.Address, first *int, after *string) (*model.VehicleConnection, error) {
	return r.Repo.GetAccessibleVehicles(ctx, address, first, after)
}

// OwnedAftermarketDevices is the resolver for the ownedAftermarketDevices field.
func (r *queryResolver) OwnedAftermarketDevices(ctx context.Context, address common.Address, first *int, after *string) (*model.AftermarketDeviceConnection, error) {
	return r.Repo.GetOwnedAftermarketDevices(ctx, address, first, after)
}

// Vehicle is the resolver for the vehicle field.
func (r *queryResolver) Vehicle(ctx context.Context, id int) (*model.Vehicle, error) {
	return r.Repo.GetVehicle(ctx, id)
}

// AftermarketDevice is the resolver for the aftermarketDevice field.
func (r *vehicleResolver) AftermarketDevice(ctx context.Context, obj *model.Vehicle) (*model.AftermarketDevice, error) {
	return loader.GetAftermarketDeviceByVehicleID(ctx, obj.ID)
}

// Privileges is the resolver for the privileges field.
func (r *vehicleResolver) Privileges(ctx context.Context, obj *model.Vehicle, first *int, after *string) (*model.PrivilegesConnection, error) {
	return r.Repo.GetPrivilegesForVehicle(ctx, obj.ID, first, after)
}

// AftermarketDevice returns AftermarketDeviceResolver implementation.
func (r *Resolver) AftermarketDevice() AftermarketDeviceResolver {
	return &aftermarketDeviceResolver{r}
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Vehicle returns VehicleResolver implementation.
func (r *Resolver) Vehicle() VehicleResolver { return &vehicleResolver{r} }

type aftermarketDeviceResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type vehicleResolver struct{ *Resolver }

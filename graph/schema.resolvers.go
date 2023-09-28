package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.38

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/loader"
)

// Vehicle is the resolver for the vehicle field.
func (r *aftermarketDeviceResolver) Vehicle(ctx context.Context, obj *model.AftermarketDevice) (*model.Vehicle, error) {
	if obj.VehicleID == nil {
		return nil, nil
	}
	return loader.GetVehicleByID(ctx, *obj.VehicleID)
}

// Vehicle is the resolver for the vehicle field.
func (r *dCNResolver) Vehicle(ctx context.Context, obj *model.Dcn) (*model.Vehicle, error) {
	if obj.VehicleID == nil {
		return nil, nil
	}
	return loader.GetVehicleByID(ctx, *obj.VehicleID)
}

// Vehicles is the resolver for the vehicles field.
func (r *queryResolver) Vehicles(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *model.VehiclesFilter) (*model.VehicleConnection, error) {
	return r.Repo.GetVehicles(ctx, first, after, last, before, filterBy)
}

// AftermarketDevices is the resolver for the aftermarketDevices field.
func (r *queryResolver) AftermarketDevices(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *model.AftermarketDevicesFilter) (*model.AftermarketDeviceConnection, error) {
	return r.Repo.GetAftermarketDevices(ctx, first, after, last, before, filterBy)
}

// Vehicle is the resolver for the vehicle field.
func (r *queryResolver) Vehicle(ctx context.Context, id int) (*model.Vehicle, error) {
	return r.Repo.GetVehicle(ctx, id)
}

// AftermarketDevice is the resolver for the aftermarketDevice field.
func (r *queryResolver) AftermarketDevice(ctx context.Context, by model.AftermarketDeviceBy) (*model.AftermarketDevice, error) {
	return r.Repo.GetAftermarketDevice(ctx, by)
}

// Dcn is the resolver for the dcn field.
func (r *queryResolver) Dcn(ctx context.Context, by model.DCNBy) (*model.Dcn, error) {
	return r.Repo.GetDCN(ctx, by)
}

// AftermarketDevice is the resolver for the aftermarketDevice field.
func (r *vehicleResolver) AftermarketDevice(ctx context.Context, obj *model.Vehicle) (*model.AftermarketDevice, error) {
	return loader.GetAftermarketDeviceByVehicleID(ctx, obj.ID)
}

// Privileges is the resolver for the privileges field.
func (r *vehicleResolver) Privileges(ctx context.Context, obj *model.Vehicle, filterBy model.PrivilegeFilterBy) (*model.PrivilegesConnection, error) {
	return r.Repo.GetPrivilegesForVehicle(ctx, obj.ID, filterBy)
}

// SyntheticDevice is the resolver for the syntheticDevice field.
func (r *vehicleResolver) SyntheticDevice(ctx context.Context, obj *model.Vehicle) (*model.SyntheticDevice, error) {
	return loader.GetSyntheticDeviceByVehicleID(ctx, obj.ID)
}

// Dcn is the resolver for the dcn field.
func (r *vehicleResolver) Dcn(ctx context.Context, obj *model.Vehicle) (*model.Dcn, error) {
	return loader.GetDCNByVehicleID(ctx, obj.ID)
}

// AftermarketDevice returns AftermarketDeviceResolver implementation.
func (r *Resolver) AftermarketDevice() AftermarketDeviceResolver {
	return &aftermarketDeviceResolver{r}
}

// DCN returns DCNResolver implementation.
func (r *Resolver) DCN() DCNResolver { return &dCNResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Vehicle returns VehicleResolver implementation.
func (r *Resolver) Vehicle() VehicleResolver { return &vehicleResolver{r} }

type aftermarketDeviceResolver struct{ *Resolver }
type dCNResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type vehicleResolver struct{ *Resolver }

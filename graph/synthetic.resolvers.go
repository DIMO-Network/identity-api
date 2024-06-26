package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.49

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/loader"
)

// SyntheticDevice is the resolver for the syntheticDevice field.
func (r *queryResolver) SyntheticDevice(ctx context.Context, by model.SyntheticDeviceBy) (*model.SyntheticDevice, error) {
	return r.synthetic.GetSyntheticDevice(ctx, by)
}

// SyntheticDevices is the resolver for the syntheticDevices field.
func (r *queryResolver) SyntheticDevices(ctx context.Context, first *int, last *int, after *string, before *string, filterBy *model.SyntheticDevicesFilter) (*model.SyntheticDeviceConnection, error) {
	return r.synthetic.GetSyntheticDevices(ctx, first, last, after, before, filterBy)
}

// Vehicle is the resolver for the vehicle field.
func (r *syntheticDeviceResolver) Vehicle(ctx context.Context, obj *model.SyntheticDevice) (*model.Vehicle, error) {
	return loader.GetVehicleByID(ctx, obj.VehicleID)
}

// SyntheticDevice returns SyntheticDeviceResolver implementation.
func (r *Resolver) SyntheticDevice() SyntheticDeviceResolver { return &syntheticDeviceResolver{r} }

type syntheticDeviceResolver struct{ *Resolver }

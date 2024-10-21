package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.55

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/loader"
)

// Manufacturer is the resolver for the manufacturer field.
func (r *aftermarketDeviceResolver) Manufacturer(ctx context.Context, obj *model.AftermarketDevice) (*model.Manufacturer, error) {
	return loader.GetManufacturerID(ctx, obj.ManufacturerID)
}

// Vehicle is the resolver for the vehicle field.
func (r *aftermarketDeviceResolver) Vehicle(ctx context.Context, obj *model.AftermarketDevice) (*model.Vehicle, error) {
	if obj.VehicleID == nil {
		return nil, nil
	}
	return loader.GetVehicleByID(ctx, *obj.VehicleID)
}

// Earnings is the resolver for the earnings field.
func (r *aftermarketDeviceResolver) Earnings(ctx context.Context, obj *model.AftermarketDevice) (*model.AftermarketDeviceEarnings, error) {
	return r.reward.GetEarningsByAfterMarketDeviceID(ctx, obj.TokenID)
}

// History is the resolver for the history field.
func (r *aftermarketDeviceEarningsResolver) History(ctx context.Context, obj *model.AftermarketDeviceEarnings, first *int, after *string, last *int, before *string) (*model.EarningsConnection, error) {
	return r.reward.PaginateAftermarketDeviceEarningsByID(ctx, obj, first, after, last, before)
}

// AftermarketDevice is the resolver for the aftermarketDevice field.
func (r *queryResolver) AftermarketDevice(ctx context.Context, by model.AftermarketDeviceBy) (*model.AftermarketDevice, error) {
	return r.aftermarket.GetAftermarketDevice(ctx, by)
}

// AftermarketDevices is the resolver for the aftermarketDevices field.
func (r *queryResolver) AftermarketDevices(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *model.AftermarketDevicesFilter) (*model.AftermarketDeviceConnection, error) {
	return r.aftermarket.GetAftermarketDevices(ctx, first, after, last, before, filterBy)
}

// AftermarketDevice returns AftermarketDeviceResolver implementation.
func (r *Resolver) AftermarketDevice() AftermarketDeviceResolver {
	return &aftermarketDeviceResolver{r}
}

// AftermarketDeviceEarnings returns AftermarketDeviceEarningsResolver implementation.
func (r *Resolver) AftermarketDeviceEarnings() AftermarketDeviceEarningsResolver {
	return &aftermarketDeviceEarningsResolver{r}
}

type aftermarketDeviceResolver struct{ *Resolver }
type aftermarketDeviceEarningsResolver struct{ *Resolver }

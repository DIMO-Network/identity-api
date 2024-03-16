package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.39

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
)

// AftermarketDevices is the resolver for the aftermarketDevices field on the manufacturer object.
func (r *manufacturerResolver) AftermarketDevices(ctx context.Context, obj *model.Manufacturer, first *int, after *string, last *int, before *string, filterBy *model.AftermarketDevicesFilter) (*model.AftermarketDeviceConnection, error) {
	return r.aftermarket.GetAftermarketDevicesForManufacturer(ctx, obj, first, after, last, before, filterBy)
}

// Manufacturer is the resolver for the manufacturer field.
func (r *queryResolver) Manufacturer(ctx context.Context, by model.ManufacturerBy) (*model.Manufacturer, error) {
	return r.manufacturer.GetManufacturer(ctx, by)
}

// Manufacturer returns ManufacturerResolver implementation.
func (r *Resolver) Manufacturer() ManufacturerResolver { return &manufacturerResolver{r} }

type manufacturerResolver struct{ *Resolver }

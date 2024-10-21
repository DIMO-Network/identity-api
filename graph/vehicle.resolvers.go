package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.55

import (
	"context"
	"errors"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Vehicle is the resolver for the vehicle field.
func (r *queryResolver) Vehicle(ctx context.Context, tokenID int) (*model.Vehicle, error) {
	v, err := r.vehicle.GetVehicle(ctx, tokenID)
	if errors.Is(err, repositories.ErrNotFound) {
		return nil, graphql.ErrorOnPath(ctx, &gqlerror.Error{
			Message: fmt.Sprintf("No vehicle with token id %d.", tokenID),
			Extensions: map[string]interface{}{
				"code": "NOT_FOUND",
			},
		})
	}
	return v, err
}

// Vehicles is the resolver for the vehicles field.
func (r *queryResolver) Vehicles(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *model.VehiclesFilter) (*model.VehicleConnection, error) {
	return r.vehicle.GetVehicles(ctx, first, after, last, before, filterBy)
}

// Manufacturer is the resolver for the manufacturer field.
func (r *vehicleResolver) Manufacturer(ctx context.Context, obj *model.Vehicle) (*model.Manufacturer, error) {
	return loader.GetManufacturerID(ctx, obj.ManufacturerID)
}

// AftermarketDevice is the resolver for the aftermarketDevice field.
func (r *vehicleResolver) AftermarketDevice(ctx context.Context, obj *model.Vehicle) (*model.AftermarketDevice, error) {
	return loader.GetAftermarketDeviceByVehicleID(ctx, obj.TokenID)
}

// Privileges is the resolver for the privileges field.
func (r *vehicleResolver) Privileges(ctx context.Context, obj *model.Vehicle, first *int, after *string, last *int, before *string, filterBy *model.PrivilegeFilterBy) (*model.PrivilegesConnection, error) {
	return r.vehicleprivilege.GetPrivilegesForVehicle(ctx, obj.TokenID, first, after, last, before, filterBy)
}

// Sacds is the resolver for the sacds field.
func (r *vehicleResolver) Sacds(ctx context.Context, obj *model.Vehicle, first *int, after *string, last *int, before *string) (*model.SacdConnection, error) {
	return r.vehiclesacd.GetSacdsForVehicle(ctx, obj.TokenID, first, after, last, before)
}

// SyntheticDevice is the resolver for the syntheticDevice field.
func (r *vehicleResolver) SyntheticDevice(ctx context.Context, obj *model.Vehicle) (*model.SyntheticDevice, error) {
	return loader.GetSyntheticDeviceByVehicleID(ctx, obj.TokenID)
}

// Dcn is the resolver for the dcn field.
func (r *vehicleResolver) Dcn(ctx context.Context, obj *model.Vehicle) (*model.Dcn, error) {
	return loader.GetDCNByVehicleID(ctx, obj.TokenID)
}

// Earnings is the resolver for the earnings field.
func (r *vehicleResolver) Earnings(ctx context.Context, obj *model.Vehicle) (*model.VehicleEarnings, error) {
	return r.reward.GetEarningsByVehicleID(ctx, obj.TokenID)
}

// History is the resolver for the history field.
func (r *vehicleEarningsResolver) History(ctx context.Context, obj *model.VehicleEarnings, first *int, after *string, last *int, before *string) (*model.EarningsConnection, error) {
	return r.reward.PaginateVehicleEarningsByID(ctx, obj, first, after, last, before)
}

// Vehicle returns VehicleResolver implementation.
func (r *Resolver) Vehicle() VehicleResolver { return &vehicleResolver{r} }

// VehicleEarnings returns VehicleEarningsResolver implementation.
func (r *Resolver) VehicleEarnings() VehicleEarningsResolver { return &vehicleEarningsResolver{r} }

type vehicleResolver struct{ *Resolver }
type vehicleEarningsResolver struct{ *Resolver }

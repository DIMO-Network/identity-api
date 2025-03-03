package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.66

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/loader"
)

// Stakes is the resolver for the stakes field.
func (r *queryResolver) Stakes(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *model.StakeFilterBy) (*model.StakeConnection, error) {
	return r.stake.GetStakes(ctx, first, after, last, before, filterBy)
}

// Vehicle is the resolver for the vehicle field.
func (r *stakeResolver) Vehicle(ctx context.Context, obj *model.Stake) (*model.Vehicle, error) {
	if obj.VehicleID == nil {
		return nil, nil
	}
	return loader.GetVehicleByID(ctx, *obj.VehicleID)
}

// Stake returns StakeResolver implementation.
func (r *Resolver) Stake() StakeResolver { return &stakeResolver{r} }

type stakeResolver struct{ *Resolver }

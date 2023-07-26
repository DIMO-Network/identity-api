package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.34

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/ethereum/go-ethereum/common"
)

// AccessibleVehicles is the resolver for the accessibleVehicles field.
func (r *queryResolver) AccessibleVehicles(ctx context.Context, address common.Address, first *int, after *string) (*model.VehicleConnection, error) {
	return r.Repo.GetAccessibleVehicles(ctx, address, first, after)
}

// OwnedAftermarketDevices is the resolver for the ownedAftermarketDevices field.
func (r *queryResolver) OwnedAftermarketDevices(ctx context.Context, address common.Address, first *int, after *string) (*model.AftermarketDeviceConnection, error) {
	return r.Repo.GetOwnedAftermarketDevices(ctx, address, first, after)
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }

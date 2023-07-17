package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.34

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	repo "github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/ethereum/go-ethereum/common"
)

// OwnedVehicles is the resolver for the ownedVehicles field.
func (r *queryResolver) OwnedVehicles(ctx context.Context, address common.Address, first *int, after *string) (*model.VehicleConnection, error) {
	vr := repo.NewVehiclesRepo(ctx, r.DB)
	return vr.GetOwnedVehicles(address, first, after)
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }

package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.54

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/loader"
)

// Vehicle is the resolver for the vehicle field.
func (r *dCNResolver) Vehicle(ctx context.Context, obj *model.Dcn) (*model.Vehicle, error) {
	if obj.VehicleID == nil {
		return nil, nil
	}
	return loader.GetVehicleByID(ctx, *obj.VehicleID)
}

// Dcn is the resolver for the dcn field.
func (r *queryResolver) Dcn(ctx context.Context, by model.DCNBy) (*model.Dcn, error) {
	return r.dcn.GetDCN(ctx, by)
}

// Dcns is the resolver for the dcns field.
func (r *queryResolver) Dcns(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *model.DCNFilter) (*model.DCNConnection, error) {
	return r.dcn.GetDCNs(ctx, first, after, last, before, filterBy)
}

// DCN returns DCNResolver implementation.
func (r *Resolver) DCN() DCNResolver { return &dCNResolver{r} }

type dCNResolver struct{ *Resolver }

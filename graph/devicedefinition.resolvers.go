package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.66

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
)

// DeviceDefinition is the resolver for the deviceDefinition field.
func (r *queryResolver) DeviceDefinition(ctx context.Context, by model.DeviceDefinitionBy) (*model.DeviceDefinition, error) {
	return r.deviceDefinition.GetDeviceDefinition(ctx, by)
}

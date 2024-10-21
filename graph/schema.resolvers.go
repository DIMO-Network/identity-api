package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.55

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/aftermarket"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/repositories/dcn"
	"github.com/DIMO-Network/identity-api/internal/repositories/manufacturer"
	"github.com/DIMO-Network/identity-api/internal/repositories/synthetic"
	"github.com/DIMO-Network/identity-api/internal/repositories/vehicle"
)

// Node is the resolver for the node field.
func (r *queryResolver) Node(ctx context.Context, id string) (model.Node, error) {
	prefix, objID, err := base.DecodeGlobalTokenID(id)
	if err != nil {
		return nil, fmt.Errorf("could not decode global id: %w", err)
	}
	switch prefix {
	case vehicle.TokenPrefix:
		return r.vehicle.GetVehicle(ctx, objID)
	case aftermarket.TokenPrefix:
		return r.aftermarket.GetAftermarketDevice(ctx, model.AftermarketDeviceBy{TokenID: &objID})
	case manufacturer.TokenPrefix:
		return r.manufacturer.GetManufacturer(ctx, model.ManufacturerBy{TokenID: &objID})
	case dcn.TokenPrefix:
		b := big.NewInt(int64(objID)).Bytes()
		return r.dcn.GetDCN(ctx, model.DCNBy{Node: b})
	case synthetic.TokenPrefix:
		return r.synthetic.GetSyntheticDevice(ctx, model.SyntheticDeviceBy{TokenID: &objID})
	default:
		return nil, errors.New("unrecognized global id") // TODO(elffs): Fix.
	}
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }

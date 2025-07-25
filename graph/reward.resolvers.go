package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.76

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/ethereum/go-ethereum/common"
)

// AftermarketDevice is the resolver for the aftermarketDevice field.
func (r *earningResolver) AftermarketDevice(ctx context.Context, obj *model.Earning) (*model.AftermarketDevice, error) {
	if obj.AftermarketDeviceID == nil {
		return nil, nil
	}

	return loader.GetAftermarketDeviceByID(ctx, *obj.AftermarketDeviceID)
}

// SyntheticDevice is the resolver for the syntheticDevice field.
func (r *earningResolver) SyntheticDevice(ctx context.Context, obj *model.Earning) (*model.SyntheticDevice, error) {
	if obj.SyntheticDeviceID == nil {
		return nil, nil
	}

	return loader.GetSyntheticDeviceByID(ctx, *obj.SyntheticDeviceID)
}

// Vehicle is the resolver for the vehicle field.
func (r *earningResolver) Vehicle(ctx context.Context, obj *model.Earning) (*model.Vehicle, error) {
	return loader.GetVehicleByID(ctx, obj.VehicleID)
}

// Rewards is the resolver for the rewards field.
func (r *queryResolver) Rewards(ctx context.Context, user common.Address) (*model.UserRewards, error) {
	return r.reward.GetEarningsByUserAddress(ctx, user)
}

// History is the resolver for the history field.
func (r *userRewardsResolver) History(ctx context.Context, obj *model.UserRewards, first *int, after *string, last *int, before *string) (*model.EarningsConnection, error) {
	return r.reward.PaginateGetEarningsByUsersDevices(ctx, obj, first, after, last, before)
}

// Earning returns EarningResolver implementation.
func (r *Resolver) Earning() EarningResolver { return &earningResolver{r} }

// UserRewards returns UserRewardsResolver implementation.
func (r *Resolver) UserRewards() UserRewardsResolver { return &userRewardsResolver{r} }

type earningResolver struct{ *Resolver }
type userRewardsResolver struct{ *Resolver }

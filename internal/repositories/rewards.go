package repositories

import (
	"context"
	"fmt"
	"math/big"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/dbtypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"golang.org/x/exp/slices"
)

type RewardsCursor struct {
	Week      int
	VehicleID int
}

func RewardToAPI(reward models.Reward) gmodel.Earning {
	stEarn := dbtypes.NullDecimalToInt(reward.StreakEarnings)
	adEarn := dbtypes.NullDecimalToInt(reward.AftermarketEarnings)
	syEarn := dbtypes.NullDecimalToInt(reward.SyntheticEarnings)

	return gmodel.Earning{
		Week:                    reward.IssuanceWeek,
		Beneficiary:             common.BytesToAddress(*reward.ReceivedByAddress.Ptr()),
		ConnectionStreak:        reward.ConnectionStreak.Ptr(),
		StreakTokens:            stEarn,
		AftermarketDeviceID:     reward.AftermarketTokenID.Ptr(),
		AftermarketDeviceTokens: adEarn,
		SyntheticDeviceID:       reward.SyntheticTokenID.Ptr(),
		SyntheticDeviceTokens:   syEarn,
		SentAt:                  reward.EarnedAt,
		VehicleID:               reward.VehicleID,
	}
}

func (r *Repository) paginateRewards(ctx context.Context, conditions []qm.QueryMod, first *int, after *string, last *int, before *string, limit int) (*gmodel.EarningsConnection, error) {

	queryMods := []qm.QueryMod{}
	queryMods = append(queryMods, conditions...)

	if after != nil {
		afterID, err := helpers.CursorToID(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.RewardWhere.IssuanceWeek.LT(afterID))
	}

	if before != nil {
		beforeID, err := helpers.CursorToID(*before)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.RewardWhere.IssuanceWeek.GT(beforeID))
	}

	orderBy := "DESC"
	if last != nil {
		orderBy = "ASC"
	}

	queryMods = append(queryMods,
		// Use limit + 1 here to check if there's another page.
		qm.Limit(limit+1),
		qm.OrderBy(models.RewardColumns.IssuanceWeek+" "+orderBy),
	)

	all, err := models.Rewards(queryMods...).All(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	// We assume that cursors come from real elements.
	hasNext := before != nil
	hasPrevious := after != nil

	if first != nil && len(all) == limit+1 {
		hasNext = true
		all = all[:limit]
	} else if last != nil && len(all) == limit+1 {
		hasPrevious = true
		all = all[:limit]
	}

	if last != nil {
		slices.Reverse(all)
	}

	edges := make([]*gmodel.EarningsEdge, len(all))
	nodes := make([]*gmodel.Earning, len(all))

	for i, rw := range all {
		reward := RewardToAPI(*rw)

		edges[i] = &gmodel.EarningsEdge{
			Node:   &reward,
			Cursor: helpers.IDToCursor(reward.Week),
		}

		nodes[i] = &reward
	}

	var endCursor, startCursor *string

	if len(all) != 0 {
		ec := helpers.IDToCursor(all[len(all)-1].IssuanceWeek)
		endCursor = &ec

		sc := helpers.IDToCursor(all[0].IssuanceWeek)
		startCursor = &sc
	}

	return &gmodel.EarningsConnection{
		Edges: edges,
		Nodes: nodes,
		PageInfo: &gmodel.PageInfo{
			StartCursor:     startCursor,
			EndCursor:       endCursor,
			HasPreviousPage: hasPrevious,
			HasNextPage:     hasNext,
		},
	}, nil
}

func (r *Repository) GetEarningsByVehicleID(ctx context.Context, tokenID int) (*gmodel.VehicleEarnings, error) {
	type rewardStats struct {
		TokenSum   types.NullDecimal `boil:"token_sum"`
		TotalCount int               `boil:"total_count"`
	}
	var stats rewardStats

	err := models.Rewards(
		qm.Select(
			fmt.Sprintf(
				`sum(%s + %s + %s) as token_sum`,
				models.RewardColumns.StreakEarnings, models.RewardColumns.AftermarketEarnings, models.RewardColumns.SyntheticEarnings,
			),
			"count(*) as total_count",
		),
		models.RewardWhere.VehicleID.EQ(tokenID),
	).Bind(ctx, r.pdb.DBS().Reader, &stats)
	if err != nil {
		return nil, err
	}

	if stats.TokenSum.IsZero() {
		return &gmodel.VehicleEarnings{
			TotalTokens: &big.Int{},
			History:     &gmodel.EarningsConnection{},
			VehicleID:   tokenID,
		}, nil
	}

	earningsConn := &gmodel.EarningsConnection{
		TotalCount: stats.TotalCount,
	}

	return &gmodel.VehicleEarnings{
		TotalTokens: dbtypes.NullDecimalToInt(stats.TokenSum),
		History:     earningsConn,
		VehicleID:   tokenID,
	}, nil
}

func (r *Repository) PaginateVehicleEarningsByID(ctx context.Context, vehicleEarnings *gmodel.VehicleEarnings, first *int, after *string, last *int, before *string) (*gmodel.EarningsConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, maxPageSize)
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{
		models.RewardWhere.VehicleID.EQ(vehicleEarnings.VehicleID),
	}
	vhs, err := r.paginateRewards(ctx, queryMods, first, after, last, before, limit)
	if err != nil {
		return nil, err
	}

	vehicleEarnings.History.Edges = vhs.Edges
	vehicleEarnings.History.Nodes = vhs.Nodes
	vehicleEarnings.History.PageInfo = vhs.PageInfo
	return vehicleEarnings.History, nil
}

func (r *Repository) GetEarningsByAfterMarketDeviceID(ctx context.Context, tokenID int) (*gmodel.AftermarketDeviceEarnings, error) {
	type rewardStats struct {
		TokenSum   types.NullDecimal `boil:"token_sum"`
		TotalCount int               `boil:"total_count"`
	}
	var stats rewardStats

	err := models.Rewards(
		qm.Select(
			fmt.Sprintf(
				`sum(%s + %s + %s) as token_sum`,
				models.RewardColumns.StreakEarnings, models.RewardColumns.AftermarketEarnings, models.RewardColumns.SyntheticEarnings,
			),
			"count(*) as total_count",
		),
		models.RewardWhere.AftermarketTokenID.EQ(null.IntFrom(tokenID)),
	).Bind(ctx, r.pdb.DBS().Reader, &stats)
	if err != nil {
		return nil, err
	}

	if stats.TotalCount == 0 {
		return &gmodel.AftermarketDeviceEarnings{
			TotalTokens: big.NewInt(0),
			History: &gmodel.EarningsConnection{
				PageInfo: &gmodel.PageInfo{},
			},
		}, nil
	}

	earningsConn := &gmodel.EarningsConnection{
		TotalCount: stats.TotalCount,
	}

	return &gmodel.AftermarketDeviceEarnings{
		TotalTokens:         dbtypes.NullDecimalToInt(stats.TokenSum),
		History:             earningsConn,
		AftermarketDeviceID: tokenID,
	}, nil
}

func (r *Repository) PaginateAfterMarketDeviceEarningsByID(ctx context.Context, afterMarketDeviceEarnings *gmodel.AftermarketDeviceEarnings, first *int, after *string, last *int, before *string) (*gmodel.EarningsConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, maxPageSize) // return early if both first and last are provided
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{
		models.RewardWhere.AftermarketTokenID.EQ(null.IntFrom(afterMarketDeviceEarnings.AftermarketDeviceID)),
	}

	afd, err := r.paginateRewards(ctx, queryMods, first, after, last, before, limit)
	if err != nil {
		return nil, err
	}

	afterMarketDeviceEarnings.History.Edges = afd.Edges
	afterMarketDeviceEarnings.History.Nodes = afd.Nodes
	afterMarketDeviceEarnings.History.PageInfo = afd.PageInfo
	return afterMarketDeviceEarnings.History, nil
}

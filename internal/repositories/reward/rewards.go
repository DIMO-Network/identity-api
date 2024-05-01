package reward

import (
	"context"
	"fmt"
	"math/big"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"golang.org/x/exp/slices"
)

type Repository struct {
	*base.Repository
}

type RewardsCursor struct {
	Week      int
	VehicleID int
}

type EarningsSummary struct {
	TokenSum   types.NullDecimal `boil:"token_sum"`
	TotalCount int               `boil:"total_count"`
}

var rewardsCursorColumns = "(" + models.RewardColumns.IssuanceWeek + ", " + models.RewardColumns.VehicleID + ")"

func RewardToAPI(reward models.Reward) gmodel.Earning {
	stEarn := NullDecimalToInt(reward.StreakEarnings)
	adEarn := NullDecimalToInt(reward.AftermarketEarnings)
	syEarn := NullDecimalToInt(reward.SyntheticEarnings)

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
	rwCursorHelper := &helpers.PaginationHelper[RewardsCursor]{}

	queryMods := []qm.QueryMod{}
	queryMods = append(queryMods, conditions...)

	orderBy := " DESC"
	if last != nil {
		orderBy = " ASC"
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(models.RewardColumns.IssuanceWeek+orderBy+", "+models.RewardColumns.VehicleID+orderBy),
	)

	if after != nil {
		afterT, err := rwCursorHelper.DecodeCursor(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods,
			qm.Where(rewardsCursorColumns+" < (?, ?)", afterT.Week, afterT.VehicleID),
		)
	} else if before != nil {
		beforeT, err := rwCursorHelper.DecodeCursor(*before)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods,
			qm.Where(rewardsCursorColumns+" > (?, ?)", beforeT.Week, beforeT.VehicleID),
		)
	}

	all, err := models.Rewards(queryMods...).All(ctx, r.PDB.DBS().Reader)
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

		crsr, err := rwCursorHelper.EncodeCursor(RewardsCursor{
			Week:      reward.Week,
			VehicleID: reward.VehicleID,
		})
		if err != nil {
			return nil, err
		}
		edges[i] = &gmodel.EarningsEdge{
			Node:   &reward,
			Cursor: crsr,
		}

		nodes[i] = &reward
	}

	var endCursor, startCursor *string

	if len(all) != 0 {
		endCursor = &edges[len(edges)-1].Cursor
		startCursor = &edges[0].Cursor
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

func (r *Repository) GetEarningsSummary(ctx context.Context, conditions []qm.QueryMod) (*EarningsSummary, error) {
	var summary EarningsSummary
	queryMods := []qm.QueryMod{
		qm.Select(
			fmt.Sprintf(
				"sum(COALESCE(%s, 0) + COALESCE(%s, 0) + COALESCE(%s, 0)) as token_sum",
				models.RewardColumns.StreakEarnings, models.RewardColumns.AftermarketEarnings, models.RewardColumns.SyntheticEarnings,
			),
			"count(*) as total_count",
		),
	}
	queryMods = append(queryMods, conditions...)

	err := models.Rewards(queryMods...).Bind(ctx, r.PDB.DBS().Reader, &summary)
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

func (r *Repository) GetEarningsByVehicleID(ctx context.Context, tokenID int) (*gmodel.VehicleEarnings, error) {
	summary, err := r.GetEarningsSummary(ctx, []qm.QueryMod{models.RewardWhere.VehicleID.EQ(tokenID)})
	if err != nil {
		return nil, err
	}

	earningsConn := &gmodel.EarningsConnection{
		TotalCount: summary.TotalCount,
	}

	return &gmodel.VehicleEarnings{
		TotalTokens: NullDecimalToInt(summary.TokenSum),
		History:     earningsConn,
		VehicleID:   tokenID,
	}, nil
}

func (r *Repository) PaginateVehicleEarningsByID(ctx context.Context, vehicleEarnings *gmodel.VehicleEarnings, first *int, after *string, last *int, before *string) (*gmodel.EarningsConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
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
	stats, err := r.GetEarningsSummary(ctx, []qm.QueryMod{models.RewardWhere.AftermarketTokenID.EQ(null.IntFrom(tokenID))})
	if err != nil {
		return nil, err
	}

	earningsConn := &gmodel.EarningsConnection{
		TotalCount: stats.TotalCount,
	}

	return &gmodel.AftermarketDeviceEarnings{
		TotalTokens:         NullDecimalToInt(stats.TokenSum),
		History:             earningsConn,
		AftermarketDeviceID: tokenID,
	}, nil
}

func (r *Repository) PaginateAftermarketDeviceEarningsByID(ctx context.Context, afterMarketDeviceEarnings *gmodel.AftermarketDeviceEarnings, first *int, after *string, last *int, before *string) (*gmodel.EarningsConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize) // return early if both first and last are provided
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

func (r *Repository) GetEarningsByUserAddress(ctx context.Context, user common.Address) (*gmodel.UserRewards, error) {
	summary, err := r.GetEarningsSummary(ctx, []qm.QueryMod{models.RewardWhere.ReceivedByAddress.EQ(null.BytesFrom(user.Bytes()))})
	if err != nil {
		return nil, err
	}

	earningsConn := &gmodel.EarningsConnection{
		TotalCount: summary.TotalCount,
	}

	return &gmodel.UserRewards{
		TotalTokens: NullDecimalToInt(summary.TokenSum),
		History:     earningsConn,
		User:        user,
	}, nil
}

func (r *Repository) PaginateGetEarningsByUsersDevices(ctx context.Context, userDeviceEarnings *gmodel.UserRewards, first *int, after *string, last *int, before *string) (*gmodel.EarningsConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize) // return early if both first and last are provided
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{
		models.RewardWhere.ReceivedByAddress.EQ(null.BytesFrom(userDeviceEarnings.User.Bytes())),
	}

	afd, err := r.paginateRewards(ctx, queryMods, first, after, last, before, limit)
	if err != nil {
		return nil, err
	}

	userDeviceEarnings.History.Edges = afd.Edges
	userDeviceEarnings.History.Nodes = afd.Nodes
	userDeviceEarnings.History.PageInfo = afd.PageInfo
	return userDeviceEarnings.History, nil
}

// NullDecimalToInt converts a null decimal to a non nil *big.Int.
// If the null decimal is nil, it returns a *big.Int with value 0.
func NullDecimalToInt(x types.NullDecimal) *big.Int {
	if x.IsZero() {
		return big.NewInt(0)
	}
	return x.Int(nil)
}

package repositories

import (
	"context"
	"fmt"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/dbtypes"
	"github.com/ethereum/go-ethereum/common"
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
		Beneficiary:             common.BytesToAddress(reward.ReceivedByAddress.Bytes),
		ConnectionStreak:        reward.ConnectionStreak.Int,
		StreakTokens:            stEarn,
		AftermarketDeviceID:     &reward.AftermarketTokenID.Int,
		AftermarketDeviceTokens: adEarn,
		SyntheticDeviceID:       &reward.SyntheticTokenID.Int,
		SyntheticDeviceTokens:   syEarn,
		SentAt:                  reward.EarnedAt,
		VehicleID:               reward.VehicleID,
	}
}

func (r *Repository) GetEarningsByVehicleID(ctx context.Context, tokenID int) (*gmodel.VehicleEarnings, error) {
	pHelp := helpers.PaginationHelper[RewardsCursor]{}

	type rewardStats struct {
		TokenSum   types.NullDecimal `boil:"token_sum"`
		TotalCount int               `boil:"total_count"`
	}
	var stats rewardStats

	limit := new(int)
	*limit = 100

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

	queryMods := []qm.QueryMod{
		models.RewardWhere.VehicleID.EQ(tokenID),
		qm.OrderBy(models.RewardColumns.IssuanceWeek + " desc"),
	}

	queryMods = append(queryMods, qm.Limit(*limit))

	rewards, err := models.Rewards(queryMods...).All(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	edges := make([]*gmodel.EarningsEdge, len(rewards))
	nodes := make([]*gmodel.Earning, len(rewards))

	for i, rw := range rewards {
		earning := RewardToAPI(*rw)
		cursor, err := pHelp.EncodeCursor(RewardsCursor{
			Week:      earning.Week,
			VehicleID: tokenID,
		})
		if err != nil {
			return nil, err
		}
		edges[i] = &gmodel.EarningsEdge{
			Node:   &earning,
			Cursor: cursor,
		}

		nodes[i] = &earning
	}

	if len(edges) == 0 {
		return &gmodel.VehicleEarnings{}, nil
	}

	endCursor, err := pHelp.EncodeCursor(RewardsCursor{
		Week:      edges[len(edges)-1].Node.Week,
		VehicleID: tokenID,
	})
	if err != nil {
		return nil, err
	}

	startCursor, err := pHelp.EncodeCursor(RewardsCursor{
		Week:      edges[0].Node.Week,
		VehicleID: tokenID,
	})
	if err != nil {
		return nil, err
	}

	earningsConn := &gmodel.EarningsConnection{
		TotalCount: stats.TotalCount,
		PageInfo: &gmodel.PageInfo{
			EndCursor:   &endCursor,
			StartCursor: &startCursor,
		},
		Edges: edges,
		Nodes: nodes,
	}

	return &gmodel.VehicleEarnings{
		TotalTokens: dbtypes.NullDecimalToInt(stats.TokenSum),
		History:     earningsConn,
		VehicleID:   tokenID,
	}, nil
}

func (r *Repository) PaginateVehicleEarningsByID(ctx context.Context, vehicleEarnings *gmodel.VehicleEarnings, first *int, after *string, last *int, before *string) (*gmodel.EarningsConnection, error) {
	pHelp := helpers.PaginationHelper[RewardsCursor]{}
	limit, err := pHelp.ValidateFirstLast(first, last, maxPageSize)
	if err != nil {
		return nil, err
	}

	if vehicleEarnings == nil || vehicleEarnings.History == nil {
		return &gmodel.EarningsConnection{}, nil
	}

	erningsConnectionCopy := *vehicleEarnings.History
	earningIdx := map[string]int{}
	for idx, val := range erningsConnectionCopy.Edges {
		earningIdx[val.Cursor] = idx
	}

	var edges []*gmodel.EarningsEdge
	var nodes []*gmodel.Earning

	if after != nil {
		startFrom := earningIdx[*after]

		if startFrom == len(erningsConnectionCopy.Edges)-1 {
			return &gmodel.EarningsConnection{
				PageInfo:   &gmodel.PageInfo{},
				TotalCount: vehicleEarnings.History.TotalCount,
			}, nil
		}

		if startFrom+1 < len(erningsConnectionCopy.Edges) {
			startFrom += 1
		}

		erningsConnectionCopy.Edges = erningsConnectionCopy.Edges[startFrom:]
		erningsConnectionCopy.Nodes = erningsConnectionCopy.Nodes[startFrom:]
	}

	if before != nil {
		startFrom := earningIdx[*before]

		if startFrom == 0 {
			return &gmodel.EarningsConnection{
				PageInfo:   &gmodel.PageInfo{},
				TotalCount: vehicleEarnings.History.TotalCount,
			}, nil
		}

		erningsConnectionCopy.Edges = erningsConnectionCopy.Edges[:startFrom]
		erningsConnectionCopy.Nodes = erningsConnectionCopy.Nodes[:startFrom]
	}

	if before != nil || last != nil {
		slices.Reverse(erningsConnectionCopy.Edges)
		slices.Reverse(erningsConnectionCopy.Nodes)
	}

	if *limit > len(erningsConnectionCopy.Edges) { // out of bounds protection
		*limit = len(erningsConnectionCopy.Edges)
	}
	edges = erningsConnectionCopy.Edges[:*limit]
	nodes = erningsConnectionCopy.Nodes[:*limit]

	erningsConnectionCopy.Edges = edges
	erningsConnectionCopy.Nodes = nodes

	hasNext := before != nil
	hasPrevious := after != nil

	if first != nil && len(vehicleEarnings.History.Edges) == *limit+1 {
		hasNext = true
	} else if last != nil && len(vehicleEarnings.History.Edges) == *limit+1 {
		hasPrevious = true
	}

	erningsConnectionCopy.PageInfo.HasNextPage = hasNext
	erningsConnectionCopy.PageInfo.HasPreviousPage = hasPrevious

	erningsConnectionCopy.PageInfo.StartCursor = &erningsConnectionCopy.Edges[0].Cursor
	erningsConnectionCopy.PageInfo.EndCursor = &erningsConnectionCopy.Edges[len(erningsConnectionCopy.Edges)-1].Cursor

	return &erningsConnectionCopy, nil
}

package repositories

import (
	"context"
	"math/big"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/dbtypes"
	"github.com/ethereum/go-ethereum/common"
)

func (r *Repository) GetEarningsByVehicleID(ctx context.Context, tokenID int) (*gmodel.EarningsConnection, error) {
	rewards, err := models.Rewards(
		models.RewardWhere.VehicleID.EQ(tokenID),
	).All(ctx, r.pdb.DBS().Reader)

	if err != nil {
		return nil, err
	}

	earnings := []*gmodel.EarningsEdge{}
	totalTokensEarned := big.NewInt(0)

	for _, reward := range rewards {
		stEarn := dbtypes.NullDecimalToInt(reward.StreakEarnings)
		adEarn := dbtypes.NullDecimalToInt(reward.AftermarketEarnings)
		syEarn := dbtypes.NullDecimalToInt(reward.SyntheticEarnings)

		totalTokensEarned = totalTokensEarned.Add(totalTokensEarned, stEarn)
		totalTokensEarned = totalTokensEarned.Add(totalTokensEarned, adEarn)
		totalTokensEarned = totalTokensEarned.Add(totalTokensEarned, syEarn)

		earning := &gmodel.EarningsEdge{
			Node: &gmodel.EarningNode{
				Week:                    reward.IssuanceWeek,
				Beneficiary:             common.BytesToAddress(reward.ReceivedByAddress.Bytes),
				ConnectionStreak:        reward.ConnectionStreak.Int,
				StreakTokens:            stEarn,
				AftermarketDeviceID:     &reward.AftermarketTokenID.Int,
				AftermarketDeviceTokens: adEarn,
				SyntheticDeviceID:       &reward.SyntheticTokenID.Int,
				SyntheticDeviceTokens:   syEarn,
				SentAt:                  reward.EarnedAt,
			},
		}

		earnings = append(earnings, earning)
	}

	return &gmodel.EarningsConnection{
		Edges: earnings,
	}, err
}

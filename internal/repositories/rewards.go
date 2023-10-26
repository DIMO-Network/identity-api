package repositories

import (
	"context"
	"log"
	"math/big"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/types"
)

// toUint64 takes a nullable decimal and returns uint6(0) if there is no value, or
// a reference to the uint64 value of the decimal otherwise. If the value does not
// fit then we return uint64(0) and log.
func toUint64(dec types.NullDecimal) uint64 {
	if dec.IsZero() {
		return uint64(0)
	}

	ui, ok := dec.Uint64()
	if !ok {
		log.Println(dec.String(), "Value too large for uint64.")
		return uint64(0)
	}

	return ui
}

func (r *Repository) GetEarningsByVehicleID(ctx context.Context, tokenID int) (*gmodel.VehicleEarningsConnection, error) {
	rewards, err := models.Rewards(
		models.RewardWhere.VehicleID.EQ(tokenID),
	).All(ctx, r.pdb.DBS().Reader)

	if err != nil {
		return nil, err
	}

	earnings := []*gmodel.EarningsEdge{}
	totalTokensEarned := big.NewInt(0)

	for _, reward := range rewards {
		stEarn := new(big.Int).SetUint64(toUint64(reward.StreakEarnings))
		adEarn := new(big.Int).SetUint64(toUint64(reward.AftermarketEarnings))
		syEarn := new(big.Int).SetUint64(toUint64(reward.SyntheticEarnings))

		totalTokensEarned = totalTokensEarned.Add(totalTokensEarned, stEarn)
		totalTokensEarned = totalTokensEarned.Add(totalTokensEarned, adEarn)
		totalTokensEarned = totalTokensEarned.Add(totalTokensEarned, syEarn)

		earning := &gmodel.EarningsEdge{
			Node: &gmodel.EarningNode{
				Week:                    reward.IssuanceWeek,
				Beneficiary:             common.BytesToAddress(reward.ReceivedByAddress.Bytes),
				ConnectionStreak:        reward.ConnectionStreak.Int,
				StreakTokens:            stEarn,
				AftermarketDevice:       reward.AftermarketTokenID.Int,
				AftermarketDeviceTokens: adEarn,
				SyntheticDevice:         reward.SyntheticTokenID.Int,
				SyntheticDeviceTokens:   syEarn,
				SentAt:                  reward.EarnedAt,
			},
		}

		earnings = append(earnings, earning)
	}

	return &gmodel.VehicleEarningsConnection{
		EarnedTokens:      totalTokensEarned,
		EarningsTransfers: earnings,
	}, err
}

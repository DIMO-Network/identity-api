package repositories

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	test "github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/dbtypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type RewardsRepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container testcontainers.Container
	repo      *Repository
	settings  config.Settings
}

func (r *RewardsRepoTestSuite) SetupSuite() {
	r.ctx = context.Background()
	r.pdb, r.container = test.StartContainerDatabase(r.ctx, r.T(), "../../migrations")

	r.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
	}
	r.repo = New(r.pdb)
}

// TearDownTest after each test truncate tables
func (r *RewardsRepoTestSuite) TearDownTest() {
	test.TruncateTables(r.pdb.DBS().Writer.DB, r.T())
}

// TearDownSuite cleanup at end by terminating container
func (r *RewardsRepoTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", r.container.SessionID())

	if err := r.container.Terminate(r.ctx); err != nil {
		r.T().Fatal(err)
	}
}

// Test Runner
func TestRewardsRepoTestSuite(t *testing.T) {
	suite.Run(t, new(RewardsRepoTestSuite))
}

func (r *RewardsRepoTestSuite) Test_GetEarningsByVehicleID_Success() {

	_, ben, err := test.GenerateWallet()
	r.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	payloads := []struct {
		AD  models.AftermarketDevice
		SD  models.SyntheticDevice
		Veh models.Vehicle
		RW  models.Reward
	}{
		{
			Veh: models.Vehicle{
				ID:           11,
				OwnerAddress: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
				Make:         null.StringFrom("Ford"),
				Model:        null.StringFrom("Bronco"),
				Year:         null.IntFrom(2022),
				MintedAt:     time.Now(),
			},
			AD: models.AftermarketDevice{
				ID:          1,
				Address:     common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf5").Bytes(),
				Owner:       common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes(),
				Serial:      null.StringFrom("aftermarketDeviceSerial-1"),
				Imei:        null.StringFrom("aftermarketDeviceIMEI-1"),
				MintedAt:    time.Now(),
				VehicleID:   null.IntFrom(11),
				Beneficiary: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
			},
			SD: models.SyntheticDevice{
				ID:            1,
				IntegrationID: 2,
				VehicleID:     11,
				DeviceAddress: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
				MintedAt:      time.Now(),
			},
			RW: models.Reward{
				IssuanceWeek:       1,
				VehicleID:          11,
				ConnectionStreak:   null.IntFrom(20),
				AftermarketTokenID: null.IntFrom(1),
				SyntheticTokenID:   null.IntFrom(1),
				ReceivedByAddress:  null.BytesFrom(ben.Bytes()),
				EarnedAt:           currTime,
			},
		},
	}

	totalEarned := big.NewInt(0)

	// Aftermarket Earnings
	adEarn, ok := new(big.Int).SetString("59147051345528509681", 10)
	r.NotZero(ok)
	totalEarned = totalEarned.Add(totalEarned, adEarn)

	// Synthetic Earnings
	synthEarn, ok := new(big.Int).SetString("59147051345528509682", 10)
	r.NotZero(ok)
	totalEarned = totalEarned.Add(totalEarned, synthEarn)

	// Streak Earnings
	strkEarn, ok := new(big.Int).SetString("59147051345528509684", 10)
	r.NotZero(ok)
	totalEarned = totalEarned.Add(totalEarned, strkEarn)

	for _, payload := range payloads {
		err = payload.Veh.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer()) // Insert vehicle
		r.NoError(err)

		err = payload.AD.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer()) // Insert AftermarketDevice
		r.NoError(err)

		err = payload.SD.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer()) // Insert SyntheticDevice
		r.NoError(err)

		payload.RW.AftermarketEarnings = dbtypes.NullIntToDecimal(adEarn)

		payload.RW.SyntheticEarnings = dbtypes.NullIntToDecimal(synthEarn)

		payload.RW.StreakEarnings = dbtypes.NullIntToDecimal(strkEarn)

		// Insert Reward
		err = payload.RW.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	rwrd, err := r.repo.GetEarningsByVehicleID(r.ctx, 11)
	r.NoError(err)

	r.Equal(&model.VehicleEarningsConnection{
		EarnedTokens: totalEarned,
		EarningsTransfers: []*model.EarningsEdge{
			{
				Node: &model.EarningNode{
					Week:                    payloads[0].RW.IssuanceWeek,
					Beneficiary:             *ben,
					ConnectionStreak:        payloads[0].RW.ConnectionStreak.Int,
					StreakTokens:            strkEarn,
					AftermarketDeviceID:     &payloads[0].RW.AftermarketTokenID.Int,
					AftermarketDeviceTokens: adEarn,
					SyntheticDeviceID:       &payloads[0].RW.SyntheticTokenID.Int,
					SyntheticDeviceTokens:   synthEarn,
					SentAt:                  currTime,
				},
			},
		},
	}, rwrd)
}

package repositories

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/graph/model"
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	test "github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/dbtypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
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

func (r *RewardsRepoTestSuite) createDependentRecords() {
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
		},
	}

	for _, payload := range payloads {
		err := payload.Veh.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer()) // Insert vehicle
		r.NoError(err)

		err = payload.AD.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer()) // Insert AftermarketDevice
		r.NoError(err)

		err = payload.SD.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer()) // Insert SyntheticDevice
		r.NoError(err)

	}
}

func (r *RewardsRepoTestSuite) Test_GetEarningsByVehicleID_Success() {
	pHelp := helpers.PaginationHelper[RewardsCursor]{}

	_, ben, err := test.GenerateWallet()
	r.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	r.createDependentRecords()

	totalEarned := big.NewInt(0)

	rw := []models.Reward{
		{
			IssuanceWeek:       1,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(20),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(ben.Bytes()),
			EarnedAt:           currTime,
		},
		{
			IssuanceWeek:       2,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(21),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(ben.Bytes()),
			EarnedAt:           currTime,
		},
	}

	var aftEarn types.NullDecimal
	var strkEarn types.NullDecimal
	var syntEarn types.NullDecimal

	for _, rr := range rw {
		baseAmt, ok := new(big.Int).SetString("59147051345528509681", 10)
		r.NotZero(ok)

		aftEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(30)))
		strkEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(50)))
		syntEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(10)))

		rr.AftermarketEarnings = aftEarn
		rr.StreakEarnings = strkEarn
		rr.SyntheticEarnings = syntEarn

		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(aftEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(strkEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(syntEarn))

		err = rr.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	rwrd, err := r.repo.GetEarningsByVehicleID(r.ctx, 11)
	r.NoError(err)

	endCrsr, err := pHelp.EncodeCursor(RewardsCursor{
		Week:      1,
		VehicleID: 11,
	})
	r.NoError(err)

	startCrsr, err := pHelp.EncodeCursor(RewardsCursor{
		Week:      2,
		VehicleID: 11,
	})
	r.NoError(err)

	firstCrsr, err := pHelp.EncodeCursor(RewardsCursor{
		Week:      2,
		VehicleID: 11,
	})
	r.NoError(err)

	aftID := 1
	syntID := 1

	r.Equal(&gmodel.PageInfo{
		EndCursor:       &endCrsr,
		HasNextPage:     false,
		HasPreviousPage: false,
		StartCursor:     &startCrsr,
	}, rwrd.History.PageInfo)
	r.Equal(2, rwrd.History.TotalCount)
	r.Equal(11, rwrd.VehicleID)
	r.Equal(totalEarned, rwrd.TotalTokens)
	r.Equal([]*gmodel.EarningsEdge{
		{
			Node: &gmodel.Earning{
				Week:                    2,
				Beneficiary:             common.BytesToAddress(ben.Bytes()),
				ConnectionStreak:        21,
				StreakTokens:            dbtypes.NullDecimalToInt(strkEarn),
				AftermarketDeviceID:     &aftID,
				AftermarketDeviceTokens: dbtypes.NullDecimalToInt(aftEarn),
				SyntheticDeviceID:       &syntID,
				SyntheticDeviceTokens:   dbtypes.NullDecimalToInt(syntEarn),
				SentAt:                  currTime,
				VehicleID:               11,
			},
			Cursor: firstCrsr,
		},
		{
			Node: &gmodel.Earning{
				Week:                    1,
				Beneficiary:             common.BytesToAddress(ben.Bytes()),
				ConnectionStreak:        20,
				StreakTokens:            dbtypes.NullDecimalToInt(strkEarn),
				AftermarketDeviceID:     &aftID,
				AftermarketDeviceTokens: dbtypes.NullDecimalToInt(aftEarn),
				SyntheticDeviceID:       &syntID,
				SyntheticDeviceTokens:   dbtypes.NullDecimalToInt(syntEarn),
				SentAt:                  currTime,
				VehicleID:               11,
			},
			Cursor: endCrsr,
		},
	}, rwrd.History.Edges)
}

func (r *RewardsRepoTestSuite) Test_PaginateVehicleEarningsByID_Disallow_FirstAndLast() {
	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	r.createDependentRecords()

	totalEarned := big.NewInt(0)

	rw := []models.Reward{
		{
			IssuanceWeek:       1,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(20),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
		{
			IssuanceWeek:       2,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(21),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
	}

	var aftEarn types.NullDecimal
	var strkEarn types.NullDecimal
	var syntEarn types.NullDecimal

	for _, rr := range rw {
		baseAmt, ok := new(big.Int).SetString("59147051345528509681", 10)
		r.NotZero(ok)

		aftEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(30)))
		strkEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(50)))
		syntEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(10)))

		rr.AftermarketEarnings = aftEarn
		rr.StreakEarnings = strkEarn
		rr.SyntheticEarnings = syntEarn

		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(aftEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(strkEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(syntEarn))

		err = rr.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	rwrd, err := r.repo.GetEarningsByVehicleID(r.ctx, 11)
	r.NoError(err)

	first := 1
	last := 2
	_, err = r.repo.PaginateVehicleEarningsByID(r.ctx, rwrd, &first, nil, &last, nil)
	r.EqualError(err, "pass `first` or `last`, but not both")
}

func (r *RewardsRepoTestSuite) Test_PaginateVehicleEarningsByID_FwdPagination_First() {
	pHelp := helpers.PaginationHelper[RewardsCursor]{}

	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	r.createDependentRecords()

	totalEarned := big.NewInt(0)

	rw := []models.Reward{
		{
			IssuanceWeek:       1,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(20),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
		{
			IssuanceWeek:       2,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(21),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
	}

	var aftEarn types.NullDecimal
	var strkEarn types.NullDecimal
	var syntEarn types.NullDecimal

	for _, rr := range rw {
		baseAmt, ok := new(big.Int).SetString("59147051345528509681", 10)
		r.NotZero(ok)

		aftEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(30)))
		strkEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(50)))
		syntEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(10)))

		rr.AftermarketEarnings = aftEarn
		rr.StreakEarnings = strkEarn
		rr.SyntheticEarnings = syntEarn

		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(aftEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(strkEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(syntEarn))

		err = rr.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	rwrd, err := r.repo.GetEarningsByVehicleID(r.ctx, 11)
	r.NoError(err)

	first := 1
	paginatedEarnings, err := r.repo.PaginateVehicleEarningsByID(r.ctx, rwrd, &first, nil, nil, nil)
	r.NoError(err)

	crsr, err := pHelp.EncodeCursor(RewardsCursor{
		Week:      2,
		VehicleID: 11,
	})
	r.NoError(err)
	aftID := 1
	syntID := 1

	r.Equal(&model.PageInfo{
		EndCursor:       &crsr,
		HasNextPage:     true,
		HasPreviousPage: false,
		StartCursor:     &crsr,
	}, paginatedEarnings.PageInfo)
	r.Equal(2, paginatedEarnings.TotalCount)
	r.Equal([]*model.EarningsEdge{
		{
			Node: &model.Earning{
				Week:                    2,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        21,
				StreakTokens:            dbtypes.NullDecimalToInt(strkEarn),
				AftermarketDeviceID:     &aftID,
				AftermarketDeviceTokens: dbtypes.NullDecimalToInt(aftEarn),
				SyntheticDeviceID:       &syntID,
				SyntheticDeviceTokens:   dbtypes.NullDecimalToInt(syntEarn),
				SentAt:                  currTime,
				VehicleID:               11,
			},
			Cursor: crsr,
		},
	}, paginatedEarnings.Edges)
}

func (r *RewardsRepoTestSuite) Test_PaginateVehicleEarningsByID_FwdPagination_First_After() {
	pHelp := helpers.PaginationHelper[RewardsCursor]{}

	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	r.createDependentRecords()

	totalEarned := big.NewInt(0)

	rw := []models.Reward{
		{
			IssuanceWeek:       1,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(20),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
		{
			IssuanceWeek:       2,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(21),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
		{
			IssuanceWeek:       3,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(22),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
	}

	var aftEarn types.NullDecimal
	var strkEarn types.NullDecimal
	var syntEarn types.NullDecimal

	for _, rr := range rw {
		baseAmt, ok := new(big.Int).SetString("59147051345528509681", 10)
		r.NotZero(ok)

		aftEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(30)))
		strkEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(50)))
		syntEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(10)))

		rr.AftermarketEarnings = aftEarn
		rr.StreakEarnings = strkEarn
		rr.SyntheticEarnings = syntEarn

		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(aftEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(strkEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(syntEarn))

		err = rr.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	rwrd, err := r.repo.GetEarningsByVehicleID(r.ctx, 11)
	r.NoError(err)

	first := 2
	after := "kgML"
	paginatedEarnings, err := r.repo.PaginateVehicleEarningsByID(r.ctx, rwrd, &first, &after, nil, nil)
	r.NoError(err)

	startCrsr, err := pHelp.EncodeCursor(RewardsCursor{
		Week:      2,
		VehicleID: 11,
	})
	r.NoError(err)
	endCrsr, err := pHelp.EncodeCursor(RewardsCursor{
		Week:      1,
		VehicleID: 11,
	})
	r.NoError(err)

	aftID := 1
	syntID := 1

	r.Equal(&model.PageInfo{
		EndCursor:       &endCrsr,
		HasNextPage:     true,
		HasPreviousPage: true,
		StartCursor:     &startCrsr,
	}, paginatedEarnings.PageInfo)
	r.Equal(3, paginatedEarnings.TotalCount)
	r.Equal([]*model.EarningsEdge{
		{
			Node: &model.Earning{
				Week:                    2,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        21,
				StreakTokens:            dbtypes.NullDecimalToInt(strkEarn),
				AftermarketDeviceID:     &aftID,
				AftermarketDeviceTokens: dbtypes.NullDecimalToInt(aftEarn),
				SyntheticDeviceID:       &syntID,
				SyntheticDeviceTokens:   dbtypes.NullDecimalToInt(syntEarn),
				SentAt:                  currTime,
				VehicleID:               11,
			},
			Cursor: startCrsr,
		},
		{
			Node: &model.Earning{
				Week:                    1,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        20,
				StreakTokens:            dbtypes.NullDecimalToInt(strkEarn),
				AftermarketDeviceID:     &aftID,
				AftermarketDeviceTokens: dbtypes.NullDecimalToInt(aftEarn),
				SyntheticDeviceID:       &syntID,
				SyntheticDeviceTokens:   dbtypes.NullDecimalToInt(syntEarn),
				SentAt:                  currTime,
				VehicleID:               11,
			},
			Cursor: endCrsr,
		},
	}, paginatedEarnings.Edges)
}

func (r *RewardsRepoTestSuite) Test_PaginateVehicleEarningsByID_FwdPagination_EmptyWhenOutOfBounds() {
	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	r.createDependentRecords()

	totalEarned := big.NewInt(0)

	rw := []models.Reward{
		{
			IssuanceWeek:       1,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(20),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
		{
			IssuanceWeek:       2,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(21),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
		{
			IssuanceWeek:       3,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(22),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
	}

	var aftEarn types.NullDecimal
	var strkEarn types.NullDecimal
	var syntEarn types.NullDecimal

	for _, rr := range rw {
		baseAmt, ok := new(big.Int).SetString("59147051345528509681", 10)
		r.NotZero(ok)

		aftEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(30)))
		strkEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(50)))
		syntEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(10)))

		rr.AftermarketEarnings = aftEarn
		rr.StreakEarnings = strkEarn
		rr.SyntheticEarnings = syntEarn

		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(aftEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(strkEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(syntEarn))

		err = rr.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	rwrd, err := r.repo.GetEarningsByVehicleID(r.ctx, 11)
	r.NoError(err)

	first := 2
	after := "kgEL"
	paginatedEarnings, err := r.repo.PaginateVehicleEarningsByID(r.ctx, rwrd, &first, &after, nil, nil)
	r.NoError(err)

	r.Equal(&gmodel.EarningsConnection{
		TotalCount: 3,
		Edges:      nil,
		Nodes:      nil,
		PageInfo:   &gmodel.PageInfo{},
	}, paginatedEarnings)
}

func (r *RewardsRepoTestSuite) Test_PaginateVehicleEarningsByID_BackPagination_Last() {
	pHelp := helpers.PaginationHelper[RewardsCursor]{}

	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	r.createDependentRecords()

	totalEarned := big.NewInt(0)

	rw := []models.Reward{
		{
			IssuanceWeek:       1,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(20),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
		{
			IssuanceWeek:       2,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(21),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
	}

	var aftEarn types.NullDecimal
	var strkEarn types.NullDecimal
	var syntEarn types.NullDecimal

	for _, rr := range rw {
		baseAmt, ok := new(big.Int).SetString("59147051345528509681", 10)
		r.NotZero(ok)

		aftEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(30)))
		strkEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(50)))
		syntEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(10)))

		rr.AftermarketEarnings = aftEarn
		rr.StreakEarnings = strkEarn
		rr.SyntheticEarnings = syntEarn

		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(aftEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(strkEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(syntEarn))

		err = rr.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	rwrd, err := r.repo.GetEarningsByVehicleID(r.ctx, 11)
	r.NoError(err)

	last := 1
	paginatedEarnings, err := r.repo.PaginateVehicleEarningsByID(r.ctx, rwrd, nil, nil, &last, nil)
	r.NoError(err)

	crsr, err := pHelp.EncodeCursor(RewardsCursor{
		Week:      1,
		VehicleID: 11,
	})
	r.NoError(err)
	aftID := 1
	syntID := 1

	r.Equal(&model.PageInfo{
		EndCursor:       &crsr,
		HasNextPage:     false,
		HasPreviousPage: true,
		StartCursor:     &crsr,
	}, paginatedEarnings.PageInfo)
	r.Equal(2, paginatedEarnings.TotalCount)
	r.Equal([]*model.EarningsEdge{
		{
			Node: &model.Earning{
				Week:                    1,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        20,
				StreakTokens:            dbtypes.NullDecimalToInt(strkEarn),
				AftermarketDeviceID:     &aftID,
				AftermarketDeviceTokens: dbtypes.NullDecimalToInt(aftEarn),
				SyntheticDeviceID:       &syntID,
				SyntheticDeviceTokens:   dbtypes.NullDecimalToInt(syntEarn),
				SentAt:                  currTime,
				VehicleID:               11,
			},
			Cursor: crsr,
		},
	}, paginatedEarnings.Edges)
}

func (r *RewardsRepoTestSuite) Test_PaginateVehicleEarningsByID_BackPagination_Last_Before() {
	pHelp := helpers.PaginationHelper[RewardsCursor]{}

	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	r.createDependentRecords()

	totalEarned := big.NewInt(0)

	rw := []models.Reward{
		{
			IssuanceWeek:       1,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(20),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
		{
			IssuanceWeek:       2,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(21),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
		{
			IssuanceWeek:       3,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(22),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
	}

	var aftEarn types.NullDecimal
	var strkEarn types.NullDecimal
	var syntEarn types.NullDecimal

	for _, rr := range rw {
		baseAmt, ok := new(big.Int).SetString("59147051345528509681", 10)
		r.NotZero(ok)

		aftEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(30)))
		strkEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(50)))
		syntEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(10)))

		rr.AftermarketEarnings = aftEarn
		rr.StreakEarnings = strkEarn
		rr.SyntheticEarnings = syntEarn

		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(aftEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(strkEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(syntEarn))

		err = rr.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	rwrd, err := r.repo.GetEarningsByVehicleID(r.ctx, 11)
	r.NoError(err)

	last := 2
	before := "kgEL"
	paginatedEarnings, err := r.repo.PaginateVehicleEarningsByID(r.ctx, rwrd, nil, nil, &last, &before)
	r.NoError(err)

	startCrsr, err := pHelp.EncodeCursor(RewardsCursor{
		Week:      2,
		VehicleID: 11,
	})
	r.NoError(err)
	endCrsr, err := pHelp.EncodeCursor(RewardsCursor{
		Week:      3,
		VehicleID: 11,
	})
	r.NoError(err)

	aftID := 1
	syntID := 1

	r.Equal(&model.PageInfo{
		EndCursor:       &endCrsr,
		HasNextPage:     true,
		HasPreviousPage: true,
		StartCursor:     &startCrsr,
	}, paginatedEarnings.PageInfo)
	r.Equal(3, paginatedEarnings.TotalCount)
	r.Equal([]*model.EarningsEdge{
		{
			Node: &model.Earning{
				Week:                    2,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        21,
				StreakTokens:            dbtypes.NullDecimalToInt(strkEarn),
				AftermarketDeviceID:     &aftID,
				AftermarketDeviceTokens: dbtypes.NullDecimalToInt(aftEarn),
				SyntheticDeviceID:       &syntID,
				SyntheticDeviceTokens:   dbtypes.NullDecimalToInt(syntEarn),
				SentAt:                  currTime,
				VehicleID:               11,
			},
			Cursor: startCrsr,
		},
		{
			Node: &model.Earning{
				Week:                    3,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        22,
				StreakTokens:            dbtypes.NullDecimalToInt(strkEarn),
				AftermarketDeviceID:     &aftID,
				AftermarketDeviceTokens: dbtypes.NullDecimalToInt(aftEarn),
				SyntheticDeviceID:       &syntID,
				SyntheticDeviceTokens:   dbtypes.NullDecimalToInt(syntEarn),
				SentAt:                  currTime,
				VehicleID:               11,
			},
			Cursor: endCrsr,
		},
	}, paginatedEarnings.Edges)
}

func (r *RewardsRepoTestSuite) Test_PaginateVehicleEarningsByID_BackPagination_EmptyWhenOutOfBounds() {
	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	r.createDependentRecords()

	totalEarned := big.NewInt(0)

	rw := []models.Reward{
		{
			IssuanceWeek:       1,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(20),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
		{
			IssuanceWeek:       2,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(21),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
		{
			IssuanceWeek:       3,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(22),
			AftermarketTokenID: null.IntFrom(1),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:           currTime,
		},
	}

	var aftEarn types.NullDecimal
	var strkEarn types.NullDecimal
	var syntEarn types.NullDecimal

	for _, rr := range rw {
		baseAmt, ok := new(big.Int).SetString("59147051345528509681", 10)
		r.NotZero(ok)

		aftEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(30)))
		strkEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(50)))
		syntEarn = dbtypes.NullIntToDecimal(baseAmt.Add(baseAmt, big.NewInt(10)))

		rr.AftermarketEarnings = aftEarn
		rr.StreakEarnings = strkEarn
		rr.SyntheticEarnings = syntEarn

		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(aftEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(strkEarn))
		totalEarned.Add(totalEarned, dbtypes.NullDecimalToInt(syntEarn))

		err = rr.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	rwrd, err := r.repo.GetEarningsByVehicleID(r.ctx, 11)
	r.NoError(err)

	last := 2
	before := "kgML"
	paginatedEarnings, err := r.repo.PaginateVehicleEarningsByID(r.ctx, rwrd, nil, nil, &last, &before)
	r.NoError(err)

	r.Equal(&gmodel.EarningsConnection{
		TotalCount: 3,
		Edges:      nil,
		Nodes:      nil,
		PageInfo:   &gmodel.PageInfo{},
	}, paginatedEarnings)
}

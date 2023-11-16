package repositories

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/dbtypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type RewardsRepoTestSuite struct {
	suite.Suite
	ctx              context.Context
	pdb              db.Store
	container        testcontainers.Container
	repo             *Repository
	settings         config.Settings
	paginationHelper helpers.PaginationHelper[RewardsCursor]
}

type createRewardsRecordsInput struct {
	beneficiary         common.Address
	dateTime            time.Time
	afterMarketDeviceID int
}

func (r *RewardsRepoTestSuite) SetupSuite() {
	r.ctx = context.Background()
	r.pdb, r.container = helpers.StartContainerDatabase(r.ctx, r.T(), "../../migrations")

	r.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
	}
	r.repo = New(r.pdb, r.settings)
	r.paginationHelper = helpers.PaginationHelper[RewardsCursor]{}

}

// TearDownTest after each test truncate tables
func (r *RewardsRepoTestSuite) TearDownTest() {
	helpers.TruncateTables(r.pdb.DBS().Writer.DB, r.T())
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

func (r *RewardsRepoTestSuite) createRewardsRecords(count int, args createRewardsRecordsInput) ([]models.Reward, error) {
	rewards := []models.Reward{}
	for idx := 1; idx <= count; idx++ {

		rwrd := models.Reward{
			IssuanceWeek:       idx,
			VehicleID:          11,
			ConnectionStreak:   null.IntFrom(20 + idx),
			AftermarketTokenID: null.IntFrom(args.afterMarketDeviceID),
			SyntheticTokenID:   null.IntFrom(1),
			ReceivedByAddress:  null.BytesFrom(args.beneficiary.Bytes()),
			EarnedAt:           args.dateTime,
		}

		rewards = append(rewards, rwrd)
	}

	return rewards, nil
}

func (r *RewardsRepoTestSuite) Test_GetEarningsSummary_Success() {
	_, ben, err := helpers.GenerateWallet()
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

	summary, err := r.repo.GetEarningsSummary(r.ctx, []qm.QueryMod{models.RewardWhere.VehicleID.EQ(11)})
	r.NoError(err)

	r.Equal(2, summary.TotalCount)
	r.Equal(totalEarned, dbtypes.NullDecimalToInt(summary.TokenSum))
}

func (r *RewardsRepoTestSuite) Test_GetEarningsByVehicleID_Success() {
	_, ben, err := helpers.GenerateWallet()
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

	r.Equal(&gmodel.VehicleEarnings{
		TotalTokens: rwrd.TotalTokens,
		History: &gmodel.EarningsConnection{
			TotalCount: rwrd.History.TotalCount,
			Edges:      nil,
			Nodes:      nil,
		},
		VehicleID: rwrd.VehicleID,
	}, rwrd)
}

func (r *RewardsRepoTestSuite) Test_GetEarningsByVehicleID_NoRows() {
	rwrd, err := r.repo.GetEarningsByVehicleID(r.ctx, 11)
	r.NoError(err)

	r.Equal(&gmodel.VehicleEarnings{
		TotalTokens: big.NewInt(0),
		History: &gmodel.EarningsConnection{
			TotalCount: 0,
			Edges:      nil,
			Nodes:      nil,
		},
		VehicleID: 11,
	}, rwrd)
}

func (r *RewardsRepoTestSuite) Test_PaginateVehicleEarningsByID_Disallow_FirstAndLast() {

	first := 1
	last := 2
	_, err := r.repo.PaginateVehicleEarningsByID(r.ctx, &gmodel.VehicleEarnings{}, &first, nil, &last, nil)
	r.EqualError(err, "pass `first` or `last`, but not both")
}

func (r *RewardsRepoTestSuite) Test_PaginateVehicleEarningsByID_FwdPagination_First() {
	_, beneficiary, err := helpers.GenerateWallet()
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

	crsr, err := r.paginationHelper.EncodeCursor(RewardsCursor{Week: 2})
	r.NoError(err)

	aftID := 1
	syntID := 1
	connStrk := 21

	r.Equal(&gmodel.PageInfo{
		EndCursor:       &crsr,
		HasNextPage:     true,
		HasPreviousPage: false,
		StartCursor:     &crsr,
	}, paginatedEarnings.PageInfo)
	r.Equal(2, paginatedEarnings.TotalCount)
	r.Equal([]*gmodel.EarningsEdge{
		{
			Node: &gmodel.Earning{
				Week:                    2,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        &connStrk,
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
	_, beneficiary, err := helpers.GenerateWallet()
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
	after := "kwMAxCoweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA="
	paginatedEarnings, err := r.repo.PaginateVehicleEarningsByID(r.ctx, rwrd, &first, &after, nil, nil)
	r.NoError(err)

	startCrsr, err := r.paginationHelper.EncodeCursor(RewardsCursor{Week: 2})
	r.NoError(err)
	endCrsr, err := r.paginationHelper.EncodeCursor(RewardsCursor{Week: 1})
	r.NoError(err)

	aftID := 1
	syntID := 1
	connStrk := [2]int{21, 20}

	r.Equal(&gmodel.PageInfo{
		EndCursor:       &endCrsr,
		HasNextPage:     false,
		HasPreviousPage: true,
		StartCursor:     &startCrsr,
	}, paginatedEarnings.PageInfo)
	r.Equal(3, paginatedEarnings.TotalCount)
	r.Equal([]*gmodel.EarningsEdge{
		{
			Node: &gmodel.Earning{
				Week:                    2,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        &connStrk[0],
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
			Node: &gmodel.Earning{
				Week:                    1,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        &connStrk[1],
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
	_, beneficiary, err := helpers.GenerateWallet()
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
	after := "kwEAxCoweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA="
	paginatedEarnings, err := r.repo.PaginateVehicleEarningsByID(r.ctx, rwrd, &first, &after, nil, nil)
	r.NoError(err)

	r.Equal(&gmodel.EarningsConnection{
		TotalCount: 3,
		Edges:      []*gmodel.EarningsEdge{},
		Nodes:      []*gmodel.Earning{},
		PageInfo: &gmodel.PageInfo{
			HasPreviousPage: true,
			HasNextPage:     false,
		},
	}, paginatedEarnings)
}

func (r *RewardsRepoTestSuite) Test_PaginateVehicleEarningsByID_BackPagination_Last() {
	_, beneficiary, err := helpers.GenerateWallet()
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

	crsr, err := r.paginationHelper.EncodeCursor(RewardsCursor{Week: 1})
	r.NoError(err)

	r.NoError(err)
	aftID := 1
	syntID := 1

	connStrk := 20

	r.Equal(&gmodel.PageInfo{
		EndCursor:       &crsr,
		HasNextPage:     false,
		HasPreviousPage: true,
		StartCursor:     &crsr,
	}, paginatedEarnings.PageInfo)
	r.Equal(2, paginatedEarnings.TotalCount)
	r.Equal([]*gmodel.EarningsEdge{
		{
			Node: &gmodel.Earning{
				Week:                    1,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        &connStrk,
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
	_, beneficiary, err := helpers.GenerateWallet()
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
	before := "kwEAxCoweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA="
	paginatedEarnings, err := r.repo.PaginateVehicleEarningsByID(r.ctx, rwrd, nil, nil, &last, &before)
	r.NoError(err)

	startCrsr, err := r.paginationHelper.EncodeCursor(RewardsCursor{Week: 3})
	r.NoError(err)

	endCrsr, err := r.paginationHelper.EncodeCursor(RewardsCursor{Week: 2})
	r.NoError(err)

	aftID := 1
	syntID := 1

	connStrk := [2]int{21, 22}

	r.Equal(&gmodel.PageInfo{
		EndCursor:       &endCrsr,
		HasNextPage:     true,
		HasPreviousPage: false,
		StartCursor:     &startCrsr,
	}, paginatedEarnings.PageInfo)
	r.Equal(3, paginatedEarnings.TotalCount)
	r.Equal([]*gmodel.EarningsEdge{
		{
			Node: &gmodel.Earning{
				Week:                    3,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        &connStrk[1],
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
			Node: &gmodel.Earning{
				Week:                    2,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        &connStrk[0],
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
	_, beneficiary, err := helpers.GenerateWallet()
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
	before := "kwMAxCoweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA="
	paginatedEarnings, err := r.repo.PaginateVehicleEarningsByID(r.ctx, rwrd, nil, nil, &last, &before)
	r.NoError(err)

	r.Equal(&gmodel.EarningsConnection{
		TotalCount: 3,
		Edges:      []*gmodel.EarningsEdge{},
		Nodes:      []*gmodel.Earning{},
		PageInfo: &gmodel.PageInfo{
			HasNextPage:     true,
			HasPreviousPage: false,
		},
	}, paginatedEarnings)
}

func (r *RewardsRepoTestSuite) Test_PaginateVehicleEarningsByID_NoRows() {
	rwrd, err := r.repo.GetEarningsByVehicleID(r.ctx, 11)
	r.NoError(err)

	last := 2
	before := "kwMAxCoweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA="
	paginatedEarnings, err := r.repo.PaginateVehicleEarningsByID(r.ctx, rwrd, nil, nil, &last, &before)
	r.NoError(err)

	r.Equal(&gmodel.EarningsConnection{
		TotalCount: 0,
		Edges:      []*gmodel.EarningsEdge{},
		Nodes:      []*gmodel.Earning{},
		PageInfo: &gmodel.PageInfo{
			HasNextPage:     true,
			HasPreviousPage: false,
		},
	}, paginatedEarnings)
}

func (r *RewardsRepoTestSuite) Test_PaginateAfterMarketDeviceEarningsByID_Disallow_FirstAndLast() {

	first := 1
	last := 2
	_, err := r.repo.PaginateAftermarketDeviceEarningsByID(r.ctx, &gmodel.AftermarketDeviceEarnings{}, &first, nil, &last, nil)
	r.EqualError(err, "pass `first` or `last`, but not both")
}

func (r *RewardsRepoTestSuite) Test_GetEarningsByAfterMarketDevice_FwdPagination_First() {
	_, beneficiary, err := helpers.GenerateWallet()
	r.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	r.createDependentRecords()

	totalEarned := big.NewInt(0)

	aft := models.AftermarketDevice{
		ID:          111,
		Address:     beneficiary.Bytes(),
		Beneficiary: beneficiary.Bytes(),
		Owner:       beneficiary.Bytes(),
	}
	err = aft.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
	r.NoError(err)

	rw, err := r.createRewardsRecords(2, createRewardsRecordsInput{
		beneficiary:         *beneficiary,
		dateTime:            currTime,
		afterMarketDeviceID: 111,
	})
	r.NoError(err)

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

	first := 2

	rwrd, err := r.repo.GetEarningsByAfterMarketDeviceID(r.ctx, 111)
	r.NoError(err)

	paginatedEarnings, err := r.repo.PaginateAftermarketDeviceEarningsByID(r.ctx, rwrd, &first, nil, nil, nil)
	r.NoError(err)

	startCursor, err := r.paginationHelper.EncodeCursor(RewardsCursor{Week: 2})
	r.NoError(err)

	endCursor, err := r.paginationHelper.EncodeCursor(RewardsCursor{Week: 1})
	r.NoError(err)

	syntID := 1
	aftID := 111

	r.Equal(&gmodel.PageInfo{
		EndCursor:       &endCursor,
		HasNextPage:     false,
		HasPreviousPage: false,
		StartCursor:     &startCursor,
	}, paginatedEarnings.PageInfo)
	r.Equal(2, paginatedEarnings.TotalCount)
	r.Equal([]*gmodel.EarningsEdge{
		{
			Node: &gmodel.Earning{
				Week:                    2,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        rw[1].ConnectionStreak.Ptr(),
				StreakTokens:            dbtypes.NullDecimalToInt(strkEarn),
				AftermarketDeviceID:     &aftID,
				AftermarketDeviceTokens: dbtypes.NullDecimalToInt(aftEarn),
				SyntheticDeviceID:       &syntID,
				SyntheticDeviceTokens:   dbtypes.NullDecimalToInt(syntEarn),
				SentAt:                  currTime,
				VehicleID:               11,
			},
			Cursor: startCursor,
		},
		{
			Node: &gmodel.Earning{
				Week:                    1,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        rw[0].ConnectionStreak.Ptr(),
				StreakTokens:            dbtypes.NullDecimalToInt(strkEarn),
				AftermarketDeviceID:     &aftID,
				AftermarketDeviceTokens: dbtypes.NullDecimalToInt(aftEarn),
				SyntheticDeviceID:       &syntID,
				SyntheticDeviceTokens:   dbtypes.NullDecimalToInt(syntEarn),
				SentAt:                  currTime,
				VehicleID:               11,
			},
			Cursor: endCursor,
		},
	}, paginatedEarnings.Edges)
}

func (r *RewardsRepoTestSuite) Test_GetEarningsByAfterMarketDevice_FwdPagination_First_After() {
	_, beneficiary, err := helpers.GenerateWallet()
	r.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	r.createDependentRecords()

	totalEarned := big.NewInt(0)

	aft := models.AftermarketDevice{
		ID:          111,
		Address:     beneficiary.Bytes(),
		Beneficiary: beneficiary.Bytes(),
		Owner:       beneficiary.Bytes(),
	}
	err = aft.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
	r.NoError(err)

	rw, err := r.createRewardsRecords(3, createRewardsRecordsInput{
		beneficiary:         *beneficiary,
		dateTime:            currTime,
		afterMarketDeviceID: 111,
	})
	r.NoError(err)

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

	first := 2
	after := "kwMAxCoweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA="
	aftID := 111

	rwrd, err := r.repo.GetEarningsByAfterMarketDeviceID(r.ctx, aftID)
	r.NoError(err)

	paginatedEarnings, err := r.repo.PaginateAftermarketDeviceEarningsByID(r.ctx, rwrd, &first, &after, nil, nil)
	r.NoError(err)

	startCrsr, err := r.paginationHelper.EncodeCursor(RewardsCursor{Week: 2})
	r.NoError(err)

	endCrsr, err := r.paginationHelper.EncodeCursor(RewardsCursor{Week: 1})
	r.NoError(err)

	syntID := 1
	connStrk := [2]int{22, 21}

	r.Equal(&gmodel.PageInfo{
		EndCursor:       &endCrsr,
		HasNextPage:     false,
		HasPreviousPage: true,
		StartCursor:     &startCrsr,
	}, paginatedEarnings.PageInfo)
	r.Equal(3, paginatedEarnings.TotalCount)
	r.Equal([]*gmodel.EarningsEdge{
		{
			Node: &gmodel.Earning{
				Week:                    2,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        &connStrk[0],
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
			Node: &gmodel.Earning{
				Week:                    1,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        &connStrk[1],
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

func (r *RewardsRepoTestSuite) Test_GetEarningsByAfterMarketDevice_BackPagination_Last() {
	_, beneficiary, err := helpers.GenerateWallet()
	r.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	r.createDependentRecords()

	totalEarned := big.NewInt(0)

	aft := models.AftermarketDevice{
		ID:          111,
		Address:     beneficiary.Bytes(),
		Beneficiary: beneficiary.Bytes(),
		Owner:       beneficiary.Bytes(),
	}
	err = aft.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
	r.NoError(err)

	rw, err := r.createRewardsRecords(2, createRewardsRecordsInput{
		beneficiary:         *beneficiary,
		dateTime:            currTime,
		afterMarketDeviceID: 111,
	})
	r.NoError(err)

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

	last := 1

	rwrd, err := r.repo.GetEarningsByAfterMarketDeviceID(r.ctx, 111)
	r.NoError(err)

	paginatedEarnings, err := r.repo.PaginateAftermarketDeviceEarningsByID(r.ctx, rwrd, nil, nil, &last, nil)
	r.NoError(err)

	crsr, err := r.paginationHelper.EncodeCursor(RewardsCursor{Week: 1})
	r.NoError(err)

	r.NoError(err)
	aftID := 111
	syntID := 1

	connStrk := 21

	r.Equal(&gmodel.PageInfo{
		EndCursor:       &crsr,
		HasNextPage:     false,
		HasPreviousPage: true,
		StartCursor:     &crsr,
	}, paginatedEarnings.PageInfo)
	r.Equal(2, paginatedEarnings.TotalCount)
	r.Equal([]*gmodel.EarningsEdge{
		{
			Node: &gmodel.Earning{
				Week:                    1,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        &connStrk,
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

func (r *RewardsRepoTestSuite) Test_GetEarningsByAfterMarketDevice_BackPagination_Last_Before() {
	_, beneficiary, err := helpers.GenerateWallet()
	r.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	r.createDependentRecords()

	totalEarned := big.NewInt(0)

	aft := models.AftermarketDevice{
		ID:          111,
		Address:     beneficiary.Bytes(),
		Beneficiary: beneficiary.Bytes(),
		Owner:       beneficiary.Bytes(),
	}
	err = aft.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
	r.NoError(err)

	rw, err := r.createRewardsRecords(3, createRewardsRecordsInput{
		beneficiary:         *beneficiary,
		dateTime:            currTime,
		afterMarketDeviceID: 111,
	})
	r.NoError(err)

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

	last := 2
	before := "kwEAxCoweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA="

	rwrd, err := r.repo.GetEarningsByAfterMarketDeviceID(r.ctx, 111)
	r.NoError(err)

	paginatedEarnings, err := r.repo.PaginateAftermarketDeviceEarningsByID(r.ctx, rwrd, nil, nil, &last, &before)
	r.NoError(err)

	startCrsr, err := r.paginationHelper.EncodeCursor(RewardsCursor{Week: 3})
	r.NoError(err)

	endCrsr, err := r.paginationHelper.EncodeCursor(RewardsCursor{Week: 2})
	r.NoError(err)

	aftID := 111
	syntID := 1

	connStrk := [2]int{22, 23}

	r.Equal(&gmodel.PageInfo{
		EndCursor:       &endCrsr,
		HasNextPage:     true,
		HasPreviousPage: false,
		StartCursor:     &startCrsr,
	}, paginatedEarnings.PageInfo)
	r.Equal(3, paginatedEarnings.TotalCount)
	r.Equal([]*gmodel.EarningsEdge{
		{
			Node: &gmodel.Earning{
				Week:                    3,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        &connStrk[1],
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
			Node: &gmodel.Earning{
				Week:                    2,
				Beneficiary:             common.BytesToAddress(beneficiary.Bytes()),
				ConnectionStreak:        &connStrk[0],
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

/* func (r *RewardsRepoTestSuite) Test_GenCursors() {
	for i := 1; i < 4; i++ {
		log.Println(r.paginationHelper.EncodeCursor(RewardsCursor{Week: i}))
	}
}
2023/11/15 00:20:16 kwEAxCoweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA= <nil>
2023/11/15 00:20:16 kwIAxCoweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA= <nil>
2023/11/15 00:20:16 kwMAxCoweDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA= <nil> */

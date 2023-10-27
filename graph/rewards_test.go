package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DIMO-Network/identity-api/internal/config"
	test "github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/dbtypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type RewardsQueryTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container testcontainers.Container
	settings  config.Settings
	resolver  *Resolver
	repo      *repositories.Repository
}

func (r *RewardsQueryTestSuite) SetupSuite() {
	r.ctx = context.Background()
	r.pdb, r.container = test.StartContainerDatabase(r.ctx, r.T(), migrationsDir)

	r.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
	}
	r.repo = repositories.New(r.pdb)
	r.resolver = NewResolver(r.repo)
}

// TearDownTest after each test truncate tables
func (r *RewardsQueryTestSuite) TearDownTest() {
	test.TruncateTables(r.pdb.DBS().Writer.DB, r.T())
}

// TearDownSuite cleanup at end by terminating container
func (r *RewardsQueryTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", r.container.SessionID())

	if err := r.container.Terminate(r.ctx); err != nil {
		r.T().Fatal(err)
	}
}

// Test Runner
func TestRewardsQueryTestSuite(t *testing.T) {
	suite.Run(t, new(RewardsQueryTestSuite))
}

func (r *RewardsQueryTestSuite) Test_Query_GetEarningsByVehicle() {
	currTime := time.Now().UTC().Truncate(time.Second)
	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	var vehicle = models.Vehicle{
		ID:           11,
		OwnerAddress: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Make:         null.StringFrom("Ford"),
		Model:        null.StringFrom("Bronco"),
		Year:         null.IntFrom(2022),
		MintedAt:     time.Now(),
	}

	var aftermarketDevice = models.AftermarketDevice{
		ID:          1,
		Address:     common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf5").Bytes(),
		Owner:       common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes(),
		Serial:      null.StringFrom("aftermarketDeviceSerial-1"),
		Imei:        null.StringFrom("aftermarketDeviceIMEI-1"),
		MintedAt:    time.Now(),
		VehicleID:   null.IntFrom(11),
		Beneficiary: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	}

	var syntheticDevice = models.SyntheticDevice{
		ID:            1,
		IntegrationID: 2,
		VehicleID:     11,
		DeviceAddress: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt:      time.Now(),
	}

	// Aftermarket Earnings
	adEarn, ok := new(big.Int).SetString("59147051345528509681", 10)
	r.NotZero(ok)

	// Synthetic Earnings
	synthEarn, ok := new(big.Int).SetString("59147051345528509682", 10)
	r.NotZero(ok)

	// Streak Earnings
	strkEarn, ok := new(big.Int).SetString("59147051345528509684", 10)
	r.NotZero(ok)

	var reward = models.Reward{
		IssuanceWeek:        1,
		VehicleID:           11,
		ConnectionStreak:    null.IntFrom(20),
		StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
		AftermarketTokenID:  null.IntFrom(1),
		AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
		SyntheticTokenID:    null.IntFrom(1),
		SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
		ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
		EarnedAt:            currTime,
	}

	err = vehicle.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
	r.NoError(err)

	err = aftermarketDevice.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
	r.NoError(err)

	err = syntheticDevice.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
	r.NoError(err)

	err = reward.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
	r.NoError(err)

	query := `{vehicle(tokenId: 11) {id earnings {earnedTokens earningsTransfers {node {week beneficiary connectionStreak streakTokens aftermarketDevice {id} aftermarketDeviceTokens syntheticDevice {tokenId} syntheticDeviceTokens sentAt}}}}}`

	c := client.New(loader.Middleware(r.pdb, handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: r.resolver}))))

	var resp interface{}
	c.MustPost(query, &resp)
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))

	r.JSONEq(fmt.Sprintf(`{"vehicle":{"earnings":{"earnedTokens":"177441154036585529047","earningsTransfers":[{"node":{"aftermarketDevice":null,"aftermarketDeviceTokens":"59147051345528509681","beneficiary":"%s","connectionStreak":20,"sentAt":"%s","streakTokens":"59147051345528509684","syntheticDevice":null,"syntheticDeviceTokens":"59147051345528509682","week":1}}]},"id":"V_kQs="}}`, beneficiary.Hex(), currTime.Format(time.RFC3339)), string(b))
}

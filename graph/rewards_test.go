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
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/dbtypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
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
	repo      *base.Repository
}

func (r *RewardsQueryTestSuite) SetupSuite() {
	r.ctx = context.Background()
	r.pdb, r.container = test.StartContainerDatabase(r.ctx, r.T(), migrationsDir)

	r.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
	}
	logger := zerolog.Nop()
	r.repo = base.NewRepository(r.pdb, r.settings, &logger)
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

func (r *RewardsQueryTestSuite) createDependencies() {
	var vehicle = models.Vehicle{
		ID:           1,
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
		VehicleID:   null.IntFrom(1),
		Beneficiary: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	}

	var syntheticDevice = models.SyntheticDevice{
		ID:            1,
		IntegrationID: 2,
		VehicleID:     1,
		DeviceAddress: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt:      time.Now(),
	}

	err := vehicle.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
	r.NoError(err)

	err = aftermarketDevice.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
	r.NoError(err)

	err = syntheticDevice.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
	r.NoError(err)
}

func (r *RewardsQueryTestSuite) Test_Query_GetEarningsByVehicle_FwdPaginate() {
	currTime := time.Now().UTC().Truncate(time.Second)
	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	r.createDependencies()

	// Aftermarket Earnings
	adEarn, ok := new(big.Int).SetString("59147051345528509681", 10)
	r.True(ok)

	// Synthetic Earnings
	synthEarn, ok := new(big.Int).SetString("59147051345528509682", 10)
	r.True(ok)

	// Streak Earnings
	strkEarn, ok := new(big.Int).SetString("59147051345528509684", 10)
	r.True(ok)

	var reward = models.Reward{
		IssuanceWeek:        2,
		VehicleID:           1,
		ConnectionStreak:    null.IntFrom(20),
		StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
		AftermarketTokenID:  null.IntFrom(1),
		AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
		SyntheticTokenID:    null.IntFrom(1),
		SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
		ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
		EarnedAt:            currTime,
	}

	err = reward.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
	r.NoError(err)

	query := `{
		vehicle(tokenId: 1) {
		  id
		  earnings {
			totalTokens
			history(first: 3) {
			  totalCount
			  edges {
				cursor
				node {
				  week
				  beneficiary
				  connectionStreak
				  streakTokens
				  aftermarketDevice {
					id
					tokenId
				  }
				  aftermarketDeviceTokens
				  syntheticDevice {
					tokenId
					integrationId
				  }
				  vehicle {
					id
					tokenId
				  }
				  syntheticDeviceTokens
				  sentAt
				}
			  }
			  nodes {
				week
				beneficiary
				connectionStreak
			  }
			  pageInfo {
				endCursor
				hasNextPage
				hasPreviousPage
				startCursor
			  }
			}
		  }
		}
	  }`

	c := client.New(
		loader.Middleware(
			r.pdb,
			handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: r.resolver})), r.settings,
		),
	)

	var resp interface{}
	c.MustPost(query, &resp)
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))

	r.JSONEq(fmt.Sprintf(`
	{
		"vehicle": {
		  "id": "V_kQE=",
		  "earnings": {
			"totalTokens": "177441154036585529047",
			"history": {
			  "totalCount": 1,
			  "edges": [
				{
				  "cursor": "kgIB",
				  "node": {
					"week": 2,
					"beneficiary": "%s",
					"connectionStreak": 20,
					"streakTokens": "59147051345528509684",
					"aftermarketDevice": {
					  "id": "AD_kQE=",
					  "tokenId": 1
					},
					"aftermarketDeviceTokens": "59147051345528509681",
					"syntheticDevice": {
					  "tokenId": 1,
					  "integrationId": 2
					},
					"vehicle": {
					  "id": "V_kQE=",
					  "tokenId": 1
					},
					"syntheticDeviceTokens": "59147051345528509682",
					"sentAt": "%s"
				  }
				}
			  ],
			  "nodes": [
				{
				  "week": 2,
				  "beneficiary": "%s",
				  "connectionStreak": 20
				}
			  ],
			  "pageInfo": {
				"endCursor": "kgIB",
				"hasNextPage": false,
				"hasPreviousPage": false,
				"startCursor": "kgIB"
			  }
			}
		  }
		}
	}
	`, beneficiary.Hex(), currTime.Format(time.RFC3339), beneficiary.Hex()), string(b))
}

func (r *RewardsQueryTestSuite) Test_Query_GetEarningsByVehicle_FwdPaginate_FirstAfter() {
	currTime := time.Now().UTC().Truncate(time.Second)
	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	r.createDependencies()

	// Aftermarket Earnings
	adEarn, ok := new(big.Int).SetString("59147051345528509681", 10)
	r.True(ok)

	// Synthetic Earnings
	synthEarn, ok := new(big.Int).SetString("59147051345528509682", 10)
	r.True(ok)

	// Streak Earnings
	strkEarn, ok := new(big.Int).SetString("59147051345528509684", 10)
	r.True(ok)

	var rewards = []models.Reward{
		{
			IssuanceWeek:        2,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(12),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
		{
			IssuanceWeek:        3,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(13),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
		{
			IssuanceWeek:        4,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(14),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
	}

	for _, rwd := range rewards {
		err = rwd.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	query := `{
		vehicle(tokenId: 1) {
		  id
		  earnings {
			totalTokens
			history(first: 2, after: "kgQB") {
			  totalCount
			  edges {
				cursor
				node {
				  week
				  beneficiary
				  connectionStreak
				  streakTokens
				  aftermarketDevice {
					id
					tokenId
				  }
				  aftermarketDeviceTokens
				  syntheticDevice {
					tokenId
					integrationId
				  }
				  vehicle {
					id
					tokenId
				  }
				  syntheticDeviceTokens
				  sentAt
				}
			  }
			  nodes {
				week
				beneficiary
				connectionStreak
			  }
			  pageInfo {
				endCursor
				hasNextPage
				hasPreviousPage
				startCursor
			  }
			}
		  }
		}
	  }`

	c := client.New(
		loader.Middleware(
			r.pdb,
			handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: r.resolver})), r.settings,
		),
	)

	var resp interface{}
	c.MustPost(query, &resp)
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))

	r.JSONEq(fmt.Sprintf(`
	{	
		"vehicle": {
			"id": "V_kQE=",
			"earnings": {
			"totalTokens": "532323462109756587141",
			"history": {
				"totalCount": 3,
				"edges": [
					{
						"cursor": "kgMB",
						"node": {
						"week": 3,
						"beneficiary": "%s",
						"connectionStreak": 13,
						"streakTokens": "59147051345528509684",
						"aftermarketDevice": {
							"id": "AD_kQE=",
							"tokenId": 1
						},
						"aftermarketDeviceTokens": "59147051345528509681",
						"syntheticDevice": {
							"tokenId": 1,
							"integrationId": 2
						},
						"vehicle": {
							"id": "V_kQE=",
							"tokenId": 1
						},
						"syntheticDeviceTokens": "59147051345528509682",
						"sentAt": "%s"
						}
					},
					{
						"cursor": "kgIB",
						"node": {
						"week": 2,
						"beneficiary": "%s",
						"connectionStreak": 12,
						"streakTokens": "59147051345528509684",
						"aftermarketDevice": {
							"id": "AD_kQE=",
							"tokenId": 1
						},
						"aftermarketDeviceTokens": "59147051345528509681",
						"syntheticDevice": {
							"tokenId": 1,
							"integrationId": 2
						},
						"vehicle": {
							"id": "V_kQE=",
							"tokenId": 1
						},
						"syntheticDeviceTokens": "59147051345528509682",
						"sentAt": "%s"
						}
					}
				],
				"nodes": [
				{
					"week": 3,
					"beneficiary": "%s",
					"connectionStreak": 13
				},
				{
					"week": 2,
					"beneficiary": "%s",
					"connectionStreak": 12
				}
				],
				"pageInfo": {
				"endCursor": "kgIB",
				"hasNextPage": false,
				"hasPreviousPage": true,
				"startCursor": "kgMB"
				}
			}
			}
		}
	  }
	`, beneficiary.Hex(), currTime.Format(time.RFC3339), beneficiary.Hex(), currTime.Format(time.RFC3339), beneficiary.Hex(), beneficiary.Hex()), string(b))
}

func (r *RewardsQueryTestSuite) Test_Query_GetEarningsByVehicle_BackPaginate_Last() {
	currTime := time.Now().UTC().Truncate(time.Second)
	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	r.createDependencies()

	// Aftermarket Earnings
	adEarn, ok := new(big.Int).SetString("59147051345528509681", 10)
	r.True(ok)

	// Synthetic Earnings
	synthEarn, ok := new(big.Int).SetString("59147051345528509682", 10)
	r.True(ok)

	// Streak Earnings
	strkEarn, ok := new(big.Int).SetString("59147051345528509684", 10)
	r.True(ok)

	var rewards = []models.Reward{
		{
			IssuanceWeek:        2,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(12),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
		{
			IssuanceWeek:        3,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(13),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
		{
			IssuanceWeek:        4,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(14),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
	}

	for _, rwd := range rewards {
		err = rwd.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	query := `{
		vehicle(tokenId: 1) {
		  id
		  earnings {
			totalTokens
			history(last: 2) {
			  totalCount
			  edges {
				cursor
				node {
				  week
				  beneficiary
				  connectionStreak
				  streakTokens
				  aftermarketDevice {
					id
					tokenId
				  }
				  aftermarketDeviceTokens
				  syntheticDevice {
					tokenId
					integrationId
				  }
				  vehicle {
					id
					tokenId
				  }
				  syntheticDeviceTokens
				  sentAt
				}
			  }
			  nodes {
				week
				beneficiary
				connectionStreak
			  }
			  pageInfo {
				endCursor
				hasNextPage
				hasPreviousPage
				startCursor
			  }
			}
		  }
		}
	  }`

	c := client.New(
		loader.Middleware(
			r.pdb,
			handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: r.resolver})), r.settings,
		),
	)

	var resp interface{}
	c.MustPost(query, &resp)
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))

	r.JSONEq(fmt.Sprintf(`
	{	
		"vehicle": {
			"id": "V_kQE=",
			"earnings": {
			"totalTokens": "532323462109756587141",
			"history": {
				"totalCount": 3,
				"edges": [
					{
						"cursor": "kgMB",
						"node": {
						"week": 3,
						"beneficiary": "%s",
						"connectionStreak": 13,
						"streakTokens": "59147051345528509684",
						"aftermarketDevice": {
							"id": "AD_kQE=",
							"tokenId": 1
						},
						"aftermarketDeviceTokens": "59147051345528509681",
						"syntheticDevice": {
							"tokenId": 1,
							"integrationId": 2
						},
						"vehicle": {
							"id": "V_kQE=",
							"tokenId": 1
						},
						"syntheticDeviceTokens": "59147051345528509682",
						"sentAt": "%s"
						}
					},
					{
						"cursor": "kgIB",
						"node": {
						"week": 2,
						"beneficiary": "%s",
						"connectionStreak": 12,
						"streakTokens": "59147051345528509684",
						"aftermarketDevice": {
							"id": "AD_kQE=",
							"tokenId": 1
						},
						"aftermarketDeviceTokens": "59147051345528509681",
						"syntheticDevice": {
							"tokenId": 1,
							"integrationId": 2
						},
						"vehicle": {
							"id": "V_kQE=",
							"tokenId": 1
						},
						"syntheticDeviceTokens": "59147051345528509682",
						"sentAt": "%s"
						}
					}
				],
				"nodes": [
					{
						"week": 3,
						"beneficiary": "%s",
						"connectionStreak": 13
					},
					{
						"week": 2,
						"beneficiary": "%s",
						"connectionStreak": 12
					}
				],
				"pageInfo": {
				"endCursor": "kgIB",
				"hasNextPage": false,
				"hasPreviousPage": true,
				"startCursor": "kgMB"
				}
			}
			}
		}
	  }
	`, beneficiary.Hex(), currTime.Format(time.RFC3339), beneficiary.Hex(), currTime.Format(time.RFC3339), beneficiary.Hex(), beneficiary.Hex()), string(b))
}

func (r *RewardsQueryTestSuite) Test_Query_GetEarningsByVehicle_BackPaginate_LastBefore() {
	currTime := time.Now().UTC().Truncate(time.Second)
	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	r.createDependencies()

	// Aftermarket Earnings
	adEarn, ok := new(big.Int).SetString("59147051345528509681", 10)
	r.True(ok)

	// Synthetic Earnings
	synthEarn, ok := new(big.Int).SetString("59147051345528509682", 10)
	r.True(ok)

	// Streak Earnings
	strkEarn, ok := new(big.Int).SetString("59147051345528509684", 10)
	r.True(ok)

	var rewards = []models.Reward{
		{
			IssuanceWeek:        2,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(12),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
		{
			IssuanceWeek:        3,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(13),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
		{
			IssuanceWeek:        4,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(14),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
	}

	for _, rwd := range rewards {
		err = rwd.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	query := `{
		vehicle(tokenId: 1) {
		  id
		  earnings {
			totalTokens
			history(last: 2, before: "kgIL") {
			  totalCount
			  edges {
				cursor
				node {
				  week
				  beneficiary
				  connectionStreak
				  streakTokens
				  aftermarketDevice {
					id
					tokenId
				  }
				  aftermarketDeviceTokens
				  syntheticDevice {
					tokenId
					integrationId
				  }
				  vehicle {
					id
					tokenId
				  }
				  syntheticDeviceTokens
				  sentAt
				}
			  }
			  nodes {
				week
				beneficiary
				connectionStreak
			  }
			  pageInfo {
				endCursor
				hasNextPage
				hasPreviousPage
				startCursor
			  }
			}
		  }
		}
	  }`

	c := client.New(
		loader.Middleware(
			r.pdb,
			handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: r.resolver})), r.settings,
		),
	)

	var resp interface{}
	c.MustPost(query, &resp)
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))

	r.JSONEq(fmt.Sprintf(`
	{	
		"vehicle": {
			"id": "V_kQE=",
			"earnings": {
			"totalTokens": "532323462109756587141",
			"history": {
				"totalCount": 3,
				"edges": [
					{
						"cursor": "kgQB",
						"node": {
						"week": 4,
						"beneficiary": "%s",
						"connectionStreak": 14,
						"streakTokens": "59147051345528509684",
						"aftermarketDevice": {
							"id": "AD_kQE=",
							"tokenId": 1
						},
						"aftermarketDeviceTokens": "59147051345528509681",
						"syntheticDevice": {
							"tokenId": 1,
							"integrationId": 2
						},
						"vehicle": {
							"id": "V_kQE=",
							"tokenId": 1
						},
						"syntheticDeviceTokens": "59147051345528509682",
						"sentAt": "%s"
						}
					},
					{
						"cursor": "kgMB",
						"node": {
						"week": 3,
						"beneficiary": "%s",
						"connectionStreak": 13,
						"streakTokens": "59147051345528509684",
						"aftermarketDevice": {
							"id": "AD_kQE=",
							"tokenId": 1
						},
						"aftermarketDeviceTokens": "59147051345528509681",
						"syntheticDevice": {
							"tokenId": 1,
							"integrationId": 2
						},
						"vehicle": {
							"id": "V_kQE=",
							"tokenId": 1
						},
						"syntheticDeviceTokens": "59147051345528509682",
						"sentAt": "%s"
						}
					}
				],
				"nodes": [
					{
						"week": 4,
						"beneficiary": "%s",
						"connectionStreak": 14
					},
					{
						"week": 3,
						"beneficiary": "%s",
						"connectionStreak": 13
					}
				],
				"pageInfo": {
				"endCursor": "kgMB",
				"hasNextPage": true,
				"hasPreviousPage": false,
				"startCursor": "kgQB"
				}
			}
			}
		}
	  }
	`, beneficiary.Hex(), currTime.Format(time.RFC3339), beneficiary.Hex(), currTime.Format(time.RFC3339), beneficiary.Hex(), beneficiary.Hex()), string(b))
}

func (r *RewardsQueryTestSuite) Test_Query_GetAftermarketDeviceEarnings_FwdPaginate() {
	currTime := time.Now().UTC().Truncate(time.Second)
	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	r.createDependencies()

	// Aftermarket Earnings
	adEarn, ok := new(big.Int).SetString("59147051345528509681", 10)
	r.True(ok)

	// Synthetic Earnings
	synthEarn, ok := new(big.Int).SetString("59147051345528509682", 10)
	r.True(ok)

	// Streak Earnings
	strkEarn, ok := new(big.Int).SetString("59147051345528509684", 10)
	r.True(ok)

	var reward = models.Reward{
		IssuanceWeek:        2,
		VehicleID:           1,
		ConnectionStreak:    null.IntFrom(20),
		StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
		AftermarketTokenID:  null.IntFrom(1),
		AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
		SyntheticTokenID:    null.IntFrom(1),
		SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
		ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
		EarnedAt:            currTime,
	}

	err = reward.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
	r.NoError(err)

	query := `{
		aftermarketDevices(first: 2) {
			totalCount
			edges {
			  cursor
			  node {
				id
				tokenId
				earnings {
				  totalTokens
				  history(first: 1) {
					totalCount
					pageInfo {
					  startCursor
					  endCursor
					  hasPreviousPage
					  hasNextPage
					}
					edges {
					  node {
						week
						beneficiary
						connectionStreak
					  }
					  cursor
					}
				  }
				}
			  }
			}
		  }
	  }`

	c := client.New(
		loader.Middleware(
			r.pdb,
			handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: r.resolver})), r.settings,
		),
	)

	var resp interface{}
	c.MustPost(query, &resp)
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))

	r.JSONEq(fmt.Sprintf(`
	{
		"aftermarketDevices": {
			"totalCount": 1,
			"edges": [
				{
					"cursor": "MQ==",
					"node": {
						"id": "AD_kQE=",
						"tokenId": 1,
						"earnings": {
							"totalTokens": "177441154036585529047",
							"history": {
								"totalCount": 1,
								"pageInfo": {
									"endCursor": "kgIB",
									"hasNextPage": false,
									"hasPreviousPage": false,
									"startCursor": "kgIB"
								},
								"edges": [
									{
										"node": {
										  "week": 2,
										  "beneficiary": "%s",
										  "connectionStreak": 20
										},
										"cursor": "kgIB"
									  }
								]
							}
						}
					}
				}
			]
		}
	}
	`, beneficiary.Hex()), string(b))
}

func (r *RewardsQueryTestSuite) Test_Query_GetUserRewards_FwdPaginate() {
	currTime := time.Now().UTC().Truncate(time.Second)
	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	r.createDependencies()

	// Aftermarket Earnings
	adEarn, ok := new(big.Int).SetString("59147051345528509681", 10)
	r.True(ok)

	// Synthetic Earnings
	synthEarn, ok := new(big.Int).SetString("59147051345528509682", 10)
	r.True(ok)

	// Streak Earnings
	strkEarn, ok := new(big.Int).SetString("59147051345528509684", 10)
	r.True(ok)

	var rewards = []models.Reward{
		{
			IssuanceWeek:        2,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(12),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
		{
			IssuanceWeek:        3,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(13),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
		{
			IssuanceWeek:        4,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(14),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
	}

	for _, rwd := range rewards {
		err = rwd.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	query := fmt.Sprintf(`{
		rewards(user: "%s") {
			totalTokens
			history(first: 2) {
			  totalCount
			  edges {
				node {
				  week
				  beneficiary
				  connectionStreak
				  streakTokens
				  aftermarketDevice {
					id
				  }
				  aftermarketDeviceTokens
				  syntheticDevice {
					tokenId
				  }
				  syntheticDeviceTokens
				  vehicle {
					id
					tokenId
				  }
				}
				cursor
			  }
			  pageInfo {
				startCursor
				endCursor
				hasPreviousPage
				hasNextPage
			  }
			}
		}
	  }`, beneficiary.Hex())

	c := client.New(
		loader.Middleware(
			r.pdb,
			handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: r.resolver})), r.settings,
		),
	)

	var resp interface{}
	c.MustPost(query, &resp)
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))

	r.JSONEq(fmt.Sprintf(`
	{
		"rewards": {
			"totalTokens": "532323462109756587141",
			"history": {
			  "totalCount": 3,
			  "edges": [
				{
				  "node": {
					"week": 4,
					"beneficiary": "%s",
					"connectionStreak": 14,
					"streakTokens": "59147051345528509684",
					"aftermarketDevice": {
					  "id": "AD_kQE="
					},
					"aftermarketDeviceTokens": "59147051345528509681",
					"syntheticDevice": {
					  "tokenId": 1
					},
					"syntheticDeviceTokens": "59147051345528509682",
					"vehicle": {
					  "id": "V_kQE=",
					  "tokenId": 1
					}
				  },
				  "cursor": "kgQB"
				},
				{
				  "node": {
					"week": 3,
					"beneficiary": "%s",
					"connectionStreak": 13,
					"streakTokens": "59147051345528509684",
					"aftermarketDevice": {
					  "id": "AD_kQE="
					},
					"aftermarketDeviceTokens": "59147051345528509681",
					"syntheticDevice": {
					  "tokenId": 1
					},
					"syntheticDeviceTokens": "59147051345528509682",
					"vehicle": {
					  "id": "V_kQE=",
					  "tokenId": 1
					}
				  },
				  "cursor": "kgMB"
				}
			  ],
			  "pageInfo": {
				"startCursor": "kgQB",
				"endCursor": "kgMB",
				"hasPreviousPage": false,
				"hasNextPage": true
			  }
			}
		}
	}
	`, beneficiary.Hex(), beneficiary.Hex()), string(b))
}

func (r *RewardsQueryTestSuite) Test_Query_GetUserRewards_BackPaginate_LastBefore() {
	currTime := time.Now().UTC().Truncate(time.Second)
	_, beneficiary, err := test.GenerateWallet()
	r.NoError(err)

	r.createDependencies()

	// Aftermarket Earnings
	adEarn, ok := new(big.Int).SetString("59147051345528509681", 10)
	r.True(ok)

	// Synthetic Earnings
	synthEarn, ok := new(big.Int).SetString("59147051345528509682", 10)
	r.True(ok)

	// Streak Earnings
	strkEarn, ok := new(big.Int).SetString("59147051345528509684", 10)
	r.True(ok)

	var rewards = []models.Reward{
		{
			IssuanceWeek:        2,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(12),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
		{
			IssuanceWeek:        3,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(13),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
		{
			IssuanceWeek:        4,
			VehicleID:           1,
			ConnectionStreak:    null.IntFrom(14),
			StreakEarnings:      dbtypes.NullIntToDecimal(strkEarn),
			AftermarketTokenID:  null.IntFrom(1),
			AftermarketEarnings: dbtypes.NullIntToDecimal(adEarn),
			SyntheticTokenID:    null.IntFrom(1),
			SyntheticEarnings:   dbtypes.NullIntToDecimal(synthEarn),
			ReceivedByAddress:   null.BytesFrom(beneficiary.Bytes()),
			EarnedAt:            currTime,
		},
	}

	for _, rwd := range rewards {
		err = rwd.Insert(r.ctx, r.pdb.DBS().Writer, boil.Infer())
		r.NoError(err)
	}

	query := fmt.Sprintf(`{
		rewards(user: "%s") {
			totalTokens
			history(last: 2) {
			  totalCount
			  edges {
				node {
				  week
				  beneficiary
				  connectionStreak
				  streakTokens
				  aftermarketDevice {
					id
				  }
				  aftermarketDeviceTokens
				  syntheticDevice {
					tokenId
				  }
				  syntheticDeviceTokens
				  vehicle {
					id
					tokenId
				  }
				}
				cursor
			  }
			  pageInfo {
				startCursor
				endCursor
				hasPreviousPage
				hasNextPage
			  }
			}
		}
	  }`, beneficiary.Hex())

	c := client.New(
		loader.Middleware(
			r.pdb,
			handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: r.resolver})), r.settings,
		),
	)

	var resp interface{}
	c.MustPost(query, &resp)
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))

	r.JSONEq(fmt.Sprintf(`
	{
		"rewards": {
			"totalTokens": "532323462109756587141",
			"history": {
			  "totalCount": 3,
			  "edges": [
				{
				  "node": {
					"week": 3,
					"beneficiary": "%s",
					"connectionStreak": 13,
					"streakTokens": "59147051345528509684",
					"aftermarketDevice": {
					  "id": "AD_kQE="
					},
					"aftermarketDeviceTokens": "59147051345528509681",
					"syntheticDevice": {
					  "tokenId": 1
					},
					"syntheticDeviceTokens": "59147051345528509682",
					"vehicle": {
					  "id": "V_kQE=",
					  "tokenId": 1
					}
				  },
				  "cursor": "kgMB"
				},
				{
				  "node": {
					"week": 2,
					"beneficiary": "%s",
					"connectionStreak": 12,
					"streakTokens": "59147051345528509684",
					"aftermarketDevice": {
					  "id": "AD_kQE="
					},
					"aftermarketDeviceTokens": "59147051345528509681",
					"syntheticDevice": {
					  "tokenId": 1
					},
					"syntheticDeviceTokens": "59147051345528509682",
					"vehicle": {
					  "id": "V_kQE=",
					  "tokenId": 1
					}
				  },
				  "cursor": "kgIB"
				}
			  ],
			  "pageInfo": {
				"startCursor": "kgMB",
				"endCursor": "kgIB",
				"hasPreviousPage": true,
				"hasNextPage": false
			  }
			}
		}
	}
	`, beneficiary.Hex(), beneficiary.Hex()), string(b))
}

package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type OwnedVehiclesRepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container testcontainers.Container
	repo      *Repository
	settings  config.Settings
}

// const migrationsDir = "../../migrations"

func (s *OwnedVehiclesRepoTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = helpers.StartContainerDatabase(s.ctx, s.T(), migrationsDir)

	s.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
		BaseImageURL:        "https://mockUrl.com/",
	}
	s.repo = New(s.pdb, s.settings)
}

// TearDownTest after each test truncate tables
func (s *OwnedVehiclesRepoTestSuite) TearDownTest() {
	helpers.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *OwnedVehiclesRepoTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())

	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

// Test Runner
func TestOwnedVehiclesRepoTestSuite(t *testing.T) {
	suite.Run(t, new(OwnedVehiclesRepoTestSuite))
}

/* Actual Tests */
func (s *OwnedVehiclesRepoTestSuite) Test_GetOwnedVehicles_Success() {
	_, wallet, err := helpers.GenerateWallet()
	s.NoError(err)

	_, wallet2, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:           1,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Camry"),
			Year:         null.IntFrom(2020),
			MintedAt:     currTime,
		},
		{
			ID:           2,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Camry"),
			Year:         null.IntFrom(2022),
			MintedAt:     currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
			s.NoError(err)
		}
	}

	expiresAt := currTime.Add(time.Minute + 10).UTC().Truncate(time.Second)
	privileges := []models.Privilege{
		{
			TokenID:     1,
			PrivilegeID: 1,
			UserAddress: wallet2.Bytes(),
			SetAt:       currTime,
			ExpiresAt:   expiresAt,
		},
	}

	for _, p := range privileges {
		if err := p.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
			s.NoError(err)
		}
	}

	first := 3
	res, err := s.repo.GetVehicles(s.ctx, &first, nil, nil, nil, &gmodel.VehiclesFilter{Privileged: wallet})
	s.NoError(err)

	s.Equal(2, res.TotalCount)
	s.False(res.PageInfo.HasNextPage)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:      "V_kQI=",
				TokenID: 2,
				Owner:   common.BytesToAddress(wallet.Bytes()),
				Definition: &gmodel.Definition{
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
				MintedAt:   vehicles[1].MintedAt,
				Privileges: nil,
				Image:      "https://mockUrl.com/v1/vehicle/2",
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:      "V_kQE=",
				TokenID: 1,
				Owner:   common.BytesToAddress(wallet.Bytes()),
				Definition: &gmodel.Definition{
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
				MintedAt:   vehicles[0].MintedAt,
				Privileges: nil,
				Image:      "https://mockUrl.com/v1/vehicle/1",
			},
			Cursor: "MQ==",
		},
	}

	for _, veh := range expected {
		bid := helpers.IntToBytes(veh.Node.TokenID)
		veh.Node.Name, err = helpers.CreateMnemonic(bid)

		s.NoError(err)
	}

	s.Exactly(expected, res.Edges)
}

func (s *OwnedVehiclesRepoTestSuite) Test_GetOwnedVehicles_Pagination() {
	_, wallet, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:           1,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Camry"),
			Year:         null.IntFrom(2020),
			MintedAt:     currTime,
		},
		{
			ID:           2,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Camry"),
			Year:         null.IntFrom(2022),
			MintedAt:     currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
			s.NoError(err)
		}
	}

	first := 1
	res, err := s.repo.GetVehicles(s.ctx, &first, nil, nil, nil, &gmodel.VehiclesFilter{Privileged: wallet})
	s.NoError(err)

	s.Equal(len(vehicles), res.TotalCount)
	s.Equal(res.PageInfo.HasNextPage, true)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:      "V_kQI=",
				TokenID: 2,
				Owner:   common.BytesToAddress(wallet.Bytes()),
				Definition: &gmodel.Definition{
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
				MintedAt:   vehicles[1].MintedAt,
				Privileges: nil,
				Image:      "https://mockUrl.com/v1/vehicle/2",
			},
			Cursor: "Mg==",
		},
	}

	for _, veh := range expected {
		bid := helpers.IntToBytes(veh.Node.TokenID)
		veh.Node.Name, err = helpers.CreateMnemonic(bid)

		s.NoError(err)
	}

	s.Exactly(expected, res.Edges)
}

func (s *OwnedVehiclesRepoTestSuite) Test_GetOwnedVehicles_Pagination_NextPage() {
	_, wallet, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:           1,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Camry"),
			Year:         null.IntFrom(2020),
			MintedAt:     currTime,
		},
		{
			ID:           2,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Camry"),
			Year:         null.IntFrom(2022),
			MintedAt:     currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
			s.NoError(err)
		}
	}

	first := 1
	after := "Mg=="
	res, err := s.repo.GetVehicles(s.ctx, &first, &after, nil, nil, &gmodel.VehiclesFilter{Privileged: wallet})
	s.NoError(err)

	s.Len(vehicles, res.TotalCount)
	s.False(res.PageInfo.HasNextPage)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:      "V_kQE=",
				TokenID: 1,
				Owner:   common.BytesToAddress(wallet.Bytes()),
				Definition: &gmodel.Definition{
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
				MintedAt:   vehicles[0].MintedAt,
				Privileges: nil,
				Image:      "https://mockUrl.com/v1/vehicle/1",
			},
			Cursor: "MQ==",
		},
	}

	for _, veh := range expected {
		bid := helpers.IntToBytes(veh.Node.TokenID)
		veh.Node.Name, err = helpers.CreateMnemonic(bid)

		s.NoError(err)
	}

	s.Exactly(expected, res.Edges)
}

func Test_GetOwnedVehicles_Filters(t *testing.T) {
	// Vehicle | Owner | Privileged Users
	// --------+-------+-----------------
	// 1       | A     | B
	// 2       | B     | C
	// 3       | A     |
	ctx := context.Background()
	assert := assert.New(t)
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)
	first := 3

	repo := New(pdb, config.Settings{})
	_, walletA, err := helpers.GenerateWallet()
	assert.NoError(err)
	_, walletB, err := helpers.GenerateWallet()
	assert.NoError(err)
	_, walletC, err := helpers.GenerateWallet()
	assert.NoError(err)

	for _, v := range []struct {
		TokenID    int
		Owner      *common.Address
		Privileged *common.Address
	}{
		{
			TokenID:    1,
			Owner:      walletA,
			Privileged: walletB,
		},
		{
			TokenID:    2,
			Owner:      walletB,
			Privileged: walletC,
		},
		{
			TokenID: 3,
			Owner:   walletA,
		},
	} {
		vehicle := models.Vehicle{
			ID:           v.TokenID,
			OwnerAddress: v.Owner.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Camry"),
			Year:         null.IntFrom(2022),
			MintedAt:     time.Now(),
		}
		if err := vehicle.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			assert.NoError(err)
		}

		if v.Privileged != nil {
			privileges := models.Privilege{
				TokenID:     v.TokenID,
				PrivilegeID: 1,
				UserAddress: v.Privileged.Bytes(),
				SetAt:       time.Now(),
				ExpiresAt:   time.Now().Add(5 * time.Hour),
			}

			if err := privileges.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
				assert.NoError(err)
			}
		}
	}

	// Filter By Privileged
	// Where Priv wallet has both implied (owner) and explicit (granted) privileges
	// Owner | Privileged | Result
	// ------+------------+-------
	// 	     | B          | 1, 2
	res, err := repo.GetVehicles(ctx, &first, nil, nil, nil, &gmodel.VehiclesFilter{Privileged: walletB})
	assert.NoError(err)

	assert.Equal(2, res.TotalCount)
	assert.Equal(2, res.Edges[0].Node.TokenID)
	assert.Equal(1, res.Edges[1].Node.TokenID)
	assert.Equal(walletB, &res.Edges[0].Node.Owner)
	assert.Equal(walletA, &res.Edges[1].Node.Owner)

	// Filter By Owner
	// Owner | Privileged | Result
	// ------+------------+-------
	// A     |            | 1, 3
	res, err = repo.GetVehicles(ctx, &first, nil, nil, nil, &gmodel.VehiclesFilter{Owner: walletA})
	assert.NoError(err)

	assert.Equal(2, res.TotalCount)
	assert.Equal(3, res.Edges[0].Node.TokenID)
	assert.Equal(1, res.Edges[1].Node.TokenID)
	assert.Equal(walletA, &res.Edges[0].Node.Owner)
	assert.Equal(walletA, &res.Edges[1].Node.Owner)

	// Filter By Privileged & Owner
	// where result must match both criteria, valid owner and priv wallets passed
	// Owner | Privileged | Result
	// ------+------------+-------
	// A     | B          | 1
	res, err = repo.GetVehicles(ctx, &first, nil, nil, nil, &gmodel.VehiclesFilter{Owner: walletA, Privileged: walletB})
	assert.NoError(err)

	assert.Equal(1, res.TotalCount)
	assert.Equal(1, res.Edges[0].Node.TokenID)
	assert.Equal(walletA, &res.Edges[0].Node.Owner)

	// Filter By Privileged & Owner
	// where result must match both criteria, invalid owner and priv wallets passed
	// Owner | Privileged | Result
	// ------+------------+-------
	// C     | B          |
	res, err = repo.GetVehicles(ctx, &first, nil, nil, nil, &gmodel.VehiclesFilter{Owner: walletC, Privileged: walletB})
	assert.NoError(err)

	assert.Equal(0, res.TotalCount)
}

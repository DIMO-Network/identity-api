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

const migrationsDir = "../../migrations"

func (s *OwnedVehiclesRepoTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = helpers.StartContainerDatabase(s.ctx, s.T(), migrationsDir)

	s.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
	}
	s.repo = New(s.pdb)
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
	res, err := s.repo.GetAccessibleVehicles(s.ctx, *wallet, &first, nil, nil, nil)
	s.NoError(err)

	s.Equal(2, res.TotalCount)
	s.False(res.PageInfo.HasNextPage)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:         2,
				Owner:      common.BytesToAddress(wallet.Bytes()),
				Make:       &vehicles[1].Make.String,
				Model:      &vehicles[1].Model.String,
				Year:       &vehicles[1].Year.Int,
				MintedAt:   vehicles[1].MintedAt,
				Privileges: nil,
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:         1,
				Owner:      common.BytesToAddress(wallet.Bytes()),
				Make:       &vehicles[0].Make.String,
				Model:      &vehicles[0].Model.String,
				Year:       &vehicles[0].Year.Int,
				MintedAt:   vehicles[0].MintedAt,
				Privileges: nil,
			},
			Cursor: "MQ==",
		},
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
	res, err := s.repo.GetAccessibleVehicles(s.ctx, *wallet, &first, nil, nil, nil)
	s.NoError(err)

	s.Equal(len(vehicles), res.TotalCount)
	s.Equal(res.PageInfo.HasNextPage, true)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:         2,
				Owner:      common.BytesToAddress(wallet.Bytes()),
				Make:       &vehicles[1].Make.String,
				Model:      &vehicles[1].Model.String,
				Year:       &vehicles[1].Year.Int,
				MintedAt:   vehicles[1].MintedAt,
				Privileges: nil,
			},
			Cursor: "Mg==",
		},
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
	res, err := s.repo.GetAccessibleVehicles(s.ctx, *wallet, &first, &after, nil, nil)
	s.NoError(err)

	s.Len(vehicles, res.TotalCount)
	s.False(res.PageInfo.HasNextPage)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:         1,
				Owner:      common.BytesToAddress(wallet.Bytes()),
				Make:       &vehicles[0].Make.String,
				Model:      &vehicles[0].Model.String,
				Year:       &vehicles[0].Year.Int,
				MintedAt:   vehicles[0].MintedAt,
				Privileges: nil,
			},
			Cursor: "MQ==",
		},
	}

	s.Exactly(expected, res.Edges)
}

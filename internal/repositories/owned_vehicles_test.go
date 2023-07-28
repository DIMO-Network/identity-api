package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/test"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/segmentio/ksuid"
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
	repo      VehiclesRepo
	settings  config.Settings
}

func (o *OwnedVehiclesRepoTestSuite) SetupSuite() {
	o.ctx = context.Background()
	o.pdb, o.container = test.StartContainerDatabase(o.ctx, o.T(), test.MigrationsDirRelPath)

	o.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
	}
	o.repo = NewVehiclesRepo(o.ctx, o.pdb)
}

// TearDownTest after each test truncate tables
func (s *OwnedVehiclesRepoTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (o *OwnedVehiclesRepoTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", o.container.SessionID())

	if err := o.container.Terminate(o.ctx); err != nil {
		o.T().Fatal(err)
	}
}

// Test Runner
func TestOwnedVehiclesRepoTestSuite(t *testing.T) {
	suite.Run(t, new(OwnedVehiclesRepoTestSuite))
}

/* Actual Tests */
func (s *OwnedVehiclesRepoTestSuite) Test_GetOwnedVehicles_Success() {
	_, wallet, err := test.GenerateWallet()
	s.NoError(err)

	_, wallet2, err := test.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Microsecond)
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

	expiresAt := currTime.Add(time.Minute + 10).UTC().Truncate(time.Microsecond)
	privileges := []models.Privilege{
		{
			ID:          ksuid.New().String(),
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
	res, err := s.repo.GetOwnedVehicles(s.ctx, *wallet, &first, nil)
	s.NoError(err)

	s.Equal(2, res.TotalCount)
	s.False(res.PageInfo.HasNextPage)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:         "2",
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
				ID:         "1",
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
	_, wallet, err := test.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Microsecond)
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
	res, err := s.repo.GetOwnedVehicles(s.ctx, *wallet, &first, nil)
	s.NoError(err)

	s.Equal(len(vehicles), res.TotalCount)
	s.Equal(res.PageInfo.HasNextPage, true)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:         "2",
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
	_, wallet, err := test.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Microsecond)
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
	res, err := s.repo.GetOwnedVehicles(s.ctx, *wallet, &first, &after)
	s.NoError(err)

	s.Equal(len(vehicles), res.TotalCount)
	s.False(res.PageInfo.HasNextPage)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:         "1",
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

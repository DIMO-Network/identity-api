package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/test"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type VehiclesPrivilegesRepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container testcontainers.Container
	repo      VehiclesRepo
	settings  config.Settings
}

func (s *VehiclesPrivilegesRepoTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), test.MigrationsDirRelPath)

	s.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
	}
	s.repo = NewVehiclesRepo(s.ctx, s.pdb)
}

// TearDownTest after each test truncate tables
func (s *VehiclesPrivilegesRepoTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (o *VehiclesPrivilegesRepoTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", o.container.SessionID())

	if err := o.container.Terminate(o.ctx); err != nil {
		o.T().Fatal(err)
	}
}

// Test Runner
func TestVehiclesPrivilegesRepoTestSuite(t *testing.T) {
	suite.Run(t, new(VehiclesPrivilegesRepoTestSuite))
}

func (s *VehiclesPrivilegesRepoTestSuite) Test_GetVehiclePrivileges_Success() {
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
	}

	for _, v := range vehicles {
		if err := v.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
			s.NoError(err)
		}
	}

	expiresAt := currTime.Add(time.Minute + 10).UTC().Truncate(time.Microsecond)
	privileges := []*models.Privilege{
		{
			ID:          ksuid.New().String(),
			TokenID:     1,
			PrivilegeID: 1,
			UserAddress: wallet.Bytes(),
			SetAt:       currTime,
			ExpiresAt:   expiresAt,
		},
	}

	for _, p := range privileges {
		if err := p.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
			s.NoError(err)
		}
	}

	make := "Toyota"
	mdl := "Camry"
	year := 2020

	veh := &model.Vehicle{
		ID:         "1",
		Owner:      *wallet,
		Make:       &make,
		Model:      &mdl,
		Year:       &year,
		MintedAt:   currTime,
		Privileges: nil,
	}

	res, err := s.repo.GetPrivilegesForVehicles(s.ctx, veh)
	s.NoError(err)

	s.Exactly([]*model.Privilege{
		{
			ID:        1,
			User:      *wallet,
			SetAt:     currTime,
			ExpiresAt: expiresAt,
		},
	}, res)
}

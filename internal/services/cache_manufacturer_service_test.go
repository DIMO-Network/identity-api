package services

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	test "github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type ManufacturerCacheServiceTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container *postgres.PostgresContainer
	settings  config.Settings
	logger    zerolog.Logger
}

func (o *ManufacturerCacheServiceTestSuite) SetupSuite() {
	o.ctx = context.Background()
	o.pdb, o.container = test.StartContainerDatabase(o.ctx, o.T(), "../../migrations")

	o.settings = config.Settings{}

	o.logger = zerolog.New(os.Stdout).With().Timestamp().Str("app", test.DBSettings.Name).Logger()
}

// TearDownTest after each test truncate tables
func (s *ManufacturerCacheServiceTestSuite) TearDownTest() {
	s.Require().NoError(s.container.Restore(s.ctx))
}

// TearDownSuite cleanup at end by terminating container
func (o *ManufacturerCacheServiceTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", o.container.SessionID())

	if err := o.container.Terminate(o.ctx); err != nil {
		o.T().Fatal(err)
	}
}

// Test Runner
func TestManufacturerCacheServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ManufacturerCacheServiceTestSuite))
}

func (o *ManufacturerCacheServiceTestSuite) Test_Manufacturer_All_Success() {
	manufacturers := []string{"ford", "tesla", "kia", "acura", "honda", "jeep"}

	for i := 0; i < 6; i++ {
		m := models.Manufacturer{
			ID:       i,
			Name:     manufacturers[i],
			Owner:    common.FromHex("3232323232323232323232323232323232323232"),
			MintedAt: time.Now(),
		}

		err := m.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer())
		o.NoError(err)
	}

	cache := NewManufacturerCacheService(o.pdb, &o.logger, &o.settings)
	all, err := cache.GetAllManufacturers(o.ctx)
	o.NoError(err)
	o.Len(all, len(manufacturers))
}

func (o *ManufacturerCacheServiceTestSuite) Test_Manufacturer_Empty() {
	cache := NewManufacturerCacheService(o.pdb, &o.logger, &o.settings)

	all, err := cache.GetAllManufacturers(o.ctx)
	o.NoError(err)
	o.Len(all, 0)
}

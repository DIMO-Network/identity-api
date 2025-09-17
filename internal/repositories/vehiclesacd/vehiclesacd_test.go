package vehiclesacd

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type VehiclesSacdRepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container *postgres.PostgresContainer
	repo      *Repository
	settings  config.Settings
}

func (s *VehiclesSacdRepoTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = helpers.StartContainerDatabase(s.ctx, s.T(), "../../../migrations")

	s.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
	}
	logger := zerolog.Nop()
	s.repo = &Repository{base.NewRepository(s.pdb, s.settings, &logger)}
}

// TearDownTest after each test truncate tables
func (s *VehiclesSacdRepoTestSuite) TearDownTest() {
	s.Require().NoError(s.container.Restore(s.ctx))
}

// TearDownSuite cleanup at end by terminating container
func (s *VehiclesSacdRepoTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())

	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

// Test Runner
func TestVehiclesPrivilegesRepoTestSuite(t *testing.T) {
	suite.Run(t, new(VehiclesSacdRepoTestSuite))
}

func (s *VehiclesSacdRepoTestSuite) TestSacdToAPIResponse_WithoutTemplate() {
	sacd := &models.VehicleSacd{
		VehicleID:   1,
		Grantee:     common.HexToAddress("0x1234567890123456789012345678901234567890").Bytes(),
		Permissions: "1010",
		Source:      "test-source",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Hour),
	}

	result, err := sacdToAPIResponse(sacd)
	s.NoError(err)
	s.NotNil(result)
	s.Nil(result.Template)
	s.Equal("0xa", result.Permissions) // 1010 binary = a hex
}

func (s *VehiclesSacdRepoTestSuite) TestSacdToAPIResponse_WithTemplate() {
	templateIdStr := "39432737238797479986393736422353506685818102154178527542856781838379111614015"
	templateId := new(big.Int)
	templateId.SetString(templateIdStr, 10)
	tokenIdBytes, err := helpers.ConvertTokenIDToID(templateId)
	assert.NoError(s.T(), err)

	creator := common.HexToAddress("0x1111111111111111111111111111111111111111")
	asset := common.HexToAddress("0x2222222222222222222222222222222222222222")

	template := &models.Template{
		ID:          tokenIdBytes,
		Creator:     creator.Bytes(),
		Asset:       asset.Bytes(),
		Permissions: "1111",
		Cid:         "QmTestCID123",
		CreatedAt:   time.Now(),
	}

	sacd := &models.VehicleSacd{
		VehicleID:   1,
		Grantee:     common.HexToAddress("0x1234567890123456789012345678901234567890").Bytes(),
		Permissions: "1010",
		Source:      "test-source",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	// Manually set the relationship for testing
	sacd.R = sacd.R.NewStruct()
	sacd.R.Template = template

	result, err := sacdToAPIResponse(sacd)
	s.NoError(err)
	s.NotNil(result)
	s.NotNil(result.Template)

	s.Equal(new(big.Int).SetBytes(tokenIdBytes), result.Template.TokenID)
	s.Equal(creator, result.Template.Creator)
	s.Equal(asset, result.Template.Asset)
	s.Equal("1111", result.Template.Permissions)
	s.Equal("QmTestCID123", result.Template.Cid)
	s.Equal(template.CreatedAt, result.Template.CreatedAt)
}

func (s *VehiclesSacdRepoTestSuite) TestSacdToAPIResponse_WithRButNoTemplate() {
	sacd := &models.VehicleSacd{
		VehicleID:   1,
		Grantee:     common.HexToAddress("0x1234567890123456789012345678901234567890").Bytes(),
		Permissions: "1010",
		Source:      "test-source",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	// Manually set the relationship for testing with nil template
	sacd.R = sacd.R.NewStruct()
	sacd.R.Template = nil

	result, err := sacdToAPIResponse(sacd)
	s.NoError(err)
	s.NotNil(result)
	s.Nil(result.Template)
}

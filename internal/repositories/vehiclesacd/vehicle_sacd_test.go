package vehiclesacd

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type VehiclesSacdRepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container *postgres.PostgresContainer
	repo      *Repository
	settings  config.Settings
}

func (v *VehiclesSacdRepoTestSuite) SetupSuite() {
	v.ctx = context.Background()
	v.pdb, v.container = helpers.StartContainerDatabase(v.ctx, v.T(), "../../../migrations")

	v.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
	}
	logger := zerolog.Nop()
	v.repo = &Repository{base.NewRepository(v.pdb, v.settings, &logger)}
}

// TearDownTest after each test truncate tables
func (v *VehiclesSacdRepoTestSuite) TearDownTest() {
	v.Require().NoError(v.container.Restore(v.ctx))
}

// TearDownSuite cleanup at end by terminating container
func (v *VehiclesSacdRepoTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", v.container.SessionID())

	if err := v.container.Terminate(v.ctx); err != nil {
		v.T().Fatal(err)
	}
}

// Test Runner
func TestVehiclesSacdRepoTestSuite(t *testing.T) {
	suite.Run(t, new(VehiclesSacdRepoTestSuite))
}

func (v *VehiclesSacdRepoTestSuite) Test_GetVehicleSacd_Success() {
	grantee := common.BigToAddress(big.NewInt(123))
	manufOwner := common.BigToAddress(big.NewInt(456))
	vehicleOwner := common.BigToAddress(big.NewInt(789))

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    manufOwner.Bytes(),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(v.ctx, v.pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(v.T(), err)
	}

	vehicles := []models.Vehicle{
		{
			ID:             1,
			ManufacturerID: 131,
			OwnerAddress:   vehicleOwner.Bytes(),
			Make:           null.StringFrom("Toyota"),
			Model:          null.StringFrom("Camry"),
			Year:           null.IntFrom(2020),
			MintedAt:       currTime,
		},
	}

	for _, veh := range vehicles {
		if err := veh.Insert(v.ctx, v.pdb.DBS().Writer, boil.Infer()); err != nil {
			v.NoError(err)
		}
	}

	expiresAt := currTime.Add(time.Minute * 10).UTC().Truncate(time.Second)
	sacd := []*models.Sacd{
		{
			TokenID:     1,
			Permissions: "0011101100101100",
			Grantee:     grantee.Bytes(),
			CreatedAt:   currTime,
			ExpiresAt:   expiresAt,
		},
	}

	for _, s := range sacd {
		if err := s.Insert(v.ctx, v.pdb.DBS().Writer, boil.Infer()); err != nil {
			v.NoError(err)
		}
	}

	res, err := v.repo.GetSacdsForVehicle(v.ctx, 1, nil, nil, nil, nil, nil)
	v.NoError(err)

	pHelp := &helpers.PaginationHelper[SacdCursor]{}
	cursor, err := pHelp.EncodeCursor(SacdCursor{
		CreatedAt:   sacd[len(sacd)-1].CreatedAt,
		Permissions: sacd[len(sacd)-1].Permissions,
		Grantee:     sacd[len(sacd)-1].Grantee,
	})
	v.NoError(err)

	expected := &model.SacdsConnection{
		Edges: []*model.SacdEdge{
			{
				Node: &model.Sacd{
					Permissions: "0x3b2c",
					Grantee:     grantee,
					CreatedAt:   currTime,
					ExpiresAt:   expiresAt,
				},
				Cursor: cursor,
			},
		},
		Nodes: []*model.Sacd{
			{
				Permissions: "0x3b2c",
				Grantee:     grantee,
				CreatedAt:   currTime,
				ExpiresAt:   expiresAt,
			},
		},
		PageInfo: &model.PageInfo{
			EndCursor:   &cursor,
			HasNextPage: false,
		},
		TotalCount: 1,
	}
	v.Exactly(expected.Edges, res.Edges)
}

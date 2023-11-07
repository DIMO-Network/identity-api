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
	ctx := context.Background()
	assert := assert.New(t)
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	repo := New(pdb)
	_, wallet, err := helpers.GenerateWallet()
	assert.NoError(err)
	_, wallet2, err := helpers.GenerateWallet()
	assert.NoError(err)

	vehicleTbl := []models.Vehicle{
		{
			ID:           1,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Camry"),
			Year:         null.IntFrom(2022),
			MintedAt:     time.Now(),
		},
		{
			ID:           2,
			OwnerAddress: wallet2.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Camry"),
			Year:         null.IntFrom(2020),
			MintedAt:     time.Now(),
		},
	}

	for _, v := range vehicleTbl {
		if err := v.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			assert.NoError(err)
		}
	}

	privTbl := []models.Privilege{
		{
			TokenID:     1,
			PrivilegeID: 1,
			UserAddress: wallet2.Bytes(),
			SetAt:       time.Now(),
			ExpiresAt:   time.Now().Add(5 * time.Hour),
		},
	}

	for _, p := range privTbl {
		if err := p.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			assert.NoError(err)
		}
	}

	if a, err := models.Privileges().All(ctx, pdb.DBS().Reader); err != nil {
		assert.NoError(err)
	} else {
		for _, b := range a {
			fmt.Println(b.PrivilegeID, b.UserAddress, b.TokenID)
		}
	}

	// Filter by priv expected response
	privExpectedResp := []gmodel.VehicleEdge{
		{Node: &gmodel.Vehicle{
			TokenID: 2,
			Owner:   common.BytesToAddress(wallet2.Bytes()),
		}},
		{Node: &gmodel.Vehicle{
			TokenID: 1,
			Owner:   common.BytesToAddress(wallet.Bytes()),
		}},
	}

	first := 3
	res, err := repo.GetVehicles(ctx, &first, nil, nil, nil, &gmodel.VehiclesFilter{Privileged: wallet2})
	assert.NoError(err)

	assert.Equal(len(privExpectedResp), res.TotalCount)
	assert.Equal(privExpectedResp[0].Node.TokenID, res.Edges[0].Node.TokenID)
	assert.Equal(privExpectedResp[0].Node.Owner, res.Edges[0].Node.Owner)
	assert.Equal(privExpectedResp[len(privExpectedResp)-1].Node.TokenID, res.Edges[res.TotalCount-1].Node.TokenID)
	assert.Equal(privExpectedResp[len(privExpectedResp)-1].Node.Owner, res.Edges[res.TotalCount-1].Node.Owner)

	// Filter by owner expected response
	ownerExpectedResp := []gmodel.VehicleEdge{
		{Node: &gmodel.Vehicle{
			TokenID: 2,
			Owner:   common.BytesToAddress(wallet2.Bytes()),
		}},
	}

	res, err = repo.GetVehicles(ctx, &first, nil, nil, nil, &gmodel.VehiclesFilter{Owner: wallet2})
	assert.NoError(err)

	assert.Equal(len(ownerExpectedResp), res.TotalCount)
	assert.Equal(ownerExpectedResp[0].Node.TokenID, res.Edges[0].Node.TokenID)
	assert.Equal(ownerExpectedResp[0].Node.Owner, res.Edges[0].Node.Owner)
}

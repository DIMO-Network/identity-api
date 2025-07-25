package vehicleprivilege

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type VehiclesPrivilegesRepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container *postgres.PostgresContainer
	repo      *Repository
	settings  config.Settings
}

func (s *VehiclesPrivilegesRepoTestSuite) SetupSuite() {
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
func (s *VehiclesPrivilegesRepoTestSuite) TearDownTest() {
	s.Require().NoError(s.container.Restore(s.ctx))
}

// TearDownSuite cleanup at end by terminating container
func (s *VehiclesPrivilegesRepoTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())

	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

// Test Runner
func TestVehiclesPrivilegesRepoTestSuite(t *testing.T) {
	suite.Run(t, new(VehiclesPrivilegesRepoTestSuite))
}

func (s *VehiclesPrivilegesRepoTestSuite) Test_GetVehiclePrivileges_Success() {
	_, wallet, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(s.T(), err)
	}

	vehicles := []models.Vehicle{
		{
			ID:             1,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           null.StringFrom("Toyota"),
			Model:          null.StringFrom("Camry"),
			Year:           null.IntFrom(2020),
			MintedAt:       currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
			s.NoError(err)
		}
	}

	expiresAt := currTime.Add(time.Minute * 10).UTC().Truncate(time.Second)
	privileges := []*models.Privilege{
		{
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

	res, err := s.repo.GetPrivilegesForVehicle(s.ctx, 1, nil, nil, nil, nil, nil)
	s.NoError(err)

	pHelp := &helpers.PaginationHelper[PrivilegeCursor]{}
	cursor, err := pHelp.EncodeCursor(PrivilegeCursor{
		SetAt:       currTime,
		PrivilegeID: 1,
		User:        wallet.Bytes(),
	})
	s.NoError(err)

	expected := &model.PrivilegesConnection{
		Edges: []*model.PrivilegeEdge{
			{
				Node: &model.Privilege{
					ID:        1,
					User:      *wallet,
					SetAt:     currTime,
					ExpiresAt: expiresAt,
				},
				Cursor: cursor,
			},
		},
		Nodes: []*model.Privilege{
			{
				ID:        1,
				User:      *wallet,
				SetAt:     currTime,
				ExpiresAt: expiresAt,
			},
		},
		PageInfo: &model.PageInfo{
			EndCursor:   &cursor,
			HasNextPage: false,
		},
		TotalCount: 1,
	}
	s.Exactly(expected, res)
}

func (s *VehiclesPrivilegesRepoTestSuite) Test_Privileges_NoExpiredPrivilege_Pagination() {
	_, wallet, err := helpers.GenerateWallet()
	s.NoError(err)
	_, wallet2, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(s.T(), err)
	}

	vehicles := []models.Vehicle{
		{
			ID:             1,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           null.StringFrom("Toyota"),
			Model:          null.StringFrom("Camry"),
			Year:           null.IntFrom(2020),
			MintedAt:       currTime,
		},
		{
			ID:             2,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           null.StringFrom("Toyota"),
			Model:          null.StringFrom("Camry"),
			Year:           null.IntFrom(2021),
			MintedAt:       currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
			s.NoError(err)
		}
	}

	expiresAt := currTime.Add(time.Minute * 10).UTC().Truncate(time.Second)

	privileges := []*models.Privilege{
		{
			TokenID:     1,
			PrivilegeID: 1,
			UserAddress: wallet2.Bytes(),
			SetAt:       currTime,
			ExpiresAt:   expiresAt,
		},
		{
			TokenID:     2,
			PrivilegeID: 2,
			UserAddress: wallet2.Bytes(),
			SetAt:       currTime,
			ExpiresAt:   currTime.Add(-time.Minute * 10).UTC().Truncate(time.Second),
		},
	}

	for _, p := range privileges {
		if err := p.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
			s.NoError(err)
		}
	}

	limit := 2
	res, err := s.repo.GetPrivilegesForVehicle(s.ctx, 1, &limit, nil, nil, nil, nil)
	s.NoError(err)

	pHelp := &helpers.PaginationHelper[PrivilegeCursor]{}
	cursor, err := pHelp.EncodeCursor(PrivilegeCursor{
		SetAt:       currTime,
		PrivilegeID: 1,
		User:        wallet2.Bytes(),
	})
	s.NoError(err)

	expected := &model.PrivilegesConnection{
		Edges: []*model.PrivilegeEdge{
			{
				Node: &model.Privilege{
					ID:        1,
					User:      *wallet2,
					SetAt:     currTime,
					ExpiresAt: expiresAt,
				},
				Cursor: cursor,
			},
		},
		Nodes: []*model.Privilege{
			{
				ID:        1,
				User:      *wallet2,
				SetAt:     currTime,
				ExpiresAt: expiresAt,
			},
		},
		PageInfo: &model.PageInfo{
			EndCursor:   &cursor,
			HasNextPage: false,
		},
		TotalCount: 1,
	}
	s.Exactly(expected, res)
}

func (s *VehiclesPrivilegesRepoTestSuite) Test_Privileges_Pagination_Success() {
	_, wallet, err := helpers.GenerateWallet()
	s.NoError(err)
	_, wallet2, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(s.T(), err)
	}

	vehicles := []models.Vehicle{
		{
			ID:             1,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           null.StringFrom("Toyota"),
			Model:          null.StringFrom("Camry"),
			Year:           null.IntFrom(2020),
			MintedAt:       currTime,
		},
		{
			ID:             2,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           null.StringFrom("Toyota"),
			Model:          null.StringFrom("Camry"),
			Year:           null.IntFrom(2021),
			MintedAt:       currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
			s.NoError(err)
		}
	}

	expiresAt := currTime.Add(time.Minute * 10).UTC().Truncate(time.Second)

	privileges := []*models.Privilege{
		{
			TokenID:     1,
			PrivilegeID: 1,
			UserAddress: wallet2.Bytes(),
			SetAt:       currTime,
			ExpiresAt:   expiresAt,
		},
		{
			TokenID:     1,
			PrivilegeID: 2,
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

	limit := 1
	res, err := s.repo.GetPrivilegesForVehicle(s.ctx, 1, &limit, nil, nil, nil, nil)
	s.NoError(err)

	pHelp := &helpers.PaginationHelper[PrivilegeCursor]{}
	cursor, err := pHelp.EncodeCursor(PrivilegeCursor{
		SetAt:       currTime,
		PrivilegeID: 1,
		User:        wallet2.Bytes(),
	})
	s.NoError(err)

	expected := &model.PrivilegesConnection{
		Edges: []*model.PrivilegeEdge{
			{
				Node: &model.Privilege{
					ID:        1,
					User:      *wallet2,
					SetAt:     currTime,
					ExpiresAt: expiresAt,
				},
				Cursor: cursor,
			},
		},
		Nodes: []*model.Privilege{
			{
				ID:        1,
				User:      *wallet2,
				SetAt:     currTime,
				ExpiresAt: expiresAt,
			},
		},
		PageInfo: &model.PageInfo{
			EndCursor:   &cursor,
			HasNextPage: true,
		},
		TotalCount: 2,
	}
	s.Exactly(expected, res)

	res, err = s.repo.GetPrivilegesForVehicle(s.ctx, 1, &limit, res.PageInfo.EndCursor, nil, nil, nil)
	s.NoError(err)

	cursor, err = pHelp.EncodeCursor(PrivilegeCursor{
		SetAt:       currTime,
		PrivilegeID: 2,
		User:        wallet2.Bytes(),
	})
	s.NoError(err)

	expected = &model.PrivilegesConnection{
		Edges: []*model.PrivilegeEdge{
			{
				Node: &model.Privilege{
					ID:        2,
					User:      *wallet2,
					SetAt:     currTime,
					ExpiresAt: expiresAt,
				},
				Cursor: cursor,
			},
		},
		Nodes: []*model.Privilege{
			{
				ID:        2,
				User:      *wallet2,
				SetAt:     currTime,
				ExpiresAt: expiresAt,
			},
		},
		PageInfo: &model.PageInfo{
			EndCursor:   &cursor,
			HasNextPage: false,
		},
		TotalCount: 2,
	}
	s.Exactly(expected, res)
}

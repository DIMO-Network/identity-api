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
	"github.com/DIMO-Network/shared/db"
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
	s.repo = &Repository{base.NewRepository(s.pdb, s.settings)}
}

// TearDownTest after each test truncate tables
func (s *VehiclesPrivilegesRepoTestSuite) TearDownTest() {
	helpers.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
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
			Year:         null.IntFrom(2021),
			MintedAt:     currTime,
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
			Year:         null.IntFrom(2021),
			MintedAt:     currTime,
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

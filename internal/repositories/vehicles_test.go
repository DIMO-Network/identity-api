package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	test "github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type AccessibleVehiclesRepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container testcontainers.Container
	repo      *Repository
	settings  config.Settings
}

func (o *AccessibleVehiclesRepoTestSuite) SetupSuite() {
	o.ctx = context.Background()
	o.pdb, o.container = test.StartContainerDatabase(o.ctx, o.T(), "../../migrations")

	o.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
	}
	o.repo = New(o.pdb)
}

// TearDownTest after each test truncate tables
func (s *AccessibleVehiclesRepoTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (o *AccessibleVehiclesRepoTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", o.container.SessionID())

	if err := o.container.Terminate(o.ctx); err != nil {
		o.T().Fatal(err)
	}
}

// Test Runner
func TestAccessibleVehiclesRepoTestSuite(t *testing.T) {
	suite.Run(t, new(AccessibleVehiclesRepoTestSuite))
}

/* Actual Tests */
func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Success() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	_, wallet2, err := test.GenerateWallet()
	o.NoError(err)

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
		if err := v.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	privileges := []models.Privilege{
		{
			TokenID:     1,
			PrivilegeID: 1,
			UserAddress: wallet2.Bytes(),
			SetAt:       currTime,
			ExpiresAt:   currTime,
		},
	}

	for _, p := range privileges {
		if err := p.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	first := 3
	res, err := o.repo.GetAccessibleVehicles(o.ctx, *wallet, &first, nil)
	o.NoError(err)

	o.Equal(2, res.TotalCount)
	o.Equal(res.PageInfo.HasNextPage, false)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:       2,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				Make:     &vehicles[1].Make.String,
				Model:    &vehicles[1].Model.String,
				Year:     &vehicles[1].Year.Int,
				MintedAt: vehicles[1].MintedAt,
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:       1,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				Make:     &vehicles[0].Make.String,
				Model:    &vehicles[0].Model.String,
				Year:     &vehicles[0].Year.Int,
				MintedAt: vehicles[0].MintedAt,
			},
			Cursor: "MQ==",
		},
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

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
		if err := v.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	first := 1
	res, err := o.repo.GetAccessibleVehicles(o.ctx, *wallet, &first, nil)
	o.NoError(err)

	o.Equal(len(vehicles), res.TotalCount)
	o.Equal(res.PageInfo.HasNextPage, true)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:       2,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				Make:     &vehicles[1].Make.String,
				Model:    &vehicles[1].Model.String,
				Year:     &vehicles[1].Year.Int,
				MintedAt: vehicles[1].MintedAt,
			},
			Cursor: "Mg==",
		},
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination_NextPage() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

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
		if err := v.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	first := 1
	after := "Mg=="
	res, err := o.repo.GetAccessibleVehicles(o.ctx, *wallet, &first, &after)
	o.NoError(err)

	o.Equal(len(vehicles), res.TotalCount)
	o.Equal(res.PageInfo.HasNextPage, false)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:       1,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				Make:     &vehicles[0].Make.String,
				Model:    &vehicles[0].Model.String,
				Year:     &vehicles[0].Year.Int,
				MintedAt: vehicles[0].MintedAt,
			},
			Cursor: "MQ==",
		},
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_OwnedByUser_And_ForPrivilegesGranted() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	_, wallet2, err := test.GenerateWallet()
	o.NoError(err)

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
			OwnerAddress: wallet2.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Camry"),
			Year:         null.IntFrom(2022),
			MintedAt:     currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	privileges := []models.Privilege{
		{
			TokenID:     2,
			PrivilegeID: 1,
			UserAddress: wallet.Bytes(),
			SetAt:       currTime,
			ExpiresAt:   currTime.Add(time.Hour),
		},
	}

	for _, p := range privileges {
		if err := p.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	first := 3
	res, err := o.repo.GetAccessibleVehicles(o.ctx, *wallet, &first, nil)
	o.NoError(err)

	o.Equal(2, res.TotalCount)
	o.Equal(res.PageInfo.HasNextPage, false)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:       2,
				Owner:    common.BytesToAddress(wallet2.Bytes()),
				Make:     &vehicles[1].Make.String,
				Model:    &vehicles[1].Model.String,
				Year:     &vehicles[1].Year.Int,
				MintedAt: vehicles[1].MintedAt,
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:       1,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				Make:     &vehicles[0].Make.String,
				Model:    &vehicles[0].Model.String,
				Year:     &vehicles[0].Year.Int,
				MintedAt: vehicles[0].MintedAt,
			},
			Cursor: "MQ==",
		},
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) TestVehiclesMultiplePrivsOnOne() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	_, wallet2, err := test.GenerateWallet()
	o.NoError(err)

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
			OwnerAddress: wallet2.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Camry"),
			Year:         null.IntFrom(2022),
			MintedAt:     currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	privileges := []models.Privilege{
		{
			TokenID:     2,
			PrivilegeID: 1,
			UserAddress: wallet.Bytes(),
			SetAt:       currTime,
			ExpiresAt:   currTime.Add(time.Hour),
		},
		{
			TokenID:     2,
			PrivilegeID: 2,
			UserAddress: wallet.Bytes(),
			SetAt:       currTime,
			ExpiresAt:   currTime.Add(time.Hour),
		},
	}

	for _, p := range privileges {
		if err := p.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	first := 3
	res, err := o.repo.GetAccessibleVehicles(o.ctx, *wallet, &first, nil)
	o.NoError(err)

	o.Equal(2, res.TotalCount)
	o.Equal(res.PageInfo.HasNextPage, false)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:       2,
				Owner:    common.BytesToAddress(wallet2.Bytes()),
				Make:     &vehicles[1].Make.String,
				Model:    &vehicles[1].Model.String,
				Year:     &vehicles[1].Year.Int,
				MintedAt: vehicles[1].MintedAt,
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:       1,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				Make:     &vehicles[0].Make.String,
				Model:    &vehicles[0].Model.String,
				Year:     &vehicles[0].Year.Int,
				MintedAt: vehicles[0].MintedAt,
			},
			Cursor: "MQ==",
		},
	}

	o.Exactly(expected, res.Edges)
}

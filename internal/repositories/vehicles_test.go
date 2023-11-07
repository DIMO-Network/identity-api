package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	test "github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const migrationsDir = "../../migrations"

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

	ddfUrl := []string{
		"http://some-url.com",
		"http://some-url-2.com",
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:            1,
			OwnerAddress:  wallet.Bytes(),
			Make:          null.StringFrom("Toyota"),
			Model:         null.StringFrom("Camry"),
			Year:          null.IntFrom(2020),
			MintedAt:      currTime,
			DefinitionURI: null.StringFrom(ddfUrl[0]),
		},
		{
			ID:            2,
			OwnerAddress:  wallet.Bytes(),
			Make:          null.StringFrom("Toyota"),
			Model:         null.StringFrom("Camry"),
			Year:          null.IntFrom(2022),
			MintedAt:      currTime,
			DefinitionURI: null.StringFrom(ddfUrl[1]),
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
	res, err := o.repo.GetVehicles(o.ctx, &first, nil, nil, nil, &gmodel.VehiclesFilter{Privileged: wallet})
	o.NoError(err)

	o.Equal(2, res.TotalCount)
	o.Equal(res.PageInfo.HasNextPage, false)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:       "V_kQI=",
				TokenID:  2,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				MintedAt: vehicles[1].MintedAt,
				Definition: &gmodel.Definition{
					URI:   &ddfUrl[1],
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:       "V_kQE=",
				TokenID:  1,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				MintedAt: vehicles[0].MintedAt,
				Definition: &gmodel.Definition{
					URI:   &ddfUrl[0],
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
			},
			Cursor: "MQ==",
		},
	}

	for _, veh := range expected {
		bid := helpers.IntToBytes(veh.Node.TokenID)
		veh.Node.Name, err = helpers.CreateMnemonic(bid)

		o.NoError(err)
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	ddfUrl := []string{
		"http://some-url.com",
		"http://some-url-2.com",
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:            1,
			OwnerAddress:  wallet.Bytes(),
			Make:          null.StringFrom("Toyota"),
			Model:         null.StringFrom("Camry"),
			Year:          null.IntFrom(2020),
			MintedAt:      currTime,
			DefinitionURI: null.StringFrom(ddfUrl[0]),
		},
		{
			ID:            2,
			OwnerAddress:  wallet.Bytes(),
			Make:          null.StringFrom("Toyota"),
			Model:         null.StringFrom("Camry"),
			Year:          null.IntFrom(2022),
			MintedAt:      currTime,
			DefinitionURI: null.StringFrom(ddfUrl[1]),
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	first := 1
	res, err := o.repo.GetVehicles(o.ctx, &first, nil, nil, nil, &gmodel.VehiclesFilter{Privileged: wallet})
	o.NoError(err)

	o.Equal(len(vehicles), res.TotalCount)
	o.Equal(res.PageInfo.HasNextPage, true)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:       "V_kQI=",
				TokenID:  2,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				MintedAt: vehicles[1].MintedAt,
				Definition: &gmodel.Definition{
					URI:   &ddfUrl[1],
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
			},
			Cursor: "Mg==",
		},
	}

	for _, veh := range expected {
		bid := helpers.IntToBytes(veh.Node.TokenID)
		veh.Node.Name, err = helpers.CreateMnemonic(bid)

		o.NoError(err)
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination_NextPage() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	ddfUrl := []string{
		"http://some-url.com",
		"http://some-url-2.com",
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:            1,
			OwnerAddress:  wallet.Bytes(),
			Make:          null.StringFrom("Toyota"),
			Model:         null.StringFrom("Camry"),
			Year:          null.IntFrom(2020),
			MintedAt:      currTime,
			DefinitionURI: null.StringFrom(ddfUrl[0]),
		},
		{
			ID:            2,
			OwnerAddress:  wallet.Bytes(),
			Make:          null.StringFrom("Toyota"),
			Model:         null.StringFrom("Camry"),
			Year:          null.IntFrom(2022),
			MintedAt:      currTime,
			DefinitionURI: null.StringFrom(ddfUrl[1]),
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	first := 1
	after := "Mg=="
	res, err := o.repo.GetVehicles(o.ctx, &first, &after, nil, nil, &gmodel.VehiclesFilter{Privileged: wallet})
	o.NoError(err)

	o.Equal(len(vehicles), res.TotalCount)
	o.Equal(res.PageInfo.HasNextPage, false)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:       "V_kQE=",
				TokenID:  1,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				MintedAt: vehicles[0].MintedAt,
				Definition: &gmodel.Definition{
					URI:   &ddfUrl[0],
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
			},
			Cursor: "MQ==",
		},
	}

	for _, veh := range expected {
		bid := helpers.IntToBytes(veh.Node.TokenID)
		veh.Node.Name, err = helpers.CreateMnemonic(bid)

		o.NoError(err)
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_OwnedByUser_And_ForPrivilegesGranted() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	_, wallet2, err := test.GenerateWallet()
	o.NoError(err)

	ddfUrl := []string{
		"http://some-url.com",
		"http://some-url-2.com",
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:            1,
			OwnerAddress:  wallet.Bytes(),
			Make:          null.StringFrom("Toyota"),
			Model:         null.StringFrom("Camry"),
			Year:          null.IntFrom(2020),
			MintedAt:      currTime,
			DefinitionURI: null.StringFrom(ddfUrl[0]),
		},
		{
			ID:            2,
			OwnerAddress:  wallet2.Bytes(),
			Make:          null.StringFrom("Toyota"),
			Model:         null.StringFrom("Camry"),
			Year:          null.IntFrom(2022),
			MintedAt:      currTime,
			DefinitionURI: null.StringFrom(ddfUrl[1]),
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
	res, err := o.repo.GetVehicles(o.ctx, &first, nil, nil, nil, &gmodel.VehiclesFilter{Privileged: wallet})
	o.NoError(err)

	o.Equal(2, res.TotalCount)
	o.Equal(res.PageInfo.HasNextPage, false)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:       "V_kQI=",
				TokenID:  2,
				Owner:    common.BytesToAddress(wallet2.Bytes()),
				MintedAt: vehicles[1].MintedAt,
				Definition: &gmodel.Definition{
					URI:   &ddfUrl[1],
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:       "V_kQE=",
				TokenID:  1,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				MintedAt: vehicles[0].MintedAt,
				Definition: &gmodel.Definition{
					URI:   &ddfUrl[0],
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
			},
			Cursor: "MQ==",
		},
	}

	for _, veh := range expected {
		bid := helpers.IntToBytes(veh.Node.TokenID)
		veh.Node.Name, err = helpers.CreateMnemonic(bid)

		o.NoError(err)
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) TestVehiclesMultiplePrivsOnOne() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	_, wallet2, err := test.GenerateWallet()
	o.NoError(err)

	ddfUrl := []string{
		"http://some-url.com",
		"http://some-url-2.com",
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:            1,
			OwnerAddress:  wallet.Bytes(),
			Make:          null.StringFrom("Toyota"),
			Model:         null.StringFrom("Camry"),
			Year:          null.IntFrom(2020),
			MintedAt:      currTime,
			DefinitionURI: null.StringFrom(ddfUrl[0]),
		},
		{
			ID:            2,
			OwnerAddress:  wallet2.Bytes(),
			Make:          null.StringFrom("Toyota"),
			Model:         null.StringFrom("Camry"),
			Year:          null.IntFrom(2022),
			MintedAt:      currTime,
			DefinitionURI: null.StringFrom(ddfUrl[1]),
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
	res, err := o.repo.GetVehicles(o.ctx, &first, nil, nil, nil, &gmodel.VehiclesFilter{Privileged: wallet})
	o.NoError(err)

	o.Equal(2, res.TotalCount)
	o.Equal(res.PageInfo.HasNextPage, false)

	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:       "V_kQI=",
				TokenID:  2,
				Owner:    common.BytesToAddress(wallet2.Bytes()),
				MintedAt: vehicles[1].MintedAt,
				Definition: &gmodel.Definition{
					URI:   &ddfUrl[1],
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:       "V_kQE=",
				TokenID:  1,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				MintedAt: vehicles[0].MintedAt,
				Definition: &gmodel.Definition{
					URI:   &ddfUrl[0],
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
			},
			Cursor: "MQ==",
		},
	}

	for _, veh := range expected {
		bid := helpers.IntToBytes(veh.Node.TokenID)
		veh.Node.Name, err = helpers.CreateMnemonic(bid)

		o.NoError(err)
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination_PreviousPage() {
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
			Model:        null.StringFrom("Rav4"),
			Year:         null.IntFrom(2022),
			MintedAt:     currTime,
		},
		{
			ID:           3,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Corolla"),
			Year:         null.IntFrom(2023),
			MintedAt:     currTime,
		},
		{
			ID:           4,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Highlander"),
			Year:         null.IntFrom(2018),
			MintedAt:     currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	last := 2
	before := "MQ=="
	startCrsr := "Mw=="
	endCrsr := "Mg=="
	res, err := o.repo.GetVehicles(o.ctx, nil, nil, &last, &before, &gmodel.VehiclesFilter{Privileged: wallet})
	o.NoError(err)

	o.Len(res.Edges, 2)
	o.Equal(res.TotalCount, 4)
	o.Equal(res.PageInfo, &gmodel.PageInfo{
		StartCursor:     &startCrsr,
		EndCursor:       &endCrsr,
		HasPreviousPage: true,
		HasNextPage:     true,
	})
	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:                "V_kQM=",
				TokenID:           3,
				Owner:             common.BytesToAddress(wallet.Bytes()),
				MintedAt:          vehicles[2].MintedAt,
				AftermarketDevice: nil,
				Privileges:        nil,
				SyntheticDevice:   nil,
				Definition: &gmodel.Definition{
					URI:   nil,
					Make:  &vehicles[2].Make.String,
					Model: &vehicles[2].Model.String,
					Year:  &vehicles[2].Year.Int,
				},
			},
			Cursor: "Mw==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:                "V_kQI=",
				TokenID:           2,
				Owner:             common.BytesToAddress(wallet.Bytes()),
				MintedAt:          vehicles[1].MintedAt,
				AftermarketDevice: nil,
				Privileges:        nil,
				SyntheticDevice:   nil,
				Definition: &gmodel.Definition{
					URI:   nil,
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
			},
			Cursor: "Mg==",
		},
	}

	for _, veh := range expected {
		bid := helpers.IntToBytes(veh.Node.TokenID)
		veh.Node.Name, err = helpers.CreateMnemonic(bid)

		o.NoError(err)
	}
	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination_AfterBefore() {
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
			Model:        null.StringFrom("Rav4"),
			Year:         null.IntFrom(2022),
			MintedAt:     currTime,
		},
		{
			ID:           3,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Corolla"),
			Year:         null.IntFrom(2023),
			MintedAt:     currTime,
		},
		{
			ID:           4,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Highlander"),
			Year:         null.IntFrom(2018),
			MintedAt:     currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	// Order is 4 3 2 1
	//            ^   ^
	//            |   |
	//        after   before

	last := 2
	after := "Mw=="     // 3
	before := "MQ=="    // 1
	startCrsr := "Mg==" // 2
	endCrsr := "Mg=="
	res, err := o.repo.GetVehicles(o.ctx, nil, &after, &last, &before, &gmodel.VehiclesFilter{Privileged: wallet})
	o.NoError(err)

	o.Len(res.Edges, 1)
	o.Equal(res.TotalCount, 4)
	o.Equal(&gmodel.PageInfo{
		StartCursor:     &startCrsr,
		EndCursor:       &endCrsr,
		HasPreviousPage: true,
		HasNextPage:     true,
	}, res.PageInfo)
	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:                "V_kQI=",
				TokenID:           2,
				Owner:             common.BytesToAddress(wallet.Bytes()),
				MintedAt:          vehicles[0].MintedAt,
				AftermarketDevice: nil,
				Privileges:        nil,
				SyntheticDevice:   nil,
				Definition: &gmodel.Definition{
					URI:   nil,
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
			},
			Cursor: "Mg==",
		},
	}

	for _, veh := range expected {
		bid := helpers.IntToBytes(veh.Node.TokenID)
		veh.Node.Name, err = helpers.CreateMnemonic(bid)

		o.NoError(err)
	}
	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination_AfterLast() {
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
			Model:        null.StringFrom("Rav4"),
			Year:         null.IntFrom(2022),
			MintedAt:     currTime,
		},
		{
			ID:           3,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Corolla"),
			Year:         null.IntFrom(2023),
			MintedAt:     currTime,
		},
		{
			ID:           4,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Highlander"),
			Year:         null.IntFrom(2018),
			MintedAt:     currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	last := 2
	after := "NA=="     // 4. Doesn't have much of an effect.
	startCrsr := "Mg==" // 2.
	endCrsr := "MQ=="   // 1.
	res, err := o.repo.GetVehicles(o.ctx, nil, &after, &last, nil, &gmodel.VehiclesFilter{Privileged: wallet})
	o.NoError(err)

	o.Len(res.Edges, 2)
	o.Equal(res.TotalCount, 4)
	o.Equal(res.PageInfo, &gmodel.PageInfo{
		StartCursor:     &startCrsr,
		EndCursor:       &endCrsr,
		HasPreviousPage: true,
		HasNextPage:     false,
	})
	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:       "V_kQI=",
				TokenID:  2,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				MintedAt: vehicles[1].MintedAt,
				Definition: &gmodel.Definition{
					URI:   nil,
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:       "V_kQE=",
				TokenID:  1,
				Owner:    common.BytesToAddress(wallet.Bytes()),
				MintedAt: vehicles[0].MintedAt,
				Definition: &gmodel.Definition{
					URI:   nil,
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
			},
			Cursor: "MQ==",
		},
	}
	for _, veh := range expected {
		bid := helpers.IntToBytes(veh.Node.TokenID)
		veh.Node.Name, err = helpers.CreateMnemonic(bid)

		o.NoError(err)
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination_BeforeFirst() {
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
			Model:        null.StringFrom("Rav4"),
			Year:         null.IntFrom(2022),
			MintedAt:     currTime,
		},
		{
			ID:           3,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Corolla"),
			Year:         null.IntFrom(2023),
			MintedAt:     currTime,
		},
		{
			ID:           4,
			OwnerAddress: wallet.Bytes(),
			Make:         null.StringFrom("Toyota"),
			Model:        null.StringFrom("Highlander"),
			Year:         null.IntFrom(2018),
			MintedAt:     currTime,
		},
	}

	for _, v := range vehicles {
		if err := v.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	first := 2
	before := "Mg=="
	startCrsr := "NA=="
	endCrsr := "Mw=="
	res, err := o.repo.GetVehicles(o.ctx, &first, nil, nil, &before, &gmodel.VehiclesFilter{Privileged: wallet})
	o.NoError(err)

	o.Len(res.Edges, 2)
	o.Equal(res.TotalCount, 4)
	o.Equal(res.PageInfo, &gmodel.PageInfo{
		StartCursor:     &startCrsr,
		EndCursor:       &endCrsr,
		HasPreviousPage: false,
		HasNextPage:     true,
	})
	expected := []*gmodel.VehicleEdge{
		{
			Node: &gmodel.Vehicle{
				ID:                "V_kQQ=",
				TokenID:           4,
				Owner:             common.BytesToAddress(wallet.Bytes()),
				MintedAt:          vehicles[3].MintedAt,
				AftermarketDevice: nil,
				Privileges:        nil,
				SyntheticDevice:   nil,
				Definition: &gmodel.Definition{
					URI:   nil,
					Make:  &vehicles[3].Make.String,
					Model: &vehicles[3].Model.String,
					Year:  &vehicles[3].Year.Int,
				},
			},
			Cursor: "NA==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:                "V_kQM=",
				TokenID:           3,
				Owner:             common.BytesToAddress(wallet.Bytes()),
				MintedAt:          vehicles[2].MintedAt,
				AftermarketDevice: nil,
				Privileges:        nil,
				SyntheticDevice:   nil,
				Definition: &gmodel.Definition{
					URI:   nil,
					Make:  &vehicles[2].Make.String,
					Model: &vehicles[2].Model.String,
					Year:  &vehicles[2].Year.Int,
				},
			},
			Cursor: "Mw==",
		},
	}

	for _, veh := range expected {
		bid := helpers.IntToBytes(veh.Node.TokenID)
		veh.Node.Name, err = helpers.CreateMnemonic(bid)

		o.NoError(err)
	}
	o.Exactly(expected, res.Edges)
}

package vehicle

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"net/http"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	test "github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/repositories/devicedefinition"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/mnemonic"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jarcoal/httpmock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	toyota     = null.StringFrom("Toyota")
	honda      = null.StringFrom("Honda")
	camry      = null.StringFrom("Camry")
	highlander = null.StringFrom("Highlander")
	rav4       = null.StringFrom("RAV4")
	corolla    = null.StringFrom("Corolla")
	civic      = null.StringFrom("Civic")
	accord     = null.StringFrom("Accord")
	year2018   = null.IntFrom(2018)
	year2020   = null.IntFrom(2020)
	year2022   = null.IntFrom(2022)
	year2023   = null.IntFrom(2023)
)

type AccessibleVehiclesRepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container *postgres.PostgresContainer
	repo      *Repository
	settings  config.Settings
}

func (o *AccessibleVehiclesRepoTestSuite) SetupSuite() {
	o.ctx = context.Background()
	o.pdb, o.container = test.StartContainerDatabase(o.ctx, o.T(), "../../../migrations")

	o.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
		VehicleNFTAddr:      "0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8",
		BaseImageURL:        "https://mockUrl.com/v1",
		BaseVehicleDataURI:  "https://dimoData/vehicles/",
		TablelandAPIGateway: "http://local/",
	}
	logger := zerolog.Nop()
	baseRepo := base.NewRepository(o.pdb, o.settings, &logger)

	// Create a real device definition repository with HTTP mocks
	tablelandAPI := services.NewTablelandApiService(&logger, &o.settings)
	deviceDefRepo := devicedefinition.New(baseRepo, tablelandAPI)
	o.repo = New(baseRepo, deviceDefRepo)
}

// TearDownTest after each test truncate tables
func (s *AccessibleVehiclesRepoTestSuite) TearDownTest() {
	s.Require().NoError(s.container.Restore(s.ctx))
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

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
		o.NoError(err)
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:             1,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          camry,
			Year:           year2020,
			MintedAt:       currTime,
		},
		{
			ID:             2,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          camry,
			Year:           year2022,
			MintedAt:       currTime,
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
				ID:             "V_kQI=",
				ManufacturerID: 131,
				TokenID:        2,
				TokenDID:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:2",
				Owner:          common.BytesToAddress(wallet.Bytes()),
				MintedAt:       vehicles[1].MintedAt,
				Definition: &gmodel.Definition{
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/2/image",
				Image:    "https://mockUrl.com/v1/vehicle/2/image",
				DataURI:  "https://dimoData/vehicles/2",
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:             "V_kQE=",
				ManufacturerID: 131,
				TokenID:        1,
				TokenDID:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:1",
				Owner:          common.BytesToAddress(wallet.Bytes()),
				MintedAt:       vehicles[0].MintedAt,
				Definition: &gmodel.Definition{
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/1/image",
				Image:    "https://mockUrl.com/v1/vehicle/1/image",
				DataURI:  "https://dimoData/vehicles/1",
			},
			Cursor: "MQ==",
		},
	}

	for _, veh := range expected {
		veh.Node.Name = strings.Join(mnemonic.FromInt32WithObfuscation(int32(veh.Node.TokenID)), " ")
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
		o.NoError(err)
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:             1,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          camry,
			Year:           year2020,
			MintedAt:       currTime,
		},
		{
			ID:             2,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          camry,
			Year:           year2022,
			MintedAt:       currTime,
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
				ID:             "V_kQI=",
				TokenID:        2,
				TokenDID:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:2",
				ManufacturerID: 131,
				Owner:          common.BytesToAddress(wallet.Bytes()),
				MintedAt:       vehicles[1].MintedAt,
				Definition: &gmodel.Definition{
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/2/image",
				Image:    "https://mockUrl.com/v1/vehicle/2/image",
				DataURI:  "https://dimoData/vehicles/2",
			},
			Cursor: "Mg==",
		},
	}

	for _, veh := range expected {
		veh.Node.Name = strings.Join(mnemonic.FromInt32WithObfuscation(int32(veh.Node.TokenID)), " ")
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination_NextPage() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
		o.NoError(err)
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:             1,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          camry,
			Year:           year2020,
			MintedAt:       currTime,
		},
		{
			ID:             2,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          camry,
			Year:           year2022,
			MintedAt:       currTime,
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
				ID:             "V_kQE=",
				TokenID:        1,
				TokenDID:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:1",
				ManufacturerID: 131,
				Owner:          common.BytesToAddress(wallet.Bytes()),
				MintedAt:       vehicles[0].MintedAt,
				Definition: &gmodel.Definition{
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/1/image",
				Image:    "https://mockUrl.com/v1/vehicle/1/image",
				DataURI:  "https://dimoData/vehicles/1",
			},
			Cursor: "MQ==",
		},
	}

	for _, veh := range expected {
		veh.Node.Name = strings.Join(mnemonic.FromInt32WithObfuscation(int32(veh.Node.TokenID)), " ")
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_OwnedByUser_And_ForPrivilegesGranted() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	_, wallet2, err := test.GenerateWallet()
	o.NoError(err)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
		o.Require().NoError(err)
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:             1,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          camry,
			Year:           year2020,
			MintedAt:       currTime,
		},
		{
			ID:             2,
			ManufacturerID: 131,
			OwnerAddress:   wallet2.Bytes(),
			Make:           toyota,
			Model:          camry,
			Year:           year2022,
			MintedAt:       currTime,
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
				ID:             "V_kQI=",
				TokenID:        2,
				TokenDID:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:2",
				Owner:          common.BytesToAddress(wallet2.Bytes()),
				ManufacturerID: 131,
				MintedAt:       vehicles[1].MintedAt,
				Definition: &gmodel.Definition{
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/2/image",
				Image:    "https://mockUrl.com/v1/vehicle/2/image",
				DataURI:  "https://dimoData/vehicles/2",
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:             "V_kQE=",
				TokenID:        1,
				TokenDID:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:1",
				ManufacturerID: 131,
				Owner:          common.BytesToAddress(wallet.Bytes()),
				MintedAt:       vehicles[0].MintedAt,
				Definition: &gmodel.Definition{
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/1/image",
				Image:    "https://mockUrl.com/v1/vehicle/1/image",
				DataURI:  "https://dimoData/vehicles/1",
			},
			Cursor: "MQ==",
		},
	}

	for _, veh := range expected {
		veh.Node.Name = strings.Join(mnemonic.FromInt32WithObfuscation(int32(veh.Node.TokenID)), " ")
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) TestVehiclesMultiplePrivsOnOne() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	_, wallet2, err := test.GenerateWallet()
	o.NoError(err)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
	}

	if err := m.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
		o.Require().NoError(err)
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:             1,
			OwnerAddress:   wallet.Bytes(),
			ManufacturerID: 131,
			Make:           toyota,
			Model:          camry,
			Year:           year2020,
			MintedAt:       currTime,
		},
		{
			ID:             2,
			OwnerAddress:   wallet2.Bytes(),
			ManufacturerID: 131,
			Make:           toyota,
			Model:          camry,
			Year:           year2022,
			MintedAt:       currTime,
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
				ID:             "V_kQI=",
				TokenID:        2,
				TokenDID:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:2",
				ManufacturerID: 131,
				Owner:          common.BytesToAddress(wallet2.Bytes()),
				MintedAt:       vehicles[1].MintedAt,
				Definition: &gmodel.Definition{
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/2/image",
				Image:    "https://mockUrl.com/v1/vehicle/2/image",
				DataURI:  "https://dimoData/vehicles/2",
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:             "V_kQE=",
				TokenID:        1,
				TokenDID:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:1",
				ManufacturerID: 131,
				Owner:          common.BytesToAddress(wallet.Bytes()),
				MintedAt:       vehicles[0].MintedAt,
				Definition: &gmodel.Definition{
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/1/image",
				Image:    "https://mockUrl.com/v1/vehicle/1/image",
				DataURI:  "https://dimoData/vehicles/1",
			},
			Cursor: "MQ==",
		},
	}

	for _, veh := range expected {
		veh.Node.Name = strings.Join(mnemonic.FromInt32WithObfuscation(int32(veh.Node.TokenID)), " ")
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination_PreviousPage() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
	}

	if err := m.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
		o.NoError(err)
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:             1,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          camry,
			Year:           year2020,
			MintedAt:       currTime,
		},
		{
			ID:             2,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          rav4,
			Year:           year2022,
			MintedAt:       currTime,
		},
		{
			ID:             3,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          corolla,
			Year:           year2023,
			MintedAt:       currTime,
		},
		{
			ID:             4,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          highlander,
			Year:           year2018,
			MintedAt:       currTime,
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
				ManufacturerID:    131,
				TokenID:           3,
				TokenDID:          "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:3",
				Owner:             common.BytesToAddress(wallet.Bytes()),
				MintedAt:          vehicles[2].MintedAt,
				AftermarketDevice: nil,
				Privileges:        nil,
				SyntheticDevice:   nil,
				Definition: &gmodel.Definition{
					Make:  &vehicles[2].Make.String,
					Model: &vehicles[2].Model.String,
					Year:  &vehicles[2].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/3/image",
				Image:    "https://mockUrl.com/v1/vehicle/3/image",
				DataURI:  "https://dimoData/vehicles/3",
			},
			Cursor: "Mw==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:                "V_kQI=",
				ManufacturerID:    131,
				TokenID:           2,
				TokenDID:          "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:2",
				Owner:             common.BytesToAddress(wallet.Bytes()),
				MintedAt:          vehicles[1].MintedAt,
				AftermarketDevice: nil,
				Privileges:        nil,
				SyntheticDevice:   nil,
				Definition: &gmodel.Definition{
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/2/image",
				Image:    "https://mockUrl.com/v1/vehicle/2/image",
				DataURI:  "https://dimoData/vehicles/2",
			},
			Cursor: "Mg==",
		},
	}

	for _, veh := range expected {
		veh.Node.Name = strings.Join(mnemonic.FromInt32WithObfuscation(int32(veh.Node.TokenID)), " ")
	}
	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination_AfterBefore() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
	}

	if err := m.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
		o.NoError(err)
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:             1,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          camry,
			Year:           year2020,
			MintedAt:       currTime,
		},
		{
			ID:             2,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          rav4,
			Year:           year2022,
			MintedAt:       currTime,
		},
		{
			ID:             3,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          corolla,
			Year:           year2023,
			MintedAt:       currTime,
		},
		{
			ID:             4,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          highlander,
			Year:           year2018,
			MintedAt:       currTime,
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
				TokenDID:          "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:2",
				Owner:             common.BytesToAddress(wallet.Bytes()),
				MintedAt:          vehicles[0].MintedAt,
				AftermarketDevice: nil,
				ManufacturerID:    131,
				Privileges:        nil,
				SyntheticDevice:   nil,
				Definition: &gmodel.Definition{
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/2/image",
				Image:    "https://mockUrl.com/v1/vehicle/2/image",
				DataURI:  "https://dimoData/vehicles/2",
			},
			Cursor: "Mg==",
		},
	}

	for _, veh := range expected {
		veh.Node.Name = strings.Join(mnemonic.FromInt32WithObfuscation(int32(veh.Node.TokenID)), " ")
	}
	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination_AfterLast() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
	}

	if err := m.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
		o.NoError(err)
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:             1,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          camry,
			Year:           year2020,
			MintedAt:       currTime,
		},
		{
			ID:             2,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          rav4,
			Year:           year2022,
			MintedAt:       currTime,
		},
		{
			ID:             3,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          corolla,
			Year:           year2023,
			MintedAt:       currTime,
		},
		{
			ID:             4,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          highlander,
			Year:           year2018,
			MintedAt:       currTime,
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
				ID:             "V_kQI=",
				TokenID:        2,
				TokenDID:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:2",
				ManufacturerID: 131,
				Owner:          common.BytesToAddress(wallet.Bytes()),
				MintedAt:       vehicles[1].MintedAt,
				Definition: &gmodel.Definition{
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/2/image",
				Image:    "https://mockUrl.com/v1/vehicle/2/image",
				DataURI:  "https://dimoData/vehicles/2",
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:             "V_kQE=",
				TokenID:        1,
				TokenDID:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:1",
				Owner:          common.BytesToAddress(wallet.Bytes()),
				MintedAt:       vehicles[0].MintedAt,
				ManufacturerID: 131,
				Definition: &gmodel.Definition{
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/1/image",
				Image:    "https://mockUrl.com/v1/vehicle/1/image",
				DataURI:  "https://dimoData/vehicles/1",
			},
			Cursor: "MQ==",
		},
	}
	for _, veh := range expected {
		veh.Node.Name = strings.Join(mnemonic.FromInt32WithObfuscation(int32(veh.Node.TokenID)), " ")
	}

	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehicles_Pagination_BeforeFirst() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
	}

	if err := m.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
		o.NoError(err)
	}

	currTime := time.Now().UTC().Truncate(time.Second)
	vehicles := []models.Vehicle{
		{
			ID:             1,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          camry,
			Year:           year2020,
			MintedAt:       currTime,
		},
		{
			ID:             2,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          rav4,
			Year:           year2022,
			MintedAt:       currTime,
		},
		{
			ID:             3,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          corolla,
			Year:           year2023,
			MintedAt:       currTime,
		},
		{
			ID:             4,
			ManufacturerID: 131,
			OwnerAddress:   wallet.Bytes(),
			Make:           toyota,
			Model:          highlander,
			Year:           year2018,
			MintedAt:       currTime,
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
				TokenDID:          "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:4",
				Owner:             common.BytesToAddress(wallet.Bytes()),
				MintedAt:          vehicles[3].MintedAt,
				ManufacturerID:    131,
				AftermarketDevice: nil,
				Privileges:        nil,
				SyntheticDevice:   nil,
				Definition: &gmodel.Definition{
					Make:  &vehicles[3].Make.String,
					Model: &vehicles[3].Model.String,
					Year:  &vehicles[3].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/4/image",
				Image:    "https://mockUrl.com/v1/vehicle/4/image",
				DataURI:  "https://dimoData/vehicles/4",
			},
			Cursor: "NA==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:                "V_kQM=",
				TokenID:           3,
				TokenDID:          "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:3",
				Owner:             common.BytesToAddress(wallet.Bytes()),
				MintedAt:          vehicles[2].MintedAt,
				ManufacturerID:    131,
				AftermarketDevice: nil,
				Privileges:        nil,
				SyntheticDevice:   nil,
				Definition: &gmodel.Definition{
					Make:  &vehicles[2].Make.String,
					Model: &vehicles[2].Model.String,
					Year:  &vehicles[2].Year.Int,
				},
				ImageURI: "https://mockUrl.com/v1/vehicle/3/image",
				Image:    "https://mockUrl.com/v1/vehicle/3/image",
				DataURI:  "https://dimoData/vehicles/3",
			},
			Cursor: "Mw==",
		},
	}

	for _, veh := range expected {
		veh.Node.Name = strings.Join(mnemonic.FromInt32WithObfuscation(int32(veh.Node.TokenID)), " ")
	}
	o.Exactly(expected, res.Edges)
}

func (o *AccessibleVehiclesRepoTestSuite) Test_GetAccessibleVehiclesFilters() {
	_, wallet1, err := test.GenerateWallet()
	o.NoError(err)

	_, wallet2, err := test.GenerateWallet()
	o.NoError(err)

	toyotaMfr := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
		TableID:  null.IntFrom(1),
	}

	hondaMfr := models.Manufacturer{
		ID:       48,
		Name:     "Honda",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDff"),
		MintedAt: time.Now(),
		Slug:     "honda",
		TableID:  null.IntFrom(2),
	}

	nissanMfr := models.Manufacturer{
		ID:       49,
		Name:     "Nissan",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf5"),
		MintedAt: time.Now(),
		Slug:     "nissan",
		TableID:  null.IntFrom(3),
	}

	cadillacMfr := models.Manufacturer{
		ID:       50,
		Name:     "Cadillac",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf6"),
		MintedAt: time.Now(),
		Slug:     "cadillac",
		TableID:  null.IntFrom(4),
	}

	mazdaMfr := models.Manufacturer{
		ID:       51,
		Name:     "Mazda",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf7"),
		MintedAt: time.Now(),
		Slug:     "mazda",
		TableID:  null.IntFrom(5),
	}

	currTime := time.Now().UTC().Truncate(time.Second)

	mfrs := []models.Manufacturer{toyotaMfr, hondaMfr, nissanMfr, cadillacMfr, mazdaMfr}
	for _, v := range mfrs {
		if err := v.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.Require().NoError(err)
		}
	}

	// Set up HTTP mocks for device definition calls
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	const baseURL = "http://local/"

	// Mock the query for nissan manufacturer (table _30001_3)
	queryURLNissan := "api/v1/query?statement=SELECT+%2A+FROM+%22_80001_3%22+WHERE+%28%22id%22+IN+%28%27nissan_gt-r_2020%27%29%29"
	respQueryBodyNissan := `[
	  {
		"id": "nissan_gt-r_2020",
		"deviceType": "vehicle",
		"imageURI": "https://image",
		"ksuid": "26G3iFH7Xc9Wvsw7pg6sD7uzoSS",
		"model": "GT R",
		"year": 2020,
		"metadata": {
		  "device_attributes": [
			{
			  "name": "powertrain_type",
			  "value": "ICE"
			}
		  ]
		}
	  }
	]`
	httpmock.RegisterResponder(http.MethodGet, baseURL+queryURLNissan, httpmock.NewStringResponder(200, respQueryBodyNissan))

	// Mock the query for toyota manufacturer
	queryURLToyota := "api/v1/query?statement=SELECT+%2A+FROM+%22_80001_1%22+WHERE+%28%22id%22+IN+%28%27toyota_rav4_2022%27%2C+%27toyota_camry_2020%27%29%29"
	respQueryBodyToyota := `[
	  {
		"id": "toyota_rav4_2022",
		"deviceType": "vehicle",
		"imageURI": "https://image",
		"ksuid": "26G3iFH7Xc9Wvsw7pg6sD7uzoSS",
		"model": "RAV4",
		"year": 2022,
		"metadata": {
		  "device_attributes": [
			{
			  "name": "powertrain_type",
			  "value": "ICE"
			}
		  ]
		}
	  },	  
      {
		"id": "toyota_camry_2020",
		"deviceType": "vehicle",
		"imageURI": "https://image",
		"ksuid": "26G3iFH7Xc9Wvsw7pg6sD7uzoSS",
		"model": "Camry",
		"year": 2020,
		"metadata": {
		  "device_attributes": [
			{
			  "name": "powertrain_type",
			  "value": "ICE"
			}
		  ]
		}
	  }
	]`
	httpmock.RegisterResponder(http.MethodGet, baseURL+queryURLToyota, httpmock.NewStringResponder(200, respQueryBodyToyota))

	// Mock the query for cadillac manufacturer (table _30001_4)
	queryURLCadillac := "api/v1/query?statement=SELECT+%2A+FROM+%22_80001_4%22+WHERE+%28%22id%22+IN+%28%27cadillac_ats-v-coupe_2019%27%29%29"
	respQueryBodyCadillac := `[
	  {
		"id": "cadillac_ats-v-coupe_2019",
		"deviceType": "vehicle",
		"imageURI": "https://image",
		"ksuid": "12G3iFH7Xc9Wvsw7pg6sD7uzoKK",
		"metadata": "",
		"model": "ATS V Coupe",
		"year": 2019
	  }
	]`
	httpmock.RegisterResponder(http.MethodGet, baseURL+queryURLCadillac, httpmock.NewStringResponder(200, respQueryBodyCadillac))

	// Mock the query for mazda manufacturer (table _30001_5)
	queryURLMazda := "api/v1/query?statement=SELECT+%2A+FROM+%22_80001_5%22+WHERE+%28%22id%22+IN+%28%27mazda_cx-5_2023%27%29%29"
	respQueryBodyMazda := `[
	  {
		"id": "mazda_cx-5_2023",
		"deviceType": "vehicle",
		"imageURI": "https://image",
		"ksuid": "34G3iFH7Xc9Wvsw7pg6sD7uzoLL",
		"metadata": "",
		"model": "CX 5",
		"year": 2023
	  }
	]`
	httpmock.RegisterResponder(http.MethodGet, baseURL+queryURLMazda, httpmock.NewStringResponder(200, respQueryBodyMazda))

	// Mock the query for honda manufacturer
	queryURLHonda := "api/v1/query?statement=SELECT+%2A+FROM+%22_80001_2%22+WHERE+%28%22id%22+IN+%28%27honda_accord_2020%27%2C+%27honda_civic_2022%27%29%29"
	respQueryBodyHonda := `[
	  {
		"id": "honda_civic_2022",
		"deviceType": "vehicle",
		"imageURI": "https://image",
		"ksuid": "34G3iFH7Xc9Wvsw7pg6sD7uzoLL",
		"metadata": "",
		"model": "Civic",
		"year": 2022
	  },
	  {
		"id": "honda_accord_2020",
		"deviceType": "vehicle",
		"imageURI": "https://image",
		"ksuid": "34G3iFH7Xc9Wvsw7pg6sD7uzoLL",
		"metadata": "",
		"model": "Accord",
		"year": 2020
	  }
	]`
	httpmock.RegisterResponder(http.MethodGet, baseURL+queryURLHonda, httpmock.NewStringResponder(200, respQueryBodyHonda))

	toyotaCamryTID1 := models.Vehicle{
		ID:                 1,
		ManufacturerID:     131,
		OwnerAddress:       wallet1.Bytes(),
		Make:               toyota,
		Model:              camry,
		Year:               year2020,
		MintedAt:           currTime,
		DeviceDefinitionID: null.StringFrom("toyota_camry_2020"),
	}
	vehicle1ImageURL, err := DefaultImageURI(o.settings.BaseImageURL, toyotaCamryTID1.ID)
	o.Require().NoError(err)
	vehicle1DataURI, err := GetVehicleDataURI(o.settings.BaseVehicleDataURI, toyotaCamryTID1.ID)
	o.Require().NoError(err)
	vehicle1AsAPI, err := o.repo.ToAPI(&toyotaCamryTID1, vehicle1ImageURL, vehicle1DataURI, nil)
	o.NoError(err)

	hondaCivicTID2 := models.Vehicle{
		ID:                 2,
		OwnerAddress:       wallet1.Bytes(),
		Make:               honda,
		Model:              civic,
		ManufacturerID:     48,
		Year:               year2022,
		MintedAt:           currTime,
		DeviceDefinitionID: null.StringFrom("honda_civic_2022"),
	}
	vehicle2ImageURL, err := DefaultImageURI(o.settings.BaseImageURL, hondaCivicTID2.ID)
	o.Require().NoError(err)
	vehicle2DataURI, err := GetVehicleDataURI(o.settings.BaseVehicleDataURI, hondaCivicTID2.ID)
	o.Require().NoError(err)
	vehicle2AsAPI, err := o.repo.ToAPI(&hondaCivicTID2, vehicle2ImageURL, vehicle2DataURI, nil)
	o.NoError(err)

	toyotaRav4TID3 := models.Vehicle{
		ID:                 3,
		OwnerAddress:       wallet2.Bytes(),
		Make:               toyota,
		ManufacturerID:     131,
		Model:              rav4,
		Year:               year2022,
		MintedAt:           currTime,
		DeviceDefinitionID: null.StringFrom("toyota_rav4_2022"),
	}
	vehicle3ImageURL, err := DefaultImageURI(o.settings.BaseImageURL, toyotaRav4TID3.ID)
	o.Require().NoError(err)
	vehicle3DataURI, err := GetVehicleDataURI(o.settings.BaseVehicleDataURI, toyotaRav4TID3.ID)
	o.Require().NoError(err)
	vehicle3AsAPI, err := o.repo.ToAPI(&toyotaRav4TID3, vehicle3ImageURL, vehicle3DataURI, nil)
	o.NoError(err)

	hondaAccordTID4 := models.Vehicle{
		ID:                 4,
		OwnerAddress:       wallet2.Bytes(),
		Make:               honda,
		Model:              accord,
		ManufacturerID:     48,
		Year:               year2020,
		MintedAt:           currTime,
		DeviceDefinitionID: null.StringFrom("honda_accord_2020"),
	}
	vehicle4ImageURL, err := DefaultImageURI(o.settings.BaseImageURL, hondaAccordTID4.ID)
	o.Require().NoError(err)
	vehicle4DataURI, err := GetVehicleDataURI(o.settings.BaseVehicleDataURI, hondaAccordTID4.ID)
	o.Require().NoError(err)
	vehicle4AsAPI, err := o.repo.ToAPI(&hondaAccordTID4, vehicle4ImageURL, vehicle4DataURI, nil)
	o.Require().NoError(err)

	vehicles := []models.Vehicle{toyotaCamryTID1, hondaCivicTID2, toyotaRav4TID3, hondaAccordTID4}
	first := len(vehicles)
	for _, v := range vehicles {
		if err := v.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.Require().NoError(err)
		}
	}

	privileges := []models.Privilege{
		{
			TokenID:     toyotaRav4TID3.ID,
			PrivilegeID: 1,
			UserAddress: wallet1.Bytes(),
			SetAt:       currTime,
			ExpiresAt:   currTime.Add(time.Hour),
		},
	}

	for _, p := range privileges {
		if err := p.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			o.NoError(err)
		}
	}

	// create table of tests for testing the o.repo.GetVehicles function
	tests := []struct {
		name    string
		filter  *gmodel.VehiclesFilter
		results []*gmodel.VehicleEdge
	}{
		{
			name:   "No filters",
			filter: nil,
			results: []*gmodel.VehicleEdge{
				{Node: vehicle1AsAPI},
				{Node: vehicle2AsAPI},
				{Node: vehicle3AsAPI},
				{Node: vehicle4AsAPI},
			},
		},
		{
			name:   "Empty filters",
			filter: &gmodel.VehiclesFilter{},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle1AsAPI},
				{Node: vehicle2AsAPI},
				{Node: vehicle3AsAPI},
				{Node: vehicle4AsAPI},
			},
		},
		{
			name: "Filter by owner",
			filter: &gmodel.VehiclesFilter{
				Owner: wallet1,
			},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle1AsAPI},
				{Node: vehicle2AsAPI},
			},
		},
		{
			name: "Filter by Privileged",
			filter: &gmodel.VehiclesFilter{
				Privileged: wallet1,
			},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle1AsAPI},
				{Node: vehicle2AsAPI},
				{Node: vehicle3AsAPI},
			},
		},
		{
			name: "Filter by Make",
			filter: &gmodel.VehiclesFilter{
				Make: toyota.Ptr(),
			},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle1AsAPI},
				{Node: vehicle3AsAPI},
			},
		},
		{
			name: "Filter by Model",
			filter: &gmodel.VehiclesFilter{
				Model: camry.Ptr(),
			},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle1AsAPI},
			},
		},
		{
			name: "Filter by Year",
			filter: &gmodel.VehiclesFilter{
				Year: year2022.Ptr(),
			},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle2AsAPI},
				{Node: vehicle3AsAPI},
			},
		},
		{
			name: "Filter by Manufacturer",
			filter: &gmodel.VehiclesFilter{
				ManufacturerTokenID: &toyotaMfr.ID,
			},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle1AsAPI},
				{Node: vehicle3AsAPI},
			},
		},
		{
			name: "Filter by Owner and Privileged same address",
			filter: &gmodel.VehiclesFilter{
				Privileged: wallet1,
				Owner:      wallet1,
			},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle1AsAPI},
				{Node: vehicle2AsAPI},
			},
		},
		{
			name: "Filter by Owner and Privileged different addresses",
			filter: &gmodel.VehiclesFilter{
				Privileged: wallet1,
				Owner:      wallet2,
			},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle3AsAPI},
			},
		},
		{
			name: "Filter by Privileged and Make",
			filter: &gmodel.VehiclesFilter{
				Privileged: wallet1,
				Make:       toyota.Ptr(),
			},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle1AsAPI},
				{Node: vehicle3AsAPI},
			},
		},
		{
			name: "Filter by Owner and Manufacturer",
			filter: &gmodel.VehiclesFilter{
				Owner:               wallet2,
				ManufacturerTokenID: &toyotaMfr.ID,
			},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle3AsAPI},
			},
		},
		{
			name: "Filter by Privileged and Model",
			filter: &gmodel.VehiclesFilter{
				Privileged: wallet1,
				Model:      camry.Ptr(),
			},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle1AsAPI},
			},
		},
		{
			name: "Filter by Privileged and Year",
			filter: &gmodel.VehiclesFilter{
				Privileged: wallet1,
				Year:       year2022.Ptr(),
			},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle2AsAPI},
				{Node: vehicle3AsAPI},
			},
		},
		{
			name: "Filter No Results",
			filter: &gmodel.VehiclesFilter{
				Owner:      wallet1,
				Privileged: wallet2,
			},
			results: []*gmodel.VehicleEdge{},
		},
		{
			name: "Filter by Device Definition ID",
			filter: &gmodel.VehiclesFilter{
				DeviceDefinitionID: &toyotaCamryTID1.DeviceDefinitionID.String,
			},
			results: []*gmodel.VehicleEdge{
				{Node: vehicle1AsAPI},
				{Node: vehicle3AsAPI},
			},
		},
	}

	for i := range tests {
		tt := tests[i]
		o.Run(tt.name, func() {
			res, err := o.repo.GetVehicles(o.ctx, &first, nil, nil, nil, tt.filter)
			o.Require().NoError(err)
			o.Require().NotNil(res)
			requireEqualVehicles(o.T(), tt.results, res.Edges) // vehicles aren't matching up, got Mazda, expected Honda
			o.Require().Equalf(len(tt.results), res.TotalCount, "Test %s: expected total count to be %d, got %d", tt.name, len(tt.results), res.TotalCount)
		})
	}
}

// requireEqualVehicles is a helper function to compare two slices of VehicleEdges
func requireEqualVehicles(t *testing.T, expected, actual []*gmodel.VehicleEdge) {
	t.Helper()
	require.Len(t, actual, len(expected))
	slices.SortFunc(expected, func(a, b *gmodel.VehicleEdge) int {
		return cmp.Compare(a.Node.ID, b.Node.ID) * -1
	})
	for i := range expected {
		require.Equal(t, expected[i].Node, actual[i].Node)
	}
}

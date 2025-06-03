package vehicle

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/mnemonic"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type OwnedVehiclesRepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container *postgres.PostgresContainer
	repo      *Repository
	settings  config.Settings
}

const migrationsDir = "../../../migrations"

func (s *OwnedVehiclesRepoTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = helpers.StartContainerDatabase(s.ctx, s.T(), migrationsDir)

	s.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
		VehicleNFTAddr:      "0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8",
		BaseImageURL:        "https://mockUrl.com/v1",
		BaseVehicleDataURI:  "https://dimoData/vehicles/",
	}
	logger := zerolog.Nop()
	s.repo = New(base.NewRepository(s.pdb, s.settings, &logger))
}

// TearDownTest after each test truncate tables
func (s *OwnedVehiclesRepoTestSuite) TearDownTest() {
	s.Require().NoError(s.container.Restore(s.ctx))
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

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	currTime := time.Now().UTC().Truncate(time.Second)
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
			Year:           null.IntFrom(2022),
			MintedAt:       currTime,
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
				ID:             "V_kQI=",
				TokenID:        2,
				TokenDid:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:2",
				ManufacturerID: 131,
				Owner:          common.BytesToAddress(wallet.Bytes()),
				OwnerDid:       "did:ethr:80001:" + common.BytesToAddress(wallet.Bytes()).Hex(),
				Definition: &gmodel.Definition{
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
				MintedAt:   vehicles[1].MintedAt,
				Privileges: nil,
				Image:      "https://mockUrl.com/v1/vehicle/2/image",
				ImageURI:   "https://mockUrl.com/v1/vehicle/2/image",
				DataURI:    "https://dimoData/vehicles/2",
			},
			Cursor: "Mg==",
		},
		{
			Node: &gmodel.Vehicle{
				ID:             "V_kQE=",
				TokenID:        1,
				TokenDid:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:1",
				ManufacturerID: 131,
				Owner:          common.BytesToAddress(wallet.Bytes()),
				OwnerDid:       "did:ethr:80001:" + common.BytesToAddress(wallet.Bytes()).Hex(),
				Definition: &gmodel.Definition{
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
				MintedAt:   vehicles[0].MintedAt,
				Privileges: nil,
				Image:      "https://mockUrl.com/v1/vehicle/1/image",
				ImageURI:   "https://mockUrl.com/v1/vehicle/1/image",
				DataURI:    "https://dimoData/vehicles/1",
			},
			Cursor: "MQ==",
		},
	}

	for _, veh := range expected {
		veh.Node.Name = strings.Join(mnemonic.FromInt32WithObfuscation(int32(veh.Node.TokenID)), " ")
	}

	s.Exactly(expected, res.Edges)
}

func (s *OwnedVehiclesRepoTestSuite) Test_GetOwnedVehicles_Pagination() {
	_, wallet, err := helpers.GenerateWallet()
	s.NoError(err)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	currTime := time.Now().UTC().Truncate(time.Second)
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
			Year:           null.IntFrom(2022),
			MintedAt:       currTime,
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
				ID:             "V_kQI=",
				TokenID:        2,
				TokenDid:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:2",
				ManufacturerID: 131,
				Owner:          common.BytesToAddress(wallet.Bytes()),
				OwnerDid:       "did:ethr:80001:" + common.BytesToAddress(wallet.Bytes()).Hex(),
				Definition: &gmodel.Definition{
					Make:  &vehicles[1].Make.String,
					Model: &vehicles[1].Model.String,
					Year:  &vehicles[1].Year.Int,
				},
				MintedAt:   vehicles[1].MintedAt,
				Privileges: nil,
				Image:      "https://mockUrl.com/v1/vehicle/2/image",
				ImageURI:   "https://mockUrl.com/v1/vehicle/2/image",
				DataURI:    "https://dimoData/vehicles/2",
			},
			Cursor: "Mg==",
		},
	}

	for _, veh := range expected {
		veh.Node.Name = strings.Join(mnemonic.FromInt32WithObfuscation(int32(veh.Node.TokenID)), " ")
	}

	s.Exactly(expected, res.Edges)
}

func (s *OwnedVehiclesRepoTestSuite) Test_GetOwnedVehicles_Pagination_NextPage() {
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
		s.NoError(err)
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
			Year:           null.IntFrom(2022),
			MintedAt:       currTime,
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
				ID:             "V_kQE=",
				ManufacturerID: 131,
				TokenID:        1,
				TokenDid:       "did:erc721:80001:0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8:1",
				Owner:          common.BytesToAddress(wallet.Bytes()),
				OwnerDid:       "did:ethr:80001:" + common.BytesToAddress(wallet.Bytes()).Hex(),
				Definition: &gmodel.Definition{
					Make:  &vehicles[0].Make.String,
					Model: &vehicles[0].Model.String,
					Year:  &vehicles[0].Year.Int,
				},
				MintedAt:   vehicles[0].MintedAt,
				Privileges: nil,
				ImageURI:   "https://mockUrl.com/v1/vehicle/1/image",
				Image:      "https://mockUrl.com/v1/vehicle/1/image",
				DataURI:    "https://dimoData/vehicles/1",
			},
			Cursor: "MQ==",
		},
	}

	for _, veh := range expected {
		veh.Node.Name = strings.Join(mnemonic.FromInt32WithObfuscation(int32(veh.Node.TokenID)), " ")
	}

	s.Exactly(expected, res.Edges)
}

func Test_GetOwnedVehicles_Filters(t *testing.T) {
	// Vehicle | Owner | Privileged Users
	// --------+-------+-----------------
	// 1       | A     | B
	// 2       | B     | C
	// 3       | A     |
	ctx := context.Background()
	assert := assert.New(t)
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	logger := zerolog.Nop()
	repo := New(base.NewRepository(pdb, config.Settings{}, &logger))
	_, walletA, err := helpers.GenerateWallet()
	assert.NoError(err)
	_, walletB, err := helpers.GenerateWallet()
	assert.NoError(err)
	_, walletC, err := helpers.GenerateWallet()
	assert.NoError(err)

	data := []struct {
		TokenID    int
		Owner      *common.Address
		Privileged *common.Address
	}{
		{
			TokenID:    1,
			Owner:      walletA,
			Privileged: walletB,
		},
		{
			TokenID:    2,
			Owner:      walletB,
			Privileged: walletC,
		},
		{
			TokenID: 3,
			Owner:   walletA,
		},
	}

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(err)
	}

	for _, v := range data {
		vehicle := models.Vehicle{
			ID:             v.TokenID,
			ManufacturerID: 131,
			OwnerAddress:   v.Owner.Bytes(),
			Make:           null.StringFrom("Toyota"),
			Model:          null.StringFrom("Camry"),
			Year:           null.IntFrom(2022),
			MintedAt:       time.Now(),
		}
		if err := vehicle.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			assert.NoError(err)
		}

		if v.Privileged != nil {
			privileges := models.Privilege{
				TokenID:     v.TokenID,
				PrivilegeID: 1,
				UserAddress: v.Privileged.Bytes(),
				SetAt:       time.Now(),
				ExpiresAt:   time.Now().Add(5 * time.Hour),
			}

			if err := privileges.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
				assert.NoError(err)
			}
		}
	}

	for _, testCases := range []struct {
		Name           string
		Description    string
		First          int
		Filter         gmodel.VehiclesFilter
		ExpectedTotal  int
		ExpectedTokens []int
		ExpectedOwner  []*common.Address
	}{
		{
			Description: "Filter By Privileged: priv wallet has both implied (owner) and explicit (granted) privileges",
			// Owner | Privileged | Result
			// ------+------------+-------
			// 	     | B          | 1, 2
			First:          3,
			Filter:         gmodel.VehiclesFilter{Privileged: walletB},
			ExpectedTotal:  2,
			ExpectedTokens: []int{2, 1},
			ExpectedOwner:  []*common.Address{walletB, walletA},
		},
		{
			Description: "Filter By Owner",
			// Owner | Privileged | Result
			// ------+------------+-------
			// A     |            | 1, 3
			First:          3,
			Filter:         gmodel.VehiclesFilter{Owner: walletA},
			ExpectedTotal:  2,
			ExpectedTokens: []int{3, 1},
			ExpectedOwner:  []*common.Address{walletA, walletA},
		},
		{
			Description: "Filter By Privileged & Owner, result must match both criteria (valid combo)",
			// Owner | Privileged | Result
			// ------+------------+-------
			// A     | B          | 1
			First:          3,
			Filter:         gmodel.VehiclesFilter{Owner: walletA, Privileged: walletB},
			ExpectedTotal:  1,
			ExpectedTokens: []int{1},
			ExpectedOwner:  []*common.Address{walletA},
		},
		{
			Description: "Filter By Privileged & Owner, result must match both criteria (invalid combo)",
			// Owner | Privileged | Result
			// ------+------------+-------
			// C     | B          |
			First:          3,
			Filter:         gmodel.VehiclesFilter{Owner: walletC, Privileged: walletB},
			ExpectedTotal:  0,
			ExpectedTokens: []int{},
			ExpectedOwner:  []*common.Address{},
		},
	} {
		res, err := repo.GetVehicles(ctx, &testCases.First, nil, nil, nil, &testCases.Filter)
		assert.NoError(err)

		assert.Equal(testCases.ExpectedTotal, res.TotalCount)
		for idx, tknID := range testCases.ExpectedTokens {
			assert.Equal(tknID, res.Edges[idx].Node.TokenID)
		}
		for idx, addr := range testCases.ExpectedOwner {
			assert.Equal(*addr, res.Edges[idx].Node.Owner)
		}
	}
}

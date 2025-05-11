package synthetic

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var (
	toyota   = null.StringFrom("Toyota")
	tacoma   = null.StringFrom("Tacoma")
	honda    = null.StringFrom("Honda")
	civic    = null.StringFrom("Civic")
	year2020 = null.IntFrom(2020)
	year2023 = null.IntFrom(2023)
)

func Test_SyntheticDeviceToAPI(t *testing.T) {
	t.Parallel()
	_, wallet, err := helpers.GenerateWallet()
	require.NoError(t, err)

	currTime := time.Now()
	sd := &models.SyntheticDevice{
		ID:            1,
		IntegrationID: 2,
		VehicleID:     1,
		DeviceAddress: wallet.Bytes(),
		MintedAt:      currTime,
	}

	res, err := ToAPI(sd)
	require.NoError(t, err)

	encodedID, err := base.EncodeGlobalTokenID(TokenPrefix, sd.ID)
	require.NoError(t, err)

	assert.Exactly(t, &model.SyntheticDevice{
		ID:            encodedID,
		Name:          "learn island zoo",
		TokenID:       1,
		IntegrationID: 2,
		Address:       *wallet,
		MintedAt:      currTime,
		VehicleID:     sd.VehicleID,
	}, res)
}

type SyntheticTestSuite struct {
	suite.Suite
	pdb       db.Store
	container *postgres.PostgresContainer
	repo      *Repository
	settings  config.Settings

	// Test data
	toyota     models.Manufacturer
	honda      models.Manufacturer
	vehicle1   models.Vehicle
	vehicle2   models.Vehicle
	vehicle3   models.Vehicle
	synthetic1 models.SyntheticDevice
	synthetic2 models.SyntheticDevice
	synthetic3 models.SyntheticDevice
}

func TestSyntheticTestSuite(t *testing.T) {
	suite.Run(t, new(SyntheticTestSuite))
}

func (s *SyntheticTestSuite) SetupSuite() {
	ctx := context.Background()
	s.pdb, s.container = helpers.StartContainerDatabase(ctx, s.T(), "../../../migrations")

	s.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
		BaseImageURL:        "https://mockUrl.com/v1",
		BaseVehicleDataURI:  "https://dimoData/vehicles/",
	}
	logger := zerolog.New(os.Stderr)
	s.repo = &Repository{base.NewRepository(s.pdb, s.settings, &logger)}

	// Create test data
	currTime := time.Now().UTC().Truncate(time.Second)

	_, vehicle1And2Owner, err := helpers.GenerateWallet()
	s.Require().NoError(err)
	_, vehicle3Owner, err := helpers.GenerateWallet()
	s.Require().NoError(err)
	_, synthetic1Addr, err := helpers.GenerateWallet()
	s.Require().NoError(err)
	_, synthetic2Addr, err := helpers.GenerateWallet()
	s.Require().NoError(err)
	_, synthetic3Addr, err := helpers.GenerateWallet()
	s.Require().NoError(err)

	s.toyota = models.Manufacturer{
		ID:    131,
		Name:  "Toyota",
		Owner: vehicle1And2Owner.Bytes(),
		Slug:  "toyota",
	}

	s.honda = models.Manufacturer{
		ID:    48,
		Name:  "Honda",
		Owner: vehicle1And2Owner.Bytes(),
		Slug:  "honda",
	}

	s.vehicle1 = models.Vehicle{
		ID:             1,
		OwnerAddress:   vehicle1And2Owner.Bytes(),
		ManufacturerID: 131,
		Make:           toyota,
		Model:          tacoma,
		Year:           year2023,
		MintedAt:       currTime,
	}

	s.vehicle2 = models.Vehicle{
		ManufacturerID: 48,
		ID:             2,
		OwnerAddress:   vehicle1And2Owner.Bytes(),
		Make:           honda,
		Model:          civic,
		Year:           year2020,
		MintedAt:       currTime,
	}

	s.vehicle3 = models.Vehicle{
		ManufacturerID: 48,
		ID:             3,
		OwnerAddress:   vehicle3Owner.Bytes(),
		Make:           honda,
		Model:          civic,
		Year:           year2020,
		MintedAt:       currTime,
	}

	s.synthetic1 = models.SyntheticDevice{
		ID:            1,
		IntegrationID: 1,
		VehicleID:     s.vehicle1.ID,
		DeviceAddress: synthetic1Addr.Bytes(),
		MintedAt:      currTime,
	}

	s.synthetic2 = models.SyntheticDevice{
		ID:            2,
		IntegrationID: 2,
		VehicleID:     s.vehicle2.ID,
		DeviceAddress: synthetic2Addr.Bytes(),
		MintedAt:      currTime,
	}

	s.synthetic3 = models.SyntheticDevice{
		ID:            3,
		IntegrationID: 1,
		VehicleID:     s.vehicle3.ID,
		DeviceAddress: synthetic3Addr.Bytes(),
		MintedAt:      currTime,
	}

}

// TearDownTest after each test truncate tables.
func (s *SyntheticTestSuite) TearDownTest() {
	s.Require().NoError(s.container.Restore(context.TODO()))
}

// TearDownSuite cleanup at end by terminating container.
func (s *SyntheticTestSuite) TearDownSuite() {
	err := s.container.Terminate(context.Background())
	s.Require().NoErrorf(err, "failed to terminate container: %v", err)
}

func (s *SyntheticTestSuite) Test_GetSyntheticDevices() {
	ctx := context.Background()
	var err error
	// insert test data into the database
	err = s.toyota.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)
	err = s.honda.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)
	err = s.vehicle1.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)
	err = s.vehicle2.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)
	err = s.vehicle3.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)
	err = s.synthetic1.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)
	err = s.synthetic2.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)
	err = s.synthetic3.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	s1API, err := ToAPI(&s.synthetic1)
	s.Require().NoError(err)
	s2API, err := ToAPI(&s.synthetic2)
	s.Require().NoError(err)
	s3API, err := ToAPI(&s.synthetic3)
	s.Require().NoError(err)

	tests := []struct {
		name               string
		filter             *model.SyntheticDevicesFilter
		results            []*model.SyntheticDevice
		expectedTotalCount *int
		first              *int
		last               *int
		after              *string
		before             *string
	}{
		{
			name: "No params",
			results: []*model.SyntheticDevice{
				s3API,
				s2API,
				s1API,
			},
			first: Ptr(10),
		},
		{
			name:               "Total Count",
			results:            nil,
			first:              Ptr(0),
			expectedTotalCount: Ptr(3),
		},
		{
			name: "first 2",
			results: []*model.SyntheticDevice{
				s3API,
				s2API,
			},
			first: Ptr(2),
		},
		{
			name: "last 2",
			results: []*model.SyntheticDevice{
				s2API,
				s1API,
			},
			last: Ptr(2),
		},
		{
			name: "Filter by owner",
			filter: &model.SyntheticDevicesFilter{
				Owner: Ptr(common.BytesToAddress(s.vehicle1.OwnerAddress)),
			},
			results: []*model.SyntheticDevice{
				s2API,
				s1API,
			},
			first: Ptr(10),
		},
		{
			name: "Filter by integrationID",
			filter: &model.SyntheticDevicesFilter{
				IntegrationID: Ptr(1),
			},
			results: []*model.SyntheticDevice{
				s3API,
				s1API,
			},
			first: Ptr(10),
		},
		{
			name: "Filter by owner and integrationID",
			filter: &model.SyntheticDevicesFilter{
				Owner:         Ptr(common.BytesToAddress(s.vehicle1.OwnerAddress)),
				IntegrationID: Ptr(1),
			},
			results: []*model.SyntheticDevice{
				s1API,
			},
			first: Ptr(10),
		},
	}
	for i := range tests {
		test := tests[i]
		s.Run(test.name, func() {
			actual, err := s.repo.GetSyntheticDevices(ctx, test.first, test.last, test.after, test.before, test.filter)
			if test.expectedTotalCount != nil {
				require.Equal(s.T(), *test.expectedTotalCount, actual.TotalCount)
			}
			s.Require().NoError(err, "failed to get synthetic devices")
			requireEqualSyntheticDevices(s.T(), test.results, actual.Nodes)
		})
	}
}

func Ptr[T any](t T) *T {
	return &t
}

// Test_GetSyntheticDevice tests the GetSyntheticDevice method.
func requireEqualSyntheticDevices(t *testing.T, expected, actual []*model.SyntheticDevice) {
	t.Helper()
	require.Len(t, actual, len(expected))
	for i := range expected {
		require.Equal(t, expected[i], actual[i])
	}
}

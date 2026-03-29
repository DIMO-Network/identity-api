package loader

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/repositories/vehicle"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubVehicleDefinitionFetcher struct {
	docs map[string]*services.DeviceDefinitionDoc
	errs map[string]error
}

func (s stubVehicleDefinitionFetcher) GetVehicleDefinitionDoc(_ context.Context, vehicleDID string) (*services.DeviceDefinitionDoc, error) {
	if err := s.errs[vehicleDID]; err != nil {
		return nil, err
	}

	return s.docs[vehicleDID], nil
}

func TestVehicleLoaderBatchGetVehicleByID_OverridesAndPreservesOrder(t *testing.T) {
	ctx := context.Background()
	pdb, cont := helpers.StartContainerDatabase(ctx, t, migrationsDir)
	defer cont.Terminate(t.Context()) //nolint

	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)
	settings := config.Settings{
		DIMORegistryChainID: 80001,
		VehicleNFTAddr:      "0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8",
		BaseImageURL:        "https://images.example",
		BaseVehicleDataURI:  "https://data.example/vehicles",
	}
	baseRepo := base.NewRepository(pdb, settings, &logger)
	repo := vehicle.New(baseRepo)

	mfr := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}
	require.NoError(t, mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

	veh1 := &models.Vehicle{
		ID:             1,
		ManufacturerID: 131,
		OwnerAddress:   common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Make:           null.StringFrom("DB Make 1"),
		Model:          null.StringFrom("DB Model 1"),
		Year:           null.IntFrom(2020),
		MintedAt:       time.Now(),
	}
	veh2 := &models.Vehicle{
		ID:             2,
		ManufacturerID: 131,
		OwnerAddress:   common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Make:           null.StringFrom("DB Make 2"),
		Model:          null.StringFrom("DB Model 2"),
		Year:           null.IntFrom(2021),
		MintedAt:       time.Now(),
	}
	require.NoError(t, veh1.Insert(ctx, pdb.DBS().Writer, boil.Infer()))
	require.NoError(t, veh2.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

	veh1API, err := repo.ToAPI(veh1, "https://images.example/vehicle/1/image", "https://data.example/vehicles/1")
	require.NoError(t, err)
	veh2API, err := repo.ToAPI(veh2, "https://images.example/vehicle/2/image", "https://data.example/vehicles/2")
	require.NoError(t, err)

	loader := NewVehicleLoader(repo, stubVehicleDefinitionFetcher{
		docs: map[string]*services.DeviceDefinitionDoc{
			veh1API.TokenDID: {ID: "fetch-1", Make: "Fetch Make 1", Model: "Fetch Model 1", Year: 2030},
			veh2API.TokenDID: {ID: "fetch-2", Make: "Fetch Make 2", Model: "Fetch Model 2", Year: 2031},
		},
	}, &logger)

	results := loader.BatchGetVehicleByID(ctx, []int{2, 1, 2})
	require.Len(t, results, 3)

	require.NoError(t, results[0].Error)
	assert.Equal(t, 2, results[0].Data.TokenID)
	assert.Equal(t, "fetch-2", *results[0].Data.Definition.ID)
	assert.Equal(t, "Fetch Make 2", *results[0].Data.Definition.Make)
	assert.Equal(t, "Fetch Model 2", *results[0].Data.Definition.Model)
	assert.Equal(t, 2031, *results[0].Data.Definition.Year)

	require.NoError(t, results[1].Error)
	assert.Equal(t, 1, results[1].Data.TokenID)
	assert.Equal(t, "fetch-1", *results[1].Data.Definition.ID)
	assert.Equal(t, "Fetch Make 1", *results[1].Data.Definition.Make)
	assert.Equal(t, "Fetch Model 1", *results[1].Data.Definition.Model)
	assert.Equal(t, 2030, *results[1].Data.Definition.Year)

	require.NoError(t, results[2].Error)
	assert.Equal(t, 2, results[2].Data.TokenID)
	assert.Equal(t, "fetch-2", *results[2].Data.Definition.ID)
}

func TestVehicleLoaderBatchGetVehicleByID_FallsBackOnFetchError(t *testing.T) {
	ctx := context.Background()
	pdb, cont := helpers.StartContainerDatabase(ctx, t, migrationsDir)
	defer cont.Terminate(t.Context()) //nolint

	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)
	settings := config.Settings{
		DIMORegistryChainID: 80001,
		VehicleNFTAddr:      "0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8",
		BaseImageURL:        "https://images.example",
		BaseVehicleDataURI:  "https://data.example/vehicles",
	}
	baseRepo := base.NewRepository(pdb, settings, &logger)
	repo := vehicle.New(baseRepo)

	mfr := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}
	require.NoError(t, mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

	veh := &models.Vehicle{
		ID:             1,
		ManufacturerID: 131,
		OwnerAddress:   common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Make:           null.StringFrom("DB Make"),
		Model:          null.StringFrom("DB Model"),
		Year:           null.IntFrom(2020),
		MintedAt:       time.Now(),
	}
	require.NoError(t, veh.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

	vehAPI, err := repo.ToAPI(veh, "https://images.example/vehicle/1/image", "https://data.example/vehicles/1")
	require.NoError(t, err)

	loader := NewVehicleLoader(repo, stubVehicleDefinitionFetcher{
		errs: map[string]error{vehAPI.TokenDID: errors.New("boom")},
	}, &logger)

	results := loader.BatchGetVehicleByID(ctx, []int{1})
	require.Len(t, results, 1)
	require.NoError(t, results[0].Error)
	assert.Equal(t, "DB Make", *results[0].Data.Definition.Make)
	assert.Equal(t, "DB Model", *results[0].Data.Definition.Model)
	assert.Equal(t, 2020, *results[0].Data.Definition.Year)
	assert.Nil(t, results[0].Data.Definition.ID)
	assert.True(t, strings.Contains(logBuf.String(), "fetch-api definition lookup failed, using DB values"))
}

package devicedefinition

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const migrationsDir = "../../../migrations"

// manifest with three BMW definitions and one Alfa Romeo; note metadata ""
// tolerance for backfilled legacy rows.
const manifestBody = `{
  "updatedAt": "2026-08-19T00:00:00.000Z",
  "count": 4,
  "definitions": [
    {
      "id": "alfa-romeo_147_2007",
      "ksuid": "26G3iFH7Xc9Wvsw7pg6sD7uzoSS",
      "model": "147",
      "year": 2007,
      "devicetype": "vehicle",
      "imageuri": "https://image",
      "metadata": {"device_attributes": [{"name": "powertrain_type", "value": "ICE"}]},
      "manufacturer": {"tokenId": 137, "slug": "alfa-romeo", "name": "Alfa Romeo"}
    },
    {
      "id": "bmw_x5_2019",
      "ksuid": "12G3iFH7Xc9Wvsw7pg6sD7uzoKK",
      "model": "X5",
      "year": 2019,
      "devicetype": "vehicle",
      "imageuri": "https://image",
      "metadata": "",
      "manufacturer": {"tokenId": 13, "slug": "bmw", "name": "BMW"}
    },
    {
      "id": "bmw_x6_2019",
      "ksuid": "12G3iFH7Xc9Wvsw7pg6sD7uzoLL",
      "model": "X6",
      "year": 2019,
      "devicetype": "vehicle",
      "imageuri": "https://image",
      "metadata": null,
      "manufacturer": {"tokenId": 13, "slug": "bmw", "name": "BMW"}
    },
    {
      "id": "bmw_x7_2020",
      "ksuid": "12G3iFH7Xc9Wvsw7pg6sD7uzoMM",
      "model": "X7",
      "year": 2020,
      "devicetype": "vehicle",
      "imageuri": "https://image",
      "metadata": null,
      "manufacturer": {"tokenId": 13, "slug": "bmw", "name": "BMW"}
    }
  ]
}`

func startCatalog(t *testing.T) *services.DefinitionsCatalogService {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/manifest.json", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(manifestBody))
	}))
	t.Cleanup(srv.Close)
	logger := zerolog.Nop()
	return services.NewDefinitionsCatalogService(&logger, &config.Settings{DefinitionsCatalogURL: srv.URL})
}

func Test_GetDeviceDefinitions_Query(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	mfr := models.Manufacturer{
		ID:    137,
		Name:  "Alfa Romeo",
		Owner: common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:  "alfa-romeo",
	}
	require.NoError(t, mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer()))
	mfr2 := models.Manufacturer{
		ID:    13,
		Name:  "BMW",
		Owner: common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:  "bmw",
	}
	require.NoError(t, mfr2.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

	logger := zerolog.Nop()
	repo := base.NewRepository(pdb, config.Settings{}, &logger)
	adController := New(repo, startCatalog(t))

	// Last 2 BMW definitions before the cursor for bmw_x7_2020.
	last := 2
	before := idToCursor("bmw_x7_2020")

	res, err := adController.GetDeviceDefinitions(ctx, 13, nil, nil, &last, &before, &model.DeviceDefinitionFilter{})
	require.NoError(t, err)

	assert.Len(t, res.Edges, 2)
	assert.Equal(t, 3, res.TotalCount)

	assert.Equal(t, "bmw_x5_2019", res.Edges[0].Node.DeviceDefinitionID)
	assert.Equal(t, "12G3iFH7Xc9Wvsw7pg6sD7uzoKK", *res.Edges[0].Node.LegacyID)
	assert.Equal(t, "vehicle", *res.Edges[0].Node.DeviceType)
	assert.Equal(t, "BMW", res.Edges[0].Node.Manufacturer.Name)
	assert.Equal(t, 13, res.Edges[0].Node.Manufacturer.TokenID)
	assert.Equal(t, "bmw_x6_2019", res.Edges[1].Node.DeviceDefinitionID)

	// Year filter.
	first := 10
	res, err = adController.GetDeviceDefinitions(ctx, 13, &first, nil, nil, nil, &model.DeviceDefinitionFilter{Year: &[]int{2020}[0]})
	require.NoError(t, err)
	assert.Equal(t, 1, res.TotalCount)
	assert.Equal(t, "bmw_x7_2020", res.Nodes[0].DeviceDefinitionID)
}

func Test_GetDeviceDefinition_Query(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	mfr := models.Manufacturer{
		ID:    137,
		Name:  "Alfa Romeo",
		Owner: common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:  "alfa-romeo",
	}
	require.NoError(t, mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

	logger := zerolog.Nop()
	repo := base.NewRepository(pdb, config.Settings{}, &logger)
	adController := New(repo, startCatalog(t))

	res, err := adController.GetDeviceDefinition(ctx, model.DeviceDefinitionBy{ID: "alfa-romeo_147_2007"})
	require.NoError(t, err)

	assert.Equal(t, "alfa-romeo_147_2007", res.DeviceDefinitionID)
	assert.Equal(t, "26G3iFH7Xc9Wvsw7pg6sD7uzoSS", *res.LegacyID)
	assert.Equal(t, "vehicle", *res.DeviceType)
	assert.Equal(t, "https://image", *res.ImageURI)
	assert.Equal(t, "Alfa Romeo", res.Manufacturer.Name)
	assert.Equal(t, 137, res.Manufacturer.TokenID)
	require.Len(t, res.Attributes, 1)
	assert.Equal(t, "powertrain_type", res.Attributes[0].Name)
	assert.Equal(t, "ICE", res.Attributes[0].Value)
}

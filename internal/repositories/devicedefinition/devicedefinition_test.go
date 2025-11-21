package devicedefinition

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jarcoal/httpmock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const migrationsDir = "../../../migrations"

func Test_GetDeviceDefinitions_Query(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	mfr := models.Manufacturer{
		ID:      137,
		Name:    "Alfa Romeo",
		Owner:   common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:    "alfa-romeo",
		TableID: null.IntFrom(1),
	}
	err := mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	mfr2 := models.Manufacturer{
		ID:      13,
		Name:    "BMW",
		Owner:   common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:    "bmw",
		TableID: null.IntFrom(2),
	}
	err = mfr2.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	logger := zerolog.Nop()

	const baseURL = "http://local"

	repo := base.NewRepository(pdb, config.Settings{DIMORegistryChainID: 30001, TablelandAPIGateway: baseURL}, &logger)

	tablelandAPI := services.NewTablelandApiService(&logger, &config.Settings{
		TablelandAPIGateway: baseURL,
	})

	adController := New(repo, tablelandAPI)
	last := 2
	before := "MQ=="

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	countURL := "api/v1/query?statement=SELECT+COUNT%28%2A%29+FROM+%22_30001_1%22"
	respCountBody := `[{"count(*)": 4}]`
	var modelCountTablelandResponse []DeviceDefinitionTablelandCountModel
	_ = json.Unmarshal([]byte(respCountBody), &modelCountTablelandResponse)

	httpmock.RegisterResponder(http.MethodGet, baseURL+countURL, httpmock.NewStringResponder(200, respCountBody))

	queryURL := "api/v1/query?statement=SELECT+%2A+FROM+%22_30001_1%22+WHERE+%28%22id%22+%3C+%271%27%29+ORDER+BY+%22id%22+DESC+LIMIT+3"
	respQueryBody := `[
	  {
		"id": "alfa-romeo_147_2007",
		"deviceType": "vehicle",
		"imageURI": "https://image",
		"ksuid": "26G3iFH7Xc9Wvsw7pg6sD7uzoSS",
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
		"id": "bmw_x5_2019",
		"deviceType": "vehicle",
		"imageURI": "https://image",
		"ksuid": "12G3iFH7Xc9Wvsw7pg6sD7uzoKK",
		"metadata": ""
	  }
	]`
	// when we query against tableland, if metadata is not set, it is returned as an empty string ""
	var modelQueryTablelandResponse []DeviceDefinitionTablelandModel
	errUm := json.Unmarshal([]byte(respQueryBody), &modelQueryTablelandResponse)
	require.NoError(t, errUm)

	httpmock.RegisterResponder(http.MethodGet, baseURL+queryURL, httpmock.NewStringResponder(200, respQueryBody))

	res, err := adController.GetDeviceDefinitions(ctx, &mfr.TableID.Int, nil, nil, &last, &before, &model.DeviceDefinitionFilter{})
	require.NoError(t, err)

	assert.Len(t, res.Edges, 2)
	assert.Equal(t, res.TotalCount, 4)

	deviceType := "vehicle"
	legacyID1 := "26G3iFH7Xc9Wvsw7pg6sD7uzoSS"
	legacyID2 := "12G3iFH7Xc9Wvsw7pg6sD7uzoKK"

	expected := []*model.DeviceDefinitionEdge{
		{
			Node: &model.DeviceDefinition{
				DeviceDefinitionID: "bmw_x5_2019",
				LegacyID:           &legacyID2,
				DeviceType:         &deviceType,
				Manufacturer: &model.Manufacturer{
					Name:    "BMW",
					TokenID: 13,
				},
			},
			Cursor: "Mg==",
		},
		{
			Node: &model.DeviceDefinition{
				DeviceDefinitionID: "alfa-romeo_147_2007",
				LegacyID:           &legacyID1,
				DeviceType:         &deviceType,
				Manufacturer: &model.Manufacturer{
					Name:    "Alfa Romeo",
					TokenID: 137,
				},
			},
			Cursor: "Mw==",
		},
	}

	for i, e := range expected {
		assert.Equal(t, e.Node.DeviceDefinitionID, res.Edges[i].Node.DeviceDefinitionID)
		assert.Equal(t, e.Node.LegacyID, res.Edges[i].Node.LegacyID)
		assert.Equal(t, e.Node.DeviceType, res.Edges[i].Node.DeviceType)
		assert.Equal(t, e.Node.Manufacturer.Name, res.Edges[i].Node.Manufacturer.Name)
		assert.Equal(t, e.Node.Manufacturer.TokenID, res.Edges[i].Node.Manufacturer.TokenID)
	}
}

func Test_GetDeviceDefinition_Query(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	mfr := models.Manufacturer{
		ID:      137,
		Name:    "Toyota",
		Owner:   common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:    "toyota",
		TableID: null.IntFrom(1),
	}
	err := mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	logger := zerolog.Nop()

	const baseURL = "http://local"

	repo := base.NewRepository(pdb, config.Settings{DIMORegistryChainID: 30001, TablelandAPIGateway: baseURL}, &logger)

	tablelandAPI := services.NewTablelandApiService(&logger, &config.Settings{
		TablelandAPIGateway: baseURL,
	})

	adController := New(repo, tablelandAPI)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	queryURL := "api/v1/query?statement=SELECT+%2A+FROM+%22_30001_1%22+WHERE+%28%22id%22+%3D+%27toyota_camry_2007%27%29"
	respQueryBody := `
	  [{
		"id": "toyota_camry_2007",
		"deviceType": "vehicle",
		"imageURI": "https://image",
		"ksuid": "26G3iFH7Xc9Wvsw7pg6sD7uzoSS",
		"metadata": {
		  "device_attributes": [
			{
			  "name": "powertrain_type",
			  "value": "ICE"
			}
		  ]
		}
	  }]
	`
	// when we query against tableland, if metadata is not set, it is returned as an empty string ""
	var modelQueryTablelandResponse []DeviceDefinitionTablelandModel
	errUm := json.Unmarshal([]byte(respQueryBody), &modelQueryTablelandResponse)
	require.NoError(t, errUm)

	httpmock.RegisterResponder(http.MethodGet, baseURL+queryURL, httpmock.NewStringResponder(200, respQueryBody))

	res, err := adController.GetDeviceDefinition(ctx, model.DeviceDefinitionBy{ID: "toyota_camry_2007"})
	require.NoError(t, err)

	deviceType := "vehicle"
	legacyID1 := "26G3iFH7Xc9Wvsw7pg6sD7uzoSS"

	assert.Equal(t, "toyota_camry_2007", res.DeviceDefinitionID)
	assert.Equal(t, legacyID1, *res.LegacyID)
	assert.Equal(t, deviceType, *res.DeviceType)
	assert.Equal(t, "https://image", *res.ImageURI)
	assert.Equal(t, "Toyota", res.Manufacturer.Name)
	assert.Equal(t, 137, res.Manufacturer.TokenID)
}

func Test_GetDeviceDefinitionsByIDs_Query(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	mfr := models.Manufacturer{
		ID:      137,
		Name:    "Alfa Romeo",
		Owner:   common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:    "alfa-romeo",
		TableID: null.IntFrom(1),
	}
	err := mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	mfr2 := models.Manufacturer{
		ID:      13,
		Name:    "BMW",
		Owner:   common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:    "bmw",
		TableID: null.IntFrom(2),
	}
	err = mfr2.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	logger := zerolog.Nop()

	const baseURL = "http://local"

	repo := base.NewRepository(pdb, config.Settings{DIMORegistryChainID: 30001, TablelandAPIGateway: baseURL}, &logger)

	tablelandAPI := services.NewTablelandApiService(&logger, &config.Settings{
		TablelandAPIGateway: baseURL,
	})

	adController := New(repo, tablelandAPI)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock the query for alfa-romeo manufacturer (table _30001_1)
	queryURLAlfaRomeo := "api/v1/query?statement=SELECT+%2A+FROM+%22_30001_1%22+WHERE+%28%22id%22+IN+%28%27alfa-romeo_147_2007%27%29%29"
	respQueryBodyAlfaRomeo := `[
	  {
		"id": "alfa-romeo_147_2007",
		"deviceType": "vehicle",
		"imageURI": "https://image",
		"ksuid": "26G3iFH7Xc9Wvsw7pg6sD7uzoSS",
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
	httpmock.RegisterResponder(http.MethodGet, baseURL+queryURLAlfaRomeo, httpmock.NewStringResponder(200, respQueryBodyAlfaRomeo))

	// Mock the query for bmw manufacturer (table _30001_2)
	queryURLBMW := "api/v1/query?statement=SELECT+%2A+FROM+%22_30001_2%22+WHERE+%28%22id%22+IN+%28%27bmw_x5_2019%27%29%29"
	respQueryBodyBMW := `[
	  {
		"id": "bmw_x5_2019",
		"deviceType": "vehicle",
		"imageURI": "https://image",
		"ksuid": "12G3iFH7Xc9Wvsw7pg6sD7uzoKK",
		"metadata": ""
	  }
	]`
	httpmock.RegisterResponder(http.MethodGet, baseURL+queryURLBMW, httpmock.NewStringResponder(200, respQueryBodyBMW))

	res, err := adController.GetDeviceDefinitionsByIDs(ctx, []string{"alfa-romeo_147_2007", "bmw_x5_2019"})
	require.NoError(t, err)

	assert.Len(t, res, 2)
	assert.Equal(t, res[0].DeviceDefinitionID, "alfa-romeo_147_2007")
	assert.Equal(t, res[1].DeviceDefinitionID, "bmw_x5_2019")

	deviceType := "vehicle"
	legacyID1 := "26G3iFH7Xc9Wvsw7pg6sD7uzoSS"
	legacyID2 := "12G3iFH7Xc9Wvsw7pg6sD7uzoKK"

	expected := []*model.DeviceDefinition{
		{
			DeviceDefinitionID: "alfa-romeo_147_2007",
			LegacyID:           &legacyID1,
			DeviceType:         &deviceType,
			Manufacturer: &model.Manufacturer{
				Name:    "Alfa Romeo",
				TokenID: 137,
			},
		},
		{
			DeviceDefinitionID: "bmw_x5_2019",
			LegacyID:           &legacyID2,
			DeviceType:         &deviceType,
			Manufacturer: &model.Manufacturer{
				Name:    "BMW",
				TokenID: 13,
			},
		},
	}

	for i, e := range expected {
		assert.Equal(t, e.DeviceDefinitionID, res[i].DeviceDefinitionID)
		assert.Equal(t, e.LegacyID, res[i].LegacyID)
		assert.Equal(t, e.DeviceType, res[i].DeviceType)
		assert.Equal(t, e.Manufacturer.Name, res[i].Manufacturer.Name)
		assert.Equal(t, e.Manufacturer.TokenID, res[i].Manufacturer.TokenID)
	}
}

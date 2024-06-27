package devicedefinition

import (
	"context"
	"encoding/json"
	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jarcoal/httpmock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"math/big"
	"net/http"
	"testing"
)

const migrationsDir = "../../../migrations"

var aftermarketDeviceNodeMintedArgs = services.AftermarketDeviceNodeMintedData{
	AftermarketDeviceAddress: common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	ManufacturerID:           big.NewInt(137),
	Owner:                    common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	TokenID:                  big.NewInt(42),
}

func Test_GetDeviceDefinitions_Query(t *testing.T) {
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

	adController := Repository{Repository: repo, TablelandApiService: tablelandAPI}
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
		"metadata": {
		  "device_attributes": [
			{
			  "name": "powertrain_type",
			  "value": "EV"
			}
		  ]
		}
	  }
	]`
	var modelQueryTablelandResponse []DeviceDefinitionTablelandModel
	_ = json.Unmarshal([]byte(respQueryBody), &modelQueryTablelandResponse)

	httpmock.RegisterResponder(http.MethodGet, baseURL+queryURL, httpmock.NewStringResponder(200, respQueryBody))

	res, err := adController.GetDeviceDefinitions(ctx, &mfr.TableID.Int, nil, nil, &last, &before, &model.DeviceDefinitionFilter{})
	assert.NoError(t, err)

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
			},
			Cursor: "Mg==",
		},
		{
			Node: &model.DeviceDefinition{
				DeviceDefinitionID: "alfa-romeo_147_2007",
				LegacyID:           &legacyID1,
				DeviceType:         &deviceType,
			},
			Cursor: "Mw==",
		},
	}

	for i, e := range expected {
		assert.Equal(t, e.Node.DeviceDefinitionID, res.Edges[i].Node.DeviceDefinitionID)
		assert.Equal(t, e.Node.LegacyID, res.Edges[i].Node.LegacyID)
		assert.Equal(t, e.Node.DeviceType, res.Edges[i].Node.DeviceType)
	}
}

package services

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/jarcoal/httpmock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func Test_Query_Tableland_Success(t *testing.T) {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	const baseURL = "http://local"

	tablelandAPI := NewTablelandApiService(&logger, &config.Settings{
		TablelandAPIGateway: baseURL,
	})

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	statement := "SELECT * FROM table"
	expectedURL := "api/v1/query?statement=SELECT+%2A+FROM+table"

	respBody := `[
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

	httpmock.RegisterResponder(http.MethodGet, baseURL+expectedURL, httpmock.NewStringResponder(200, respBody))

	var result []map[string]interface{}
	err := tablelandAPI.Query(context.Background(), statement, &result)

	require.NoError(t, err)
	require.NotEmpty(t, result)
}

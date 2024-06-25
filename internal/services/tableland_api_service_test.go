package services

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/url"
	"os"
	"testing"
)

type MockHTTPClientWrapper struct {
	mock.Mock
}

func (m *MockHTTPClientWrapper) ExecuteRequest(url, method string, body interface{}) (*http.Response, error) {
	args := m.Called(url, method, body)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestTablelandApiService_Query(t *testing.T) {
	// Setup
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	settings := &config.Settings{
		TablelandAPIGateway: "https://example.com",
	}

	mockHTTPClient := new(MockHTTPClientWrapper)
	apiService := &TablelandApiService{
		log:        &logger,
		settings:   settings,
		httpClient: mockHTTPClient,
		url:        &url.URL{Path: tablelandQueryPath},
	}

	// Test data
	statement := "SELECT * FROM table"
	expectedResult := map[string]interface{}{"key": "value"}

	// Mock response
	responseBody, _ := json.Marshal(expectedResult)
	resp := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
	}

	// Expectations
	mockHTTPClient.On("ExecuteRequest", "https://example.com/api/v1/query?statement=SELECT+%2A+FROM+table", "GET", nil).
		Return(resp, nil)

	// Execute
	var result map[string]interface{}
	err := apiService.Query(context.Background(), statement, &result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockHTTPClient.AssertExpectations(t)
}

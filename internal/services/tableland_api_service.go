package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
)

type TablelandApiService struct {
	log        *zerolog.Logger
	settings   *config.Settings
	httpClient shared.HTTPClientWrapper
	url        *url.URL
}

func NewTablelandApiService(log *zerolog.Logger, settings *config.Settings) *TablelandApiService {
	httpClient, _ := shared.NewHTTPClientWrapper(settings.TablelandAPIGateway, "", 10*time.Second, nil, true)

	qu, _ := url.Parse(tablelandQueryPath)

	return &TablelandApiService{
		log:        log,
		httpClient: httpClient,
		settings:   settings,
		url:        qu,
	}
}

const tablelandQueryPath = "/api/v1/query"

func (r *TablelandApiService) Query(_ context.Context, statement string, result any) error {
	v := url.Values{}
	v.Add("statement", statement)
	q := v.Encode()

	// Copy
	url := *r.url
	url.RawQuery = q

	resp, err := r.httpClient.ExecuteRequest(url.String(), "GET", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(result)
}

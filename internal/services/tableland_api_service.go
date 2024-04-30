package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
)

type TablelandApiService struct {
	log        *zerolog.Logger
	settings   *config.Settings
	httpClient shared.HTTPClientWrapper
}

func NewTablelandApiService(log *zerolog.Logger, settings *config.Settings) *TablelandApiService {
	httpClient, _ := shared.NewHTTPClientWrapper(settings.TablelandAPIGateway, "", 10*time.Second, nil, true)

	return &TablelandApiService{
		log:        log,
		httpClient: httpClient,
		settings:   settings,
	}
}

func (r *TablelandApiService) Query(ctx context.Context, queryParams map[string]string, result interface{}) error {
	//if queryParams != nil {
	//	values := fullURL.Query()
	//	for key, value := range queryParams {
	//		values.Set(key, value)
	//	}
	//	fullURL.RawQuery = values.Encode()
	//}

	//req, err := r.httpClient.ExecuteRequest((ctx, http.MethodGet, fullURL.String(), nil)
	resp, err := r.httpClient.ExecuteRequest("path", "GET", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}

	return nil
}

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

func (r *TablelandApiService) Query(_ context.Context, queryParams map[string]string, result interface{}) error {
	var queryString string = "api/v1/query?"
	if len(queryParams) > 0 {
		queryParamsList := make([]string, 0, len(queryParams))
		for key, value := range queryParams {
			queryParamsList = append(queryParamsList, key+"="+value)
		}
		queryString += strings.Join(queryParamsList, "&")
	}

	resp, err := r.httpClient.ExecuteRequest(queryString, "GET", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}

	return nil
}

package services

import (
	"context"
	"encoding/json"
	"fmt"

	"net/http"
	"net/url"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/rs/zerolog"
)

type TablelandApiService struct {
	log      *zerolog.Logger
	settings *config.Settings
}

func NewTablelandApiService(log *zerolog.Logger, settings *config.Settings) *TablelandApiService {
	return &TablelandApiService{
		log:      log,
		settings: settings,
	}
}

func (r *TablelandApiService) Query(ctx context.Context, queryParams map[string]string, result interface{}) error {
	fullURL, err := url.Parse(r.settings.TablelandAPIGateway)
	if err != nil {
		return err
	}

	fullURL = fullURL.JoinPath(fullURL.Path, "api/v1/query")

	if queryParams != nil {
		values := fullURL.Query()
		for key, value := range queryParams {
			values.Set(key, value)
		}
		fullURL.RawQuery = values.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to complete request: %w", err)
	}

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}

	return nil
}

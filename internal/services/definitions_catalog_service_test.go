package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A refresh that returns a corrupt body (or an error status) must serve the
// existing snapshot, not fail reads.
func TestCatalogServesStaleOnBadRefresh(t *testing.T) {
	good := `{"updatedAt":"t","count":1,"definitions":[
	  {"id":"toyota_camry_2020","ksuid":"K","model":"Camry","year":2020,
	   "devicetype":"vehicle","imageuri":"","metadata":null,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`

	mode := "good"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		switch mode {
		case "good":
			_, _ = w.Write([]byte(good))
		case "garbage":
			_, _ = w.Write([]byte(`{"definitions": [{"id": tru`))
		default:
			w.WriteHeader(http.StatusBadGateway)
		}
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDefinitionsCatalogService(&logger, &config.Settings{DefinitionsCatalogURL: srv.URL})
	ctx := context.Background()

	d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)

	// Force refresh against a corrupt 200 body: reads keep working.
	mode = "garbage"
	svc.lastFetch = svc.lastFetch.Add(-2 * svc.refreshInterval)
	d, err = svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	assert.NotNil(t, d)

	// Same for a 502.
	mode = "bad"
	svc.lastFetch = svc.lastFetch.Add(-2 * svc.refreshInterval)
	d, err = svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	assert.NotNil(t, d)
}

package services

import (
	"context"
	"os"
	"testing"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Runs only against a live catalog (local wrangler dev or deployed worker):
//
//	CATALOG_E2E_URL=http://localhost:8787 go test ./internal/services/ -run TestCatalogE2E
func TestCatalogE2E(t *testing.T) {
	url := os.Getenv("CATALOG_E2E_URL")
	if url == "" {
		t.Skip("CATALOG_E2E_URL not set")
	}
	ctx := context.Background()
	logger := zerolog.Nop()
	svc := NewDefinitionsCatalogService(&logger, &config.Settings{DefinitionsCatalogURL: url})

	def, err := svc.GetDefinitionByID(ctx, "dodge_town-&-country_2012")
	require.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, "Town & Country", def.Model)
	assert.Equal(t, 33, def.Manufacturer.TokenID)
	assert.Equal(t, "dodge", def.Manufacturer.Slug)

	bmw, err := svc.DefinitionsByManufacturer(ctx, 13)
	require.NoError(t, err)
	assert.Greater(t, len(bmw), 1500, "BMW should have >1500 definitions")
	for i := 1; i < len(bmw); i++ {
		require.Less(t, bmw[i-1].ID, bmw[i].ID, "definitions must be sorted by id")
	}

	missing, err := svc.GetDefinitionByID(ctx, "not_areal_2020")
	require.NoError(t, err)
	assert.Nil(t, missing)
}

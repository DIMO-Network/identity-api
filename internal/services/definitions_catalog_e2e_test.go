package services

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Runs only against a live catalog: local wrangler dev, or a deployed worker.
//
//	CATALOG_E2E_URL=http://localhost:8787 go test ./internal/services/ -run TestCatalogE2E
//
// The data expectations are picked by host, because the deployed catalogs
// belong to different chains: the same slug carries a different manufacturer
// token id on each, so the prod numbers this test used to hard-code failed
// against the dev catalog it also tells you to point at. Any other host, a
// local bucket or a preview deployment included, is checked for shape only.
//
// It works against either source. Whether the host serves a template index or
// only the flat manifest is the service's business, not the test's.
func TestCatalogE2E(t *testing.T) {
	base := os.Getenv("CATALOG_E2E_URL")
	if base == "" {
		t.Skip("CATALOG_E2E_URL not set")
	}
	parsed, err := url.Parse(base)
	require.NoError(t, err)

	// Dodge's token id, and the fewest definitions BMW (token 13) should have.
	var dodgeToken, bmwAtLeast int
	switch parsed.Host {
	case "definitions.dimo.org":
		dodgeToken, bmwAtLeast = 33, 1500
	case "definitions.dev.dimo.org":
		dodgeToken, bmwAtLeast = 32, 900
	}

	ctx := context.Background()
	logger := zerolog.Nop()
	svc, err := NewDefinitionsCatalogService(&logger, &config.Settings{DefinitionsCatalogURL: base})
	require.NoError(t, err)
	t.Cleanup(svc.Close)

	def, err := svc.GetDefinitionByID(ctx, "dodge_town-&-country_2012")
	require.NoError(t, err)
	require.NotNil(t, def)
	t.Logf("%s serves the %s source", parsed.Host, svc.snap.Load().source)

	assert.Equal(t, "Town & Country", def.Model)
	assert.Equal(t, "dodge", def.Manufacturer.Slug)
	assert.Positive(t, def.Manufacturer.TokenID)
	if dodgeToken != 0 {
		assert.Equal(t, dodgeToken, def.Manufacturer.TokenID, "%s is chain %d's catalog", parsed.Host, dodgeToken)
	}

	// BMW on both deployed catalogs; on an unknown host, whatever manufacturer
	// the definition above belongs to.
	listToken := 13
	if dodgeToken == 0 {
		listToken = def.Manufacturer.TokenID
	}
	list, err := svc.DefinitionsByManufacturer(ctx, listToken)
	require.NoError(t, err)
	require.NotEmpty(t, list)
	if bmwAtLeast != 0 {
		assert.Greater(t, len(list), bmwAtLeast, "BMW should have more than %d definitions", bmwAtLeast)
	}
	for i, d := range list {
		require.NotEmpty(t, d.ID)
		require.Positive(t, d.Manufacturer.TokenID)
		require.NotEmpty(t, d.Manufacturer.Slug)
		if i > 0 {
			require.Less(t, list[i-1].ID, d.ID, "definitions must be sorted by id")
		}
	}

	missing, err := svc.GetDefinitionByID(ctx, "not_areal_2020")
	require.NoError(t, err, "an id that does not exist is not an error, in either source")
	assert.Nil(t, missing)
}

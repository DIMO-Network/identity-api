package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The shared loader assigns strings verbatim and turns a malformed int into 0
// with no error, so both settings are validated here and startup fails on a
// bad value instead of running with the floor silently disabled or with a URL
// every request fails to parse.
func TestDefinitionsCatalogSettings(t *testing.T) {
	t.Run("url", func(t *testing.T) {
		valid := map[string]string{
			"https://definitions.dimo.org":         "https://definitions.dimo.org",
			"https://definitions.dimo.org/":        "https://definitions.dimo.org",
			"https://definitions.dimo.org///":      "https://definitions.dimo.org",
			"http://localhost:8787":                "http://localhost:8787",
			"HTTPS://definitions.dimo.org":         "HTTPS://definitions.dimo.org",
			"https://example.org/catalog/":         "https://example.org/catalog",
			"https://definitions.dev.dimo.org:443": "https://definitions.dev.dimo.org:443",
		}
		for raw, want := range valid {
			got, err := (&Settings{DefinitionsCatalogURL: raw}).DefinitionsCatalog()
			require.NoError(t, err, raw)
			assert.Equal(t, want, got.URL, raw)
		}

		invalid := []string{
			"",
			"https://definitions.dimo.org ",
			" https://definitions.dimo.org",
			"https://definitions.dimo.org\n",
			"https://defi\tnitions.dimo.org",
			"https://definitions.dimo.org\x7f",
			"https://definitions.dimo.org ",
			"ftp://definitions.dimo.org",
			"definitions.dimo.org",
			"https://",
			"https:///catalog",
			"https://definitions.dimo.org?x=1",
			"https://definitions.dimo.org/?",
			"https://definitions.dimo.org#top",
			"https://definitions.dimo.org/%zz",
		}
		for _, raw := range invalid {
			_, err := (&Settings{DefinitionsCatalogURL: raw}).DefinitionsCatalog()
			assert.Error(t, err, "%q must be refused", raw)
		}
	})

	t.Run("max staleness", func(t *testing.T) {
		valid := map[string]time.Duration{
			"":      DefaultDefinitionsMaxStaleness,
			"0":     0,
			"24h":   24 * time.Hour,
			"90m":   90 * time.Minute,
			"1h30m": 90 * time.Minute,
		}
		for raw, want := range valid {
			got, err := (&Settings{DefinitionsCatalogURL: "https://definitions.dimo.org", DefinitionsMaxStaleness: raw}).DefinitionsCatalog()
			require.NoError(t, err, raw)
			assert.Equal(t, want, got.MaxStaleness, raw)
		}

		for _, raw := range []string{"-1h", "24 h", " 24h", "24hours", "day", "24"} {
			_, err := (&Settings{DefinitionsCatalogURL: "https://definitions.dimo.org", DefinitionsMaxStaleness: raw}).DefinitionsCatalog()
			assert.Error(t, err, "%q must be refused", raw)
		}
	})

	t.Run("max build age", func(t *testing.T) {
		valid := map[string]time.Duration{
			"":     DefaultDefinitionsMaxBuildAge,
			"0":    0,
			"72h":  72 * time.Hour,
			"168h": 168 * time.Hour,
		}
		for raw, want := range valid {
			got, err := (&Settings{DefinitionsCatalogURL: "https://definitions.dimo.org", DefinitionsMaxBuildAge: raw}).DefinitionsCatalog()
			require.NoError(t, err, raw)
			assert.Equal(t, want, got.MaxBuildAge, raw)
		}

		for _, raw := range []string{"-1h", "72 h", " 72h", "3days", "72"} {
			_, err := (&Settings{DefinitionsCatalogURL: "https://definitions.dimo.org", DefinitionsMaxBuildAge: raw}).DefinitionsCatalog()
			assert.Error(t, err, "%q must be refused", raw)
		}
	})

	t.Run("min count", func(t *testing.T) {
		valid := map[string]int{"": 0, "0": 0, "8000": 8000, "15000": 15000, "007": 7}
		for raw, want := range valid {
			got, err := (&Settings{DefinitionsCatalogURL: "https://definitions.dimo.org", DefinitionsMinCount: raw}).DefinitionsCatalog()
			require.NoError(t, err, raw)
			assert.Equal(t, want, got.MinCount, raw)
		}

		for _, raw := range []string{"15,000", "15_000", "15000 ", " 15000", "15000\n", "1.5e4", "-1", "+5", "abc", "99999999999999999999999"} {
			_, err := (&Settings{DefinitionsCatalogURL: "https://definitions.dimo.org", DefinitionsMinCount: raw}).DefinitionsCatalog()
			assert.Error(t, err, "%q must be refused", raw)
		}
	})
}

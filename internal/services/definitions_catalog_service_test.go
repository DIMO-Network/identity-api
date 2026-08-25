package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

// A well-formed 200 whose body is empty (or whose count disagrees with the
// definitions it carries) must not replace a good snapshot. The worker can
// produce a truncated manifest -- a lost read of manifest.json makes it rewrite
// the object from a single change -- and adopting that silently turns every
// device-definition query into a successful, empty answer rather than an error.
func TestCatalogRejectsDegenerateManifest(t *testing.T) {
	good := `{"updatedAt":"t","count":1,"definitions":[
	  {"id":"toyota_camry_2020","ksuid":"K","model":"Camry","year":2020,
	   "devicetype":"vehicle","imageuri":"","metadata":null,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`

	mode := "good"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		switch mode {
		case "good":
			_, _ = w.Write([]byte(good))
		case "empty":
			_, _ = w.Write([]byte(`{"updatedAt":"t","count":0,"definitions":[]}`))
		default: // count disagrees with the payload
			_, _ = w.Write([]byte(`{"updatedAt":"t","count":9000,"definitions":[]}`))
		}
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDefinitionsCatalogService(&logger, &config.Settings{DefinitionsCatalogURL: srv.URL})
	ctx := context.Background()

	d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)

	for _, m := range []string{"empty", "mismatch"} {
		mode = m
		svc.mu.Lock()
		svc.lastFetch = time.Time{} // force a refresh
		svc.etag = ""
		svc.mu.Unlock()

		d, err = svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.NoError(t, err, m)
		require.NotNil(t, d, "%s: snapshot must survive a degenerate manifest", m)
		assert.Equal(t, "Camry", d.Model, m)
	}
}

// One malformed definition must not cost the whole catalog. Decoding the
// manifest as a single document means a single bad element aborts everything:
// warm pods freeze on a stale snapshot and a cold pod fails every query.
func TestCatalogSkipsMalformedDefinitions(t *testing.T) {
	body := `{"updatedAt":"t","count":3,"definitions":[
	  {"id":"toyota_camry_2020","ksuid":"K","model":"Camry","year":2020,
	   "devicetype":"vehicle","imageuri":"","metadata":null,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}},
	  {"id":"toyota_broken_2020","year":"not-a-number"},
	  {"id":"toyota_supra_2021","ksuid":"K2","model":"Supra","year":2021,
	   "devicetype":"vehicle","imageuri":"","metadata":null,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDefinitionsCatalogService(&logger, &config.Settings{DefinitionsCatalogURL: srv.URL})
	ctx := context.Background()

	good, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, good, "a good definition must survive a malformed sibling")
	assert.Equal(t, "Camry", good.Model)

	other, err := svc.GetDefinitionByID(ctx, "toyota_supra_2021")
	require.NoError(t, err)
	require.NotNil(t, other, "definitions after the malformed one must still load")
}

// Every neighbouring catalog setting in values.yaml carries a trailing slash.
func TestCatalogURLTolerantOfTrailingSlash(t *testing.T) {
	body := `{"updatedAt":"t","count":1,"definitions":[
	  {"id":"toyota_camry_2020","ksuid":"K","model":"Camry","year":2020,
	   "devicetype":"vehicle","imageuri":"","metadata":null,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDefinitionsCatalogService(&logger, &config.Settings{DefinitionsCatalogURL: srv.URL + "/"})
	d, err := svc.GetDefinitionByID(context.Background(), "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)
	assert.Equal(t, "/manifest.json", gotPath)
}

// The degenerate-manifest guard and the per-element decode were added together
// and defeated each other: the guard counted raw elements while the snapshot is
// built from decoded survivors. A producer-side type change fails every element,
// so the guard sees a full manifest, the skip loop discards all of it, and the
// catalog is replaced with an empty map -- a fleet-wide outage from a good
// snapshot voluntarily thrown away.
func TestCatalogKeepsSnapshotWhenEveryDefinitionFailsToDecode(t *testing.T) {
	good := `{"updatedAt":"t","count":1,"definitions":[
	  {"id":"toyota_camry_2020","ksuid":"K","model":"Camry","year":2020,
	   "devicetype":"vehicle","imageuri":"","metadata":null,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`

	// Same element count, but every one is undecodable (year became a string).
	allBad := `{"updatedAt":"t","count":1,"definitions":[
	  {"id":"toyota_camry_2020","year":"2020"}]}`

	mode := "good"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if mode == "good" {
			_, _ = w.Write([]byte(good))
			return
		}
		_, _ = w.Write([]byte(allBad))
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDefinitionsCatalogService(&logger, &config.Settings{DefinitionsCatalogURL: srv.URL})
	ctx := context.Background()

	d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)

	mode = "bad"
	svc.mu.Lock()
	svc.lastFetch = time.Time{}
	svc.etag = ""
	svc.mu.Unlock()

	d, err = svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d, "a manifest whose definitions all fail to decode must not empty the catalog")
	assert.Equal(t, "Camry", d.Model)
}

// Rate-limiting a failed refresh must not make the failure look like success.
// lastFetch feeds the freshness short-circuit, so setting it on an error path
// suppressed the error for the rest of the interval while the catalog was still
// empty: one request per minute got a real error and every other request got a
// confident "no such definition" with no error at all.
func TestCatalogKeepsReportingAColdStartFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDefinitionsCatalogService(&logger, &config.Settings{DefinitionsCatalogURL: srv.URL})
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.Error(t, err, "request %d must report the outage, not answer emptily", i)
		assert.Nil(t, d)
	}

	defs, err := svc.DefinitionsByManufacturer(ctx, 131)
	require.Error(t, err, "an unreachable catalog must not look like a manufacturer with no definitions")
	assert.Empty(t, defs)
}

// A floor the operator sets, checked on every pod including a cold one. The
// proportional guard could not protect a cold pod (nothing held to compare
// against), which is exactly the pod that adopts a one-definition manifest
// published during a rebuild.
func TestCatalogRefusesAManifestBelowTheConfiguredFloor(t *testing.T) {
	one := `{"updatedAt":"t","count":1,"definitions":[
	  {"id":"toyota_camry_2020","model":"Camry","year":2020,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(one))
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDefinitionsCatalogService(&logger, &config.Settings{
		DefinitionsCatalogURL: srv.URL,
		DefinitionsMinCount:   100,
	})

	// Cold pod: nothing held, so there is no proportional comparison to make.
	d, err := svc.GetDefinitionByID(context.Background(), "toyota_camry_2020")
	require.Error(t, err, "a manifest below the floor must be refused even with nothing held")
	assert.Nil(t, d)
	assert.Contains(t, err.Error(), "below the configured minimum")
}

func TestCatalogAcceptsAManifestAtOrAboveTheFloor(t *testing.T) {
	two := `{"updatedAt":"t","count":2,"definitions":[
	  {"id":"toyota_camry_2020","model":"Camry","year":2020,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}},
	  {"id":"toyota_supra_2021","model":"Supra","year":2021,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(two))
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDefinitionsCatalogService(&logger, &config.Settings{
		DefinitionsCatalogURL: srv.URL,
		DefinitionsMinCount:   2,
	})

	d, err := svc.GetDefinitionByID(context.Background(), "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)
}

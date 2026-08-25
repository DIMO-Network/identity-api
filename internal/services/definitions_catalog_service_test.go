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

// A large shrink must be refused once, then adopted if the catalog still
// reports it. Refusing indefinitely means a legitimate purge is never picked up
// by running pods while any pod that restarts adopts it, so the same query
// answers differently depending on which replica serves it -- with no escape
// short of an undocumented rollout restart.
func TestCatalogAdoptsALargeShrinkOnceItIsConfirmed(t *testing.T) {
	big := `{"updatedAt":"t","count":4,"definitions":[
	  {"id":"toyota_a_2020","model":"A","year":2020,"manufacturer":{"tokenId":1,"slug":"t","name":"T"}},
	  {"id":"toyota_b_2020","model":"B","year":2020,"manufacturer":{"tokenId":1,"slug":"t","name":"T"}},
	  {"id":"toyota_c_2020","model":"C","year":2020,"manufacturer":{"tokenId":1,"slug":"t","name":"T"}},
	  {"id":"toyota_d_2020","model":"D","year":2020,"manufacturer":{"tokenId":1,"slug":"t","name":"T"}}]}`
	small := `{"updatedAt":"t","count":1,"definitions":[
	  {"id":"toyota_a_2020","model":"A","year":2020,"manufacturer":{"tokenId":1,"slug":"t","name":"T"}}]}`

	body := big
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDefinitionsCatalogService(&logger, &config.Settings{DefinitionsCatalogURL: srv.URL})
	ctx := context.Background()

	d, err := svc.GetDefinitionByID(ctx, "toyota_d_2020")
	require.NoError(t, err)
	require.NotNil(t, d)

	forceRefresh := func() {
		svc.mu.Lock()
		svc.lastFetch = time.Time{}
		svc.etag = ""
		svc.mu.Unlock()
	}

	// First sighting of the shrink is refused; the old catalog still serves.
	body = small
	forceRefresh()
	d, err = svc.GetDefinitionByID(ctx, "toyota_d_2020")
	require.NoError(t, err)
	require.NotNil(t, d, "a large shrink must not be adopted on first sight")

	// Still reporting the same size: it is real, so adopt it.
	forceRefresh()
	d, err = svc.GetDefinitionByID(ctx, "toyota_d_2020")
	require.NoError(t, err)
	assert.Nil(t, d, "a confirmed shrink must be adopted rather than refused forever")

	kept, err := svc.GetDefinitionByID(ctx, "toyota_a_2020")
	require.NoError(t, err)
	require.NotNil(t, kept)
}

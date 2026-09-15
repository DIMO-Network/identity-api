package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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

	svc := newTestCatalog(t, config.Settings{DefinitionsCatalogURL: srv.URL})
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

	svc := newTestCatalog(t, config.Settings{DefinitionsCatalogURL: srv.URL})
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

	svc := newTestCatalog(t, config.Settings{DefinitionsCatalogURL: srv.URL})
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

	svc := newTestCatalog(t, config.Settings{DefinitionsCatalogURL: srv.URL + "/"})
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

	svc := newTestCatalog(t, config.Settings{DefinitionsCatalogURL: srv.URL})
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

	svc := newTestCatalog(t, config.Settings{DefinitionsCatalogURL: srv.URL})
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

	svc := newTestCatalog(t, config.Settings{
		DefinitionsCatalogURL: srv.URL,
		DefinitionsMinCount:   "100",
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

	svc := newTestCatalog(t, config.Settings{
		DefinitionsCatalogURL: srv.URL,
		DefinitionsMinCount:   "2",
	})

	d, err := svc.GetDefinitionByID(context.Background(), "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)
}

// The refresh is shared by every caller queued behind the lock, so it must not
// die with the caller that happened to trigger it. On a cold pod the first
// query after a deploy often comes from a client with a short timeout; when
// net/http cancelled that request's context mid-download the context.Canceled
// was recorded as a failed attempt and the backoff answered every request with
// "context canceled" for a full interval without re-fetching.
func TestCatalogColdRefreshSurvivesTheFirstCallerDisconnecting(t *testing.T) {
	good := `{"updatedAt":"t","count":1,"definitions":[
	  {"id":"toyota_camry_2020","ksuid":"K","model":"Camry","year":2020,
	   "devicetype":"vehicle","imageuri":"","metadata":null,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`

	var (
		hits    atomic.Int32
		started = make(chan struct{})
		release = make(chan struct{})
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) == 1 {
			// Hold the first download until the caller has given up on it.
			close(started)
			<-release
		}
		_, _ = w.Write([]byte(good))
	}))
	defer srv.Close()

	svc := newTestCatalog(t, config.Settings{DefinitionsCatalogURL: srv.URL})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-started
		cancel()
		close(release)
	}()

	d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err, "the shared refresh must finish even though the caller that started it went away")
	require.NotNil(t, d)

	svc.mu.RLock()
	assert.NoError(t, svc.lastErr, "a caller's own cancellation must never be recorded as a failed attempt")
	assert.True(t, svc.lastAttempt.IsZero(), "a caller's own cancellation must never rate-limit the retry")
	svc.mu.RUnlock()

	// The next caller, with a healthy context, is served from the freshly
	// loaded catalog: no backoff error and no second download.
	d, err = svc.GetDefinitionByID(context.Background(), "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)
	assert.Equal(t, int32(1), hits.Load(), "the catalog loaded by the interrupted refresh must be served, not fetched again")
}

// Same disconnect on a warm pod: the stale-serve path used to see the caller's
// context.Canceled, mark the cache fresh and silently skip the refresh for an
// interval. The refresh must complete and the new manifest must be adopted.
func TestCatalogWarmRefreshSurvivesTheCallerDisconnecting(t *testing.T) {
	v1 := `{"updatedAt":"t","count":1,"definitions":[
	  {"id":"toyota_camry_2020","ksuid":"K","model":"Camry","year":2020,
	   "devicetype":"vehicle","imageuri":"","metadata":null,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`
	v2Head := `{"updatedAt":"t2","count":2,"definitions":[
	  {"id":"toyota_camry_2020","ksuid":"K","model":"Camry","year":2020,
	   "devicetype":"vehicle","imageuri":"","metadata":null,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}},`
	v2Tail := `
	  {"id":"toyota_supra_2021","ksuid":"K2","model":"Supra","year":2021,
	   "devicetype":"vehicle","imageuri":"","metadata":null,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`

	var (
		hits    atomic.Int32
		started = make(chan struct{})
		release = make(chan struct{})
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) == 1 {
			_, _ = w.Write([]byte(v1))
			return
		}
		// Second download: send half the body, then hold the rest until the
		// caller has given up, so the cancellation lands mid-decode.
		_, _ = w.Write([]byte(v2Head))
		w.(http.Flusher).Flush()
		close(started)
		<-release
		_, _ = w.Write([]byte(v2Tail))
	}))
	defer srv.Close()

	svc := newTestCatalog(t, config.Settings{DefinitionsCatalogURL: srv.URL})

	d, err := svc.GetDefinitionByID(context.Background(), "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)

	svc.mu.Lock()
	svc.lastFetch = time.Time{}
	svc.etag = ""
	svc.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-started
		cancel()
		close(release)
	}()
	_, err = svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)

	svc.mu.RLock()
	assert.NoError(t, svc.lastErr)
	svc.mu.RUnlock()

	// The interrupted refresh finished and its manifest was adopted, so the
	// definition it added is visible without another download.
	supra, err := svc.GetDefinitionByID(context.Background(), "toyota_supra_2021")
	require.NoError(t, err)
	require.NotNil(t, supra, "the refresh a disconnecting caller started must still be adopted")
	assert.Equal(t, int32(2), hits.Load())
}

// Detaching from the caller must not mean waiting forever: the refresh carries
// its own deadline, and running past it is a real failure that is recorded so
// the cold-start backoff engages.
func TestCatalogRefreshHonoursItsOwnTimeout(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	svc := newTestCatalog(t, config.Settings{DefinitionsCatalogURL: srv.URL})
	svc.refreshTimeout = 50 * time.Millisecond

	start := time.Now()
	d, err := svc.GetDefinitionByID(context.Background(), "toyota_camry_2020")
	require.Error(t, err)
	assert.Nil(t, d)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 2*time.Second, "the refresh must give up at its own deadline")

	svc.mu.RLock()
	assert.Error(t, svc.lastErr, "a refresh that ran past its deadline is a failed attempt")
	svc.mu.RUnlock()

	// The failure rate-limits the retry like any other cold-start failure.
	_, err = svc.GetDefinitionByID(context.Background(), "toyota_camry_2020")
	require.Error(t, err)
	assert.Equal(t, int32(1), hits.Load())
}

// The non-2xx branch passed failedAttempt the nil error left over from a
// successful client.Do, so lastErr stayed nil and the cold-start backoff never
// engaged: every query on a cold pod re-downloaded the manifest during an
// outage. A 5xx must be recorded so the retry is rate-limited, and once a
// snapshot is held a 5xx must keep serving it.
func TestCatalogBacksOffAfterAColdStartServerError(t *testing.T) {
	good := `{"updatedAt":"t","count":1,"definitions":[
	  {"id":"toyota_camry_2020","ksuid":"K","model":"Camry","year":2020,
	   "devicetype":"vehicle","imageuri":"","metadata":null,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`

	var hits atomic.Int32
	mode := "bad"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		if mode == "good" {
			_, _ = w.Write([]byte(good))
			return
		}
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	svc := newTestCatalog(t, config.Settings{DefinitionsCatalogURL: srv.URL})
	ctx := context.Background()

	// Cold pod: the first 502 is recorded and the next queries back off on it.
	for i := 1; i <= 3; i++ {
		d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.Error(t, err, "request %d must report the outage", i)
		assert.Nil(t, d)
		assert.Contains(t, err.Error(), "returned 502")
	}
	assert.Equal(t, int32(1), hits.Load(), "a cold pod must not re-download on every query during an outage")
	svc.mu.RLock()
	assert.Error(t, svc.lastErr, "a 5xx on a cold pod must be recorded so the backoff engages")
	svc.mu.RUnlock()

	// Outage over, backoff window lifted: the catalog loads.
	mode = "good"
	svc.mu.Lock()
	svc.lastAttempt = time.Time{}
	svc.mu.Unlock()
	d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)

	// Warm pod: a 5xx on refresh keeps serving the held snapshot.
	mode = "bad"
	svc.mu.Lock()
	svc.lastFetch = time.Time{}
	svc.etag = ""
	svc.mu.Unlock()
	d, err = svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err, "a 5xx must not fail reads once a snapshot is held")
	require.NotNil(t, d)
	assert.Equal(t, "Camry", d.Model)
	assert.Equal(t, int32(3), hits.Load())
}

func newTestCatalog(t *testing.T, settings config.Settings) *DefinitionsCatalogService {
	t.Helper()
	logger := zerolog.Nop()
	svc, err := NewDefinitionsCatalogService(&logger, &settings)
	require.NoError(t, err)
	return svc
}

// A malformed DEFINITIONS_CATALOG_URL used to reach every query as a request
// that failed to build: no backoff, no context, no log. It now fails
// construction, and with it startup, as does a floor the loader cannot parse.
func TestCatalogRefusesMalformedSettingsAtConstruction(t *testing.T) {
	logger := zerolog.Nop()
	for _, raw := range []string{"", "https://definitions.dimo.org ", "https://definitions.dimo.org\n", "definitions.dimo.org", "ftp://definitions.dimo.org", "https://"} {
		svc, err := NewDefinitionsCatalogService(&logger, &config.Settings{DefinitionsCatalogURL: raw})
		assert.Error(t, err, "%q must be refused", raw)
		assert.Nil(t, svc)
	}

	_, err := NewDefinitionsCatalogService(&logger, &config.Settings{
		DefinitionsCatalogURL: "https://definitions.dimo.org",
		DefinitionsMinCount:   "15,000",
	})
	assert.ErrorContains(t, err, "DEFINITIONS_MIN_COUNT")
}

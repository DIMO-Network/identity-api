package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	camryDoc = `{"id":"toyota_camry_2020","ksuid":"K","model":"Camry","year":2020,
	   "devicetype":"vehicle","imageuri":"","metadata":null,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}`
	supraDoc = `{"id":"toyota_supra_2021","ksuid":"K2","model":"Supra","year":2021,
	   "devicetype":"vehicle","imageuri":"","metadata":null,
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}`
)

func manifestOf(defs ...string) string {
	return fmt.Sprintf(`{"updatedAt":"t","count":%d,"definitions":[%s]}`, len(defs), strings.Join(defs, ","))
}

// catalogServer stands in for the definitions worker: it routes by path and
// counts what was asked for, so a test can assert what a refresh fetched.
// Handlers run on the server's goroutines and never assert.
type catalogServer struct {
	*httptest.Server
	mu     sync.Mutex
	routes map[string]http.HandlerFunc
	hits   map[string]int
}

func newCatalogServer(t *testing.T) *catalogServer {
	t.Helper()
	c := &catalogServer{routes: map[string]http.HandlerFunc{}, hits: map[string]int{}}
	c.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.mu.Lock()
		c.hits[r.URL.Path]++
		h := c.routes[r.URL.Path]
		c.mu.Unlock()
		if h == nil {
			http.NotFound(w, r)
			return
		}
		h(w, r)
	}))
	t.Cleanup(c.Close)
	return c
}

func (c *catalogServer) handle(path string, h http.HandlerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.routes[path] = h
}

func (c *catalogServer) serve(path, body string) {
	c.handle(path, func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, body) })
}

func (c *catalogServer) fail(path string, status int) {
	c.handle(path, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) })
}

func (c *catalogServer) hitsFor(path string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.hits[path]
}

func newTestCatalog(t *testing.T, settings config.Settings) *DefinitionsCatalogService {
	t.Helper()
	logger := zerolog.Nop()
	svc, err := NewDefinitionsCatalogService(&logger, &settings)
	require.NoError(t, err)
	t.Cleanup(svc.Close)
	return svc
}

// manualCatalog builds a service whose refresh loop never starts, so the test
// drives each refresh with refreshOnce and no assertion races a goroutine.
func manualCatalog(t *testing.T, srv *catalogServer, settings config.Settings) *DefinitionsCatalogService {
	t.Helper()
	if settings.DefinitionsCatalogURL == "" {
		settings.DefinitionsCatalogURL = srv.URL
	}
	svc := newTestCatalog(t, settings)
	svc.startOnce.Do(func() {}) // claim the once: Start and reads no longer start the loop
	svc.coldWait = 10 * time.Millisecond
	return svc
}

func consecutiveFailures(svc *DefinitionsCatalogService) int {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	return svc.failures
}

// A refresh that fails, for any reason, must leave reads answering from the
// snapshot the pod already holds.
func TestCatalogServesTheHeldSnapshotWhenARefreshFails(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(camryDoc))
	svc := manualCatalog(t, srv, config.Settings{})
	ctx := context.Background()

	require.NoError(t, svc.refreshOnce())
	d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)

	for _, tc := range []struct {
		name   string
		break_ func()
	}{
		{"a corrupt 200 body", func() { srv.serve(legacyManifestPath, `{"definitions": [{"id": tru`) }},
		{"a 502", func() { srv.fail(legacyManifestPath, http.StatusBadGateway) }},
		{"a manifest with nothing valid in it", func() { srv.serve(legacyManifestPath, `{"definitions":[null,null,null]}`) }},
	} {
		tc.break_()
		require.Error(t, svc.refreshOnce(), tc.name)

		d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.NoError(t, err, tc.name)
		require.NotNil(t, d, tc.name)
		assert.Equal(t, "Camry", d.Model, tc.name)
	}
	assert.Equal(t, 3, consecutiveFailures(svc), "each failure is recorded, so the staleness bound and the logs can count them")
}

// A well-formed 200 carrying no definitions must not replace a good snapshot.
// Adopting it turns every device-definition query into a successful, empty
// answer rather than an error, and nothing alerts on that.
//
// The manifest's count plays no part in the decision: it is not read at all,
// so a manifest claiming 9000 definitions while carrying none is refused by
// the empty rule and by nothing else. The earlier version of this test read
// that refusal as a count check and asserted a guarantee that never existed.
func TestCatalogRefusesAnEmptyManifestOverAHeldCatalog(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(camryDoc))
	svc := manualCatalog(t, srv, config.Settings{})
	ctx := context.Background()

	require.NoError(t, svc.refreshOnce())

	for name, body := range map[string]string{
		"count agrees with the empty payload": `{"updatedAt":"t","count":0,"definitions":[]}`,
		"count claims 9000":                   `{"updatedAt":"t","count":9000,"definitions":[]}`,
	} {
		srv.serve(legacyManifestPath, body)
		require.ErrorContains(t, svc.refreshOnce(), "catalog is empty", name)

		d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.NoError(t, err, name)
		require.NotNil(t, d, "%s: the snapshot must survive an empty manifest", name)
		assert.Equal(t, "Camry", d.Model, name)
	}
}

// count and updatedAt are neither read nor strictly decoded, so a manifest is
// judged on the definitions it carries. One carrying fewer than the snapshot
// held, with a count that disagrees with both and an updatedAt of a different
// type, is adopted: the configured floor is what refuses a short manifest, and
// a producer type change in those two fields cannot fail the whole decode.
func TestCatalogIgnoresManifestCountAndUpdatedAt(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(camryDoc, supraDoc))
	svc := manualCatalog(t, srv, config.Settings{})
	ctx := context.Background()

	require.NoError(t, svc.refreshOnce())
	supra, err := svc.GetDefinitionByID(ctx, "toyota_supra_2021")
	require.NoError(t, err)
	require.NotNil(t, supra)

	srv.serve(legacyManifestPath, `{"updatedAt":1756124820,"count":"9000","definitions":[`+camryDoc+`]}`)
	require.NoError(t, svc.refreshOnce())

	camry, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, camry, "a manifest whose count and updatedAt changed type must still decode")
	supra, err = svc.GetDefinitionByID(ctx, "toyota_supra_2021")
	require.NoError(t, err)
	assert.Nil(t, supra, "the shorter manifest is adopted: no rule compares it with the count")
}

// One malformed definition must not cost the whole catalog.
func TestCatalogSkipsMalformedDefinitions(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(camryDoc, `{"id":"toyota_broken_2020","year":"not-a-number"}`, supraDoc))
	svc := manualCatalog(t, srv, config.Settings{})
	ctx := context.Background()

	require.NoError(t, svc.refreshOnce())

	good, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, good, "a good definition must survive a malformed sibling")
	assert.Equal(t, "Camry", good.Model)

	other, err := svc.GetDefinitionByID(ctx, "toyota_supra_2021")
	require.NoError(t, err)
	require.NotNil(t, other, "definitions after the malformed one must still load")
}

// Elements that decode without a type error but carry no id, no positive
// manufacturer token id or no slug used to be adopted as zero-value
// definitions: they counted toward the floor, cleared the empty guard, and
// were indexed under byID[""] and byMfrToken[0].
func TestCatalogSkipsInvalidElementsAndRefusesPastTheThreshold(t *testing.T) {
	srv := newCatalogServer(t)
	ctx := context.Background()

	t.Run("a few invalid elements are skipped", func(t *testing.T) {
		srv.serve(legacyManifestPath, manifestOf(camryDoc, "null", "{}", `{"id":"x_y_2020","manufacturer":{"tokenId":0,"slug":"x"}}`, supraDoc))
		svc := manualCatalog(t, srv, config.Settings{})
		require.NoError(t, svc.refreshOnce())

		blank, err := svc.GetDefinitionByID(ctx, "")
		require.NoError(t, err)
		assert.Nil(t, blank, "a null element must not become the definition with the empty id")

		orphans, err := svc.DefinitionsByManufacturer(ctx, 0)
		require.NoError(t, err)
		assert.Empty(t, orphans, "an element with no token id must not become manufacturer 0's catalog")

		toyotas, err := svc.DefinitionsByManufacturer(ctx, 131)
		require.NoError(t, err)
		assert.Len(t, toyotas, 2)
	})

	t.Run("too many invalid elements refuse the catalog", func(t *testing.T) {
		srv.serve(legacyManifestPath, manifestOf(camryDoc, supraDoc))
		svc := manualCatalog(t, srv, config.Settings{DefinitionsMinCount: "2"})
		require.NoError(t, svc.refreshOnce())

		srv.serve(legacyManifestPath, manifestOf(strings.Split(strings.Repeat("null,", 11), ",")...))
		require.Error(t, svc.refreshOnce())

		d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.NoError(t, err)
		require.NotNil(t, d, "a manifest of nulls must not replace a good snapshot")
	})
}

// A floor the operator sets, checked on every pod including a cold one, which
// is exactly the pod that adopts a stub manifest published during a rebuild.
func TestCatalogRefusesAManifestBelowTheConfiguredFloor(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(camryDoc))
	svc := manualCatalog(t, srv, config.Settings{DefinitionsMinCount: "100"})

	require.ErrorContains(t, svc.refreshOnce(), "below the configured minimum")

	d, err := svc.GetDefinitionByID(context.Background(), "toyota_camry_2020")
	require.Error(t, err, "a manifest below the floor must be refused even with nothing held")
	assert.Nil(t, d)
	assert.ErrorContains(t, err, "below the configured minimum")
}

func TestCatalogAcceptsAManifestAtOrAboveTheFloor(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(camryDoc, supraDoc))
	svc := manualCatalog(t, srv, config.Settings{DefinitionsMinCount: "2"})

	require.NoError(t, svc.refreshOnce())
	d, err := svc.GetDefinitionByID(context.Background(), "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)
}

// Every neighbouring catalog setting in values.yaml carries a trailing slash,
// and "//manifest.json" does not match the worker's exact route.
func TestCatalogURLTolerantOfTrailingSlash(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(camryDoc))
	svc := manualCatalog(t, srv, config.Settings{DefinitionsCatalogURL: srv.URL + "//"})

	require.NoError(t, svc.refreshOnce())
	assert.Equal(t, 1, srv.hitsFor(legacyManifestPath))
	assert.Zero(t, srv.hitsFor("//manifest.json"))
}

// A producer type change fails every element, so the skip loop would discard
// the whole manifest and replace a good catalog with an empty one.
func TestCatalogKeepsSnapshotWhenEveryDefinitionFailsToDecode(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(camryDoc))
	svc := manualCatalog(t, srv, config.Settings{})
	ctx := context.Background()

	require.NoError(t, svc.refreshOnce())

	srv.serve(legacyManifestPath, manifestOf(`{"id":"toyota_camry_2020","year":"2020"}`))
	require.Error(t, svc.refreshOnce())

	d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d, "a manifest whose definitions all fail to decode must not empty the catalog")
	assert.Equal(t, "Camry", d.Model)
}

// A pod that has never loaded a catalog must report the outage rather than
// answer "no such definition" and "this manufacturer has no definitions".
func TestCatalogReportsAColdStartFailure(t *testing.T) {
	srv := newCatalogServer(t)
	srv.fail(legacyManifestPath, http.StatusBadGateway)
	svc := manualCatalog(t, srv, config.Settings{})
	ctx := context.Background()

	require.Error(t, svc.refreshOnce())

	d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.Error(t, err, "an unreachable catalog must not look like a missing definition")
	assert.Nil(t, d)
	assert.ErrorContains(t, err, "returned 502", "the reader gets the wrapped error the refresh recorded")

	defs, err := svc.DefinitionsByManufacturer(ctx, 131)
	require.Error(t, err, "an unreachable catalog must not look like a manufacturer with no definitions")
	assert.Empty(t, defs)
}

// A refresh carries its own deadline: it is nobody's request, so no caller can
// cancel it, and a stalled origin cannot hold it open.
func TestCatalogRefreshRunsOnItsOwnDeadline(t *testing.T) {
	srv := newCatalogServer(t)
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	srv.handle(legacyManifestPath, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	})

	svc := manualCatalog(t, srv, config.Settings{})
	svc.refreshTimeout = 50 * time.Millisecond

	start := time.Now()
	err := svc.refreshOnce()
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 2*time.Second, "the refresh must give up at its own deadline")

	// The caller's context is never the refresh's: a read with a healthy
	// context still reports the recorded failure rather than starting its own.
	_, err = svc.GetDefinitionByID(context.Background(), "toyota_camry_2020")
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, 1, srv.hitsFor(legacyManifestPath), "a read must not fetch the manifest itself")
}

// The refresh used to hold the write lock across the download and the decode,
// so every query on the pod queued behind a stalled origin for up to 30s.
func TestCatalogReaderIsNotBlockedByAStalledRefresh(t *testing.T) {
	srv := newCatalogServer(t)
	stall := make(chan struct{})
	t.Cleanup(func() { close(stall) })
	var hits int
	var mu sync.Mutex
	srv.handle(legacyManifestPath, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		hits++
		first := hits == 1
		mu.Unlock()
		if first {
			_, _ = io.WriteString(w, manifestOf(camryDoc))
			return
		}
		select {
		case <-stall:
		case <-r.Context().Done():
		}
	})

	svc := newTestCatalog(t, config.Settings{DefinitionsCatalogURL: srv.URL})
	svc.refreshInterval = 20 * time.Millisecond
	svc.refreshTimeout = 5 * time.Second
	ctx := context.Background()

	svc.Start()
	require.Eventually(t, func() bool {
		d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		return err == nil && d != nil
	}, 5*time.Second, time.Millisecond, "the catalog must load")
	require.Eventually(t, func() bool { return srv.hitsFor(legacyManifestPath) > 1 }, 5*time.Second, time.Millisecond,
		"a refresh must be in flight and stalled")

	start := time.Now()
	for range 50 {
		d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.NoError(t, err)
		require.NotNil(t, d)
		defs, err := svc.DefinitionsByManufacturer(ctx, 131)
		require.NoError(t, err)
		require.Len(t, defs, 1)
	}
	assert.Less(t, time.Since(start), 500*time.Millisecond, "reads must not wait on the stalled refresh")
}

// A cold pod that meets one 502 used to answer every device-definition query
// with it for a full refresh interval, without re-fetching. The retry now
// backs off from a second, so a transient failure costs seconds.
func TestCatalogColdStartRecoversFromATransientFailure(t *testing.T) {
	srv := newCatalogServer(t)
	var mu sync.Mutex
	var hits int
	srv.handle(legacyManifestPath, func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		hits++
		failing := hits <= 2
		mu.Unlock()
		if failing {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = io.WriteString(w, manifestOf(camryDoc))
	})

	svc := newTestCatalog(t, config.Settings{DefinitionsCatalogURL: srv.URL})
	svc.refreshInterval = time.Minute // a whole interval is what must not be waited out
	svc.initialBackoff = 20 * time.Millisecond
	svc.coldWait = 50 * time.Millisecond
	ctx := context.Background()

	start := time.Now()
	svc.Start()

	_, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.Error(t, err, "while the catalog has not loaded, reads report the failure")
	assert.ErrorContains(t, err, "returned 502")

	require.Eventually(t, func() bool {
		d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		return err == nil && d != nil
	}, 10*time.Second, 5*time.Millisecond, "the retry must recover the catalog")
	assert.Less(t, time.Since(start), 30*time.Second, "recovery must take seconds, not a refresh interval")
}

// A warm replica used to serve arbitrarily stale data with no signal while a
// restarted replica failed every query, so the same query answered differently
// depending on which replica took it.
func TestCatalogStalenessBoundTurnsIntoErrors(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(camryDoc))
	svc := manualCatalog(t, srv, config.Settings{})
	svc.maxStaleness = 50 * time.Millisecond
	ctx := context.Background()

	require.NoError(t, svc.refreshOnce())
	d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)

	// Upstream breaks permanently: the snapshot is held, and ages.
	srv.fail(legacyManifestPath, http.StatusNotFound)
	require.Error(t, svc.refreshOnce())
	time.Sleep(60 * time.Millisecond)

	_, err = svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.Error(t, err, "past the bound a stale catalog must fail, not answer from frozen data")
	assert.ErrorContains(t, err, "DEFINITIONS_MAX_STALENESS")
	assert.ErrorContains(t, err, "returned 404", "the staleness error carries the failure that caused it")
	_, err = svc.DefinitionsByManufacturer(ctx, 131)
	assert.Error(t, err)

	// Zero disables the bound: the same snapshot is served again.
	svc.maxStaleness = 0
	d, err = svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)

	// A successful refresh restores freshness even when nothing changed.
	svc.maxStaleness = 50 * time.Millisecond
	srv.serve(legacyManifestPath, manifestOf(camryDoc))
	require.NoError(t, svc.refreshOnce())
	d, err = svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)
}

// Malformed settings must fail construction, and with it startup: a bad URL
// used to reach every query as a request that failed to build, with no
// backoff, no context and no log.
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

	// A bound shorter than the refresh interval would report every replica
	// stale between refreshes.
	_, err = NewDefinitionsCatalogService(&logger, &config.Settings{
		DefinitionsCatalogURL:   "https://definitions.dimo.org",
		DefinitionsMaxStaleness: "30s",
	})
	assert.ErrorContains(t, err, "shorter than")
}

// The request-building error was the one failure path that skipped the failure
// record, so a URL that could not be turned into a request produced no
// backoff, no log and no context.
func TestCatalogRecordsARequestThatCannotBeBuilt(t *testing.T) {
	srv := newCatalogServer(t)
	svc := manualCatalog(t, srv, config.Settings{})
	svc.baseURL = "http://bad host"

	err := svc.refreshOnce()
	require.ErrorContains(t, err, "failed to build a request")

	_, err = svc.GetDefinitionByID(context.Background(), "toyota_camry_2020")
	assert.ErrorContains(t, err, "failed to build a request", "readers get the recorded failure")
	assert.Equal(t, 1, consecutiveFailures(svc))
}

func TestCatalogPublishesMetrics(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(camryDoc, supraDoc))
	svc := manualCatalog(t, srv, config.Settings{})

	failuresBefore := testutil.ToFloat64(catalogRefreshFailures.WithLabelValues(sourceLegacy))
	require.NoError(t, svc.refreshOnce())
	assert.Equal(t, float64(2), testutil.ToFloat64(catalogDefinitions.WithLabelValues(sourceLegacy)))
	assert.GreaterOrEqual(t, testutil.ToFloat64(catalogSnapshotAge.WithLabelValues(sourceLegacy)), float64(0))

	srv.fail(legacyManifestPath, http.StatusBadGateway)
	require.Error(t, svc.refreshOnce())
	assert.Equal(t, failuresBefore+1, testutil.ToFloat64(catalogRefreshFailures.WithLabelValues(sourceLegacy)))
}

// A service nobody started still works: the first read starts the refresh and
// waits for it, and the loop keeps the catalog current afterwards.
func TestCatalogStartsLazilyOnFirstRead(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(camryDoc))
	svc := newTestCatalog(t, config.Settings{DefinitionsCatalogURL: srv.URL})
	svc.refreshInterval = 20 * time.Millisecond
	ctx := context.Background()

	d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err, "a read must start the refresh and wait for it")
	require.NotNil(t, d)

	srv.serve(legacyManifestPath, manifestOf(camryDoc, supraDoc))
	require.Eventually(t, func() bool {
		supra, err := svc.GetDefinitionByID(ctx, "toyota_supra_2021")
		return err == nil && supra != nil
	}, 5*time.Second, time.Millisecond, "the loop must pick up a changed manifest")
}

package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
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

func newTestCatalog(t *testing.T, settings config.Settings, opts ...DefinitionsCatalogOption) *DefinitionsCatalogService {
	t.Helper()
	logger := zerolog.Nop()
	svc, err := NewDefinitionsCatalogService(&logger, &settings, opts...)
	require.NoError(t, err)
	t.Cleanup(svc.Close)
	return svc
}

// manualCatalog builds a service whose refresh loop never starts, so the test
// drives each refresh with refreshOnce and no assertion races a goroutine.
func manualCatalog(t *testing.T, srv *catalogServer, settings config.Settings, opts ...DefinitionsCatalogOption) *DefinitionsCatalogService {
	t.Helper()
	if settings.DefinitionsCatalogURL == "" {
		settings.DefinitionsCatalogURL = srv.URL
	}
	svc := newTestCatalog(t, settings, opts...)
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
	assert.ErrorContains(t, err, "both answered 404", "the staleness error carries the failure that caused it")
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

// DEFINITIONS_MAX_STALENESS is a reachability bound: confirm() restarts the
// snapshot's age on every 304 and every unchanged build, so a worker whose
// scheduled rebuild has been failing for three weeks -- while idx/current.json
// stays perfectly reachable -- used to read as fresh on every replica. The
// build's own createdAt is the data age, and it needs a bound of its own.
func TestCatalogBoundsTheAgeOfTheBuildItIsServing(t *testing.T) {
	ctx := context.Background()
	recent := time.Now().Add(-25 * time.Hour).UTC().Format(time.RFC3339)
	ancient := time.Now().Add(-21 * 24 * time.Hour).UTC().Format(time.RFC3339)

	t.Run("a build a healthy worker would publish is served", func(t *testing.T) {
		// The worker rebuilds when the live build is 23h old and the walk takes
		// about two hours, so 25h is what a working catalog looks like.
		srv := newCatalogServer(t)
		srv.serve(catalogIndexPath, buildIndexAt("b1", recent, shardKeyFor("b1", 0)))
		srv.serve(shardPathFor("b1", 0), "["+camryTemplate+"]")
		svc := manualCatalog(t, srv, config.Settings{})

		require.NoError(t, svc.refreshOnce())
		d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.NoError(t, err)
		assert.NotNil(t, d)
	})

	t.Run("a build published three weeks ago fails, however reachable the index is", func(t *testing.T) {
		srv := newCatalogServer(t)
		srv.serve(catalogIndexPath, buildIndexAt("b1", ancient, shardKeyFor("b1", 0)))
		srv.serve(shardPathFor("b1", 0), "["+camryTemplate+"]")
		svc := manualCatalog(t, srv, config.Settings{})

		// Every refresh succeeds: the index is up, it simply names a build
		// nobody has replaced.
		require.NoError(t, svc.refreshOnce())
		require.NoError(t, svc.refreshOnce())
		assert.Zero(t, consecutiveFailures(svc))

		_, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.Error(t, err, "data this old must fail rather than be served silently")
		assert.ErrorContains(t, err, "DEFINITIONS_MAX_BUILD_AGE")
		assert.ErrorContains(t, err, "b1")
		_, err = svc.DefinitionsByManufacturer(ctx, 131)
		assert.Error(t, err)

		assert.Greater(t, testutil.ToFloat64(catalogBuildAge.WithLabelValues(sourceTemplate)), float64(20*24*time.Hour/time.Second),
			"the build age is its own metric, so the alert does not wait for queries to fail")

		// Zero disables the bound; the reachability bound is untouched.
		svc.maxBuildAge = 0
		d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.NoError(t, err)
		assert.NotNil(t, d)
	})

	t.Run("a new build restores freshness", func(t *testing.T) {
		srv := newCatalogServer(t)
		srv.serve(catalogIndexPath, buildIndexAt("b1", ancient, shardKeyFor("b1", 0)))
		srv.serve(shardPathFor("b1", 0), "["+camryTemplate+"]")
		svc := manualCatalog(t, srv, config.Settings{})
		require.NoError(t, svc.refreshOnce())
		_, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.Error(t, err)

		srv.serve(catalogIndexPath, buildIndexAt("b2", recent, shardKeyFor("b2", 0)))
		srv.serve(shardPathFor("b2", 0), "["+camryTemplate+"]")
		require.NoError(t, svc.refreshOnce())
		d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.NoError(t, err)
		assert.NotNil(t, d)
	})

	t.Run("the flat manifest has no data age, so only reachability bounds it", func(t *testing.T) {
		// updatedAt moves only when someone writes a definition, so a quiet
		// week and a dead producer look identical. Bounding on it would fail
		// every query on a catalog nobody happened to edit.
		srv := newCatalogServer(t)
		srv.serve(legacyManifestPath, manifestOf(camryDoc))
		svc := manualCatalog(t, srv, config.Settings{})

		require.NoError(t, svc.refreshOnce())
		d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.NoError(t, err)
		require.NotNil(t, d)
		assert.Equal(t, float64(0), testutil.ToFloat64(catalogBuildAge.WithLabelValues(sourceLegacy)),
			"legacy mode publishes no build age rather than a made-up one")
	})

	t.Run("a build with no usable createdAt is served, and bounded only by reachability", func(t *testing.T) {
		srv := newCatalogServer(t)
		srv.serve(catalogIndexPath, buildIndexAt("b1", "not a timestamp", shardKeyFor("b1", 0)))
		srv.serve(shardPathFor("b1", 0), "["+camryTemplate+"]")
		svc := manualCatalog(t, srv, config.Settings{})

		require.NoError(t, svc.refreshOnce(), "a producer that changes the field type must not take the catalog down")
		d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.NoError(t, err)
		assert.NotNil(t, d)
	})

	t.Run("a bound a healthy worker would trip is refused at construction", func(t *testing.T) {
		logger := zerolog.Nop()
		_, err := NewDefinitionsCatalogService(&logger, &config.Settings{
			DefinitionsCatalogURL:  "https://definitions.dimo.org",
			DefinitionsMaxBuildAge: "12h",
		})
		assert.ErrorContains(t, err, "DEFINITIONS_MAX_BUILD_AGE")
	})
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

const (
	camryTemplate = `{"id":"toyota_camry_2020","deviceType":"vehicle",
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"},
	   "model":"Camry","year":2020,"imageURI":"https://img/camry","hardwareTemplateId":"130",
	   "attributes":{"powertrain_type":"ICE","fuel_tank_capacity_gal":15.8,"driven_wheels":4,"has_tow_package":true},
	   "trims":[{"name":"XSE V6","attributes":{"number_of_doors":4}}],
	   "version":3,"author":"0x0000000000000000000000000000000000000001","createdAt":"t","updatedAt":"t"}`
	supraTemplate = `{"id":"toyota_supra_2021","deviceType":"vehicle",
	   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"},
	   "model":"Supra","year":2021,"attributes":{},"trims":[{"name":"3.0","attributes":{}}],
	   "version":1,"createdAt":"t","updatedAt":"t"}`
)

// grenadierTemplate is a definition created since the current build: the only
// kind the by-id fallback ever answers for. The model varies so a test can
// serve a curator's correction of one.
func grenadierTemplate(model string) string {
	return fmt.Sprintf(`{"id":"ineos_grenadier_2026","deviceType":"vehicle",
	   "manufacturer":{"tokenId":142,"slug":"ineos","name":"INEOS"},
	   "model":%q,"year":2026,"attributes":{},"trims":[],
	   "version":1,"createdAt":"t","updatedAt":"t"}`, model)
}

func shardKeyFor(build string, n int) string  { return fmt.Sprintf("idx/%s/shard-%d.json", build, n) }
func shardPathFor(build string, n int) string { return "/" + shardKeyFor(build, n) }

func buildIndexOf(build string, shards ...string) string {
	return buildIndexAt(build, "2026-09-14T00:00:00.000Z", shards...)
}

// buildIndexAt is the index the worker publishes: a build id, the timestamp
// that publish stamped on it, and the shards it named.
func buildIndexAt(build, createdAt string, shards ...string) string {
	quoted := make([]string, len(shards))
	for i, key := range shards {
		quoted[i] = fmt.Sprintf("%q", key)
	}
	return fmt.Sprintf(`{"build":%q,"createdAt":%q,"count":%d,"shards":[%s]}`, build, createdAt, len(shards), strings.Join(quoted, ","))
}

// serveBuild publishes a build: the index that names it and one shard holding
// the given template documents.
func serveBuild(srv *catalogServer, build string, templates ...string) {
	srv.serve(catalogIndexPath, buildIndexOf(build, shardKeyFor(build, 0)))
	srv.serve(shardPathFor(build, 0), "["+strings.Join(templates, ",")+"]")
}

// The published build is the catalog once the worker serves an index, and its
// templates flatten exactly as device-definitions-api flattens them.
func TestCatalogAdoptsATemplateBuild(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(catalogIndexPath, buildIndexOf("b1", shardKeyFor("b1", 0), shardKeyFor("b1", 1)))
	srv.serve(shardPathFor("b1", 0), "["+camryTemplate+"]")
	srv.serve(shardPathFor("b1", 1), "["+supraTemplate+"]")
	svc := manualCatalog(t, srv, config.Settings{})
	ctx := context.Background()

	require.NoError(t, svc.refreshOnce())
	assert.Zero(t, srv.hitsFor(legacyManifestPath), "the flat manifest must not be read when an index exists")

	d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)
	assert.Equal(t, "Camry", d.Model)
	assert.Equal(t, 2020, d.Year)
	assert.Equal(t, "vehicle", d.DeviceType)
	assert.Equal(t, "https://img/camry", d.ImageURI)
	assert.Equal(t, CatalogManufacturer{TokenID: 131, Slug: "toyota", Name: "Toyota"}, d.Manufacturer)
	assert.Empty(t, d.KSUID, "templates have no ksuid, and inventing one would fill an identifier field")
	assert.Equal(t, []CatalogDefinitionAttribute{
		{Name: "driven_wheels", Value: "4"},
		{Name: "fuel_tank_capacity_gal", Value: "15.8"},
		{Name: "has_tow_package", Value: "true"},
		{Name: "powertrain_type", Value: "ICE"},
	}, d.Metadata.DeviceAttributes, "template attributes only, sorted by name, rendered as dd-api renders them")

	defs, err := svc.DefinitionsByManufacturer(ctx, 131)
	require.NoError(t, err)
	require.Len(t, defs, 2, "every shard of the build is part of the catalog")
	assert.Equal(t, "toyota_camry_2020", defs[0].ID)
	assert.Equal(t, "toyota_supra_2021", defs[1].ID)
	assert.Equal(t, sourceTemplate, svc.snap.Load().source)
}

// Shards are immutable, so an index that names the same build, published at
// the same moment, with the same shards, means there is nothing to fetch.
// Refreshing a 17,000-definition catalog every minute because the index was
// re-read is exactly the download this design removes.
func TestCatalogSkipsTheShardsWhenTheBuildIsUnchanged(t *testing.T) {
	srv := newCatalogServer(t)
	index := buildIndexOf("b1", shardKeyFor("b1", 0))
	srv.handle(catalogIndexPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Etag", `W/"b1"`)
		if r.Header.Get("If-None-Match") == `W/"b1"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		_, _ = io.WriteString(w, index)
	})
	srv.serve(shardPathFor("b1", 0), "["+camryTemplate+"]")
	svc := manualCatalog(t, srv, config.Settings{})

	require.NoError(t, svc.refreshOnce())
	require.Equal(t, 1, srv.hitsFor(shardPathFor("b1", 0)))

	// A 304 on the index: unchanged, and the snapshot's age restarts.
	require.NoError(t, svc.refreshOnce())
	assert.Equal(t, 1, srv.hitsFor(shardPathFor("b1", 0)), "a 304 must not refetch the build")

	// A 200 naming the build already held: still nothing to fetch.
	srv.serve(catalogIndexPath, index)
	require.NoError(t, svc.refreshOnce())
	assert.Equal(t, 1, srv.hitsFor(shardPathFor("b1", 0)), "an unchanged index must not refetch the build")
	assert.Equal(t, 3, srv.hitsFor(catalogIndexPath))

	// A new build is fetched.
	serveBuild(srv, "b2", camryTemplate, supraTemplate)
	require.NoError(t, svc.refreshOnce())
	assert.Equal(t, 1, srv.hitsFor(shardPathFor("b2", 0)))
	assert.Equal(t, "b2", svc.snap.Load().build)
}

// A build id is not a version: idx/current.json and idx/<build>/manifest.json
// are rewritten on every publish of the same id (definitions-worker
// src/idx.ts, "Not write-once, so not immutable ... a publish repeated after
// more pages landed names more shards"). Keying freshness on the id alone
// served the short catalog of an early publish forever.
func TestCatalogReloadsARepublishedBuild(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(catalogIndexPath, buildIndexAt("b1", "2026-09-14T00:00:00.000Z", shardKeyFor("b1", 0)))
	srv.serve(shardPathFor("b1", 0), "["+camryTemplate+"]")
	srv.serve("/t/toyota_ghost_2020.json", "")
	svc := manualCatalog(t, srv, config.Settings{})
	ctx := context.Background()

	require.NoError(t, svc.refreshOnce())
	require.Equal(t, 1, svc.snap.Load().size())

	// The by-id fallback learns that a definition is missing. The longer
	// catalog below holds it, so a republish has to clear what it learned.
	srv.fail("/t/toyota_supra_2021.json", http.StatusNotFound)
	d, err := svc.GetDefinitionByID(ctx, "toyota_supra_2021")
	require.NoError(t, err)
	require.Nil(t, d)

	// The operator published b1 before the walk finished, then finished it and
	// published b1 again: the same id, a new createdAt, a longer shard list.
	srv.serve(catalogIndexPath, buildIndexAt("b1", "2026-09-14T02:00:00.000Z", shardKeyFor("b1", 0), shardKeyFor("b1", 1)))
	srv.serve(shardPathFor("b1", 1), "["+supraTemplate+"]")
	require.NoError(t, svc.refreshOnce())

	assert.Equal(t, 2, svc.snap.Load().size(), "a republished build must be reloaded")
	assert.Equal(t, 1, srv.hitsFor(shardPathFor("b1", 1)))
	d, err = svc.GetDefinitionByID(ctx, "toyota_supra_2021")
	require.NoError(t, err)
	require.NotNil(t, d, "the republished build must clear what the by-id fallback learned about the old one")

	t.Run("a retried publish that adds no shards still reloads", func(t *testing.T) {
		before := srv.hitsFor(shardPathFor("b1", 0))
		srv.serve(catalogIndexPath, buildIndexAt("b1", "2026-09-14T03:00:00.000Z", shardKeyFor("b1", 0), shardKeyFor("b1", 1)))
		require.NoError(t, svc.refreshOnce())
		assert.Equal(t, before+1, srv.hitsFor(shardPathFor("b1", 0)), "a new createdAt is a new publish")
	})

	t.Run("an index naming the same build unchanged fetches nothing", func(t *testing.T) {
		before := srv.hitsFor(shardPathFor("b1", 0))
		require.NoError(t, svc.refreshOnce())
		require.NoError(t, svc.refreshOnce())
		assert.Equal(t, before, srv.hitsFor(shardPathFor("b1", 0)),
			"an identical index must not refetch a 17,000-definition catalog every minute")
	})
}

// The deployed worker serves only manifest.json and answers 404 for the index,
// so identity has to read either layout and pick per refresh. A 5xx is not a
// missing index and must never flip the source.
func TestCatalogFallsBackToTheLegacyManifestWhenTheIndexIsMissing(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(camryDoc))
	svc := manualCatalog(t, srv, config.Settings{})
	ctx := context.Background()

	// No index deployed yet: the flat manifest is the catalog.
	require.NoError(t, svc.refreshOnce())
	assert.Equal(t, sourceLegacy, svc.snap.Load().source)
	d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)

	// The worker is deployed: the build takes over.
	serveBuild(srv, "b1", camryTemplate, supraTemplate)
	require.NoError(t, svc.refreshOnce())
	assert.Equal(t, sourceTemplate, svc.snap.Load().source)
	assert.Equal(t, "b1", svc.snap.Load().build)

	// The origin breaks: the held build stays, and the legacy manifest is not
	// read behind its back.
	legacyHits := srv.hitsFor(legacyManifestPath)
	srv.fail(catalogIndexPath, http.StatusServiceUnavailable)
	require.ErrorContains(t, svc.refreshOnce(), "definitions index returned 503")
	assert.Equal(t, legacyHits, srv.hitsFor(legacyManifestPath), "a 5xx on the index must not flip the source")
	assert.Equal(t, sourceTemplate, svc.snap.Load().source)
	d, err = svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)
}

// The fallback may bootstrap a pod, and may never downgrade one. idx/current.json
// is served with a minute of max-age and an R2 get can return null while a
// publish rewrites it, so a single 404 on it is a refresh that failed, not a
// deployment that publishes no builds -- while /manifest.json can still answer
// 200 from an edge that cached it before the route was deleted. Reading it
// would swap the build this pod serves for the pre-cutover flat documents,
// turn the by-id template fallback off, and stop the data-age bound measuring
// anything, all at once: legacy mode has no build timestamp, and every
// successful legacy refresh restarts the reachability bound, so nothing after
// the swap would ever say so.
func TestCatalogNeverDowngradesAHeldTemplateBuildToTheLegacyManifest(t *testing.T) {
	srv := newCatalogServer(t)
	serveBuild(srv, "b1", camryTemplate, supraTemplate)
	srv.serve(legacyManifestPath, manifestOf(camryDoc, supraDoc))
	svc := manualCatalog(t, srv, config.Settings{})
	ctx := context.Background()

	require.NoError(t, svc.refreshOnce())
	require.Equal(t, sourceTemplate, svc.snap.Load().source)
	legacyHits := srv.hitsFor(legacyManifestPath)

	srv.fail(catalogIndexPath, http.StatusNotFound)
	err := svc.refreshOnce()
	require.Error(t, err, "once a build has been adopted, a missing index is a refresh failure")
	assert.ErrorIs(t, err, errCatalogWouldDowngrade)
	assert.ErrorContains(t, err, srv.URL+catalogIndexPath, "the failure names what answered 404")
	assert.ErrorContains(t, err, "b1", "and the build it is refusing to give up")
	assert.Equal(t, legacyHits, srv.hitsFor(legacyManifestPath), "the flat manifest must not even be read")
	assert.Equal(t, 1, consecutiveFailures(svc),
		"the failure is recorded, so the staleness bound turns a lasting one into errors")

	snap := svc.snap.Load()
	assert.Equal(t, sourceTemplate, snap.source)
	assert.Equal(t, "b1", snap.build)
	assert.False(t, snap.builtAt.IsZero(), "the data-age bound still has a timestamp to measure")
	d, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, d)
	assert.Equal(t, "https://img/camry", d.ImageURI, "still the build's data, not the flat manifest's")

	t.Run("the by-id fallback still answers while the index is missing", func(t *testing.T) {
		srv.serve("/t/ineos_grenadier_2026.json", grenadierTemplate("Grenadier"))
		d, err := svc.GetDefinitionByID(ctx, "ineos_grenadier_2026")
		require.NoError(t, err)
		require.NotNil(t, d, "the pod is still in template mode, so a miss is still looked up")
	})

	t.Run("the index coming back is a recovery, not a new build", func(t *testing.T) {
		before := srv.hitsFor(shardPathFor("b1", 0))
		serveBuild(srv, "b1", camryTemplate, supraTemplate)
		require.NoError(t, svc.refreshOnce())
		assert.Zero(t, consecutiveFailures(svc))
		assert.Equal(t, sourceTemplate, svc.snap.Load().source)
		assert.Equal(t, before, srv.hitsFor(shardPathFor("b1", 0)), "the same build is confirmed, not refetched")
	})

	// The other direction: the gate must not keep a pod that holds nothing
	// from bootstrapping, which is the state every pod is in before the
	// cutover and after every restart until the first index answers.
	t.Run("a pod that holds no snapshot still bootstraps from the flat manifest", func(t *testing.T) {
		srv.fail(catalogIndexPath, http.StatusNotFound)
		cold := manualCatalog(t, srv, config.Settings{})

		require.NoError(t, cold.refreshOnce())
		require.Equal(t, sourceLegacy, cold.snap.Load().source)
		d, err := cold.GetDefinitionByID(ctx, "toyota_camry_2020")
		require.NoError(t, err)
		require.NotNil(t, d)

		// And keeps refreshing it, so the gate costs the pre-cutover
		// deployment nothing.
		before := srv.hitsFor(legacyManifestPath)
		require.NoError(t, cold.refreshOnce())
		assert.Equal(t, before+1, srv.hitsFor(legacyManifestPath))
		assert.Equal(t, sourceLegacy, cold.snap.Load().source)
	})
}

// Between the worker deploy and the first hand-run publish, both surfaces
// answer 404: the index does not exist yet and /manifest.json is a route the
// same migration deleted. The 404-means-legacy branch turned a missing index
// into a second guaranteed failure, and reported it as "definitions catalog
// returned 404 for manifest" -- which names neither URL and reads like a
// transient origin error rather than a deploy run out of order.
func TestCatalogSaysSoWhenNothingIsPublished(t *testing.T) {
	srv := newCatalogServer(t)
	svc := manualCatalog(t, srv, config.Settings{})

	err := svc.refreshOnce()
	require.Error(t, err)
	assert.ErrorContains(t, err, srv.URL+catalogIndexPath)
	assert.ErrorContains(t, err, srv.URL+legacyManifestPath)
	assert.ErrorContains(t, err, "no build published")
	assert.ErrorIs(t, err, errNoCatalogPublished)

	_, err = svc.GetDefinitionByID(context.Background(), "toyota_camry_2020")
	assert.ErrorIs(t, err, errNoCatalogPublished, "a cold pod reports why it has no catalog")

	// The fallback is still the real source before the cutover, so a manifest
	// that exists is read as before.
	srv.serve(legacyManifestPath, manifestOf(camryDoc))
	require.NoError(t, svc.refreshOnce())
	assert.Equal(t, sourceLegacy, svc.snap.Load().source)
}

// A listing only changes when a build is published, so a definition created
// since then is missing from the snapshot. By id, it is fetched from the
// template document the worker wrote when it was created.
func TestCatalogLooksUpAMissingDefinitionByID(t *testing.T) {
	srv := newCatalogServer(t)
	serveBuild(srv, "b1", camryTemplate)
	svc := manualCatalog(t, srv, config.Settings{})
	ctx := context.Background()
	require.NoError(t, svc.refreshOnce())

	t.Run("found, then answered from the cache", func(t *testing.T) {
		srv.serve("/t/toyota_supra_2021.json", supraTemplate)
		for range 3 {
			d, err := svc.GetDefinitionByID(ctx, "toyota_supra_2021")
			require.NoError(t, err)
			require.NotNil(t, d)
			assert.Equal(t, "Supra", d.Model)
			assert.Equal(t, 131, d.Manufacturer.TokenID)
		}
		assert.Equal(t, 1, srv.hitsFor("/t/toyota_supra_2021.json"), "a found template is cached for this build")
	})

	t.Run("missing, then answered from the negative cache", func(t *testing.T) {
		for range 3 {
			d, err := svc.GetDefinitionByID(ctx, "toyota_ghost_2020")
			require.NoError(t, err, "a definition that does not exist is not an error")
			assert.Nil(t, d)
		}
		assert.Equal(t, 1, srv.hitsFor("/t/toyota_ghost_2020.json"), "a 404 is remembered for its TTL")
	})

	t.Run("an id that needs escaping", func(t *testing.T) {
		srv.serve("/t/dodge_town-&-country_2012.json", strings.Replace(camryTemplate, "toyota_camry_2020", "dodge_town-&-country_2012", 1))
		d, err := svc.GetDefinitionByID(ctx, "dodge_town-&-country_2012")
		require.NoError(t, err)
		require.NotNil(t, d)
		assert.Equal(t, "dodge_town-&-country_2012", d.ID)
	})

	t.Run("any other failure is reported", func(t *testing.T) {
		srv.fail("/t/toyota_broken_2020.json", http.StatusInternalServerError)
		_, err := svc.GetDefinitionByID(ctx, "toyota_broken_2020")
		require.ErrorContains(t, err, "returned 500")

		srv.serve("/t/toyota_mislabelled_2020.json", supraTemplate)
		_, err = svc.GetDefinitionByID(ctx, "toyota_mislabelled_2020")
		require.ErrorContains(t, err, "carries the id")
	})

	// A manufacturer token id is required, and the producer side enforces it.
	// A stored document that lacks one is still bad data rather than a bad
	// request: it is skipped in a listing, so by id it is missing too, not a
	// 500 for the caller to make sense of.
	t.Run("a template that fails validation is missing, not an error", func(t *testing.T) {
		srv.serve("/t/toyota_invalid_2020.json", `{"id":"toyota_invalid_2020","model":"X","manufacturer":{"slug":"toyota"}}`)
		for range 3 {
			d, err := svc.GetDefinitionByID(ctx, "toyota_invalid_2020")
			require.NoError(t, err)
			assert.Nil(t, d)
		}
		assert.Equal(t, 1, srv.hitsFor("/t/toyota_invalid_2020.json"),
			"an unservable document is remembered like a 404, so it is not refetched per query")
	})

	t.Run("a new build clears what the fallback learned", func(t *testing.T) {
		before := srv.hitsFor("/t/toyota_supra_2021.json")
		serveBuild(srv, "b2", camryTemplate)
		require.NoError(t, svc.refreshOnce())

		d, err := svc.GetDefinitionByID(ctx, "toyota_supra_2021")
		require.NoError(t, err)
		require.NotNil(t, d)
		assert.Equal(t, before+1, srv.hitsFor("/t/toyota_supra_2021.json"))

		d, err = svc.GetDefinitionByID(ctx, "toyota_ghost_2020")
		require.NoError(t, err)
		assert.Nil(t, d)
		assert.Equal(t, 2, srv.hitsFor("/t/toyota_ghost_2020.json"))
	})

	t.Run("expired negative entries are refetched", func(t *testing.T) {
		// The TTL is stamped on the entry when it is written, so an id that has
		// not been looked up yet is the one that shows expiry working.
		svc.missingTTL = -time.Second
		for range 2 {
			d, err := svc.GetDefinitionByID(ctx, "toyota_phantom_2020")
			require.NoError(t, err)
			assert.Nil(t, d)
		}
		assert.Equal(t, 2, srv.hitsFor("/t/toyota_phantom_2020.json"), "an expired entry must not answer a later query")
	})
	// The positive entry needs a bound of its own. This path only ever answers
	// for a definition newer than the current build -- exactly the one a
	// curator is still correcting -- and the worker purges the CDN on the
	// write so that this fallback sees the edit. With no expiry the first
	// version answered was served until the next build was published, up to
	// 23 hours, and only on the replicas that had answered once, so the same
	// query returned different documents depending on which pod took it.
	t.Run("a found entry expires and is refetched", func(t *testing.T) {
		const id = "ineos_grenadier_2026"
		const path = "/t/" + id + ".json"
		srv.serve(path, grenadierTemplate("Grenadeer"))

		// The lifetime is stamped on the entry when it is written, so it has
		// to be set before the first lookup of this id.
		t.Cleanup(func() { svc.foundTTL = catalogFoundTTL })
		svc.foundTTL = -time.Second

		d, err := svc.GetDefinitionByID(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, d)
		require.Equal(t, "Grenadeer", d.Model)

		// A curator fixes the model name; the worker writes v2 and purges.
		srv.serve(path, grenadierTemplate("Grenadier"))
		d, err = svc.GetDefinitionByID(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, d)
		assert.Equal(t, "Grenadier", d.Model, "an expired entry must be refetched, not answered from memory")
		assert.Equal(t, 2, srv.hitsFor(path))
	})
}

// In legacy mode there is nothing to fall back to: the manifest is the whole
// catalog, and a miss is a miss.
func TestCatalogDoesNotLookUpTemplatesInLegacyMode(t *testing.T) {
	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(camryDoc))
	srv.serve("/t/toyota_supra_2021.json", supraTemplate)
	svc := manualCatalog(t, srv, config.Settings{})
	require.NoError(t, svc.refreshOnce())

	d, err := svc.GetDefinitionByID(context.Background(), "toyota_supra_2021")
	require.NoError(t, err)
	assert.Nil(t, d)
	assert.Zero(t, srv.hitsFor("/t/toyota_supra_2021.json"))
}

func TestCatalogRefusesABrokenBuild(t *testing.T) {
	srv := newCatalogServer(t)
	ctx := context.Background()

	for name, tc := range map[string]struct {
		index  string
		expect string
	}{
		"no build":      {index: `{"shards":["idx/b1/shard-0.json"]}`, expect: "names no build"},
		"no shards":     {index: `{"build":"b1","shards":[]}`, expect: "lists no shards"},
		"foreign shard": {index: buildIndexOf("b1", "idx/b2/shard-0.json"), expect: "not one of its shard keys"},
		"absolute url":  {index: buildIndexOf("b1", "https://elsewhere.example/shard-0.json"), expect: "not one of its shard keys"},
		"path escape":   {index: buildIndexOf("b1", "idx/b1/../../secret.json"), expect: "not one of its shard keys"},
		"broken index":  {index: `{"build":`, expect: "failed to decode the definitions index"},
	} {
		srv.serve(catalogIndexPath, tc.index)
		svc := manualCatalog(t, srv, config.Settings{})
		require.ErrorContains(t, svc.refreshOnce(), tc.expect, name)
		_, err := svc.GetDefinitionByID(ctx, "toyota_camry_2020")
		assert.Error(t, err, "%s: a pod with no catalog reports why", name)
	}

	t.Run("a shard that cannot be read fails the build", func(t *testing.T) {
		srv.serve(catalogIndexPath, buildIndexOf("b1", shardKeyFor("b1", 0), shardKeyFor("b1", 1)))
		srv.serve(shardPathFor("b1", 0), "["+camryTemplate+"]")
		srv.fail(shardPathFor("b1", 1), http.StatusInternalServerError)
		svc := manualCatalog(t, srv, config.Settings{})
		require.ErrorContains(t, svc.refreshOnce(), "returned 500 for shard")

		srv.serve(shardPathFor("b1", 1), `[{"id":`)
		svc = manualCatalog(t, srv, config.Settings{})
		require.ErrorContains(t, svc.refreshOnce(), "failed to decode shard")
	})
}

// The catalog carries no chain marker, and nothing ties DEFINITIONS_CATALOG_URL
// to DIMO_REGISTRY_CHAIN_ID. 108 of the 123 slugs dev and prod share have
// different token ids, so a deployment pointed at the other chain's catalog
// used to list another make's definitions under every manufacturer, silently.
func TestCatalogChecksACandidateAgainstThisChain(t *testing.T) {
	chain := map[int]string{}
	docs := make([]string, 0, 40)
	for token := 1; token <= 40; token++ {
		slug := fmt.Sprintf("mfr%d", token)
		chain[token] = slug
		docs = append(docs, fmt.Sprintf(`{"id":"%s_model_2020","model":"M","year":2020,
		  "manufacturer":{"tokenId":%d,"slug":%q,"name":"M"}}`, slug, token, slug))
	}

	srv := newCatalogServer(t)
	srv.serve(legacyManifestPath, manifestOf(docs...))
	ctx := context.Background()

	// renamed returns this chain's manufacturers with n of them renamed, as a
	// catalog from another chain would look: the same token ids, other slugs.
	renamed := func(n int) map[int]string {
		out := make(map[int]string, len(chain))
		maps.Copy(out, chain)
		for token := 1; token <= n; token++ {
			out[token] = fmt.Sprintf("other%d", token)
		}
		return out
	}
	lookup := func(slugs map[int]string, err error) ManufacturerSlugs {
		return func(context.Context) (map[int]string, error) { return slugs, err }
	}
	adopted := func(t *testing.T, svc *DefinitionsCatalogService) {
		t.Helper()
		d, err := svc.GetDefinitionByID(ctx, "mfr1_model_2020")
		require.NoError(t, err)
		assert.NotNil(t, d)
	}

	t.Run("a catalog that agrees is adopted", func(t *testing.T) {
		svc := manualCatalog(t, srv, config.Settings{}, WithManufacturerSlugs(lookup(chain, nil)))
		require.NoError(t, svc.refreshOnce())
		adopted(t, svc)
	})

	t.Run("a renamed slug or two is drift, not another chain", func(t *testing.T) {
		svc := manualCatalog(t, srv, config.Settings{}, WithManufacturerSlugs(lookup(renamed(2), nil)))
		require.NoError(t, svc.refreshOnce())
		adopted(t, svc)
	})

	t.Run("a catalog from another chain is refused", func(t *testing.T) {
		svc := manualCatalog(t, srv, config.Settings{}, WithManufacturerSlugs(lookup(renamed(3), nil)))
		err := svc.refreshOnce()
		require.ErrorContains(t, err, "carry a different slug")
		assert.ErrorContains(t, err, `token 1 is "other1" here and "mfr1" in the catalog`)

		_, err = svc.GetDefinitionByID(ctx, "mfr1_model_2020")
		assert.Error(t, err, "nothing from the other chain is served")
	})

	t.Run("only manufacturers this chain has are compared", func(t *testing.T) {
		// Five shared token ids, all agreeing: the other 35 say nothing about
		// which chain the catalog belongs to.
		few := map[int]string{1: "mfr1", 2: "mfr2", 3: "mfr3", 4: "mfr4", 5: "mfr5"}
		svc := manualCatalog(t, srv, config.Settings{}, WithManufacturerSlugs(lookup(few, nil)))
		require.NoError(t, svc.refreshOnce())
		adopted(t, svc)

		// Three of those five disagreeing is past the limit of two.
		few[1], few[2], few[3] = "other1", "other2", "other3"
		svc = manualCatalog(t, srv, config.Settings{}, WithManufacturerSlugs(lookup(few, nil)))
		require.ErrorContains(t, svc.refreshOnce(), "3 of the 5 manufacturers shared with this chain")
	})

	t.Run("a lookup failure adopts anyway", func(t *testing.T) {
		svc := manualCatalog(t, srv, config.Settings{}, WithManufacturerSlugs(lookup(nil, errors.New("database is down"))))
		require.NoError(t, svc.refreshOnce(), "the database is this check's second opinion, not the catalog's source")
		adopted(t, svc)
	})

	t.Run("no lookup configured adopts", func(t *testing.T) {
		svc := manualCatalog(t, srv, config.Settings{})
		require.NoError(t, svc.refreshOnce())
		adopted(t, svc)
	})
}

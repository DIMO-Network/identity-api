package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"math/rand/v2"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

// The two catalog layouts this service can read.
const (
	// sourceTemplate is the published template build: an index naming a build,
	// and that build's immutable shards of Template documents.
	sourceTemplate = "template"
	// sourceLegacy is the flat manifest.json the deployed worker serves today.
	sourceLegacy = "legacy"
)

const (
	catalogIndexPath   = "/idx/current.json"
	legacyManifestPath = "/manifest.json"
	templatePathPrefix = "/t/"
)

// catalogShardConcurrency is how many shards of a build are fetched at once. A
// build runs to about a hundred shards, and a refresh has refreshTimeout for
// all of them.
const catalogShardConcurrency = 8

// Bounds on the by-id fallback caches. They keep a burst of lookups for the
// same few ids off the origin; they do not mirror the catalog.
const (
	catalogFoundCacheSize   = 1024
	catalogMissingCacheSize = 4096
	// catalogMissingTTL is short because an id that misses is usually one that
	// was just created, and it becomes reachable the moment it is published.
	catalogMissingTTL = 5 * time.Minute
	// catalogLookupTimeout bounds one by-id fetch whatever the caller's
	// deadline: the fetch is shared, so it is not the first caller's to cancel.
	catalogLookupTimeout = 10 * time.Second
)

// shardKeyRE matches the shard keys a build manifest lists,
// idx/<build>/shard-<n>.json (definitions-worker src/build.ts, shardKey).
var shardKeyRE = regexp.MustCompile(`^idx/([A-Za-z0-9_-]+)/shard-[0-9]+\.json$`)

// buildManifest is the part of the worker's BuildManifest (definitions-worker
// src/build.ts) this service reads. count and createdAt are deliberately not
// decoded: nothing here depends on them, and decoding them strictly would let a
// type change in either fail the whole index.
type buildManifest struct {
	Build  string   `json:"build"`
	Shards []string `json:"shards"`
}

// templateLookups caches what the by-id fallback learned about one build. A new
// build gets a new set: its listing already carries whatever the fallback went
// to find, and a definition deleted since must not survive in a cache.
type templateLookups struct {
	found   *lru.Cache[string, *CatalogDefinition]
	missing *lru.Cache[string, time.Time]
}

func newTemplateLookups() *templateLookups {
	// lru.New only fails on a size below one, and both sizes are constants.
	found, _ := lru.New[string, *CatalogDefinition](catalogFoundCacheSize)
	missing, _ := lru.New[string, time.Time](catalogMissingCacheSize)
	return &templateLookups{found: found, missing: missing}
}

// catalogFailuresBeforeError is how many refreshes must fail in a row before
// the failure is logged at Error. One is noise: an origin hiccup, a rolling
// deploy. Five in a row is an outage that the staleness bound will turn into
// failing queries.
const catalogFailuresBeforeError = 5

var (
	catalogDefinitions = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "identity",
		Subsystem: "definitions_catalog",
		Name:      "definitions",
		Help:      "Device definitions in the snapshot being served.",
	}, []string{"source"})

	catalogSnapshotAge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "identity",
		Subsystem: "definitions_catalog",
		Name:      "snapshot_age_seconds",
		Help:      "Age of the snapshot being served, as of the last refresh attempt.",
	}, []string{"source"})

	catalogRefreshFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "identity",
		Subsystem: "definitions_catalog",
		Name:      "refresh_failures_total",
		Help:      "Refresh attempts that produced no usable catalog, by the source being read.",
	}, []string{"source"})
)

// catalogSnapshot is an immutable view of the catalog. The refresh goroutine
// builds a new one and swaps it in; a reader loads the pointer and takes no
// lock, so a refresh cannot block a query however long it runs.
type catalogSnapshot struct {
	byID       map[string]*CatalogDefinition
	byMfrToken map[int][]*CatalogDefinition
	// source is which catalog layout this was read from.
	source string
	// build is the template build this was read from; empty in legacy mode.
	build string
	// etag conditions the next read of whatever document produced it.
	etag        string
	lastSuccess time.Time
}

func newCatalogSnapshot(defs []CatalogDefinition, source, build, etag string) *catalogSnapshot {
	byID := make(map[string]*CatalogDefinition, len(defs))
	byMfr := map[int][]*CatalogDefinition{}
	for i := range defs {
		d := &defs[i]
		byID[d.ID] = d
		byMfr[d.Manufacturer.TokenID] = append(byMfr[d.Manufacturer.TokenID], d)
	}
	for _, list := range byMfr {
		slices.SortFunc(list, func(a, b *CatalogDefinition) int { return strings.Compare(a.ID, b.ID) })
	}
	return &catalogSnapshot{
		byID:        byID,
		byMfrToken:  byMfr,
		source:      source,
		build:       build,
		etag:        etag,
		lastSuccess: time.Now(),
	}
}

func (s *catalogSnapshot) size() int {
	if s == nil {
		return 0
	}
	return len(s.byID)
}

// etagFor returns the ETag a read of source may be conditioned on, or "" when
// this snapshot did not come from that source.
func (s *catalogSnapshot) etagFor(source string) string {
	if s == nil || s.source != source {
		return ""
	}
	return s.etag
}

// ManufacturerSlugs returns every manufacturer's slug keyed by token id, as
// this deployment's chain knows them.
type ManufacturerSlugs func(ctx context.Context) (map[int]string, error)

// DefinitionsCatalogOption configures the service at construction.
type DefinitionsCatalogOption func(*DefinitionsCatalogService)

// WithManufacturerSlugs gives the service its chain check: a candidate catalog
// is compared with the manufacturers this deployment has before it is adopted.
func WithManufacturerSlugs(lookup ManufacturerSlugs) DefinitionsCatalogOption {
	return func(s *DefinitionsCatalogService) { s.manufacturerSlugs = lookup }
}

// DefinitionsCatalogService keeps the R2 device definitions catalog in memory.
// One goroutine refreshes it and swaps in an immutable snapshot; reads are
// served from that snapshot and never wait on network I/O. It replaces the
// per-query Tableland proxy.
type DefinitionsCatalogService struct {
	log     *zerolog.Logger
	baseURL string
	client  *http.Client

	minCount     int
	maxStaleness time.Duration

	refreshInterval time.Duration
	// refreshTimeout bounds one refresh. A refresh is nobody's request: it runs
	// on its own context, so a caller that goes away cannot cancel it and a
	// stalled origin cannot hold it forever.
	refreshTimeout time.Duration
	// initialBackoff is the delay before the first retry after a failed
	// refresh. It doubles, with jitter, up to refreshInterval, so a transient
	// failure costs about a second rather than a whole interval.
	initialBackoff time.Duration
	// coldWait bounds how long a read on a pod that has never loaded a snapshot
	// waits for the refresh in flight. Without it a persistently failing
	// catalog would hang every device-definition query instead of failing it.
	coldWait time.Duration
	// lookupTimeout and missingTTL bound the by-id fallback.
	lookupTimeout time.Duration
	missingTTL    time.Duration
	// manufacturerSlugs is the chain check's view of this deployment. Optional:
	// without it a candidate is adopted unchecked.
	manufacturerSlugs ManufacturerSlugs

	startOnce sync.Once
	loopCtx   context.Context
	stop      context.CancelFunc

	snap atomic.Pointer[catalogSnapshot]

	// lookups serves the by-id fallback for the build snap holds.
	lookups     atomic.Pointer[templateLookups]
	lookupGroup singleflight.Group

	mu sync.Mutex
	// lastErr is the latest failure, stored exactly as callers receive it.
	lastErr error
	// failures counts refreshes that have failed in a row.
	failures int
	// attemptDone is closed when the refresh in flight ends, then replaced.
	// A read on a pod with no snapshot waits on it.
	attemptDone chan struct{}
	// metricSource is the source the gauges carry a series for. Only the
	// refresh goroutine touches it.
	metricSource string
}

// NewDefinitionsCatalogService validates the DEFINITIONS_* settings and builds
// the service. A malformed catalog URL, floor or staleness bound fails
// construction, and with it startup, instead of starting a pod whose every
// device-definition query fails or whose floor is silently disabled.
//
// It contacts nothing: call Start at process start so a pod warms before it
// takes traffic. A first read starts the refresh too, so a caller that never
// calls Start still works.
func NewDefinitionsCatalogService(log *zerolog.Logger, settings *config.Settings, opts ...DefinitionsCatalogOption) (*DefinitionsCatalogService, error) {
	cfg, err := settings.DefinitionsCatalog()
	if err != nil {
		return nil, err
	}
	loopCtx, stop := context.WithCancel(context.Background())
	s := &DefinitionsCatalogService{
		log:     log,
		baseURL: cfg.URL,
		// No client timeout: every request carries a context deadline, and a
		// second bound only makes it a mystery which one fired.
		client:          &http.Client{},
		minCount:        cfg.MinCount,
		maxStaleness:    cfg.MaxStaleness,
		refreshInterval: time.Minute,
		refreshTimeout:  30 * time.Second,
		initialBackoff:  time.Second,
		coldWait:        10 * time.Second,
		lookupTimeout:   catalogLookupTimeout,
		missingTTL:      catalogMissingTTL,
		loopCtx:         loopCtx,
		stop:            stop,
		attemptDone:     make(chan struct{}),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.lookups.Store(newTemplateLookups())
	if s.maxStaleness > 0 && s.maxStaleness < s.refreshInterval {
		stop()
		return nil, fmt.Errorf("DEFINITIONS_MAX_STALENESS %s is shorter than the %s refresh interval: every replica would report a stale catalog between refreshes", s.maxStaleness, s.refreshInterval)
	}
	return s, nil
}

// Start begins refreshing the catalog in the background.
func (s *DefinitionsCatalogService) Start() {
	s.startOnce.Do(func() { go s.run() })
}

// Close stops the refresh goroutine.
func (s *DefinitionsCatalogService) Close() {
	s.stop()
}

// GetDefinitionByID returns the definition with the given slug id, or nil.
//
// In template mode the listing only changes when a build is published, so a
// definition created since that build is not in the snapshot. A miss falls back
// to the template document itself, which the worker writes as soon as the
// definition is created.
func (s *DefinitionsCatalogService) GetDefinitionByID(ctx context.Context, id string) (*CatalogDefinition, error) {
	snap, err := s.snapshot(ctx)
	if err != nil {
		return nil, err
	}
	if d := snap.byID[id]; d != nil {
		return d, nil
	}
	if snap.source != sourceTemplate || id == "" {
		return nil, nil
	}
	return s.lookupTemplate(ctx, id)
}

// DefinitionsByManufacturer returns the manufacturer's definitions sorted by
// id. It answers from the snapshot alone, with no per-query fallback: a listing
// is a page of results, and going to the origin for it would put the catalog
// back in the request path, which is what this service exists to avoid. A
// definition published since the last build is reachable by id, and joins the
// listing at the next build.
func (s *DefinitionsCatalogService) DefinitionsByManufacturer(ctx context.Context, manufacturerTokenID int) ([]*CatalogDefinition, error) {
	snap, err := s.snapshot(ctx)
	if err != nil {
		return nil, err
	}
	return snap.byMfrToken[manufacturerTokenID], nil
}

// snapshot returns the catalog to read from. A pod that holds one never waits:
// it loads the pointer and checks its age. A pod that has never loaded one
// waits for the refresh in flight to finish, bounded by the caller's context
// and by coldWait, and reports the failure when that attempt fails.
func (s *DefinitionsCatalogService) snapshot(ctx context.Context) (*catalogSnapshot, error) {
	s.Start()
	snap := s.snap.Load()
	if snap == nil {
		s.mu.Lock()
		done := s.attemptDone
		s.mu.Unlock()
		// The attempt may have finished between the load and the lock. It
		// stores the snapshot before closing the channel, so re-read first.
		if snap = s.snap.Load(); snap == nil {
			timer := time.NewTimer(s.coldWait)
			defer timer.Stop()
			select {
			case <-done:
			case <-timer.C:
			case <-ctx.Done():
				return nil, s.unavailable(ctx.Err())
			}
			snap = s.snap.Load()
		}
	}
	if snap == nil {
		return nil, s.unavailable(nil)
	}
	if err := s.stale(snap); err != nil {
		return nil, err
	}
	return snap, nil
}

// unavailable is what a read gets when no snapshot has ever loaded: the latest
// refresh failure, so the caller is told why rather than that the catalog
// happens to be empty.
func (s *DefinitionsCatalogService) unavailable(waitErr error) error {
	last := s.latestErr()
	switch {
	case last != nil && waitErr != nil:
		return fmt.Errorf("definitions catalog has not loaded: %w (while waiting for a refresh: %w)", last, waitErr)
	case last != nil:
		return fmt.Errorf("definitions catalog has not loaded: %w", last)
	case waitErr != nil:
		return fmt.Errorf("definitions catalog has not loaded yet: %w", waitErr)
	}
	return errors.New("definitions catalog has not loaded yet")
}

// stale refuses a snapshot older than DEFINITIONS_MAX_STALENESS. Without a
// bound, a warm replica serves arbitrarily old data with no signal while a
// restarted replica fails every query, so the same request answers differently
// depending on which replica takes it.
func (s *DefinitionsCatalogService) stale(snap *catalogSnapshot) error {
	if s.maxStaleness <= 0 {
		return nil
	}
	age := time.Since(snap.lastSuccess)
	if age <= s.maxStaleness {
		return nil
	}
	cause := s.latestErr()
	if cause == nil {
		cause = errors.New("no refresh has completed since")
	}
	return fmt.Errorf("definitions catalog last refreshed %s ago, beyond DEFINITIONS_MAX_STALENESS %s: %w",
		age.Round(time.Second), s.maxStaleness, cause)
}

func (s *DefinitionsCatalogService) latestErr() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastErr
}

// run refreshes on a fixed interval, and retries a failure with exponential
// backoff and jitter instead of waiting out a whole interval.
func (s *DefinitionsCatalogService) run() {
	backoff := s.initialBackoff
	for {
		wait := s.refreshInterval
		if err := s.refreshOnce(); err != nil {
			wait = jitter(backoff)
			backoff = min(2*backoff, s.refreshInterval)
		} else {
			backoff = s.initialBackoff
		}
		timer := time.NewTimer(wait)
		select {
		case <-s.loopCtx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

// jitter spreads a retry over the second half of its backoff window, so
// replicas that failed together do not retry together.
func jitter(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	return d/2 + rand.N(d/2+1)
}

// refreshOnce runs one refresh and records its outcome. It is the only place a
// failure is recorded, so the error a caller receives, the error the log
// carries and the error the next reader is answered with are the same one.
func (s *DefinitionsCatalogService) refreshOnce() error {
	ctx, cancel := context.WithTimeout(s.loopCtx, s.refreshTimeout)
	defer cancel()

	source, err := s.refresh(ctx)
	if err != nil && s.loopCtx.Err() != nil {
		// Shutting down cancelled the refresh. That is not a catalog failure,
		// but readers waiting on this attempt still have to be released.
		s.wake()
		return err
	}
	s.recordAttempt(source, err)
	return err
}

func (s *DefinitionsCatalogService) recordAttempt(source string, err error) {
	s.mu.Lock()
	prior := s.failures
	if err != nil {
		s.lastErr = err
		s.failures++
	} else {
		s.lastErr = nil
		s.failures = 0
	}
	failures := s.failures
	s.mu.Unlock()
	s.wake()

	switch {
	case err != nil:
		catalogRefreshFailures.WithLabelValues(source).Inc()
		level := zerolog.WarnLevel
		if failures >= catalogFailuresBeforeError {
			level = zerolog.ErrorLevel
		}
		s.log.WithLevel(level).Err(err).Int("consecutiveFailures", failures).Str("source", source).
			Msg("definitions catalog refresh failed")
	case prior > 0:
		s.log.Info().Int("failures", prior).Str("source", source).
			Msg("definitions catalog refresh recovered")
	}
	s.observe()
}

// wake releases the readers waiting on the attempt that just ended and arms
// the next one.
func (s *DefinitionsCatalogService) wake() {
	s.mu.Lock()
	close(s.attemptDone)
	s.attemptDone = make(chan struct{})
	s.mu.Unlock()
}

// observe publishes the size and age of the snapshot being served. It runs
// after every attempt, so the age gauge is as coarse as the refresh cadence,
// which is enough to alert on a catalog that stopped updating hours ago.
func (s *DefinitionsCatalogService) observe() {
	snap := s.snap.Load()
	if snap == nil {
		return
	}
	if s.metricSource != "" && s.metricSource != snap.source {
		catalogDefinitions.DeleteLabelValues(s.metricSource)
		catalogSnapshotAge.DeleteLabelValues(s.metricSource)
	}
	s.metricSource = snap.source
	catalogDefinitions.WithLabelValues(snap.source).Set(float64(snap.size()))
	catalogSnapshotAge.WithLabelValues(snap.source).Set(time.Since(snap.lastSuccess).Seconds())
}

// refresh reads the catalog once and adopts what it finds, returning the source
// it read and the failure, if any.
//
// The index is asked for first. A 200 means this deployment publishes template
// builds; a 404 means the worker still serves only the flat manifest, which is
// read instead. Any other answer is a failure: an origin error must never be
// read as "this deployment has no index" and flip a pod to the other source.
func (s *DefinitionsCatalogService) refresh(ctx context.Context) (string, error) {
	held := s.snap.Load()
	resp, err := s.get(ctx, s.baseURL+catalogIndexPath, held.etagFor(sourceTemplate))
	if err != nil {
		return sourceTemplate, err
	}
	defer drain(resp.Body)

	switch resp.StatusCode {
	case http.StatusNotFound:
		return sourceLegacy, s.refreshLegacy(ctx, held)
	case http.StatusNotModified:
		if held == nil || held.source != sourceTemplate {
			return sourceTemplate, errors.New("definitions index answered 304 for a build that is not held")
		}
		s.confirm(held, resp.Header.Get("Etag"))
		return sourceTemplate, nil
	case http.StatusOK:
		return sourceTemplate, s.refreshTemplates(ctx, held, resp)
	default:
		return sourceTemplate, fmt.Errorf("definitions index returned %d", resp.StatusCode)
	}
}

// refreshTemplates adopts the build the index names. A build id that has not
// changed means the catalog has not changed: shards are immutable, so there is
// nothing to fetch.
func (s *DefinitionsCatalogService) refreshTemplates(ctx context.Context, held *catalogSnapshot, resp *http.Response) error {
	var m buildManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return fmt.Errorf("failed to decode the definitions index: %w", err)
	}
	if m.Build == "" {
		return errors.New("the definitions index names no build")
	}
	etag := resp.Header.Get("Etag")
	if held != nil && held.source == sourceTemplate && held.build == m.Build {
		s.confirm(held, etag)
		return nil
	}

	c, err := s.loadBuild(ctx, &m)
	if err != nil {
		return err
	}
	return s.adopt(ctx, held, c, sourceTemplate, m.Build, etag)
}

// loadBuild fetches every shard of a build and flattens the templates in it.
func (s *DefinitionsCatalogService) loadBuild(ctx context.Context, m *buildManifest) (*catalogCandidate, error) {
	if len(m.Shards) == 0 {
		return nil, fmt.Errorf("build %s lists no shards", m.Build)
	}
	// The index is data from the network: a key is only ever read as this
	// build's own shard, never as a URL of the producer's choosing.
	for _, key := range m.Shards {
		match := shardKeyRE.FindStringSubmatch(key)
		if match == nil || match[1] != m.Build {
			return nil, fmt.Errorf("build %s lists %q, which is not one of its shard keys", m.Build, key)
		}
	}

	parts := make([]*catalogCandidate, len(m.Shards))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(catalogShardConcurrency)
	for i, key := range m.Shards {
		g.Go(func() error {
			part, err := s.fetchShard(gctx, key)
			if err != nil {
				return err
			}
			parts[i] = part
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// A build is vetted whole: a shard that lost half its templates is a
	// catalog that lost them, not a shard to serve around.
	c := &catalogCandidate{}
	for _, part := range parts {
		c.merge(part)
	}
	return c, nil
}

func (s *DefinitionsCatalogService) fetchShard(ctx context.Context, key string) (*catalogCandidate, error) {
	resp, err := s.get(ctx, s.baseURL+"/"+key, "")
	if err != nil {
		return nil, err
	}
	defer drain(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("definitions catalog returned %d for shard %s", resp.StatusCode, key)
	}

	c := &catalogCandidate{where: key + " "}
	if err := decodeShard(resp.Body, c.visit); err != nil {
		return nil, fmt.Errorf("failed to decode shard %s: %w", key, err)
	}
	return c, nil
}

// lookupTemplate answers a by-id miss from the template document itself.
func (s *DefinitionsCatalogService) lookupTemplate(ctx context.Context, id string) (*CatalogDefinition, error) {
	lookups := s.lookups.Load()
	if d, ok := lookups.found.Get(id); ok {
		return d, nil
	}
	if expiry, ok := lookups.missing.Get(id); ok {
		if time.Now().Before(expiry) {
			return nil, nil
		}
		lookups.missing.Remove(id)
	}

	// One fetch per id however many callers arrive at once, and it outlives the
	// caller that started it, because the others are waiting on it.
	shared := s.lookupGroup.DoChan(id, func() (any, error) {
		fetchCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.lookupTimeout)
		defer cancel()
		return s.fetchTemplate(fetchCtx, lookups, id)
	})
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("looking up definition %q: %w", id, ctx.Err())
	case res := <-shared:
		if res.Err != nil {
			return nil, res.Err
		}
		d, _ := res.Val.(*CatalogDefinition)
		return d, nil
	}
}

func (s *DefinitionsCatalogService) fetchTemplate(ctx context.Context, lookups *templateLookups, id string) (*CatalogDefinition, error) {
	target := s.baseURL + templatePathPrefix + url.PathEscape(id) + ".json"
	resp, err := s.get(ctx, target, "")
	if err != nil {
		return nil, fmt.Errorf("looking up definition %q: %w", id, err)
	}
	defer drain(resp.Body)

	switch resp.StatusCode {
	case http.StatusNotFound:
		// Remembered briefly, so a client polling for a definition that does
		// not exist does not turn every query into a request.
		lookups.missing.Add(id, time.Now().Add(s.missingTTL))
		return nil, nil
	case http.StatusOK:
	default:
		return nil, fmt.Errorf("definitions catalog returned %d for template %q", resp.StatusCode, id)
	}

	d, err := decodeTemplate(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode template %q: %w", id, err)
	}
	if reason := invalidReason(&d); reason != "" {
		return nil, fmt.Errorf("template %q cannot be served: %s", id, reason)
	}
	if d.ID != id {
		return nil, fmt.Errorf("template requested as %q carries the id %q", id, d.ID)
	}
	lookups.found.Add(id, &d)
	return &d, nil
}

// refreshLegacy reads the flat manifest.json.
func (s *DefinitionsCatalogService) refreshLegacy(ctx context.Context, held *catalogSnapshot) error {
	resp, err := s.get(ctx, s.baseURL+legacyManifestPath, held.etagFor(sourceLegacy))
	if err != nil {
		return err
	}
	defer drain(resp.Body)

	switch resp.StatusCode {
	case http.StatusNotModified:
		if held == nil || held.source != sourceLegacy {
			return errors.New("definitions catalog answered 304 for a manifest that is not held")
		}
		s.confirm(held, resp.Header.Get("Etag"))
		return nil
	case http.StatusOK:
	default:
		return fmt.Errorf("definitions catalog returned %d for manifest", resp.StatusCode)
	}

	c := &catalogCandidate{}
	if err := decodeLegacyManifest(resp.Body, c.visit); err != nil {
		return fmt.Errorf("failed to decode definitions manifest: %w", err)
	}
	return s.adopt(ctx, held, c, sourceLegacy, "", resp.Header.Get("Etag"))
}

// get issues one conditional GET. A request that cannot even be built is a
// failure like any other, and reaches the caller wrapped.
func (s *DefinitionsCatalogService) get(ctx context.Context, url, etag string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build a request for %s: %w", url, err)
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	return resp, nil
}

// drain reads what is left of a body so the connection can be reused, then
// closes it.
func drain(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, io.LimitReader(body, 64<<10))
	_ = body.Close()
}

// confirm records that the held snapshot is still the current one: the catalog
// was read successfully, it had simply not changed. Its age restarts, so an
// unchanged catalog never trips the staleness bound.
func (s *DefinitionsCatalogService) confirm(held *catalogSnapshot, etag string) {
	next := *held
	next.lastSuccess = time.Now()
	if etag != "" {
		next.etag = etag
	}
	s.snap.Store(&next)
}

// adopt vets a candidate and swaps it in, or reports why it was refused.
func (s *DefinitionsCatalogService) adopt(ctx context.Context, held *catalogSnapshot, c *catalogCandidate, source, build, etag string) error {
	if c.invalid > 0 {
		s.log.Warn().Int("invalid", c.invalid).Int("kept", len(c.defs)).Strs("first", c.examples).
			Msg("skipped invalid elements in the definitions catalog")
	}
	if err := c.vet(s.minCount, held.size()); err != nil {
		return fmt.Errorf("refusing the %s definitions catalog: %w", source, err)
	}
	if err := s.checkChain(ctx, c.defs); err != nil {
		return fmt.Errorf("refusing the %s definitions catalog: %w", source, err)
	}

	next := newCatalogSnapshot(c.defs, source, build, etag)
	s.snap.Store(next)
	if held == nil || held.source != next.source || held.build != next.build {
		s.lookups.Store(newTemplateLookups())
	}

	level := zerolog.DebugLevel
	if held == nil || held.source != next.source || held.build != next.build || held.size() != next.size() {
		level = zerolog.InfoLevel
	}
	s.log.WithLevel(level).Str("source", source).Str("build", build).
		Int("definitions", next.size()).Int("invalid", c.invalid).
		Msg("adopted a definitions catalog snapshot")
	// A cold pod has nothing to compare against, so an empty catalog is
	// adopted; say so loudly, because it makes every device-definition query
	// answer successfully and emptily.
	if next.size() == 0 {
		s.log.Error().Msg("adopted an empty definitions catalog: every device-definition query will return nothing")
	}
	return nil
}

// checkChain refuses a catalog that belongs to another chain. Definitions are
// grouped by the document's manufacturer token id and looked up by the token id
// this deployment's database holds, and the catalog carries no chain marker:
// 108 of the 123 slugs dev and prod share have different token ids, so a dev
// deployment pointed at the prod catalog lists another make's definitions under
// every manufacturer, with no error anywhere. The old _<chainID>_<tableID>
// table name made that mismatch fail loudly.
func (s *DefinitionsCatalogService) checkChain(ctx context.Context, defs []CatalogDefinition) error {
	if s.manufacturerSlugs == nil {
		return nil
	}
	onChain, err := s.manufacturerSlugs(ctx)
	if err != nil {
		// The database is this check's second opinion, not the catalog's
		// source. Failing the refresh on a query error would let a database
		// blip stop the catalog from updating at all.
		s.log.Warn().Err(err).
			Msg("could not read manufacturers to check the definitions catalog against this chain; adopting it unchecked")
		return nil
	}

	shared := map[int]struct{}{}
	mismatched := map[int]string{}
	for i := range defs {
		m := defs[i].Manufacturer
		chainSlug, ok := onChain[m.TokenID]
		if !ok {
			// A manufacturer this deployment does not have says nothing about
			// which chain the catalog belongs to.
			continue
		}
		shared[m.TokenID] = struct{}{}
		if m.Slug != chainSlug {
			mismatched[m.TokenID] = fmt.Sprintf("token %d is %q here and %q in the catalog", m.TokenID, chainSlug, m.Slug)
		}
	}
	if len(mismatched) == 0 {
		return nil
	}

	tokens := slices.Sorted(maps.Keys(mismatched))
	examples := make([]string, 0, maxInvalidExamples)
	for _, token := range tokens[:min(len(tokens), maxInvalidExamples)] {
		examples = append(examples, mismatched[token])
	}
	// A renamed slug or two is drift worth logging. A large share of them is a
	// catalog that belongs to another chain.
	if limit := max(2, len(shared)*5/100); len(mismatched) > limit {
		return fmt.Errorf("%d of the %d manufacturers shared with this chain carry a different slug, above the limit of %d: %s",
			len(mismatched), len(shared), limit, strings.Join(examples, "; "))
	}
	s.log.Warn().Int("mismatched", len(mismatched)).Int("shared", len(shared)).Strs("first", examples).
		Msg("definitions catalog manufacturers disagree with this chain")
	return nil
}

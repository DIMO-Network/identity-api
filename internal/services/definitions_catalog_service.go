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
// src/build.ts) this service reads. count is deliberately not decoded: nothing
// here depends on it, and decoding it strictly would let a type change fail the
// whole index.
//
// createdAt and the shard list are part of the build's identity, not
// decoration. A build id is not a version: assembleManifest rewrites
// idx/<build>/manifest.json and idx/current.json on every publish of that id,
// so a publish repeated after more pages landed names more shards under the
// same id, with a new createdAt (definitions-worker src/idx.ts: "Not
// write-once, so not immutable"). Only the shards themselves are immutable.
type buildManifest struct {
	Build     string   `json:"build"`
	CreatedAt string   `json:"createdAt"`
	Shards    []string `json:"shards"`
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

// minDefinitionsMaxBuildAge is the smallest data-age bound that does not fire
// on a deployment working exactly as designed. The worker rebuilds when the
// live build is 23h old (definitions-worker src/scheduled.ts, REBUILD_AFTER_MS)
// and that walk takes about two hours, so the build a healthy catalog serves is
// routinely 25h old.
const minDefinitionsMaxBuildAge = 26 * time.Hour

// errNoCatalogPublished marks the one state in which neither catalog surface
// exists. It is a deploy run out of order, not an origin fault, so it is
// reported at Error from the first occurrence rather than after five.
var errNoCatalogPublished = errors.New("most likely a worker deployed with no build published yet: the cutover publishes and smokes a build (definitions-worker docs/superpowers/2026-09-14-trim-template-cutover.md, Phase 2) before the Go services deploy (Phase 3)")

// errLegacyManifestMissing is refreshLegacy's 404, which only its one caller
// sees and only ever as half of that state.
var errLegacyManifestMissing = errors.New("the flat definitions manifest does not exist")

// errCatalogWouldDowngrade is the refusal to read the flat manifest on a pod
// that is already serving a template build. A deployment deliberately rolled
// back to a worker that serves only the manifest is one pod restart away from
// reading it again, which is why the message says so: a restarted pod holds no
// build and bootstraps from the manifest as it always has.
var errCatalogWouldDowngrade = errors.New("refusing to fall back to the flat manifest, which is the pre-cutover layout: it would replace the build's data with the flat documents, turn the by-id template fallback off, and leave DEFINITIONS_MAX_BUILD_AGE with no build timestamp to measure. A missing index is a failed refresh, not a catalog to switch to; if the index is gone deliberately, restart the pods and they will read the manifest again")

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

	catalogBuildAge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "identity",
		Subsystem: "definitions_catalog",
		Name:      "build_age_seconds",
		Help:      "Age of the published build being served, from the build's own createdAt. Template mode only: no series is published while the catalog carries no build timestamp.",
	}, []string{"source"})

	catalogRefreshFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "identity",
		Subsystem: "definitions_catalog",
		Name:      "refresh_failures_total",
		Help:      "Refresh attempts that produced no usable catalog, by the source being read.",
	}, []string{"source"})
)

// catalogOrigin identifies the document a snapshot was read from, completely
// enough that two origins compare equal only when there is nothing to refetch.
type catalogOrigin struct {
	// source is which catalog layout this was read from.
	source string
	// build is the template build this was read from; empty in legacy mode.
	build string
	// createdAt is the timestamp the publish stamped on that build, verbatim;
	// empty in legacy mode. Part of the build's identity because a republished
	// build keeps its id and gets a new one.
	createdAt string
	// builtAt is createdAt parsed, and the only true measure of how old the
	// data is. Zero when the catalog carries no build timestamp: in legacy
	// mode always, and in template mode when the producer wrote one this
	// cannot read.
	builtAt time.Time
	// shards is the shard list the index named, in the order it named them:
	// assembleManifest sorts them numerically before writing, so equal lists
	// mean equal sets. A republish after more pages landed names more.
	shards []string
	// etag conditions the next read of whatever document produced it.
	etag string
}

// sameBuild reports whether the index names exactly the build already held, so
// that there is nothing to fetch. The build id alone does not settle it: only
// the shards are immutable, and a republish of the same id names a longer list
// of them under a new createdAt.
func (o catalogOrigin) sameBuild(m *buildManifest) bool {
	return o.source == sourceTemplate && o.build == m.Build &&
		o.createdAt == m.CreatedAt && slices.Equal(o.shards, m.Shards)
}

// catalogSnapshot is an immutable view of the catalog. The refresh goroutine
// builds a new one and swaps it in; a reader loads the pointer and takes no
// lock, so a refresh cannot block a query however long it runs.
type catalogSnapshot struct {
	byID       map[string]*CatalogDefinition
	byMfrToken map[int][]*CatalogDefinition
	catalogOrigin
	lastSuccess time.Time
}

func newCatalogSnapshot(defs []CatalogDefinition, origin catalogOrigin) *catalogSnapshot {
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
		byID:          byID,
		byMfrToken:    byMfr,
		catalogOrigin: origin,
		lastSuccess:   time.Now(),
	}
}

// sameOrigin reports whether a snapshot was read from the document this origin
// names. A cold pod holds no snapshot, which is never the same origin.
func (s *catalogSnapshot) sameOrigin(o catalogOrigin) bool {
	if s == nil || s.source != o.source {
		return false
	}
	return s.build == o.build && s.createdAt == o.createdAt && slices.Equal(s.shards, o.shards)
}

// buildID names the build a snapshot was read from, for a message. Empty when
// there is no snapshot, and in legacy mode.
func (s *catalogSnapshot) buildID() string {
	if s == nil {
		return ""
	}
	return s.build
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
	maxBuildAge  time.Duration

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

	// adoptedTemplate records that this pod has served a template build at
	// some point. It only ever goes from false to true: the fallback to the
	// flat manifest is a bootstrap, and a pod that has read a build must never
	// be talked back down to the pre-cutover layout by a missing index.
	adoptedTemplate atomic.Bool

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
		maxBuildAge:     cfg.MaxBuildAge,
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
	if s.maxBuildAge > 0 && s.maxBuildAge < minDefinitionsMaxBuildAge {
		stop()
		return nil, fmt.Errorf("DEFINITIONS_MAX_BUILD_AGE %s is below %s: the worker rebuilds when the live build is 23h old and the walk it then runs takes about two hours, so a healthy catalog routinely serves a build 25h old and every replica would refuse it", s.maxBuildAge, minDefinitionsMaxBuildAge)
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

// stale refuses a snapshot that is past either bound. Without them, a warm
// replica serves arbitrarily old data with no signal while a restarted replica
// fails every query, so the same request answers differently depending on which
// replica takes it.
//
// The two bounds measure different failures and neither implies the other.
func (s *DefinitionsCatalogService) stale(snap *catalogSnapshot) error {
	if err := s.unreachable(snap); err != nil {
		return err
	}
	return s.tooOld(snap)
}

// unreachable refuses a snapshot this replica has not been able to confirm for
// DEFINITIONS_MAX_STALENESS. This is a bound on reachability alone: confirm()
// restarts it on every 304 and every unchanged build, so a catalog that is up
// and serving a build nobody has replaced never trips it.
func (s *DefinitionsCatalogService) unreachable(snap *catalogSnapshot) error {
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

// tooOld refuses data older than DEFINITIONS_MAX_BUILD_AGE, measured on the
// build's own createdAt. This is the failure the reachability bound cannot see:
// the worker's scheduled rebuild stops -- a Typesense outage, a walk that can
// no longer finish inside its page budget -- while idx/current.json keeps
// answering with a build published weeks ago. Every refresh succeeds, so
// nothing else here would ever say so.
//
// Legacy mode has no such timestamp and so is bounded only by reachability: the
// flat manifest's updatedAt moves when someone writes a definition, not when
// the catalog is rebuilt, so a quiet week and a dead producer are
// indistinguishable and bounding on it would fail queries against a catalog
// nobody happened to edit.
func (s *DefinitionsCatalogService) tooOld(snap *catalogSnapshot) error {
	if s.maxBuildAge <= 0 || snap.builtAt.IsZero() {
		return nil
	}
	age := time.Since(snap.builtAt)
	if age <= s.maxBuildAge {
		return nil
	}
	err := fmt.Errorf("definitions catalog build %s was published %s ago, beyond DEFINITIONS_MAX_BUILD_AGE %s: the catalog is reachable but nothing has published a newer build",
		snap.build, age.Round(time.Second), s.maxBuildAge)
	if cause := s.latestErr(); cause != nil {
		return fmt.Errorf("%w: %w", err, cause)
	}
	return err
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
		if failures >= catalogFailuresBeforeError || errors.Is(err, errNoCatalogPublished) {
			// A catalog with no surface at all is not an origin hiccup to
			// wait out: no pod in the deployment can ever load one.
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

// observe publishes the size and the two ages of the snapshot being served. It runs
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
		catalogBuildAge.DeleteLabelValues(s.metricSource)
	}
	s.metricSource = snap.source
	catalogDefinitions.WithLabelValues(snap.source).Set(float64(snap.size()))
	catalogSnapshotAge.WithLabelValues(snap.source).Set(time.Since(snap.lastSuccess).Seconds())
	// Two different quantities, so two different series: how long it has been
	// since this replica could reach the catalog, and how old the data in it
	// is. A catalog that is perfectly reachable and three weeks old reads as
	// healthy on the first and alarming on the second.
	if snap.builtAt.IsZero() {
		catalogBuildAge.DeleteLabelValues(snap.source)
	} else {
		catalogBuildAge.WithLabelValues(snap.source).Set(time.Since(snap.builtAt).Seconds())
	}
}

// refresh reads the catalog once and adopts what it finds, returning the source
// it read and the failure, if any.
//
// The index is asked for first. A 200 means this deployment publishes template
// builds. A 404 means either that the worker still serves only the flat
// manifest or that this pod caught the index mid-publish, and the two are told
// apart by what the pod already holds: the fallback bootstraps a pod that has
// never read a build, and is refused on one that has. Any other answer is a
// failure: an origin error must never be read as "this deployment has no
// index" and flip a pod to the other source.
func (s *DefinitionsCatalogService) refresh(ctx context.Context) (string, error) {
	held := s.snap.Load()
	resp, err := s.get(ctx, s.baseURL+catalogIndexPath, held.etagFor(sourceTemplate))
	if err != nil {
		return sourceTemplate, err
	}
	defer drain(resp.Body)

	switch resp.StatusCode {
	case http.StatusNotFound:
		// A pod already serving a build keeps it. idx/current.json is served
		// with a minute of max-age and an R2 get of it can return null while
		// a publish rewrites it, so one 404 is a refresh that failed -- while
		// /manifest.json may well still answer 200 from an edge that cached
		// it before the route was deleted. Answering that with a source
		// switch would trade the build for the flat documents and silence
		// both the by-id fallback and the data-age bound, leaving a single
		// adoption line as the only sign it happened.
		if s.adoptedTemplate.Load() {
			return sourceTemplate, fmt.Errorf("%s answered 404 while this pod serves template build %q: %w",
				s.baseURL+catalogIndexPath, held.buildID(), errCatalogWouldDowngrade)
		}
		// No index and no build ever read: either this deployment still
		// publishes only the flat manifest, which is what prod answers today,
		// or a worker that no longer serves one has been deployed before
		// anything was published. The fallback tells the two apart, and it is
		// the real source until the cutover, so it stays.
		err := s.refreshLegacy(ctx, held)
		if errors.Is(err, errLegacyManifestMissing) {
			return sourceLegacy, fmt.Errorf("the definitions catalog has nothing to read: %s and %s both answered 404 -- %w",
				s.baseURL+catalogIndexPath, s.baseURL+legacyManifestPath, errNoCatalogPublished)
		}
		return sourceLegacy, err
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

// refreshTemplates adopts the build the index names. An index that names the
// same build, published at the same moment, with the same shards, means the
// catalog has not changed: those shards are immutable, so there is nothing to
// fetch. Anything else is a publish this pod has not read, including a
// republish of the build it already holds.
func (s *DefinitionsCatalogService) refreshTemplates(ctx context.Context, held *catalogSnapshot, resp *http.Response) error {
	var m buildManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return fmt.Errorf("failed to decode the definitions index: %w", err)
	}
	if m.Build == "" {
		return errors.New("the definitions index names no build")
	}
	etag := resp.Header.Get("Etag")
	if held != nil && held.sameBuild(&m) {
		s.confirm(held, etag)
		return nil
	}

	c, err := s.loadBuild(ctx, &m)
	if err != nil {
		return err
	}
	return s.adopt(ctx, held, c, catalogOrigin{
		source:    sourceTemplate,
		build:     m.Build,
		createdAt: m.CreatedAt,
		builtAt:   s.buildTimestamp(&m),
		shards:    m.Shards,
		etag:      etag,
	})
}

// buildTimestamp reads the moment this build was published. The worker writes
// it with Date.prototype.toISOString (definitions-worker src/index.ts), which
// RFC 3339 covers. A value this cannot read leaves the data-age bound with
// nothing to measure -- it is not a reason to refuse a catalog whose shards
// are perfectly readable, so it is logged and the reachability bound carries
// the weight alone.
func (s *DefinitionsCatalogService) buildTimestamp(m *buildManifest) time.Time {
	if m.CreatedAt == "" {
		s.log.Warn().Str("build", m.Build).
			Msg("the definitions index names a build with no createdAt; its age cannot be bounded")
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, m.CreatedAt)
	if err != nil {
		s.log.Warn().Err(err).Str("build", m.Build).Str("createdAt", m.CreatedAt).
			Msg("the definitions index carries a createdAt this cannot read; the build's age cannot be bounded")
		return time.Time{}
	}
	return t
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
		// Bad stored data, not a bad request. A listing skips an element that
		// fails validation, so answering a by-id lookup for the same document
		// with a 500 would make one definition two different kinds of absent
		// depending on how it was asked for. Remembered like a 404 so a client
		// polling for it does not refetch it on every query, and logged so the
		// document is fixed rather than quietly vanishing.
		s.log.Warn().Str("definition", id).Str("reason", reason).
			Msg("a definitions catalog template cannot be served and is answered as missing")
		lookups.missing.Add(id, time.Now().Add(s.missingTTL))
		return nil, nil
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
	case http.StatusNotFound:
		// Reported by the caller, which knows this was reached only because
		// the index was missing too.
		return errLegacyManifestMissing
	case http.StatusOK:
	default:
		return fmt.Errorf("definitions catalog returned %d for manifest", resp.StatusCode)
	}

	c := &catalogCandidate{}
	if err := decodeLegacyManifest(resp.Body, c.visit); err != nil {
		return fmt.Errorf("failed to decode definitions manifest: %w", err)
	}
	return s.adopt(ctx, held, c, catalogOrigin{source: sourceLegacy, etag: resp.Header.Get("Etag")})
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
// was read successfully, it had simply not changed. Its reachability age
// restarts, so an unchanged catalog never trips DEFINITIONS_MAX_STALENESS. The
// build's own age is not restarted -- nothing republished it -- and remains
// bounded by DEFINITIONS_MAX_BUILD_AGE.
func (s *DefinitionsCatalogService) confirm(held *catalogSnapshot, etag string) {
	next := *held
	next.lastSuccess = time.Now()
	if etag != "" {
		next.etag = etag
	}
	s.snap.Store(&next)
}

// adopt vets a candidate and swaps it in, or reports why it was refused.
func (s *DefinitionsCatalogService) adopt(ctx context.Context, held *catalogSnapshot, c *catalogCandidate, origin catalogOrigin) error {
	source := origin.source
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

	next := newCatalogSnapshot(c.defs, origin)
	s.snap.Store(next)
	if origin.source == sourceTemplate {
		s.adoptedTemplate.Store(true)
	}
	// A republished build keeps its id while its contents change, so what the
	// by-id fallback learned about the old one -- above all that a definition
	// was missing -- has to go with it.
	sameAsHeld := held.sameOrigin(origin)
	if !sameAsHeld {
		s.lookups.Store(newTemplateLookups())
	}

	level := zerolog.DebugLevel
	if !sameAsHeld || held.size() != next.size() {
		level = zerolog.InfoLevel
	}
	s.log.WithLevel(level).Str("source", source).Str("build", origin.build).
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

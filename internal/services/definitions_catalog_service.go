package services

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/rs/zerolog"
)

// DefinitionsCatalogService keeps the R2 definitions manifest in memory,
// refreshing it periodically with an ETag-conditional fetch. It replaces the
// per-query Tableland proxy.
type DefinitionsCatalogService struct {
	log             *zerolog.Logger
	url             string
	client          *http.Client
	refreshInterval time.Duration
	// refreshTimeout bounds one manifest download. The refresh is shared by
	// every caller queued behind the lock, so it runs detached from the
	// caller's context and carries this deadline instead.
	refreshTimeout time.Duration
	minCount       int

	mu        sync.RWMutex
	etag      string
	lastFetch time.Time
	// A failed refresh rate-limits the retry through lastAttempt, never through
	// lastFetch: lastFetch drives the freshness short-circuit, so recording a
	// failure there reports the failure once and then answers emptily and
	// successfully for the rest of the interval.
	lastAttempt time.Time
	lastErr     error
	byID        map[string]*CatalogDefinition
	byMfrToken  map[int][]*CatalogDefinition
}

// NewDefinitionsCatalogService validates the DEFINITIONS_* settings and builds
// the service. A malformed catalog URL or floor fails construction, and with it
// startup, instead of starting a pod whose every device-definition query fails
// or whose floor is silently disabled.
func NewDefinitionsCatalogService(log *zerolog.Logger, settings *config.Settings) (*DefinitionsCatalogService, error) {
	cfg, err := settings.DefinitionsCatalog()
	if err != nil {
		return nil, err
	}
	return &DefinitionsCatalogService{
		log:             log,
		url:             cfg.URL,
		client:          &http.Client{Timeout: 30 * time.Second},
		refreshInterval: time.Minute,
		refreshTimeout:  30 * time.Second,
		minCount:        cfg.MinCount,
		byID:            map[string]*CatalogDefinition{},
		byMfrToken:      map[int][]*CatalogDefinition{},
	}, nil
}

// GetDefinitionByID returns the definition with the given slug id, or nil.
func (s *DefinitionsCatalogService) GetDefinitionByID(ctx context.Context, id string) (*CatalogDefinition, error) {
	if err := s.ensureFresh(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.byID[id], nil
}

// DefinitionsByManufacturer returns the manufacturer's definitions sorted by id.
func (s *DefinitionsCatalogService) DefinitionsByManufacturer(ctx context.Context, manufacturerTokenID int) ([]*CatalogDefinition, error) {
	if err := s.ensureFresh(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.byMfrToken[manufacturerTokenID], nil
}

// failedAttempt records a refresh that produced nothing usable. It rate-limits
// the retry without touching lastFetch, so callers keep seeing the error rather
// than an empty catalog reported as success. Caller holds the write lock.
func (s *DefinitionsCatalogService) failedAttempt(err error) {
	s.lastAttempt = time.Now()
	s.lastErr = err
}

func (s *DefinitionsCatalogService) ensureFresh(ctx context.Context) error {
	s.mu.RLock()
	fresh := time.Since(s.lastFetch) < s.refreshInterval
	held := len(s.byID) != 0
	backoff := !held && s.lastErr != nil && time.Since(s.lastAttempt) < s.refreshInterval
	lastErr := s.lastErr
	s.mu.RUnlock()
	if fresh {
		return nil
	}
	// Nothing loaded and the last attempt failed recently: keep reporting that
	// failure rather than re-fetching on every request -- but keep reporting it.
	if backoff {
		return lastErr
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if time.Since(s.lastFetch) < s.refreshInterval {
		return nil
	}
	if len(s.byID) == 0 && s.lastErr != nil && time.Since(s.lastAttempt) < s.refreshInterval {
		return s.lastErr
	}

	// The download serves every caller queued behind the lock, not only the
	// one that triggered it, so it must not die with that caller's request. It
	// used to: a client that gave up mid-download cancelled the shared fetch,
	// and on a cold pod the context.Canceled was recorded as a failed attempt
	// that the backoff then answered every query with for a full interval,
	// while on a warm pod the stale-serve path marked the cache fresh and
	// skipped the refresh for an interval. Detach from the caller and bound
	// the download with our own deadline instead.
	fetchCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.refreshTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(fetchCtx, http.MethodGet, s.url+"/manifest.json", nil)
	if err != nil {
		// Recorded like every other failure, so it is logged and carries context
		// instead of re-running bare on every query.
		err = fmt.Errorf("failed to build definitions manifest request: %w", err)
		s.log.Error().Err(err).Msg("definitions manifest refresh failed")
		s.failedAttempt(err)
		return err
	}
	if s.etag != "" {
		req.Header.Set("If-None-Match", s.etag)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		// Serve the stale snapshot if we have one instead of failing reads.
		if len(s.byID) != 0 {
			s.log.Warn().Err(err).Msg("definitions manifest refresh failed, serving stale catalog")
			s.lastFetch = time.Now()
			return nil
		}
		s.failedAttempt(err)
		return fmt.Errorf("failed to fetch definitions manifest: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	switch resp.StatusCode {
	case http.StatusNotModified:
		s.lastFetch = time.Now()
		return nil
	case http.StatusOK:
	default:
		// err is nil here -- client.Do succeeded -- so build the failure first;
		// recording nil left lastErr empty and the cold-start backoff off.
		statusErr := fmt.Errorf("definitions catalog returned %d for manifest", resp.StatusCode)
		if len(s.byID) != 0 {
			s.log.Warn().Int("status", resp.StatusCode).Msg("definitions manifest refresh failed, serving stale catalog")
			s.lastFetch = time.Now()
			return nil
		}
		s.failedAttempt(statusErr)
		return statusErr
	}

	c := &catalogCandidate{}
	if err := decodeLegacyManifest(resp.Body, c.visit); err != nil {
		err = fmt.Errorf("failed to decode definitions manifest: %w", err)
		// A corrupt 200 body must not take reads down when a snapshot exists.
		if len(s.byID) != 0 {
			s.log.Warn().Err(err).Msg("definitions manifest decode failed, serving stale catalog")
			s.lastFetch = time.Now()
			return nil
		}
		s.failedAttempt(err)
		return err
	}
	if c.invalid > 0 {
		s.log.Warn().Int("invalid", c.invalid).Int("kept", len(c.defs)).Strs("first", c.examples).
			Msg("skipped invalid elements in the definitions manifest")
	}
	if err := c.vet(s.minCount, len(s.byID)); err != nil {
		err = fmt.Errorf("refusing the definitions manifest: %w", err)
		if len(s.byID) != 0 {
			s.log.Error().Err(err).Int("held", len(s.byID)).Msg("serving the stale definitions catalog")
			s.lastFetch = time.Now()
			return nil
		}
		s.failedAttempt(err)
		return err
	}
	defs := c.defs
	// A cold pod has nothing to compare against, so an empty catalog is
	// adopted; say so loudly, because it makes every device-definition query
	// answer successfully and emptily.
	if len(defs) == 0 {
		s.log.Error().Msg("adopting an empty definitions catalog: every device-definition query will return nothing")
	}

	byID := make(map[string]*CatalogDefinition, len(defs))
	byMfr := map[int][]*CatalogDefinition{}
	for i := range defs {
		d := &defs[i]
		byID[d.ID] = d
		byMfr[d.Manufacturer.TokenID] = append(byMfr[d.Manufacturer.TokenID], d)
	}
	for _, list := range byMfr {
		sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	}

	s.byID = byID
	s.byMfrToken = byMfr
	s.etag = resp.Header.Get("Etag")
	s.lastFetch = time.Now()
	s.lastErr = nil
	s.log.Debug().Int("definitions", len(byID)).Msg("refreshed definitions catalog manifest")
	return nil
}

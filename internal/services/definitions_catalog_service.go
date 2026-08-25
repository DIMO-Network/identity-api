package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/rs/zerolog"
)

// CatalogManufacturer identifies the manufacturer a definition belongs to.
type CatalogManufacturer struct {
	TokenID int    `json:"tokenId"`
	Slug    string `json:"slug"`
	Name    string `json:"name"`
}

// CatalogDefinitionAttribute is a single device attribute name/value pair.
type CatalogDefinitionAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CatalogDefinitionMetadata holds the device-specific attribute list.
type CatalogDefinitionMetadata struct {
	DeviceAttributes []CatalogDefinitionAttribute `json:"device_attributes"`
}

// CatalogDefinition is one device definition document from the R2 catalog.
// JSON keys match the documents the definitions-worker writes.
type CatalogDefinition struct {
	ID           string                     `json:"id"`
	KSUID        string                     `json:"ksuid"`
	Model        string                     `json:"model"`
	Year         int                        `json:"year"`
	DeviceType   string                     `json:"devicetype"`
	ImageURI     string                     `json:"imageuri"`
	Metadata     *CatalogDefinitionMetadata `json:"metadata"`
	Manufacturer CatalogManufacturer        `json:"manufacturer"`
}

// UnmarshalJSON tolerates metadata being an empty string, which legacy
// Tableland rows carried and may survive in backfilled documents.
func (d *CatalogDefinition) UnmarshalJSON(data []byte) error {
	type alias CatalogDefinition
	aux := &struct {
		Metadata json.RawMessage `json:"metadata"`
		*alias
	}{alias: (*alias)(d)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if len(aux.Metadata) != 0 && string(aux.Metadata) != `""` && string(aux.Metadata) != "null" {
		var md CatalogDefinitionMetadata
		if err := json.Unmarshal(aux.Metadata, &md); err != nil {
			return err
		}
		d.Metadata = &md
	}
	return nil
}

type catalogManifest struct {
	UpdatedAt string `json:"updatedAt"`
	Count     int    `json:"count"`
	// Held raw so one malformed definition costs that definition rather than
	// the whole catalog: decoding the manifest as a single document lets a
	// single bad element abort everything, which freezes warm pods on a stale
	// snapshot and makes a cold pod fail every device-definition query.
	Definitions []json.RawMessage `json:"definitions"`
}

// DefinitionsCatalogService keeps the R2 definitions manifest in memory,
// refreshing it periodically with an ETag-conditional fetch. It replaces the
// per-query Tableland proxy.
type DefinitionsCatalogService struct {
	log             *zerolog.Logger
	url             string
	client          *http.Client
	refreshInterval time.Duration
	minCount        int

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

func NewDefinitionsCatalogService(log *zerolog.Logger, settings *config.Settings) *DefinitionsCatalogService {
	return &DefinitionsCatalogService{
		log: log,
		// Every neighbouring catalog setting in values.yaml carries a trailing
		// slash; "//manifest.json" does not match the worker's exact route.
		url:             strings.TrimRight(settings.DefinitionsCatalogURL, "/"),
		client:          &http.Client{Timeout: 30 * time.Second},
		refreshInterval: time.Minute,
		minCount:        settings.DefinitionsMinCount,
		byID:            map[string]*CatalogDefinition{},
		byMfrToken:      map[int][]*CatalogDefinition{},
	}
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url+"/manifest.json", nil)
	if err != nil {
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
		if len(s.byID) != 0 {
			s.log.Warn().Int("status", resp.StatusCode).Msg("definitions manifest refresh failed, serving stale catalog")
			s.lastFetch = time.Now()
			return nil
		}
		s.failedAttempt(err)
		return fmt.Errorf("definitions catalog returned %d for manifest", resp.StatusCode)
	}

	var m catalogManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		// A corrupt 200 body must not take reads down when a snapshot exists.
		if len(s.byID) != 0 {
			s.log.Warn().Err(err).Msg("definitions manifest decode failed, serving stale catalog")
			s.lastFetch = time.Now()
			return nil
		}
		s.failedAttempt(err)
		return fmt.Errorf("failed to decode definitions manifest: %w", err)
	}

	defs := make([]CatalogDefinition, 0, len(m.Definitions))
	skipped := 0
	for _, raw := range m.Definitions {
		var d CatalogDefinition
		if err := json.Unmarshal(raw, &d); err != nil {
			skipped++
			continue
		}
		defs = append(defs, d)
	}
	if skipped > 0 {
		s.log.Warn().Int("skipped", skipped).Int("kept", len(defs)).
			Msg("skipped malformed definitions in the catalog manifest")
	}

	// One operator-set floor, checked on every pod. A proportional guard cannot
	// help the pod that matters most -- a cold one has nothing to compare
	// against, and that is exactly the pod that adopts a stub manifest
	// published while the catalog is being rebuilt. Confirming a shrink across
	// refreshes does not help either: every realistic way to produce a short
	// manifest is deterministic, so it reports the same size again and gets
	// confirmed, while the one transient case is caught by the decoder first.
	if s.minCount > 0 && len(defs) < s.minCount {
		err := fmt.Errorf("manifest carries %d definitions, below the configured minimum of %d", len(defs), s.minCount)
		if len(s.byID) != 0 {
			s.log.Error().Err(err).Int("held", len(s.byID)).
				Msg("refusing a short definitions manifest, serving stale")
			s.lastFetch = time.Now()
			return nil
		}
		s.failedAttempt(err)
		return err
	}

	// Never trade a populated catalog for an empty one, floor or no floor.
	if len(defs) == 0 && len(s.byID) != 0 {
		s.log.Error().Int("skipped", skipped).Int("held", len(s.byID)).
			Msg("definitions manifest is empty, refusing to adopt it over a populated catalog")
		s.lastFetch = time.Now()
		return nil
	}
	if len(defs) == 0 && len(m.Definitions) != 0 {
		err := fmt.Errorf("every definition in the manifest failed to decode (%d elements)", len(m.Definitions))
		s.failedAttempt(err)
		return err
	}
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

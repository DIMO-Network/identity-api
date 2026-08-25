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

	mu         sync.RWMutex
	etag       string
	lastFetch  time.Time
	byID       map[string]*CatalogDefinition
	byMfrToken map[int][]*CatalogDefinition
}

func NewDefinitionsCatalogService(log *zerolog.Logger, settings *config.Settings) *DefinitionsCatalogService {
	return &DefinitionsCatalogService{
		log:             log,
		// Every neighbouring catalog setting in values.yaml carries a trailing
		// slash; "//manifest.json" does not match the worker's exact route.
		url:             strings.TrimSuffix(settings.DefinitionsCatalogURL, "/"),
		client:          &http.Client{Timeout: 30 * time.Second},
		refreshInterval: time.Minute,
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

func (s *DefinitionsCatalogService) ensureFresh(ctx context.Context) error {
	s.mu.RLock()
	fresh := time.Since(s.lastFetch) < s.refreshInterval
	s.mu.RUnlock()
	if fresh {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if time.Since(s.lastFetch) < s.refreshInterval {
		return nil
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

	// Guard on the definitions we would actually publish, not on the raw
	// elements: skipping malformed entries happens after decoding, so counting
	// raw elements would wave through a manifest whose every element failed.
	// The check is proportional because a partial manifest is internally
	// consistent -- the producer sets count to the length it wrote -- so size
	// relative to what we already hold is the only signal that it shrank.
	if held := len(s.byID); held != 0 && len(defs)*2 < held {
		s.log.Warn().
			Int("definitions", len(defs)).
			Int("skipped", skipped).
			Int("held", held).
			Msg("definitions manifest lost most of the catalog, serving stale")
		s.lastFetch = time.Now()
		return nil
	}
	if len(defs) == 0 && len(m.Definitions) != 0 {
		return fmt.Errorf("every definition in the manifest failed to decode (%d elements)", len(m.Definitions))
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
	s.log.Debug().Int("definitions", len(byID)).Msg("refreshed definitions catalog manifest")
	return nil
}

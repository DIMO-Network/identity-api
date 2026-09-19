package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// previousDefinition, previousManifest and previousDecode reproduce the decode
// this package used before streaming (4afc3e8): the whole manifest into
// []json.RawMessage, a json.Unmarshal per element, and an UnmarshalJSON that
// re-decoded the whole definition to tolerate "" metadata. They exist to hold
// the streaming decode to the same output and to benchmark against.
type previousMetadata struct {
	DeviceAttributes []CatalogDefinitionAttribute `json:"device_attributes"`
}

type previousDefinition struct {
	ID           string              `json:"id"`
	KSUID        string              `json:"ksuid"`
	Model        string              `json:"model"`
	Year         int                 `json:"year"`
	DeviceType   string              `json:"devicetype"`
	ImageURI     string              `json:"imageuri"`
	Metadata     *previousMetadata   `json:"metadata"`
	Manufacturer CatalogManufacturer `json:"manufacturer"`
}

func (d *previousDefinition) UnmarshalJSON(data []byte) error {
	type alias previousDefinition
	aux := &struct {
		Metadata json.RawMessage `json:"metadata"`
		*alias
	}{alias: (*alias)(d)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if len(aux.Metadata) != 0 && string(aux.Metadata) != `""` && string(aux.Metadata) != "null" {
		var md previousMetadata
		if err := json.Unmarshal(aux.Metadata, &md); err != nil {
			return err
		}
		d.Metadata = &md
	}
	return nil
}

type previousManifest struct {
	UpdatedAt   string            `json:"updatedAt"`
	Count       int               `json:"count"`
	Definitions []json.RawMessage `json:"definitions"`
}

func previousDecode(r io.Reader) (defs []CatalogDefinition, skipped int, err error) {
	var m previousManifest
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		return nil, 0, err
	}
	defs = make([]CatalogDefinition, 0, len(m.Definitions))
	for _, raw := range m.Definitions {
		var d previousDefinition
		if err := json.Unmarshal(raw, &d); err != nil {
			skipped++
			continue
		}
		cur := CatalogDefinition{
			ID:           d.ID,
			KSUID:        d.KSUID,
			Model:        d.Model,
			Year:         d.Year,
			DeviceType:   d.DeviceType,
			ImageURI:     d.ImageURI,
			Manufacturer: d.Manufacturer,
		}
		// The previous shape held metadata by pointer, nil for none; the
		// current one holds it by value, zero for none. Nothing reads the
		// difference between nil and a pointer to no attributes.
		if d.Metadata != nil {
			cur.Metadata = CatalogDefinitionMetadata(*d.Metadata)
		}
		defs = append(defs, cur)
	}
	return defs, skipped, nil
}

// streamingDecode collects every element the streaming decode yields, split the
// way previousDecode splits them: decoded, or skipped for a type error.
func streamingDecode(r io.Reader) (defs []CatalogDefinition, skipped int, err error) {
	err = decodeLegacyManifest(r, func(d CatalogDefinition, err error) {
		if err != nil {
			skipped++
			return
		}
		defs = append(defs, d)
	})
	return defs, skipped, err
}

const edgeCaseManifest = `{"updatedAt":"2026-08-25T12:27:00.479Z","count":9,"definitions":[
  {"id":"toyota_camry_2020","ksuid":"K1","model":"Camry","year":2020,"devicetype":"vehicle","imageuri":"https://i/1",
   "metadata":{"device_attributes":[{"name":"powertrain_type","value":"ICE"},{"name":"fuel_type","value":"gasoline"}]},
   "manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"},"createdAt":"x","updatedAt":"y"},
  {"id":"toyota_supra_2021","metadata":null,"manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}},
  {"id":"bmw_x5_2019","metadata":"","manufacturer":{"tokenId":13,"slug":"bmw","name":"BMW"}},
  {"id":"bmw_x6_2019","metadata":{},"manufacturer":{"tokenId":13,"slug":"bmw","name":"BMW"}},
  {"id":"bmw_x7_2020","metadata":{"device_attributes":[]},"manufacturer":{"tokenId":13,"slug":"bmw","name":"BMW"}},
  {"id":"toyota_broken_2020","year":"not-a-number","manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}},
  {"id":"toyota_badmeta_2020","metadata":5,"manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}},
  {"id":"toyota_badattrs_2020","metadata":{"device_attributes":"x"},"manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}},
  {"id":"citroën_c3_2019","model":"C3 ☃","manufacturer":{"tokenId":"17","slug":"citroen","name":"Citroën"}},
  null,
  {},
  {"id":"nissan_leaf_2020","unknown":{"nested":[1,2,{"x":null}]},"manufacturer":{"tokenId":7,"slug":"nissan","name":"Nissan"}}
]}`

// The streaming decode must produce exactly what the previous decode did, on
// every edge case the catalog is known to carry and on the live manifest when
// CATALOG_MANIFEST_FILE points at a copy of it.
func TestCatalogStreamingDecodeMatchesThePreviousDecode(t *testing.T) {
	inputs := map[string][]byte{"edge cases": []byte(edgeCaseManifest)}
	if path := os.Getenv("CATALOG_MANIFEST_FILE"); path != "" {
		live, err := os.ReadFile(path)
		require.NoError(t, err)
		inputs["live manifest"] = live
	}

	for name, body := range inputs {
		t.Run(name, func(t *testing.T) {
			wantDefs, wantSkipped, err := previousDecode(bytes.NewReader(body))
			require.NoError(t, err)
			gotDefs, gotSkipped, err := streamingDecode(bytes.NewReader(body))
			require.NoError(t, err)

			assert.Equal(t, wantSkipped, gotSkipped)
			require.Equal(t, len(wantDefs), len(gotDefs))
			for i := range wantDefs {
				require.Equal(t, wantDefs[i], gotDefs[i], "element %d", i)
			}
			t.Logf("%d definitions decoded identically, %d skipped by both", len(gotDefs), gotSkipped)
		})
	}

	defs, skipped, err := streamingDecode(strings.NewReader(edgeCaseManifest))
	require.NoError(t, err)
	assert.Equal(t, 4, skipped, "year, metadata and attribute type errors, and a string tokenId, each cost one element")
	assert.Len(t, defs, 8, "null and {} decode as zero definitions here; validation rejects them later")
}

// count and updatedAt are not read, so a type change in either (count sent as a
// string, updatedAt as epoch seconds) cannot fail the manifest.
func TestCatalogDecodeIgnoresCountAndUpdatedAt(t *testing.T) {
	body := `{"count":"17993","updatedAt":1756124820,"extra":{"a":[1,{"b":null}]},"definitions":[
	  {"id":"toyota_camry_2020","manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`
	c := &catalogCandidate{}
	require.NoError(t, decodeLegacyManifest(strings.NewReader(body), c.visit))
	assert.Len(t, c.defs, 1)
	assert.Zero(t, c.invalid)
}

func TestCatalogDecodeFailsOnABrokenManifest(t *testing.T) {
	good := `{"definitions":[{"id":"toyota_camry_2020","manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`
	for name, body := range map[string]string{
		"empty body":             ``,
		"truncated mid-element":  good[:40],
		"truncated after array":  good[:len(good)-1],
		"syntax error":           `{"definitions":[{"id": tru}]}`,
		"top level is an array":  `[{"id":"toyota_camry_2020"}]`,
		"definitions not array":  `{"definitions":{"id":"toyota_camry_2020"}}`,
		"definitions is numeric": `{"definitions":5}`,
	} {
		c := &catalogCandidate{}
		assert.Error(t, decodeLegacyManifest(strings.NewReader(body), c.visit), name)
	}

	c := &catalogCandidate{}
	require.NoError(t, decodeLegacyManifest(strings.NewReader(`{"definitions":null}`), c.visit))
	assert.Zero(t, c.elements, "a null definitions array is empty, and the empty rule judges it")
}

func validDef(id string) CatalogDefinition {
	return CatalogDefinition{ID: id, Manufacturer: CatalogManufacturer{TokenID: 131, Slug: "toyota", Name: "Toyota"}}
}

// null, {}, an element with no id, and a manufacturer with no positive tokenId
// or no slug all decode without a type error, so they used to be adopted as
// zero-value definitions indexed under byID[""] and byMfrToken[0]. They are
// skipped and counted; past max(10, 1% of elements) the catalog is refused.
func TestCatalogCandidateSkipsInvalidElements(t *testing.T) {
	body := `{"definitions":[
	  {"id":"toyota_camry_2020","manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}},
	  null,
	  {},
	  {"model":"No Id","manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}},
	  {"id":"toyota_notoken_2020","manufacturer":{"slug":"toyota","name":"Toyota"}},
	  {"id":"toyota_negative_2020","manufacturer":{"tokenId":-4,"slug":"toyota","name":"Toyota"}},
	  {"id":"toyota_noslug_2020","manufacturer":{"tokenId":131,"name":"Toyota"}},
	  {"id":"toyota_supra_2021","manufacturer":{"tokenId":131,"slug":"toyota","name":"Toyota"}}]}`

	c := &catalogCandidate{}
	require.NoError(t, decodeLegacyManifest(strings.NewReader(body), c.visit))
	assert.Equal(t, 8, c.elements)
	assert.Equal(t, 6, c.invalid)
	require.Len(t, c.defs, 2)
	assert.Equal(t, "toyota_camry_2020", c.defs[0].ID)
	assert.Equal(t, "toyota_supra_2021", c.defs[1].ID)
	assert.Len(t, c.examples, maxInvalidExamples)
	assert.Contains(t, c.examples[0], "element 1")
	assert.NoError(t, c.vet(0, 0), "six invalid elements are within the limit of ten")
}

func TestCatalogCandidateVet(t *testing.T) {
	candidate := func(valid, invalid int) *catalogCandidate {
		c := &catalogCandidate{}
		for i := range valid {
			c.visit(validDef(fmt.Sprintf("toyota_m%d_2020", i)), nil)
		}
		for range invalid {
			c.visit(CatalogDefinition{}, nil)
		}
		return c
	}

	cases := []struct {
		name           string
		valid, invalid int
		minCount, held int
		refuse         string
	}{
		{name: "ten invalid of twenty is at the limit", valid: 10, invalid: 10},
		{name: "eleven invalid of twenty is past it", valid: 9, invalid: 11, refuse: "11 of 20 catalog elements are invalid"},
		{name: "one percent of a large catalog", valid: 1980, invalid: 20},
		{name: "past one percent of a large catalog", valid: 1979, invalid: 21, refuse: "above the limit of 20"},
		{name: "the floor counts valid definitions", valid: 3, invalid: 5, minCount: 3},
		{name: "invalid elements do not clear the floor", valid: 2, invalid: 5, minCount: 3, refuse: "2 valid definitions, below the configured minimum of 3"},
		{name: "only invalid elements", valid: 0, invalid: 3, refuse: "none of the catalog's 3 elements"},
		{name: "only invalid elements over a held catalog", valid: 0, invalid: 3, held: 2, refuse: "none of the catalog's 3 elements"},
		{name: "empty over a held catalog", held: 2, refuse: "catalog is empty"},
		{name: "empty on a cold pod with no floor"},
		{name: "a shrink is not refused without a floor", valid: 1, held: 2},
	}
	for _, tc := range cases {
		err := candidate(tc.valid, tc.invalid).vet(tc.minCount, tc.held)
		if tc.refuse == "" {
			assert.NoError(t, err, tc.name)
			continue
		}
		assert.ErrorContains(t, err, tc.refuse, tc.name)
	}
}

// BenchmarkCatalogManifestDecode compares the previous decode with the
// streaming one on the live manifest:
//
//	CATALOG_MANIFEST_FILE=/path/to/manifest.json go test ./internal/services/ -run '^$' -bench CatalogManifestDecode -benchmem
func BenchmarkCatalogManifestDecode(b *testing.B) {
	path := os.Getenv("CATALOG_MANIFEST_FILE")
	if path == "" {
		b.Skip("CATALOG_MANIFEST_FILE not set")
	}
	body, err := os.ReadFile(path)
	require.NoError(b, err)

	b.Run("previous", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, _, err := previousDecode(bytes.NewReader(body)); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("streaming", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			c := &catalogCandidate{}
			if err := decodeLegacyManifest(bytes.NewReader(body), c.visit); err != nil {
				b.Fatal(err)
			}
		}
	})
}

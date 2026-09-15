package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
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

// CatalogDefinitionMetadata holds the device-specific attribute list. The zero
// value means the definition carries none.
type CatalogDefinitionMetadata struct {
	DeviceAttributes []CatalogDefinitionAttribute `json:"device_attributes"`
}

// UnmarshalJSON accepts null and "" as no metadata. Legacy Tableland rows
// carried "", and backfilled documents may still. The tolerance lives on this
// one field so the rest of a definition decodes in a single pass; it used to
// re-decode the whole definition to reach it.
func (m *CatalogDefinitionMetadata) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		*m = CatalogDefinitionMetadata{}
		return nil
	}
	type plain CatalogDefinitionMetadata
	return json.Unmarshal(data, (*plain)(m))
}

// CatalogDefinition is one device definition document from the R2 catalog.
// JSON keys match the documents the definitions-worker writes.
type CatalogDefinition struct {
	ID           string                    `json:"id"`
	KSUID        string                    `json:"ksuid"`
	Model        string                    `json:"model"`
	Year         int                       `json:"year"`
	DeviceType   string                    `json:"devicetype"`
	ImageURI     string                    `json:"imageuri"`
	Metadata     CatalogDefinitionMetadata `json:"metadata"`
	Manufacturer CatalogManufacturer       `json:"manufacturer"`
}

// decodeLegacyManifest streams the legacy flat manifest,
// {"updatedAt": ..., "count": ..., "definitions": [...]}, handing each element
// of definitions to visit as soon as it is decoded. The manifest is 7.7 MB;
// streaming it avoids holding every element as raw bytes and parsing each one
// twice.
//
// Only definitions is decoded. Every other member, count and updatedAt
// included, is skipped unread: nothing depends on them, and decoding them
// strictly let a type change in either fail the whole manifest.
func decodeLegacyManifest(r io.Reader, visit func(CatalogDefinition, error)) error {
	dec := json.NewDecoder(r)
	if err := expectDelim(dec, '{'); err != nil {
		return err
	}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		if key, _ := tok.(string); key != "definitions" {
			var skip json.RawMessage
			if err := dec.Decode(&skip); err != nil {
				return err
			}
			continue
		}
		if err := decodeElements(dec, visit); err != nil {
			return fmt.Errorf("definitions: %w", err)
		}
	}
	return expectDelim(dec, '}')
}

// decodeElements streams a JSON array, decoding each element into a fresh T and
// handing it to visit. A null array has no elements. An element that fails
// with a type error is handed to visit with that error and costs only itself,
// which keeps one malformed definition from taking the catalog down with it.
// Any other error, such as a syntax error or a truncated body, fails the array.
func decodeElements[T any](dec *json.Decoder, visit func(T, error)) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if tok == nil {
		return nil
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("expected an array, found %v", tok)
	}
	for dec.More() {
		var v T
		if err := dec.Decode(&v); err != nil {
			var typeErr *json.UnmarshalTypeError
			if !errors.As(err, &typeErr) {
				return err
			}
			visit(v, err)
			continue
		}
		visit(v, nil)
	}
	return expectDelim(dec, ']')
}

func expectDelim(dec *json.Decoder, want json.Delim) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if got, ok := tok.(json.Delim); !ok || got != want {
		return fmt.Errorf("expected %q, found %v", want, tok)
	}
	return nil
}

// invalidReason says why a decoded definition cannot be served, or "" when it
// can. A JSON null element decodes to the zero definition, which the id check
// catches.
func invalidReason(d *CatalogDefinition) string {
	switch {
	case d.ID == "":
		return "no id"
	case d.Manufacturer.TokenID <= 0:
		return fmt.Sprintf("manufacturer.tokenId %d is not a token id", d.Manufacturer.TokenID)
	case d.Manufacturer.Slug == "":
		return "no manufacturer.slug"
	}
	return ""
}

// maxInvalidExamples is how many invalid elements a refusal or skip names.
const maxInvalidExamples = 5

// catalogCandidate is a decoded catalog that has not been adopted yet.
type catalogCandidate struct {
	// defs holds the valid definitions only.
	defs []CatalogDefinition
	// elements counts every element seen, valid or not.
	elements int
	invalid  int
	// examples names the first few invalid elements, for the log.
	examples []string
}

// visit files one decoded element as a definition or as invalid. It has the
// signature decodeElements hands elements to.
func (c *catalogCandidate) visit(d CatalogDefinition, err error) {
	index := c.elements
	c.elements++
	reason := invalidReason(&d)
	if err != nil {
		reason = err.Error()
	}
	if reason == "" {
		c.defs = append(c.defs, d)
		return
	}
	c.invalid++
	if len(c.examples) < maxInvalidExamples {
		c.examples = append(c.examples, fmt.Sprintf("element %d (id %q): %s", index, d.ID, reason))
	}
}

// vet applies the adoption rules to the candidate and returns why it must be
// refused, or nil. held is the number of definitions in the snapshot it would
// replace.
//
// Every count is of valid definitions. Invalid elements used to be adopted as
// zero-value definitions, so {"definitions":[null,null,null]} cleared the floor
// and the empty guard and replaced a good snapshot with nothing.
func (c *catalogCandidate) vet(minCount, held int) error {
	// A few bad elements cost themselves. More than that means the producer or
	// the shape changed, and serving the survivors would silently drop the rest.
	if limit := max(10, c.elements/100); c.invalid > limit {
		return fmt.Errorf("%d of %d catalog elements are invalid, above the limit of %d; first: %s",
			c.invalid, c.elements, limit, strings.Join(c.examples, "; "))
	}
	valid := len(c.defs)
	// One operator-set floor, checked on every pod. A proportional guard cannot
	// help the pod that matters most: a cold one has nothing to compare against,
	// and that is exactly the pod that adopts a stub published while the catalog
	// is being rebuilt.
	if minCount > 0 && valid < minCount {
		return fmt.Errorf("catalog carries %d valid definitions, below the configured minimum of %d", valid, minCount)
	}
	if valid == 0 && c.elements > 0 {
		return fmt.Errorf("none of the catalog's %d elements is a valid definition; first: %s",
			c.elements, strings.Join(c.examples, "; "))
	}
	// Never trade a populated catalog for an empty one, floor or no floor.
	if valid == 0 && held > 0 {
		return fmt.Errorf("catalog is empty; refusing to replace %d held definitions", held)
	}
	return nil
}

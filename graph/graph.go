package graph

import (
	"bytes"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed schema/*.graphqls
var schemaFS embed.FS

// GetSchema returns the combined GraphQL schema from all embedded schema files.
func GetSchema() ([]byte, error) {
	entries, err := schemaFS.ReadDir("schema")
	if err != nil {
		return nil, fmt.Errorf("failed to read schema directory: %w", err)
	}

	// Sort entries to ensure consistent ordering
	var schemaFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".graphqls") {
			schemaFiles = append(schemaFiles, entry.Name())
		}
	}

	// Sort files to ensure consistent ordering (schema.graphqls first, then alphabetical)
	sort.Slice(schemaFiles, func(i, j int) bool {
		if schemaFiles[i] == "schema.graphqls" {
			return true
		}
		if schemaFiles[j] == "schema.graphqls" {
			return false
		}
		return schemaFiles[i] < schemaFiles[j]
	})

	var schemas bytes.Buffer
	for _, fileName := range schemaFiles {
		content, err := schemaFS.ReadFile("schema/" + fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema file %s: %w", fileName, err)
		}
		schemas.Write(content)
	}

	return schemas.Bytes(), nil
}

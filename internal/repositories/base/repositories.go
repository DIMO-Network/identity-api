package base

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	// MaxPageSize is the maximum page size for paginated results
	MaxPageSize = 100
)

var (
	errInvalidToken = fmt.Errorf("invalid token")

	// InternalError is a generic error message for internal errors.
	InternalError = gqlerror.Errorf("Internal error")
)

// Repository is the base repository for all repositories.
type Repository struct {
	PDB      db.Store
	Settings config.Settings
	Log      *zerolog.Logger
}

// NewRepository creates a new base repository.
func NewRepository(pdb db.Store, settings config.Settings, logger *zerolog.Logger) *Repository {
	return &Repository{
		PDB:      pdb,
		Settings: settings,
		Log:      logger,
	}
}

// CountTrue counts the number of true values in a list of booleans.
func CountTrue(ps ...bool) int {
	out := 0

	for _, p := range ps {
		if p {
			out++
		}
	}

	return out
}

type primaryKey struct {
	TokenID int `json:"primaryKeys"`
}

// EncodeGlobalTokenID encodes a global token form and ID by prefixing it with a string and encoding it to base64.
func EncodeGlobalTokenID(prefix string, id int) (string, error) {
	var buf bytes.Buffer
	e := msgpack.NewEncoder(&buf)
	e.UseArrayEncodedStructs(true)
	err := e.Encode(primaryKey{TokenID: id})
	if err != nil {
		return "", fmt.Errorf("error encoding token id: %w", err)
	}
	return encodeGlobalToken(prefix, buf.Bytes()), nil
}

// encodeGlobalToken encodes a global token by prefixing it with a string and encoding it to base64.
func encodeGlobalToken(prefix string, data []byte) string {
	return fmt.Sprintf("%s_%s", prefix, base64.StdEncoding.EncodeToString(data))
}

// DecodeGlobalTokenID decodes a global token and returns the prefix and token id.
func DecodeGlobalTokenID(token string) (string, int, error) {
	prefix, data, err := decodeGlobalToken(token)
	if err != nil {
		return "", 0, err
	}
	var pk primaryKey
	d := msgpack.NewDecoder(bytes.NewBuffer(data))
	if err := d.Decode(&pk); err != nil {
		return "", 0, fmt.Errorf("error decoding token id: %w", err)
	}
	return prefix, pk.TokenID, nil
}

// decodeGlobalToken decodes a global token by removing the prefix and decoding it from base64.
func decodeGlobalToken(token string) (string, []byte, error) {
	parts := strings.SplitN(token, "_", 2)
	if len(parts) != 2 {
		return "", nil, errInvalidToken
	}
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, errInvalidToken
	}
	return parts[0], data, nil
}

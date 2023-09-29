package helpers

import (
	"bytes"
	"encoding/base64"

	"github.com/vmihailenco/msgpack/v5"
)

type PaginationHelper[T any] struct{}

func (p PaginationHelper[T]) EncodeCursor(cursor T) (string, error) {
	var b bytes.Buffer
	e := msgpack.NewEncoder(&b)
	e.UseArrayEncodedStructs(true)

	if err := e.Encode(cursor); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func (p PaginationHelper[T]) DecodeCursor(cursor string) (*T, error) {
	bl, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}

	var res T

	if err := msgpack.Unmarshal(bl, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

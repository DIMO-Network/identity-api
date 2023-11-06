package helpers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"

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

func ValidateFirstLast(first, last *int, maxPageSize int) (int, error) {
	var limit int

	if first != nil {
		if last != nil {
			return 0, errors.New("pass `first` or `last`, but not both")
		}
		if *first < 0 {
			return 0, errors.New("the value for `first` cannot be negative")
		}
		if *first > maxPageSize {
			return 0, fmt.Errorf("the value %d for `first` exceeds the limit %d", *last, maxPageSize)
		}
		limit = *first
	} else {
		if last == nil {
			return 0, errors.New("provide `first` or `last`")
		}
		if *last < 0 {
			return 0, errors.New("the value for `last` cannot be negative")
		}
		if *last > maxPageSize {
			return 0, fmt.Errorf("the value %d for `last` exceeds the limit %d", *last, maxPageSize)
		}
		limit = *last
	}

	return limit, nil
}

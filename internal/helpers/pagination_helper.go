package helpers

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/99designs/gqlgen/graphql/errcode"
	"github.com/vektah/gqlparser/v2/gqlerror"
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

const errInvalidPagination = "INVALID_PAGINATION"

func ValidateFirstLast(first, last *int) (err *gqlerror.Error) {
	if first != nil && last != nil {
		err = &gqlerror.Error{
			Message: "Passing both `first` and `last` to paginate a connection is not supported.",
		}
		return err
	}

	for _, arg := range []*int{first, last} {
		if arg != nil && *arg <= 0 {
			err = &gqlerror.Error{
				Message: fmt.Sprintf("invalid value provided for %d. Value cannot be less than or equal to 0", arg),
			}
			errcode.Set(err, errInvalidPagination)
			return err
		}
	}
	return nil
}

package types

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/ethereum/go-ethereum/common"
)

func MarshalAddress(addr common.Address) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(addr.Hex()))
	})
}

func UnmarshalAddress(v interface{}) (common.Address, error) {
	switch v := v.(type) {
	case string:
		return common.HexToAddress(v), nil
	case byte:
		return common.BytesToAddress([]byte{v}), nil
	default:
		return common.Address{}, fmt.Errorf("%T is not a string", v)
	}
}

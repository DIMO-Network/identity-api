package types

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/ethereum/go-ethereum/common"
)

func MarshalAddress(addr []byte) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(common.BytesToAddress(addr).Hex()))
	})
}

func UnmarshalAddress(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case string:
		return common.HexToAddress(v).Bytes(), nil
	case byte:
		return []byte{v}, nil
	default:
		return nil, fmt.Errorf("%T is not a string", v)
	}
}

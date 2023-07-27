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
		io.WriteString(w, strconv.Quote(addr.Hex()))
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

func MarshalInt(x int) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(strconv.Itoa(x)))
	})
}

var UnmarshalInt = graphql.UnmarshalInt

// func UnmarshalInt(v any) (int, error) {
// 	switch v := v.(type) {
// 	case string:
// 		return strconv.Atoi(v)
// 	case int:
// 		return v, nil
// 	case int64:
// 		return int(v), nil
// 	default:
// 		return common.Address{}, fmt.Errorf("%T is not a string", v)
// 	}
// }

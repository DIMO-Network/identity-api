package types

import (
	"encoding/json"
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

func MarshalBytes(b []byte) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		out, _ := json.Marshal(b)
		_, _ = io.WriteString(w, string(out))
	})
}

func UnmarshalBytes(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case string:
		out, _ := json.Marshal(v)
		return out, nil
	case []byte:
		return v, nil
	default:
		return nil, fmt.Errorf("value must be a string")
	}
}

func MarshalInt(x int) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(strconv.Itoa(x)))
	})
}

// Is this going to work?
var UnmarshalInt = graphql.UnmarshalInt

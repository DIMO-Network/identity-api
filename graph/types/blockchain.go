package types

import (
	"encoding/hex"
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
	s, ok := v.(string)
	if !ok {
		return common.Address{}, fmt.Errorf("type %T not a string", v)
	}

	return common.HexToAddress(s), nil
}

func MarshalBytes(b []byte) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote("0x"+hex.EncodeToString(b)))
	})
}

func UnmarshalBytes(v interface{}) ([]byte, error) {
	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("type %T not a string", v)
	}

	return common.FromHex(s), nil
}

func MarshalInt(x int) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(strconv.Itoa(x)))
	})
}

// Is this going to work?
var UnmarshalInt = graphql.UnmarshalInt

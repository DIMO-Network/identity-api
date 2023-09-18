package types

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var zeroAddress common.Address

func MarshalAddress(addr common.Address) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(addr.Hex()))
	})
}

func UnmarshalAddress(v interface{}) (common.Address, error) {
	s, ok := v.(string)
	if !ok {
		return zeroAddress, fmt.Errorf("type %T not a string", v)
	}

	if !common.IsHexAddress(s) {
		return zeroAddress, errors.New("not a valid hex address")
	}

	return common.HexToAddress(s), nil
}

func MarshalBytes(b []byte) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(hexutil.Encode(b)))
	})
}

func UnmarshalBytes(v interface{}) ([]byte, error) {
	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("type %T not a string", v)
	}

	// TODO(elffjs): A bit hacky. We want the nice feedback from hexutil, but
	// we don't want to require the prefix, so we add it if we think we need it.
	if len(s) < 2 || s[0] != '0' || s[1] != 'x' && s[1] != 'X' {
		s = "0x" + s
	}

	return hexutil.Decode(s)
}

func MarshalInt(x int) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(strconv.Itoa(x)))
	})
}

// Is this going to work?
var UnmarshalInt = graphql.UnmarshalInt

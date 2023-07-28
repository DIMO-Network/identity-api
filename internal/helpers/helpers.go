package helpers

import (
	"encoding/base64"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
)

func BytesToAddr(addrB null.Bytes) *common.Address {
	var addr *common.Address
	if addrB.Valid {
		addr = (*common.Address)(*addrB.Ptr())
	}
	return addr
}

func CursorToID(cur string) (int, error) {
	b, err := base64.StdEncoding.DecodeString(cur)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(b))
}

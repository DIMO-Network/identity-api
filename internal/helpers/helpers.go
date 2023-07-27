package helpers

import (
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

package helpers

import (
	"errors"
	"math/big"
)

func ConvertTokenIDToID(tokenID *big.Int) ([]byte, error) {
	if tokenID.Sign() < 0 {
		return nil, errors.New("token id cannot be negative")
	}

	tbs := tokenID.Bytes()
	if len(tbs) > 32 {
		return nil, errors.New("token id too large")
	}

	if len(tbs) == 32 {
		// This should almost always be the case.
		return tbs, nil
	}

	tb32 := make([]byte, 32)
	copy(tb32[32-len(tbs):], tbs)

	return tb32, nil
}

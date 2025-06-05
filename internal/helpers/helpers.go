package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"math/big"

	"strconv"
)

func CursorToID(cur string) (int, error) {
	b, err := base64.StdEncoding.DecodeString(cur)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(b))
}

func IDToCursor(id int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(id)))
}

func WithSchema(tableName string) string {
	return "identity_api." + tableName
}

func GenerateDCNNode() []byte {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return b
}

// ConvertTokenIDToID converts a token ID to a 32 byte slice.
// if the token ID is less than 32 bytes, it will be padded with zeros.
// if the token ID is greater than 32 bytes, it will return an error.
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

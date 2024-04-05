package hide

import (
	"encoding/json"
	"math/big"
)

// Int64 is an alias of int64 with obfuscating/deobfuscating json marshaller
type Int64 int64

// MarshalJSON satisfies json.Marshaller and transparently obfuscates the value
// using Default prime
func (i *Int64) MarshalJSON() ([]byte, error) {
	return json.Marshal(Int64Obfuscate(int64(*i), nil, nil))
}

// UnmarshalJSON satisfies json.Marshaller and transparently deobfuscates the
// value using inverse of Default prime
func (i *Int64) UnmarshalJSON(data []byte) error {
	var obf int64
	if err := json.Unmarshal(data, &obf); err != nil {
		*i = Int64(obf)
		return err
	}
	*i = Int64(Int64Deobfuscate(obf, nil, nil))
	return nil
}

// Int64Obfuscate obfuscates int64 provided as the 1st parameter using prime
// provided as the second one. If the provided prime is nil it will fall back
// to Default prime
func Int64Obfuscate(val int64, prime, xor *big.Int) int64 {
	if prime == nil {
		prime = Default.int64prime
	}
	bg := new(big.Int).SetInt64(val)
	modularMultiplicativeInverse(bg, prime, int64Max)

	if xor == nil {
		xor = Default.int64xor
	}
	if xor != nil {
		bg.Xor(bg, xor)
	}

	return bg.Int64()
}

// Int64Deobfuscate deobfuscates int64 provided as the 1st parameter using
// inverse provided as the second one. If the provided inverse is nil it will
// fall back to Default inverse
func Int64Deobfuscate(val int64, inverse, xor *big.Int) int64 {
	if inverse == nil {
		inverse = Default.int64inverse
	}
	bg := new(big.Int).SetInt64(val)

	if xor == nil {
		xor = Default.int64xor
	}
	if xor != nil {
		bg.Xor(bg, xor)
	}

	modularMultiplicativeInverse(bg, inverse, int64Max)

	return bg.Int64()
}

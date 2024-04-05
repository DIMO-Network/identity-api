package hide

import (
	"encoding/json"
	"math/big"
)

// Int32 is an alias of int32 with obfuscating/deobfuscating json marshaller
type Int32 int32

// MarshalJSON satisfies json.Marshaller and transparently obfuscates the value
// using Default prime
func (i *Int32) MarshalJSON() ([]byte, error) {
	return json.Marshal(Int32Obfuscate(int32(*i), nil, nil))
}

// UnmarshalJSON satisfies json.Marshaller and transparently deobfuscates the
// value using inverse of Default prime
func (i *Int32) UnmarshalJSON(data []byte) error {
	var obf int32
	if err := json.Unmarshal(data, &obf); err != nil {
		*i = Int32(obf)
		return err
	}
	*i = Int32(Int32Deobfuscate(obf, nil, nil))
	return nil
}

// Int32Obfuscate obfuscates int32 provided as the 1st parameter using prime
// provided as the second one. If the provided prime is nil it will fall back
// to Default prime
func Int32Obfuscate(val int32, prime, xor *big.Int) int32 {
	if prime == nil {
		prime = Default.int32prime
	}
	bg := new(big.Int).SetInt64(int64(val))
	modularMultiplicativeInverse(bg, prime, int32Max)

	if xor == nil {
		xor = Default.int32xor
	}
	if xor != nil {
		bg.Xor(bg, xor)
	}

	return int32(bg.Int64())
}

// Int32Deobfuscate deobfuscates int32 provided as the 1st parameter using
// inverse provided as the second one. If the provided inverse is nil it will
// fall back to Default inverse
func Int32Deobfuscate(val int32, inverse, xor *big.Int) int32 {
	if inverse == nil {
		inverse = Default.int32inverse
	}
	bg := new(big.Int).SetInt64(int64(val))

	if xor == nil {
		xor = Default.int32xor
	}
	if xor != nil {
		bg.Xor(bg, xor)
	}

	modularMultiplicativeInverse(bg, inverse, int32Max)

	return int32(bg.Int64())
}

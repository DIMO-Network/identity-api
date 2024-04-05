package hide

import (
	"encoding/json"
	"math/big"
)

// Uint64 is an alias of uint64 with obfuscating/deobfuscating json marshaller
type Uint64 uint64

// MarshalJSON satisfies json.Marshaller and transparently obfuscates the value
// using Default prime
func (i *Uint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(Uint64Obfuscate(uint64(*i), nil, nil))
}

// UnmarshalJSON satisfies json.Marshaller and transparently deobfuscates the
// value using inverse of Default prime
func (i *Uint64) UnmarshalJSON(data []byte) error {
	var obf uint64
	if err := json.Unmarshal(data, &obf); err != nil {
		*i = Uint64(obf)
		return err
	}
	*i = Uint64(Uint64Deobfuscate(obf, nil, nil))
	return nil
}

// Uint64Obfuscate obfuscates uint64 provided as the 1st parameter using prime
// provided as the second one. If the provided prime is nil it will fall back
// to Default prime
func Uint64Obfuscate(val uint64, prime, xor *big.Int) uint64 {
	if prime == nil {
		prime = Default.uint64prime
	}
	bg := new(big.Int).SetUint64(val)
	modularMultiplicativeInverse(bg, prime, uint64Max)

	if xor == nil {
		xor = Default.uint64xor
	}
	if xor != nil {
		bg.Xor(bg, xor)
	}

	return bg.Uint64()
}

// Uint64Deobfuscate deobfuscates uint64 provided as the 1st parameter using
// inverse provided as the second one. If the provided inverse is nil it will
// fall back to Default inverse
func Uint64Deobfuscate(val uint64, inverse, xor *big.Int) uint64 {
	if inverse == nil {
		inverse = Default.uint64inverse
	}
	bg := new(big.Int).SetUint64(val)

	if xor == nil {
		xor = Default.uint64xor
	}
	if xor != nil {
		bg.Xor(bg, xor)
	}

	modularMultiplicativeInverse(bg, inverse, uint64Max)

	return bg.Uint64()
}

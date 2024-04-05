package hide

import (
	"encoding/json"
	"math/big"
)

// Uint32 is an alias of uint32 with obfuscating/deobfuscating json marshaller
type Uint32 uint32

// MarshalJSON satisfies json.Marshaller and transparently obfuscates the value
// using Default prime
func (i *Uint32) MarshalJSON() ([]byte, error) {
	return json.Marshal(Uint32Obfuscate(uint32(*i), nil, nil))
}

// UnmarshalJSON satisfies json.Marshaller and transparently deobfuscates the
// value using inverse of Default prime
func (i *Uint32) UnmarshalJSON(data []byte) error {
	var obf uint32
	if err := json.Unmarshal(data, &obf); err != nil {
		*i = Uint32(obf)
		return err
	}
	*i = Uint32(Uint32Deobfuscate(obf, nil, nil))
	return nil
}

// Uint32Obfuscate obfuscates uint32 provided as the 1st parameter using prime
// provided as the second one. If the provided prime is nil it will fall back
// to Default prime
func Uint32Obfuscate(val uint32, prime, xor *big.Int) uint32 {
	if prime == nil {
		prime = Default.uint32prime
	}
	bg := new(big.Int).SetUint64(uint64(val))
	modularMultiplicativeInverse(bg, prime, uint32Max)

	if xor == nil {
		xor = Default.uint32xor
	}
	if xor != nil {
		bg.Xor(bg, xor)
	}

	return uint32(bg.Uint64())
}

// Uint32Deobfuscate deobfuscates uint32 provided as the 1st parameter using
// inverse provided as the second one. If the provided inverse is nil it will
// fall back to Default inverse
func Uint32Deobfuscate(val uint32, inverse, xor *big.Int) uint32 {
	if inverse == nil {
		inverse = Default.uint32inverse
	}
	bg := new(big.Int).SetUint64(uint64(val))

	if xor == nil {
		xor = Default.uint32xor
	}
	if xor != nil {
		bg.Xor(bg, xor)
	}

	modularMultiplicativeInverse(bg, inverse, uint32Max)

	return uint32(bg.Uint64())
}

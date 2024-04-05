package mnemonic

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
)

const (
	bitsPerByte  = 8
	bitsForBip39 = 32
	wordMaskSize = 11
	hashSize     = 256
)

var (
	// last11BitsMask is used to get the last 11 bits of a big int.
	last11BitsMask = big.NewInt(0b11111111111)

	// ErrInvalidBitSize is returned when the number of bits is not a multiple of 32.
	ErrInvalidBitSize = errors.New("invalid bit size")

	// ErrInvalidWord is returned when the word is not in the word list.
	ErrInvalidWord = errors.New("invalid words")

	// ErrInvalidChecksum is returned when the checksum is invalid.
	ErrInvalidChecksum = errors.New("invalid checksum")
)

// FromBigIntFixed converts a big int to a list of mnemonic words.
// The size is the number of bits to be used and must be a nonzero multiple of 32.
// If the provided data is smaller than the desired bits, it will be padded with 0s.
// If the provided data is larger than the desired bits the rightmost bits will be truncated.
func FromBigIntFixed(data *big.Int, sizeBits int) ([]string, error) {
	if sizeBits%bitsForBip39 != 0 || sizeBits == 0 {
		return nil, fmt.Errorf("%w: %d is not a nonzero multiple of 32", ErrInvalidBitSize, sizeBits)
	}

	if data.BitLen() > sizeBits {
		// truncate the data to the desired size
		data.Rsh(data, uint(data.BitLen()-sizeBits))
	}

	// get the number of bytes needed to hold the bits
	numOfBytes := sizeBits / bitsPerByte

	// create a buffer to hold the data + padding
	buf := make([]byte, numOfBytes)

	// convert the big int to bytes and fill the buffer
	paddedEnt := data.FillBytes(buf)

	// create a checksum of the padded data
	checksum := sha256.Sum256(paddedEnt)

	// convert the checksum to a big int
	checksumInt := new(big.Int).SetBytes(checksum[:])

	// get the number of desired bits for the checksum
	// 1 bit of the checksum is used for every 32 bits of data
	numOfCheckSumBits := sizeBits / bitsForBip39

	// get the rightmost numOfCheckSumBits bits of the checksum
	// storedhash = fullhash >> (256 - numOfCheckSumBits)
	checksumInt.Rsh(checksumInt, uint(hashSize-numOfCheckSumBits))

	// bit shift the entBin by numOfCheckSumBits to the left to make room for the checksum bits
	data.Lsh(data, uint(numOfCheckSumBits))

	// add the checksum to the end of the entBin
	// binResult = binResult | checksum
	data.Or(data, checksumInt)

	// reuse the checksum big.Int for storing the index for the nexd word
	wordIdx := checksumInt

	// get the number of bits to encode including the checksum
	bitsToEncode := sizeBits + numOfCheckSumBits

	// get the number of words needed to encode the bits each word stores 11 bits
	// bitsToEncode / 11
	numOfWords := bitsToEncode / wordMaskSize

	// allocate a list to hold the words
	words := make([]string, numOfWords)

	// populate the list of words
	for idx := numOfWords - 1; idx >= 0; idx-- {
		// use masking to get the last 11 bits of the data
		wordIdx.And(data, last11BitsMask)

		// shift the data 11 bits to the right
		data.Rsh(data, wordMaskSize)

		// get the word at the index and add it to list
		words[idx] = englishList[wordIdx.Uint64()]
	}
	return words, nil
}

// FromBigInt converts a big int to a list of mnemonic words.
// The big int is padded to the next highest number divisible by 32.
func FromBigInt(entBin *big.Int) []string {
	// get the next highest number divisible by 32
	bits := NextNumberDivisibleBy32(entBin.BitLen())

	// error is ignored because the number of bits is always a multiple of 32.
	words, _ := FromBigIntFixed(entBin, bits)
	return words
}

// NextNumberDivisibleBy32 returns the next number that is divisible by 32.
// If the number is 0, it returns 32.
func NextNumberDivisibleBy32(num int) int {
	if num == 0 {
		return bitsForBip39
	}
	return (num + 31) &^ 31
}

// ToBigInt converts a list of mnemonic words to a big int.
func ToBigInt(words []string) (*big.Int, error) {
	// create a big int to hold the result
	binResult := big.NewInt(0)

	// get the number of words
	numOfWords := len(words)

	// get the number of bits encoded by the words
	totalEncodedBits := numOfWords * wordMaskSize

	// get the number of bits to use for the data
	entSize := totalEncodedBits * bitsForBip39 / 33

	if entSize%bitsForBip39 != 0 {
		return nil, fmt.Errorf("%w: invalid word list size", ErrInvalidWord)
	}

	// loop through the words
	for _, word := range words {
		// get the index of the word
		idx, ok := englishIndex[word]
		if !ok {
			return nil, fmt.Errorf("%w %q is not in the word list", ErrInvalidWord, word)
		}

		// shift the binResult by 11 to the left
		binResult.Lsh(binResult, wordMaskSize)

		// add the index to the binResult
		binResult.Add(binResult, big.NewInt(idx))
	}

	// get the number of bits to use for the checksum
	numOfCheckSumBits := entSize / bitsForBip39

	// create mask to get the checksum from the binResult
	checkSumMask := big.NewInt(1)

	// shift the checkSumMask to the left by numOfCheckSumBits and subtract 1
	// mask = (1 << numOfCheckSumBits) - 1
	checkSumMask = checkSumMask.Lsh(checkSumMask, uint(numOfCheckSumBits)).Sub(checkSumMask, big.NewInt(1))

	// get the checksum from the binResult by masking the last numOfCheckSumBits
	// checksum = binResult & mask
	checksum := new(big.Int).And(binResult, checkSumMask)

	// get the data from the binResult
	// data = binResult >> numOfCheckSumBits
	data := binResult.Rsh(binResult, uint(numOfCheckSumBits))

	paddedBuf := make([]byte, entSize/bitsPerByte)
	// hash the data
	entBytes := data.FillBytes(paddedBuf)
	hash := sha256.Sum256(entBytes)
	hashInt := new(big.Int).SetBytes(hash[:])

	// get the rightmost bits of the hash
	hashInt.Rsh(hashInt, uint(hashSize-numOfCheckSumBits))

	// validate the checksum stored in the words against the hash of the data
	if checksum.Cmp(hashInt) != 0 {
		return nil, ErrInvalidChecksum
	}

	return binResult, nil
}

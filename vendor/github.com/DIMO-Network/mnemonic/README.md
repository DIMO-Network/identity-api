# Mnemonic

![GitHub license](https://img.shields.io/badge/license-Apache%202.0-blue.svg)
[![GoDoc](https://godoc.org/github.com/DIMO-Network/mnemonic?status.svg)](https://godoc.org/github.com/DIMO-Network/mnemonic)
[![Go Report Card](https://goreportcard.com/badge/github.com/DIMO-Network/mnemonic)](https://goreportcard.com/report/github.com/DIMO-Network/mnemonic)
## Overview

The Mnemonic package provides a flexible and customizable way to encode data into a mnemonic word list. It implements the [BIP-0039](https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki) specification using [`big.Int`](https://pkg.go.dev/math/big), allowing for the use of arbitrary entropy sizes.

## Features

- **BIP-0039 Specification:** The implementation adheres to the BIP-0039 specification, providing compatibility with various cryptographic systems.

- **Arbitrary Entropy Sizes:** This package allows users to work with arbitrary entropy sizes, providing flexibility in encoding different types of data.

- **Number Obfuscation:** For enhanced usability with IDs, the package offers optional number obfuscation. This feature uses Modular Multiplicative Inverse to convert the provided number into a seemingly random number before generating the mnemonic word list using https://github.com/c2h5oh/hide

## Getting Started

### Installation

To use this package in your Go project, run the following command:

```bash
go get github.com/DIMO-Network/mnemonic
```

### Example Usage
All Examples can be found in the [go docs](https://godoc.org/github.com/DIMO-Network/mnemonic) or [examples_test.go](./examples_test.go)

```go
package main

import (
	"fmt"
	"github.com/DIMO-Network/mnemonic"
)

func main() {
	// Example usage with a number
	id := int32(1)
	words := mnemonic.FromInt(id)
	obfuscatedWords := mnemonic.FromInt32WithObfuscation(id)
	fmt.Println(words)
	fmt.Println(obfuscatedWords)
	// Output: [abandon abandon about]
	// [learn island zoo]

	// Example usage with a word list
	obfuscatedWords := []string{"learn", "island", "zoo"}
	deobfuscatedUint, err := mnemonic.ToUint32WithDeobfuscation(obfuscatedWords)
	if err != nil {
		panic(err)
	}

	fmt.Println(deobfuscatedUint)
	// Output: 1
}
```


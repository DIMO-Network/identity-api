package model

import "math/big"

// Custom model so that we can use a non-standard Go type for the GraphQL type BigInt.
type NodeRewards struct {
	Total *big.Int `json:"total"`
}

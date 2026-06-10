package merkle

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Event names emitted by the MerkleDistributor contract.
const (
	PoolCreated    = "PoolCreated"
	RootSet        = "RootSet"
	Claimed        = "Claimed"
	Funded         = "Funded"
	Swept          = "Swept"
	WeeklyLimitSet = "WeeklyLimitSet"
)

// PoolCreatedData represents PoolCreated(uint256 indexed poolId, address indexed token, address indexed admin, uint256 weeklyLimit).
type PoolCreatedData struct {
	PoolId      *big.Int       `json:"poolId"`
	Token       common.Address `json:"token"`
	Admin       common.Address `json:"admin"`
	WeeklyLimit *big.Int       `json:"weeklyLimit"`
}

// RootSetData represents RootSet(uint256 indexed poolId, uint256 indexed week, bytes32 root, uint256 allocation, string proofsURI).
type RootSetData struct {
	PoolId     *big.Int `json:"poolId"`
	Week       *big.Int `json:"week"`
	Root       [32]byte `json:"root"`
	Allocation *big.Int `json:"allocation"`
	ProofsURI  string   `json:"proofsURI"`
}

// ClaimedData represents Claimed(uint256 indexed poolId, uint256 indexed week, address indexed account, uint256 amount).
type ClaimedData struct {
	PoolId  *big.Int       `json:"poolId"`
	Week    *big.Int       `json:"week"`
	Account common.Address `json:"account"`
	Amount  *big.Int       `json:"amount"`
}

// FundedData represents Funded(uint256 indexed poolId, address indexed from, uint256 amount).
type FundedData struct {
	PoolId *big.Int       `json:"poolId"`
	From   common.Address `json:"from"`
	Amount *big.Int       `json:"amount"`
}

// SweptData represents Swept(uint256 indexed poolId, address indexed to, uint256 amount, uint256 newBalance).
type SweptData struct {
	PoolId     *big.Int       `json:"poolId"`
	To         common.Address `json:"to"`
	Amount     *big.Int       `json:"amount"`
	NewBalance *big.Int       `json:"newBalance"`
}

// WeeklyLimitSetData represents WeeklyLimitSet(uint256 indexed poolId, uint256 limit).
type WeeklyLimitSetData struct {
	PoolId *big.Int `json:"poolId"`
	Limit  *big.Int `json:"limit"`
}

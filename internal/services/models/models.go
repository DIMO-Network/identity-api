package models

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type ContractEventData struct {
	ChainID         int64           `json:"chainId"`
	EventName       string          `json:"eventName"`
	Block           Block           `json:"block,omitempty"`
	Contract        common.Address  `json:"contract"`
	TransactionHash common.Hash     `json:"transactionHash"`
	EventSignature  common.Hash     `json:"eventSignature"`
	Arguments       json.RawMessage `json:"arguments"`
}

type Block struct {
	Number *big.Int    `json:"number,omitempty"`
	Hash   common.Hash `json:"hash,omitempty"`
	Time   time.Time   `json:"time,omitempty"`
}

package model

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Vehicle struct {
	ID         string         `json:"id"`
	Owner      common.Address `json:"owner"`
	Make       *string        `json:"make,omitempty"`
	Model      *string        `json:"model,omitempty"`
	Year       *int           `json:"year,omitempty"`
	MintedAt   time.Time      `json:"mintedAt"`
	Privileges []*Privilege   `json:"privileges,omitempty"`
}

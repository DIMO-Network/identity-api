// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Vehicle struct {
	ID       string          `json:"id"`
	Owner    *common.Address `json:"owner,omitempty"`
	Make     *string         `json:"make,omitempty"`
	Model    *string         `json:"model,omitempty"`
	Year     *int            `json:"year,omitempty"`
	MintedAt *time.Time      `json:"mintedAt,omitempty"`
}

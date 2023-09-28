// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type AftermarketDevice struct {
	ID int `json:"id"`
	// The Ethereum address for the device.
	Address common.Address `json:"address"`
	Owner   common.Address `json:"owner"`
	Serial  *string        `json:"serial,omitempty"`
	// The International Mobile Equipment Identity (IMEI) for the device.
	Imei *string `json:"imei,omitempty"`
	// The time at which this device was minted.
	MintedAt time.Time `json:"mintedAt"`
	// The vehicle, if any, with which the device is paired.
	Vehicle *Vehicle `json:"vehicle,omitempty"`
	// The beneficiary for this device, who receives any associated rewards. Defaults to the owner.
	Beneficiary common.Address `json:"beneficiary"`
	VehicleID   *int           `json:"-"`
}

type AftermarketDeviceBy struct {
	ID      *int            `json:"id,omitempty"`
	Address *common.Address `json:"address,omitempty"`
	Serial  *string         `json:"serial,omitempty"`
}

type AftermarketDeviceConnection struct {
	TotalCount int                      `json:"totalCount"`
	Edges      []*AftermarketDeviceEdge `json:"edges"`
	PageInfo   *PageInfo                `json:"pageInfo"`
}

type AftermarketDeviceEdge struct {
	Cursor string             `json:"cursor"`
	Node   *AftermarketDevice `json:"node"`
}

type AftermarketDevicesFilter struct {
	Owner *common.Address `json:"owner,omitempty"`
}

type Dcn struct {
	// The namehash of the domain.
	Node []byte `json:"node"`
	// ETH address of domain owner.
	Owner common.Address `json:"owner"`
	// The block timestamp at which the domain will cease to be valid.
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	// The block timestamp of when the domain was created.
	MintedAt time.Time `json:"mintedAt"`
	// Human readable name of the domain.
	Name *string `json:"name,omitempty"`
	// Device the domain is attached to.
	Vehicle   *Vehicle `json:"vehicle,omitempty"`
	VehicleID *int     `json:"-"`
}

type DCNBy struct {
	Node []byte  `json:"node,omitempty"`
	Name *string `json:"name,omitempty"`
}

type Definition struct {
	URI   *string `json:"uri,omitempty"`
	Make  *string `json:"make,omitempty"`
	Model *string `json:"model,omitempty"`
	Year  *int    `json:"year,omitempty"`
}

type PageInfo struct {
	StartCursor     *string `json:"startCursor,omitempty"`
	EndCursor       *string `json:"endCursor,omitempty"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	HasNextPage     bool    `json:"hasNextPage"`
}

type Privilege struct {
	ID int `json:"id"`
	// The user holding the privilege.
	User common.Address `json:"user"`
	// When this privilege was last set.
	SetAt time.Time `json:"setAt"`
	// The time at which the privilege expires.
	ExpiresAt time.Time `json:"expiresAt"`
}

type PrivilegeEdge struct {
	Node   *Privilege `json:"node"`
	Cursor string     `json:"cursor"`
}

type PrivilegesConnection struct {
	TotalCount int              `json:"totalCount"`
	Edges      []*PrivilegeEdge `json:"edges"`
	PageInfo   *PageInfo        `json:"pageInfo"`
}

type SyntheticDevice struct {
	ID            int            `json:"id"`
	IntegrationID int            `json:"integrationId"`
	Address       common.Address `json:"address"`
	MintedAt      time.Time      `json:"mintedAt"`
}

type Vehicle struct {
	ID                int                   `json:"id"`
	Owner             common.Address        `json:"owner"`
	MintedAt          time.Time             `json:"mintedAt"`
	AftermarketDevice *AftermarketDevice    `json:"aftermarketDevice,omitempty"`
	Privileges        *PrivilegesConnection `json:"privileges"`
	SyntheticDevice   *SyntheticDevice      `json:"syntheticDevice,omitempty"`
	Definition        *Definition           `json:"definition,omitempty"`
	Dcn               *Dcn                  `json:"dcn,omitempty"`
}

type VehicleConnection struct {
	TotalCount int            `json:"totalCount"`
	Edges      []*VehicleEdge `json:"edges"`
	PageInfo   *PageInfo      `json:"pageInfo"`
}

type VehicleEdge struct {
	Node   *Vehicle `json:"node"`
	Cursor string   `json:"cursor"`
}

type VehiclesFilter struct {
	Privileged *common.Address `json:"privileged,omitempty"`
}

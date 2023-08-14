// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type AftermarketDeviceConnection struct {
	TotalCount int                      `json:"totalCount"`
	Edges      []*AftermarketDeviceEdge `json:"edges"`
	PageInfo   *PageInfo                `json:"pageInfo"`
}

type AftermarketDeviceEdge struct {
	Cursor string             `json:"cursor"`
	Node   *AftermarketDevice `json:"node"`
}

type PageInfo struct {
	EndCursor   *string `json:"endCursor,omitempty"`
	HasNextPage bool    `json:"hasNextPage"`
}

type Privilege struct {
	ID        int            `json:"id"`
	User      common.Address `json:"user"`
	SetAt     time.Time      `json:"setAt"`
	ExpiresAt time.Time      `json:"expiresAt"`
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
	DeviceAddress common.Address `json:"deviceAddress"`
}

type Vehicle struct {
	ID                int                   `json:"id"`
	Owner             common.Address        `json:"owner"`
	Make              *string               `json:"make,omitempty"`
	Model             *string               `json:"model,omitempty"`
	Year              *int                  `json:"year,omitempty"`
	MintedAt          time.Time             `json:"mintedAt"`
	AftermarketDevice *AftermarketDevice    `json:"aftermarketDevice,omitempty"`
	Privileges        *PrivilegesConnection `json:"privileges"`
	SyntheticDevice   *SyntheticDevice      `json:"syntheticDevice,omitempty"`
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

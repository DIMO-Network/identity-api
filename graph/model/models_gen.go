// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Node interface {
	IsNode()
	GetID() string
}

type AftermarketDevice struct {
	// An opaque global identifier for this aftermarket device.
	ID string `json:"id"`
	// The ERC-721 token id for the device.
	TokenID int `json:"tokenId"`
	// The manufacturer of this aftermarket device.
	Manufacturer *Manufacturer `json:"manufacturer,omitempty"`
	// The Ethereum address for the device.
	Address common.Address `json:"address"`
	// The Ethereum address of the owner of the device.
	Owner common.Address `json:"owner"`
	// The serial number on the side of the device. For AutoPis this is a UUID; for Macarons it is
	// a long decimal number.
	Serial *string `json:"serial,omitempty"`
	// The International Mobile Equipment Identity (IMEI) for the device.
	Imei *string `json:"imei,omitempty"`
	// Extended Unique Identifier (EUI) for LoRa devices.
	DevEui *string `json:"devEui,omitempty"`
	// The block timestamp at which this device was minted.
	MintedAt time.Time `json:"mintedAt"`
	// The block timestamp at which this device was claimed, if it has been claimed. Devices must be
	// claimed before pairing.
	ClaimedAt *time.Time `json:"claimedAt,omitempty"`
	// The vehicle, if any, with which the device is paired.
	Vehicle *Vehicle `json:"vehicle,omitempty"`
	// The beneficiary for this device, who receives any associated rewards. Defaults to the owner.
	Beneficiary common.Address `json:"beneficiary"`
	Name        string         `json:"name"`
	// The Image Url of the device
	Image string `json:"image"`
	// The earnings attached to the aftermarket device
	Earnings       *AftermarketDeviceEarnings `json:"earnings,omitempty"`
	ManufacturerID *int                       `json:"-"`
	VehicleID      *int                       `json:"-"`
}

func (AftermarketDevice) IsNode()            {}
func (this AftermarketDevice) GetID() string { return this.ID }

type AftermarketDeviceBy struct {
	TokenID *int            `json:"tokenId,omitempty"`
	Address *common.Address `json:"address,omitempty"`
	Serial  *string         `json:"serial,omitempty"`
}

type AftermarketDeviceConnection struct {
	TotalCount int                      `json:"totalCount"`
	Edges      []*AftermarketDeviceEdge `json:"edges"`
	Nodes      []*AftermarketDevice     `json:"nodes"`
	PageInfo   *PageInfo                `json:"pageInfo"`
}

type AftermarketDeviceEarnings struct {
	TotalTokens         *big.Int            `json:"totalTokens"`
	History             *EarningsConnection `json:"history"`
	AftermarketDeviceID int                 `json:"-"`
}

type AftermarketDeviceEdge struct {
	Cursor string             `json:"cursor"`
	Node   *AftermarketDevice `json:"node"`
}

type AftermarketDevicesFilter struct {
	// Filter for aftermarket devices owned by this address.
	Owner          *common.Address `json:"owner,omitempty"`
	Serial         *string         `json:"serial,omitempty"`
	Beneficiary    *common.Address `json:"beneficiary,omitempty"`
	ManufacturerID *int            `json:"manufacturerId,omitempty"`
}

// Represents a DIMO Canonical Name. Typically these are human-readable labels for
// vehicles.
type Dcn struct {
	// The namehash of the domain.
	Node []byte `json:"node"`
	// The token id for the domain. This is simply the node reinterpreted as a uint256.
	TokenID *big.Int `json:"tokenId"`
	// Ethereum address of domain owner.
	Owner common.Address `json:"owner"`
	// The block timestamp at which the domain will cease to be valid.
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	// The block timestamp at which the domain was created.
	MintedAt time.Time `json:"mintedAt"`
	// Human readable name, if any, for the domain; for example, "reddy.dimo".
	Name *string `json:"name,omitempty"`
	// Vehicle, if any, to which the domain is attached.
	Vehicle   *Vehicle `json:"vehicle,omitempty"`
	VehicleID *int     `json:"-"`
}

type DCNBy struct {
	Node []byte  `json:"node,omitempty"`
	Name *string `json:"name,omitempty"`
}

type DCNConnection struct {
	TotalCount int        `json:"totalCount"`
	Edges      []*DCNEdge `json:"edges"`
	Nodes      []*Dcn     `json:"nodes"`
	PageInfo   *PageInfo  `json:"pageInfo"`
}

type DCNEdge struct {
	Cursor string `json:"cursor"`
	Node   *Dcn   `json:"node"`
}

type DCNFilter struct {
	// Filter for DCN owned by this address.
	Owner *common.Address `json:"owner,omitempty"`
}

type Definition struct {
	URI   *string `json:"uri,omitempty"`
	Make  *string `json:"make,omitempty"`
	Model *string `json:"model,omitempty"`
	Year  *int    `json:"year,omitempty"`
}

type Earning struct {
	// Week reward was issued
	Week int `json:"week"`
	// Address of Beneficiary that received reward
	Beneficiary common.Address `json:"beneficiary"`
	// Consecutive period of which vehicle was connected
	ConnectionStreak *int `json:"connectionStreak,omitempty"`
	// Tokens earned for connection period
	StreakTokens *big.Int `json:"streakTokens"`
	// AftermarketDevice connected to vehicle
	AftermarketDevice *AftermarketDevice `json:"aftermarketDevice,omitempty"`
	// Tokens earned by aftermarketDevice
	AftermarketDeviceTokens *big.Int `json:"aftermarketDeviceTokens"`
	// SyntheticDevice connected to vehicle
	SyntheticDevice *SyntheticDevice `json:"syntheticDevice,omitempty"`
	// Tokens earned by SyntheticDevice
	SyntheticDeviceTokens *big.Int `json:"syntheticDeviceTokens"`
	// Vehicle reward is assigned to
	Vehicle *Vehicle `json:"vehicle,omitempty"`
	// When the token was earned
	SentAt              time.Time `json:"sentAt"`
	AftermarketDeviceID *int      `json:"-"`
	SyntheticDeviceID   *int      `json:"-"`
	VehicleID           int       `json:"-"`
}

type Earnings struct {
	EarnedTokens      *big.Int            `json:"earnedTokens"`
	EarningsTransfers *EarningsConnection `json:"earningsTransfers"`
}

type EarningsConnection struct {
	TotalCount int             `json:"totalCount"`
	Edges      []*EarningsEdge `json:"edges"`
	Nodes      []*Earning      `json:"nodes"`
	PageInfo   *PageInfo       `json:"pageInfo"`
}

type EarningsEdge struct {
	Node   *Earning `json:"node"`
	Cursor string   `json:"cursor"`
}

type Manufacturer struct {
	// An opaque global identifier for this manufacturer.
	ID string `json:"id"`
	// The ERC-721 token id for the manufacturer.
	TokenID int `json:"tokenId"`
	// The name of the manufacturer.
	Name string `json:"name"`
	// The Ethereum address of the owner of this manufacturer.
	Owner common.Address `json:"owner"`
	// The block timestamp at which this manufacturer was minted.
	MintedAt time.Time `json:"mintedAt"`
	// A Relay-style connection listing any aftermarket devices associated with manufacturer.
	AftermarketDevices *AftermarketDeviceConnection `json:"aftermarketDevices"`
}

func (Manufacturer) IsNode()            {}
func (this Manufacturer) GetID() string { return this.ID }

type ManufacturerBy struct {
	Name    *string `json:"name,omitempty"`
	TokenID *int    `json:"tokenId,omitempty"`
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
	// When this privilege was last set for this user.
	SetAt time.Time `json:"setAt"`
	// The block timestamp at which the privilege expires.
	ExpiresAt time.Time `json:"expiresAt"`
}

type PrivilegeEdge struct {
	Node   *Privilege `json:"node"`
	Cursor string     `json:"cursor"`
}

type PrivilegeFilterBy struct {
	User *common.Address `json:"user,omitempty"`
}

type PrivilegesConnection struct {
	TotalCount int              `json:"totalCount"`
	Edges      []*PrivilegeEdge `json:"edges"`
	Nodes      []*Privilege     `json:"nodes"`
	PageInfo   *PageInfo        `json:"pageInfo"`
}

type SyntheticDevice struct {
	TokenID       int            `json:"tokenId"`
	IntegrationID int            `json:"integrationId"`
	Address       common.Address `json:"address"`
	MintedAt      time.Time      `json:"mintedAt"`
}

type UserRewards struct {
	TotalTokens *big.Int            `json:"totalTokens"`
	History     *EarningsConnection `json:"history"`
	User        common.Address      `json:"-"`
}

type Vehicle struct {
	// An opaque global identifier for this vehicle.
	ID string `json:"id"`
	// The ERC-721 token id for the vehicle.
	TokenID int `json:"tokenId"`
	// The manufacturer of this vehicle.
	Manufacturer *Manufacturer `json:"manufacturer,omitempty"`
	// The Ethereum address of the owner of this vehicle.
	Owner common.Address `json:"owner"`
	// The block timestamp at which this vehicle was minted.
	MintedAt time.Time `json:"mintedAt"`
	// The paired aftermarket device, if any.
	AftermarketDevice *AftermarketDevice `json:"aftermarketDevice,omitempty"`
	// A Relay-style connection listing any active privilege grants on this vehicle.
	Privileges *PrivilegesConnection `json:"privileges"`
	// The paired synthetic device, if any.
	SyntheticDevice *SyntheticDevice `json:"syntheticDevice,omitempty"`
	// The device definition for this vehicle; which includes make, model, and year among
	// other things.
	Definition *Definition `json:"definition,omitempty"`
	Dcn        *Dcn        `json:"dcn,omitempty"`
	Name       string      `json:"name"`
	// The Image Url of he vehicle
	Image          string           `json:"image"`
	Earnings       *VehicleEarnings `json:"earnings,omitempty"`
	ManufacturerID *int             `json:"-"`
}

func (Vehicle) IsNode()            {}
func (this Vehicle) GetID() string { return this.ID }

type VehicleConnection struct {
	TotalCount int            `json:"totalCount"`
	Edges      []*VehicleEdge `json:"edges"`
	Nodes      []*Vehicle     `json:"nodes"`
	PageInfo   *PageInfo      `json:"pageInfo"`
}

type VehicleEarnings struct {
	TotalTokens *big.Int            `json:"totalTokens"`
	History     *EarningsConnection `json:"history"`
	VehicleID   int                 `json:"-"`
}

type VehicleEdge struct {
	Node   *Vehicle `json:"node"`
	Cursor string   `json:"cursor"`
}

type VehiclesFilter struct {
	// Filter for vehicles to which the given address has access. This includes vehicles
	// that this address owns.
	Privileged *common.Address `json:"privileged,omitempty"`
	Owner      *common.Address `json:"owner,omitempty"`
}

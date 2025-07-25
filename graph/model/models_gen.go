// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"math/big"
	"time"

	"github.com/ericlagergren/decimal"
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
	// The DID for this aftermarket device's token ID in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID string `json:"tokenDID"`
	// The manufacturer of this aftermarket device.
	Manufacturer *Manufacturer `json:"manufacturer"`
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
	DevEui *string `json:"devEUI,omitempty"`
	// Manufacturer-specific hardware revision. May not always be present.
	HardwareRevision *string `json:"hardwareRevision,omitempty"`
	// The block timestamp at which this device was minted.
	MintedAt time.Time `json:"mintedAt"`
	// The block timestamp at which this device was claimed, if it has been claimed. Devices must be
	// claimed before pairing.
	ClaimedAt *time.Time `json:"claimedAt,omitempty"`
	// The vehicle, if any, with which the device is paired.
	Vehicle *Vehicle `json:"vehicle,omitempty"`
	// The beneficiary for this device, who receives any associated rewards. Defaults to the owner.
	Beneficiary common.Address `json:"beneficiary"`
	// Encoded name of the device
	Name string `json:"name"`
	// The Image Url of the device
	Image string `json:"image"`
	// The earnings attached to the aftermarket device
	Earnings *AftermarketDeviceEarnings `json:"earnings,omitempty"`
	// The block timestamp at which this device was paired, if it is presently paired.
	PairedAt       *time.Time `json:"pairedAt,omitempty"`
	ManufacturerID int        `json:"-"`
	VehicleID      *int       `json:"-"`
}

func (AftermarketDevice) IsNode()            {}
func (this AftermarketDevice) GetID() string { return this.ID }

// The AftermarketDeviceBy input is used to specify a unique aftermarket device to query.
type AftermarketDeviceBy struct {
	// token id of the aftermarket device NFT
	TokenID *int `json:"tokenId,omitempty"`
	// The DID of the aftermarket device in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID *string `json:"tokenDID,omitempty"`
	// address of the aftermarket device
	Address *common.Address `json:"address,omitempty"`
	// serial number of the aftermarket device
	Serial *string `json:"serial,omitempty"`
	// The International Mobile Equipment Identity (IMEI) for the device if available
	Imei *string `json:"imei,omitempty"`
	// Extended Unique Identifier (EUI) for LoRa devices if available
	DevEui *string `json:"devEUI,omitempty"`
}

// The Connection type for AftermarketDevice.
type AftermarketDeviceConnection struct {
	TotalCount int                      `json:"totalCount"`
	Edges      []*AftermarketDeviceEdge `json:"edges"`
	Nodes      []*AftermarketDevice     `json:"nodes"`
	PageInfo   *PageInfo                `json:"pageInfo"`
}

type AftermarketDeviceEarnings struct {
	TotalTokens         *decimal.Big        `json:"totalTokens"`
	History             *EarningsConnection `json:"history"`
	AftermarketDeviceID int                 `json:"-"`
}

// An edge in a AftermarketDeviceConnection.
type AftermarketDeviceEdge struct {
	Cursor string             `json:"cursor"`
	Node   *AftermarketDevice `json:"node"`
}

// The AftermarketDevicesFilter input is used to specify filtering criteria for querying aftermarket devices.
// Aftermarket devices must match all of the specified criteria.
type AftermarketDevicesFilter struct {
	// Filter for aftermarket devices owned by this address.
	Owner          *common.Address `json:"owner,omitempty"`
	Beneficiary    *common.Address `json:"beneficiary,omitempty"`
	ManufacturerID *int            `json:"manufacturerId,omitempty"`
}

type Connection struct {
	// The name of the connection. This can be at most 32 bytes long.
	Name string `json:"name"`
	// The address of the connection. This is the most commonly used identifier.
	Address common.Address `json:"address"`
	// The owner of the connection. Connections are transferable, so this may change over time.
	Owner common.Address `json:"owner"`
	// The token id of the connection as an NFT. This tends to be very large.
	TokenID *big.Int `json:"tokenId"`
	// The DID for this connection's token ID in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID string `json:"tokenDID"`
	// The block timestamp for the mint of the connection.
	MintedAt      time.Time `json:"mintedAt"`
	IntegrationID *int      `json:"-"`
}

type ConnectionBy struct {
	Name    *string         `json:"name,omitempty"`
	Address *common.Address `json:"address,omitempty"`
	TokenID *big.Int        `json:"tokenId,omitempty"`
	// The DID of the connection in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID *string `json:"tokenDID,omitempty"`
}

type ConnectionConnection struct {
	TotalCount int               `json:"totalCount"`
	Edges      []*ConnectionEdge `json:"edges"`
	Nodes      []*Connection     `json:"nodes"`
	PageInfo   *PageInfo         `json:"pageInfo"`
}

type ConnectionEdge struct {
	Node   *Connection `json:"node"`
	Cursor string      `json:"cursor"`
}

// Represents a DIMO Canonical Name. This is a unique identifier for a vehicle.
type Dcn struct {
	// An opaque global identifier for this DCN.
	ID string `json:"id"`
	// The namehash of the domain.
	Node []byte `json:"node"`
	// The token id for the domain. This is simply the node reinterpreted as a uint256.
	TokenID *big.Int `json:"tokenId"`
	// The DID for this DCN's token ID in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID string `json:"tokenDID"`
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

func (Dcn) IsNode()            {}
func (this Dcn) GetID() string { return this.ID }

// Input used to specify a unique DCN to query.
type DCNBy struct {
	Node []byte `json:"node,omitempty"`
	// The DID of the DCN in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID *string `json:"tokenDID,omitempty"`
	Name     *string `json:"name,omitempty"`
}

// The Connection type for DCN.
type DCNConnection struct {
	// The total count of DCNs in the connection.
	TotalCount int `json:"totalCount"`
	// A list of edges.
	Edges []*DCNEdge `json:"edges"`
	// A list of nodes in the connection
	Nodes []*Dcn `json:"nodes"`
	// Information to aid in pagination.
	PageInfo *PageInfo `json:"pageInfo"`
}

// An edge in a DCNConnection.
type DCNEdge struct {
	// A cursor for use in pagination.
	Cursor string `json:"cursor"`
	// The item at the end of the edge.
	Node *Dcn `json:"node"`
}

// Filter for DCN.
type DCNFilter struct {
	// Filter for DCN owned by this address.
	Owner *common.Address `json:"owner,omitempty"`
}

type Definition struct {
	ID    *string `json:"id,omitempty"`
	Make  *string `json:"make,omitempty"`
	Model *string `json:"model,omitempty"`
	Year  *int    `json:"year,omitempty"`
}

type DeveloperLicense struct {
	// The token id of the license as an NFT.
	TokenID int `json:"tokenId"`
	// The DID for this license's token ID in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID string `json:"tokenDID"`
	// The owner of the license. A single owner can own multiple licenses.
	Owner common.Address `json:"owner"`
	// Serves as the client id for OAuth as well as the address of the associated contract.
	ClientID common.Address `json:"clientId"`
	// A human-readable alias for this license. Unique among all licenses if present.
	Alias *string `json:"alias,omitempty"`
	// The block timestamp for the transaction that minted this license.
	MintedAt     time.Time              `json:"mintedAt"`
	Signers      *SignerConnection      `json:"signers"`
	RedirectURIs *RedirectURIConnection `json:"redirectURIs"`
}

type DeveloperLicenseBy struct {
	ClientID *common.Address `json:"clientId,omitempty"`
	Alias    *string         `json:"alias,omitempty"`
	TokenID  *int            `json:"tokenId,omitempty"`
	// The DID of the developer license in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID *string `json:"tokenDID,omitempty"`
}

type DeveloperLicenseConnection struct {
	TotalCount int                     `json:"totalCount"`
	Edges      []*DeveloperLicenseEdge `json:"edges"`
	Nodes      []*DeveloperLicense     `json:"nodes"`
	PageInfo   *PageInfo               `json:"pageInfo"`
}

type DeveloperLicenseEdge struct {
	Node   *DeveloperLicense `json:"node"`
	Cursor string            `json:"cursor"`
}

type DeveloperLicenseFilterBy struct {
	Signer *common.Address `json:"signer,omitempty"`
	Owner  *common.Address `json:"owner,omitempty"`
}

// Represents a Device Definition.
type DeviceDefinition struct {
	// Device definition id for this device definition.
	DeviceDefinitionID string `json:"deviceDefinitionId"`
	// Legacy id for this device definition. Newer device definitions may not have one.
	LegacyID *string `json:"legacyId,omitempty"`
	// Manufacturer for this device definition.
	Manufacturer *Manufacturer `json:"manufacturer,omitempty"`
	// Model for this device definition.
	Model string `json:"model"`
	// Year for this device definition.
	Year int `json:"year"`
	// Device type for this device definition.
	DeviceType *string `json:"deviceType,omitempty"`
	// Image URI for this device definition.
	ImageURI *string `json:"imageURI,omitempty"`
	// Device attributes for this device definition.
	Attributes []*DeviceDefinitionAttribute `json:"attributes"`
}

type DeviceDefinitionAttribute struct {
	// Name for this device definition.
	Name string `json:"name"`
	// Value for this device definition.
	Value string `json:"value"`
}

// Input used to specify a unique Device Definition to query.
type DeviceDefinitionBy struct {
	// The id for the device definition.
	ID string `json:"id"`
}

// Represents a Device Definition.
type DeviceDefinitionConnection struct {
	// The total count of Device Definitions in the connection.
	TotalCount int `json:"totalCount"`
	// A list of edges.
	Edges []*DeviceDefinitionEdge `json:"edges"`
	// A list of nodes in the connection
	Nodes []*DeviceDefinition `json:"nodes"`
	// Information to aid in pagination.
	PageInfo *PageInfo `json:"pageInfo"`
}

// An edge in a Device Definition Connection.
type DeviceDefinitionEdge struct {
	// A cursor for use in pagination.
	Cursor string `json:"cursor"`
	// The item at the end of the edge.
	Node *DeviceDefinition `json:"node"`
}

// Filter for Device Definition.
type DeviceDefinitionFilter struct {
	// Model filters for device definition that are of the given model.
	// This filter performs a case insensitive match.
	Model *string `json:"model,omitempty"`
	// Year filters for device definition that are of the given year.
	Year *int `json:"year,omitempty"`
}

type Earning struct {
	// Week reward was issued
	Week int `json:"week"`
	// Address of Beneficiary that received reward
	Beneficiary common.Address `json:"beneficiary"`
	// Consecutive period of which vehicle was connected
	ConnectionStreak *int `json:"connectionStreak,omitempty"`
	// Tokens earned for connection period
	StreakTokens *decimal.Big `json:"streakTokens"`
	// AftermarketDevice connected to vehicle
	AftermarketDevice *AftermarketDevice `json:"aftermarketDevice,omitempty"`
	// Tokens earned by aftermarketDevice
	AftermarketDeviceTokens *decimal.Big `json:"aftermarketDeviceTokens"`
	// SyntheticDevice connected to vehicle
	SyntheticDevice *SyntheticDevice `json:"syntheticDevice,omitempty"`
	// Tokens earned by SyntheticDevice
	SyntheticDeviceTokens *decimal.Big `json:"syntheticDeviceTokens"`
	// Vehicle reward is assigned to
	Vehicle *Vehicle `json:"vehicle,omitempty"`
	// When the token was earned
	SentAt              time.Time `json:"sentAt"`
	AftermarketDeviceID *int      `json:"-"`
	SyntheticDeviceID   *int      `json:"-"`
	VehicleID           int       `json:"-"`
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
	// The DID for this manufacturer's token ID in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID string `json:"tokenDID"`
	// The name of the manufacturer.
	Name string `json:"name"`
	// The Ethereum address of the owner of this manufacturer.
	Owner common.Address `json:"owner"`
	// Id of the Tableland table holding the manufacturer's device definitions.
	TableID *int `json:"tableId,omitempty"`
	// The block timestamp at which this manufacturer was minted.
	MintedAt time.Time `json:"mintedAt"`
	// A Relay-style connection listing any aftermarket devices associated with manufacturer.
	AftermarketDevices *AftermarketDeviceConnection `json:"aftermarketDevices"`
	// List device definitions under this manufacturer.
	DeviceDefinitions *DeviceDefinitionConnection `json:"deviceDefinitions"`
}

func (Manufacturer) IsNode()            {}
func (this Manufacturer) GetID() string { return this.ID }

type ManufacturerBy struct {
	Name    *string `json:"name,omitempty"`
	TokenID *int    `json:"tokenId,omitempty"`
	Slug    *string `json:"slug,omitempty"`
	// The DID of the manufacturer in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID *string `json:"tokenDID,omitempty"`
}

// The Connection type for Manufacturer.
type ManufacturerConnection struct {
	TotalCount int                 `json:"totalCount"`
	Edges      []*ManufacturerEdge `json:"edges"`
	Nodes      []*Manufacturer     `json:"nodes"`
	PageInfo   *PageInfo           `json:"pageInfo"`
}

// An edge in a ManufacturerConnection.
type ManufacturerEdge struct {
	Node   *Manufacturer `json:"node"`
	Cursor string        `json:"cursor"`
}

type PageInfo struct {
	StartCursor     *string `json:"startCursor,omitempty"`
	EndCursor       *string `json:"endCursor,omitempty"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	HasNextPage     bool    `json:"hasNextPage"`
}

type Privilege struct {
	// The id of the privilege.
	ID int `json:"id"`
	// The user holding the privilege.
	User common.Address `json:"user"`
	// The block timestamp at which this privilege was last set.
	SetAt time.Time `json:"setAt"`
	// The block timestamp at which the privilege expires.
	ExpiresAt time.Time `json:"expiresAt"`
}

type PrivilegeEdge struct {
	Node   *Privilege `json:"node"`
	Cursor string     `json:"cursor"`
}

type PrivilegeFilterBy struct {
	User        *common.Address `json:"user,omitempty"`
	PrivilegeID *int            `json:"privilegeId,omitempty"`
}

// The Connection type for Privileges.
type PrivilegesConnection struct {
	TotalCount int              `json:"totalCount"`
	Edges      []*PrivilegeEdge `json:"edges"`
	Nodes      []*Privilege     `json:"nodes"`
	PageInfo   *PageInfo        `json:"pageInfo"`
}

// The root query type for the GraphQL schema.
type Query struct {
}

type RedirectURI struct {
	URI       string    `json:"uri"`
	EnabledAt time.Time `json:"enabledAt"`
}

type RedirectURIConnection struct {
	TotalCount int                `json:"totalCount"`
	Edges      []*RedirectURIEdge `json:"edges"`
	Nodes      []*RedirectURI     `json:"nodes"`
	PageInfo   *PageInfo          `json:"pageInfo"`
}

type RedirectURIEdge struct {
	Node   *RedirectURI `json:"node"`
	Cursor string       `json:"cursor"`
}

type Sacd struct {
	// Recipient of sacd permission grant
	Grantee common.Address `json:"grantee"`
	// Hex string of permissions
	Permissions string `json:"permissions"`
	// Permission source
	Source string `json:"source"`
	// The block timestamp at which this permission was set.
	CreatedAt time.Time `json:"createdAt"`
	// The block timestamp at which the permission expires.
	ExpiresAt time.Time `json:"expiresAt"`
}

// The Connection type for Sacds.
type SacdConnection struct {
	TotalCount int         `json:"totalCount"`
	Edges      []*SacdEdge `json:"edges"`
	Nodes      []*Sacd     `json:"nodes"`
	PageInfo   *PageInfo   `json:"pageInfo"`
}

type SacdEdge struct {
	Node   *Sacd  `json:"node"`
	Cursor string `json:"cursor"`
}

type Signer struct {
	Address   common.Address `json:"address"`
	EnabledAt time.Time      `json:"enabledAt"`
}

type SignerConnection struct {
	TotalCount int           `json:"totalCount"`
	Edges      []*SignerEdge `json:"edges"`
	Nodes      []*Signer     `json:"nodes"`
	PageInfo   *PageInfo     `json:"pageInfo"`
}

type SignerEdge struct {
	Node   *Signer `json:"node"`
	Cursor string  `json:"cursor"`
}

type Stake struct {
	// The token id of the license as an NFT.
	TokenID int `json:"tokenId"`
	// The DID for this stake's token ID in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID string `json:"tokenDID"`
	// The owner of the license. A single owner can own multiple licenses.
	Owner common.Address `json:"owner"`
	// The level of the stake. Presently, the levels are 0, 1, and 2. These translate
	// to Levels 2, 3, and 4 in DIP-2. See https://docs.dimo.org/governance/improvement-proposals/dip2
	Level int `json:"level"`
	// The number of points that the stake contributes to a vehicle's weekly total. The process
	// by which points become tokens is described in DIP-2.
	Points int `json:"points"`
	// The amount of $DIMO staked.
	Amount *decimal.Big `json:"amount"`
	// The block timestamp for the transaction that created this stake.
	StakedAt time.Time `json:"stakedAt"`
	// The block timestamp after which the staked tokens will become or did become unlocked.
	EndsAt time.Time `json:"endsAt"`
	// The block timestamp at which the tokens were withdrawn, if they have been withdrawn.
	WithdrawnAt *time.Time `json:"withdrawnAt,omitempty"`
	// The vehicle to which the stake is attached, if it is attached.
	Vehicle   *Vehicle `json:"vehicle,omitempty"`
	VehicleID *int     `json:"-"`
}

type StakeConnection struct {
	TotalCount int          `json:"totalCount"`
	Edges      []*StakeEdge `json:"edges"`
	Nodes      []*Stake     `json:"nodes"`
	PageInfo   *PageInfo    `json:"pageInfo"`
}

type StakeEdge struct {
	Node   *Stake `json:"node"`
	Cursor string `json:"cursor"`
}

type StakeFilterBy struct {
	Owner *common.Address `json:"owner,omitempty"`
	// Filter stakes based on attachability. A stake is considered attachable if it
	// is not presently attached to a vehicle and has not yet ended.
	Attachable *bool `json:"attachable,omitempty"`
}

type StorageNode struct {
	// The label for the storage node. This is unique.
	Label string `json:"label"`
	// The address for the storage node. This is the location of the node's deployed contract.
	Address common.Address `json:"address"`
	// The owner of the storage node. Nodes are transferable, so this may change over time.
	Owner common.Address `json:"owner"`
	// The token id of the storage node as an NFT. Since this is uint256(keccak256(bytes(label))),
	// it tends to be very large.
	TokenID *big.Int `json:"tokenId"`
	// The URI for the node. This will host the well-known URIs that tell clients how to send in
	// and retrieve data for this vehicle.
	URI string `json:"uri"`
	// The DID for this node's NFT in the format did:erc721:<chainID>:<contractAddress>:<tokenId>.
	TokenDID string `json:"tokenDID"`
	// The timestamp of the block in which this node was minted.
	MintedAt time.Time `json:"mintedAt"`
}

// The SyntheticDevice is a software connection established to connect the vehicle to the DIMO network.
type SyntheticDevice struct {
	// An opaque global identifier for this syntheticDevice.
	ID string `json:"id"`
	// Encoded name of the device
	Name string `json:"name"`
	// The ERC-721 token id for the device.
	TokenID int `json:"tokenId"`
	// The DID for this synthetic device's token ID in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID string `json:"tokenDID"`
	// Type of integration for the synthetic device.
	IntegrationID int `json:"integrationId"`
	// The Ethereum address for the device.
	Address common.Address `json:"address"`
	// The block timestamp at which this device was minted.
	MintedAt time.Time `json:"mintedAt"`
	// The vehicle with which the synthetic device is paired.
	Vehicle *Vehicle `json:"vehicle"`
	// The parent connection for the synthetic device.
	Connection   *Connection `json:"connection"`
	ConnectionID []byte      `json:"-"`
	VehicleID    int         `json:"-"`
}

func (SyntheticDevice) IsNode()            {}
func (this SyntheticDevice) GetID() string { return this.ID }

// The SyntheticDeviceBy input is used to specify a unique synthetic device to query.
type SyntheticDeviceBy struct {
	// The token id for the synthetic device.
	TokenID *int `json:"tokenId,omitempty"`
	// The DID of the synthetic device in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID *string `json:"tokenDID,omitempty"`
	// The Ethereum address for the synthetic device.
	Address *common.Address `json:"address,omitempty"`
}

// The Connection type for SyntheticDevice.
type SyntheticDeviceConnection struct {
	// The total count of SyntheticDevices in the connection.
	TotalCount int `json:"totalCount"`
	// A list of edges.
	Edges []*SyntheticDeviceEdge `json:"edges"`
	// A list of nodes in the connection (without going through the `edges` field).
	Nodes []*SyntheticDevice `json:"nodes"`
	// Information to aid in pagination.
	PageInfo *PageInfo `json:"pageInfo"`
}

// An edge in a SytheticDeviceConnection.
type SyntheticDeviceEdge struct {
	// A cursor for use in pagination.
	Cursor string `json:"cursor"`
	// The item at the end of the edge.
	Node *SyntheticDevice `json:"node"`
}

// The SyntheticDevicesFilter input is used to specify filtering criteria for querying synthetic devices.
// Synthetic devices must match all of the specified criteria.
type SyntheticDevicesFilter struct {
	// Filter for synthetic devices owned by this address.
	Owner *common.Address `json:"owner,omitempty"`
	// Filter for synthetic devices with this integration id.
	IntegrationID *int `json:"integrationId,omitempty"`
}

type UserRewards struct {
	TotalTokens *decimal.Big        `json:"totalTokens"`
	History     *EarningsConnection `json:"history"`
	User        common.Address      `json:"-"`
}

type Vehicle struct {
	// An opaque global identifier for this vehicle.
	ID string `json:"id"`
	// The ERC-721 token id for the vehicle.
	TokenID int `json:"tokenId"`
	// The DID for this vehicle's token ID in the format did:erc721:<chainID>:<contractAddress>:<tokenId>
	TokenDID string `json:"tokenDID"`
	// The manufacturer of this vehicle.
	Manufacturer *Manufacturer `json:"manufacturer"`
	// The Ethereum address of the owner of this vehicle.
	Owner common.Address `json:"owner"`
	// The block timestamp at which this vehicle was minted.
	MintedAt time.Time `json:"mintedAt"`
	// The paired aftermarket device, if any.
	AftermarketDevice *AftermarketDevice `json:"aftermarketDevice,omitempty"`
	// A Relay-style connection listing any active privilege grants on this vehicle.
	Privileges *PrivilegesConnection `json:"privileges"`
	// A Relay-style connection listing any active SACD permission grants on this vehicle.
	Sacds *SacdConnection `json:"sacds"`
	// The paired synthetic device, if any.
	SyntheticDevice *SyntheticDevice `json:"syntheticDevice,omitempty"`
	// The device definition for this vehicle; which includes make, model, and year among
	// other things.
	Definition *Definition `json:"definition,omitempty"`
	Dcn        *Dcn        `json:"dcn,omitempty"`
	// Encoded name of the device
	Name string `json:"name"`
	// A URI containing an image for the vehicle.
	ImageURI string           `json:"imageURI"`
	Image    string           `json:"image"`
	Earnings *VehicleEarnings `json:"earnings,omitempty"`
	DataURI  string           `json:"dataURI"`
	Stake    *Stake           `json:"stake,omitempty"`
	// Description of the storage node to which the vehicle's data should be sent. If this is
	// not set, then the vehicle may be attached to the original Digital Infrastructure, Inc.
	// node.
	StorageNode    *StorageNode `json:"storageNode,omitempty"`
	ManufacturerID int          `json:"-"`
	StorageNodeID  []byte       `json:"-"`
}

func (Vehicle) IsNode()            {}
func (this Vehicle) GetID() string { return this.ID }

// The Connection type for Vehicle.
type VehicleConnection struct {
	TotalCount int            `json:"totalCount"`
	Edges      []*VehicleEdge `json:"edges"`
	Nodes      []*Vehicle     `json:"nodes"`
	PageInfo   *PageInfo      `json:"pageInfo"`
}

type VehicleEarnings struct {
	TotalTokens *decimal.Big        `json:"totalTokens"`
	History     *EarningsConnection `json:"history"`
	VehicleID   int                 `json:"-"`
}

// An edge in a VehicleConnection.
type VehicleEdge struct {
	Node   *Vehicle `json:"node"`
	Cursor string   `json:"cursor"`
}

// The VehiclesFilter input is used to specify filtering criteria for querying vehicles.
// Vehicles must match all of the specified criteria.
type VehiclesFilter struct {
	// Privileged filters for vehicles to which the given address has access. This includes vehicles
	// that this address owns.
	Privileged *common.Address `json:"privileged,omitempty"`
	// Owner filters for vehicles that this address owns.
	Owner *common.Address `json:"owner,omitempty"`
	// Make filters for vehicles that are of the given make.
	Make *string `json:"make,omitempty"`
	// Model filters for vehicles that are of the given model.
	Model *string `json:"model,omitempty"`
	// Year filters for vehicles that are of the given year.
	Year *int `json:"year,omitempty"`
	// Filter for vehicles produced by a particular manufacturer, specified by manufacturer token id.
	ManufacturerTokenID *int `json:"manufacturerTokenId,omitempty"`
	// Filter for vehicles by device definition id.
	DeviceDefinitionID *string `json:"deviceDefinitionId,omitempty"`
}

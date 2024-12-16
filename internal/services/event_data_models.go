package services

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type VehicleAttributeSetData struct {
	TokenID   *big.Int
	Attribute string
	Info      string
}

type AftermarketDeviceNodeMintedData struct {
	ManufacturerID           *big.Int
	TokenID                  *big.Int
	AftermarketDeviceAddress common.Address
	Owner                    common.Address
}

type VehicleNodeMintedData struct {
	ManufacturerNode *big.Int
	TokenID          *big.Int
	Owner            common.Address
}

type VehicleNodeMintedWithDeviceDefinitionData struct {
	ManufacturerId     *big.Int
	VehicleId          *big.Int
	Owner              common.Address
	DeviceDefinitionID string
}

type ManufacturerNodeMintedData struct {
	Name    string
	TokenID *big.Int
	Owner   common.Address
}

type AftermarketDeviceAttributeSetData struct {
	TokenID   *big.Int
	Attribute string
	Info      string
}

type AftermarketDevicePairData struct {
	AftermarketDeviceNode *big.Int
	VehicleNode           *big.Int
	Owner                 common.Address
}

type AftermarketDeviceClaimedData struct {
	AftermarketDeviceNode *big.Int
	Owner                 common.Address
}

type TransferData struct {
	From    common.Address
	To      common.Address
	TokenID *big.Int
}

type PrivilegeSetData struct {
	TokenId *big.Int
	PrivId  *big.Int
	User    common.Address
	Expires *big.Int
}

type BeneficiarySetData struct {
	IdProxyAddress common.Address
	NodeId         *big.Int
	Beneficiary    common.Address
}

type SyntheticDeviceNodeMintedData struct {
	IntegrationNode        *big.Int
	SyntheticDeviceNode    *big.Int
	VehicleNode            *big.Int
	SyntheticDeviceAddress common.Address
	Owner                  common.Address
}

type SyntheticDeviceNodeBurnedData struct {
	SyntheticDeviceNode *big.Int
	VehicleNode         *big.Int
	Owner               common.Address
}

type DeviceDefinition struct {
	Type struct {
		Make  string
		Model string
		Year  int
	}
}

type NewDCNNodeData struct {
	Node  []byte
	Owner common.Address
}

type NewDCNExpirationData struct {
	Node       []byte
	Expiration int
}

type DCNNameChangedData struct {
	Node []byte
	Name string `json:"name_"`
}

type DCNVehicleIdChangedData struct {
	Node      []byte
	VehicleID *big.Int `json:"vehicleId_"`
}

type TokensTransferredForDeviceData struct {
	User           common.Address
	Amount         *big.Int `json:"_amount"`
	VehicleNodeID  *big.Int
	DeviceNftProxy common.Address
	DeviceNode     *big.Int
	Week           *big.Int
}

type TokensTransferredForConnectionStreakData struct {
	User             common.Address
	Amount           *big.Int `json:"_amount"`
	VehicleNodeID    *big.Int
	ConnectionStreak *big.Int
	Week             *big.Int
}

type AftermarketDeviceAddressResetData struct {
	ManufacturerId           *big.Int
	TokenId                  *big.Int
	AftermarketDeviceAddress common.Address
}

type ManufacturerTableSetData struct {
	ManufacturerId *big.Int
	TableId        *big.Int
}

type DeviceDefinitionTableCreatedData struct {
	TableOwner     common.Address
	ManufacturerId *big.Int
	TableId        *big.Int
}

type PermissionsSetData struct {
	Asset       common.Address
	TokenId     *big.Int
	Permissions *big.Int
	Grantee     common.Address
	Expiration  *big.Int
	Source      string
}

type DeviceDefinitionIdSetData struct {
	VehicleId *big.Int
	DDID      string
}

type IssuedData struct {
	TokenID  *big.Int
	Owner    common.Address
	ClientID common.Address
}

type RedirectUriEnabledData struct {
	TokenID *big.Int
	URI     string
}

type RedirectUriDisabledData struct {
	TokenID *big.Int
	URI     string
}

type SignerEnabledData struct {
	TokenID *big.Int
	Signer  common.Address
}

type SignerDisabledData struct {
	TokenID *big.Int
	Signer  common.Address
}

type LicenseAliasSetData struct {
	TokenID      *big.Int
	LicenseAlias []byte
}

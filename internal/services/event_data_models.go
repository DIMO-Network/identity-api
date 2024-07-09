package services

import (
	"github.com/goccy/go-json"

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

type OwnerRegisteredData struct {
	Kernal common.Address
	Owner  common.Address
}

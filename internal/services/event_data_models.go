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

type TransferEventData struct {
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

type BeneficiarySetEventData struct {
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

type NewDCNNodeEventData struct {
	Node  []byte
	Owner common.Address
}

type NewDCNExpirationEventData struct {
	Node       []byte
	Expiration int
}

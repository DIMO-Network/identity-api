package model

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type AftermarketDevice struct {
	ID          int             `json:"id"`
	Address     *common.Address `json:"address,omitempty"`
	Owner       *common.Address `json:"owner,omitempty"`
	Serial      *string         `json:"serial,omitempty"`
	IMEI        *string         `json:"imei,omitempty"` // Override gqlgen's "Imei".
	MintedAt    *time.Time      `json:"mintedAt,omitempty"`
	VehicleID   *int            `json:"vehicleId,omitempty"` // We want to pass this to resolvers, but we don't want it in the API.
	Vehicle     *Vehicle        `json:"vehicle,omitempty"`
	Beneficiary common.Address  `json:"beneficiary,omitempty"`
}

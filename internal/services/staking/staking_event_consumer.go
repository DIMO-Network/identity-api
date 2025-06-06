// Code generated by gitub.com/DIMO-Network/eventgen. DO NOT EDIT.
package staking

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/DIMO-Network/identity-api/internal/services/models"
	"github.com/ethereum/go-ethereum/common"
)

var (
	StakedEventID          = common.HexToHash("0x1b2bd648de2b69d0b405d157e36eb4660a343e8923bb8cec3f9907d94601e05e")
	WithdrawnEventID       = common.HexToHash("0x75e161b3e824b114fc1a33274bd7091918dd4e639cede50b78b15a4eea956a21")
	StakingExtendedEventID = common.HexToHash("0x8dcdca9a89d1a6ecc036e8a86410ec8f3455be8c3cf5cb51147931e819357de0")
	VehicleAttachedEventID = common.HexToHash("0xcdd382de08d657468c7c74f0f59b15bf19da9a76903582c26c18756581cc9487")
	VehicleDetachedEventID = common.HexToHash("0x2c083deb67c92daa9b9ba680f4a657ac6cf0226e2be105c23b1497d0fdc06977")
	TransferEventID        = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
)

type Staked struct {
	User          common.Address `json:"user"`
	StakeId       *big.Int       `json:"stakeId"`
	StakingBeacon common.Address `json:"stakingBeacon"`
	Level         uint8          `json:"level"`
	Amount        *big.Int       `json:"amount"`
	LockEndTime   *big.Int       `json:"lockEndTime"`
	Points        *big.Int       `json:"points"`
}
type Withdrawn struct {
	User    common.Address `json:"user"`
	StakeId *big.Int       `json:"stakeId"`
	Amount  *big.Int       `json:"amount"`
	Points  *big.Int       `json:"points"`
}
type StakingExtended struct {
	User           common.Address `json:"user"`
	StakeId        *big.Int       `json:"stakeId"`
	NewLockEndTime *big.Int       `json:"newLockEndTime"`
	Points         *big.Int       `json:"points"`
}
type VehicleAttached struct {
	User      common.Address `json:"user"`
	StakeId   *big.Int       `json:"stakeId"`
	VehicleId *big.Int       `json:"vehicleId"`
}
type VehicleDetached struct {
	User      common.Address `json:"user"`
	StakeId   *big.Int       `json:"stakeId"`
	VehicleId *big.Int       `json:"vehicleId"`
}
type Transfer struct {
	From    common.Address `json:"from"`
	To      common.Address `json:"to"`
	TokenId *big.Int       `json:"tokenId"`
}

func (h *Handler) HandleEvent(ctx context.Context, event *models.ContractEventData) error {
	switch event.EventSignature {
	case StakedEventID:
		var args Staked
		err := json.Unmarshal(event.Arguments, &args)
		if err != nil {
			return err
		}
		return h.HandleStaked(ctx, event, &args)
	case WithdrawnEventID:
		var args Withdrawn
		err := json.Unmarshal(event.Arguments, &args)
		if err != nil {
			return err
		}
		return h.HandleWithdrawn(ctx, event, &args)
	case StakingExtendedEventID:
		var args StakingExtended
		err := json.Unmarshal(event.Arguments, &args)
		if err != nil {
			return err
		}
		return h.HandleStakingExtended(ctx, event, &args)
	case VehicleAttachedEventID:
		var args VehicleAttached
		err := json.Unmarshal(event.Arguments, &args)
		if err != nil {
			return err
		}
		return h.HandleVehicleAttached(ctx, event, &args)
	case VehicleDetachedEventID:
		var args VehicleDetached
		err := json.Unmarshal(event.Arguments, &args)
		if err != nil {
			return err
		}
		return h.HandleVehicleDetached(ctx, event, &args)
	case TransferEventID:
		var args Transfer
		err := json.Unmarshal(event.Arguments, &args)
		if err != nil {
			return err
		}
		return h.HandleTransfer(ctx, event, &args)
	}

	return nil
}


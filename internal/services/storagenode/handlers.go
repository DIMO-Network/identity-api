package storagenode

import (
	"context"
	"encoding/json"

	"github.com/DIMO-Network/identity-api/internal/helpers"
	cmodels "github.com/DIMO-Network/identity-api/internal/services/models"
	dmodels "github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

//go:generate go tool eventgen eventcfg.yaml -p storagenode -o events.go

type Handler struct {
	DBS    db.Store
	Logger *zerolog.Logger
}

func (h *Handler) Handle(ctx context.Context, ev *cmodels.ContractEventData) error {
	switch ev.EventSignature {
	case StorageNodeAnchorMintedEventID:
		return h.HandleStorageNodeAnchorMinted(ctx, ev)
	case TransferEventID:
		return h.HandleTransfer(ctx, ev)
	case NodeUriUpdatedEventID:
		return h.HandleNodeUriUpdated(ctx, ev)
	case NodeSetForVehicleEventID:
		return h.HandleNodeSetForVehicle(ctx, ev)
	default:
		return nil
	}
}

func (h *Handler) HandleStorageNodeAnchorMinted(ctx context.Context, ev *cmodels.ContractEventData) error {
	var snam StorageNodeAnchorMinted
	err := json.Unmarshal(ev.Arguments, &snam)
	if err != nil {
		return err
	}

	// This is the result of uint256(keccak256(bytes(label))). We more
	// or less use bytes32 internally.
	anchorID, err := helpers.ConvertTokenIDToID(snam.NodeAnchorId)
	if err != nil {
		return err
	}

	sn := dmodels.StorageNode{
		ID:       anchorID,
		Label:    snam.NodeAnchorLabel,
		Address:  snam.NodeAnchorAddr.Bytes(),
		Owner:    snam.Account.Bytes(),
		URI:      snam.NodeAnchorUri,
		MintedAt: ev.Block.Time,
	}

	h.Logger.Info().Str("label", snam.NodeAnchorLabel).Msg("Storage node minted.")

	return sn.Insert(ctx, h.DBS.DBS().Writer, boil.Infer())
}

func (h *Handler) HandleTransfer(ctx context.Context, ev *cmodels.ContractEventData) error {
	var t Transfer
	err := json.Unmarshal(ev.Arguments, &t)
	if err != nil {
		return err
	}

	// Let StorageNodeAnchorMinted take care of mints.
	if t.From == zeroAddr {
		return nil
	}

	anchorID, err := helpers.ConvertTokenIDToID(t.Id)
	if err != nil {
		return err
	}

	sn := dmodels.StorageNode{
		ID:    anchorID,
		Owner: t.To.Bytes(),
	}

	_, err = sn.Update(ctx, h.DBS.DBS().Writer, boil.Whitelist(dmodels.StorageNodeColumns.Owner))
	return err
}

func (h *Handler) HandleNodeUriUpdated(ctx context.Context, ev *cmodels.ContractEventData) error {
	var nuu NodeUriUpdated
	err := json.Unmarshal(ev.Arguments, &nuu)
	if err != nil {
		return err
	}

	anchorID, err := helpers.ConvertTokenIDToID(nuu.NodeId)
	if err != nil {
		return err
	}

	sn := dmodels.StorageNode{
		ID:  anchorID,
		URI: nuu.NewNodeUri,
	}

	_, err = sn.Update(ctx, h.DBS.DBS().Writer, boil.Whitelist(dmodels.StorageNodeColumns.URI))
	return err
}

func (h *Handler) HandleNodeSetForVehicle(ctx context.Context, ev *cmodels.ContractEventData) error {
	var nsfv NodeSetForVehicle
	err := json.Unmarshal(ev.Arguments, &nsfv)
	if err != nil {
		return err
	}

	vehicleID := nsfv.VehicleId.Int64()

	anchorID, err := helpers.ConvertTokenIDToID(nsfv.NodeId)
	if err != nil {
		return err
	}

	sn := dmodels.Vehicle{
		ID:            int(vehicleID),
		StorageNodeID: null.BytesFrom(anchorID),
	}

	_, err = sn.Update(ctx, h.DBS.DBS().Writer, boil.Whitelist(dmodels.VehicleColumns.StorageNodeID))
	return err
}

var zeroAddr common.Address

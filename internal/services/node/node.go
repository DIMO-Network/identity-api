package node

import (
	"context"
	"encoding/json"

	"github.com/DIMO-Network/identity-api/internal/helpers"
	cmodels "github.com/DIMO-Network/identity-api/internal/services/models"
	dmodels "github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/rs/zerolog"
)

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
		return h.HandleNodeUriUpdatedEvent(ctx, ev)
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

func (h *Handler) HandleNodeUriUpdatedEvent(ctx context.Context, ev *cmodels.ContractEventData) error {
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

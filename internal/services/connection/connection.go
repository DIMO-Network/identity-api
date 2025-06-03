package connection

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/DIMO-Network/identity-api/internal/helpers"
	cmodels "github.com/DIMO-Network/identity-api/internal/services/models"
	dmodels "github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/DIMO-Network/shared/pkg/db"

	"github.com/rs/zerolog"
)

type Handler struct {
	DBS    db.Store
	Logger *zerolog.Logger
}

func (h *Handler) Handle(ctx context.Context, ev *cmodels.ContractEventData) error {
	switch ev.EventSignature {
	case ConnectionMintedEventID:
		return h.HandleLicenseMinted(ctx, ev)
	case TransferEventID:
		return h.HandleTransfer(ctx, ev)
	default:
		return nil
	}
}

func (h *Handler) HandleLicenseMinted(ctx context.Context, ev *cmodels.ContractEventData) error {
	var lm ConnectionMinted
	err := json.Unmarshal(ev.Arguments, &lm)
	if err != nil {
		return err
	}

	cb, err := helpers.ConvertTokenIDToID(lm.ConnectionId)
	if err != nil {
		return err
	}

	conn := dmodels.Connection{
		ID:       cb,
		Address:  lm.ConnectionAddr.Bytes(),
		Owner:    lm.Account.Bytes(),
		MintedAt: ev.Block.Time,
	}

	err = conn.Upsert(ctx, h.DBS.DBS().Writer, false, []string{dmodels.ConnectionColumns.ID}, boil.None(), boil.Infer())
	if err != nil {
		return err
	}

	h.Logger.Info().Msgf("New connection %q with address %s minted.", convertIDToName(cb), lm.Account.Hex())

	return nil
}

func (h *Handler) HandleTransfer(ctx context.Context, ev *cmodels.ContractEventData) error {
	var t Transfer
	err := json.Unmarshal(ev.Arguments, &t)
	if err != nil {
		return err
	}

	if t.From == zeroAddr {
		// Handle ConnectionMinted instead.
		return nil
	}

	id, err := helpers.ConvertTokenIDToID(t.TokenId)
	if err != nil {
		return err
	}

	conn := dmodels.Connection{
		ID:    id,
		Owner: t.To.Bytes(),
	}

	_, err = conn.Update(ctx, h.DBS.DBS().Writer, boil.Whitelist(dmodels.ConnectionColumns.Owner))
	return err
}

func convertIDToName(id []byte) string {
	return string(bytes.TrimRight(id, "\x00"))
}

var zeroAddr common.Address

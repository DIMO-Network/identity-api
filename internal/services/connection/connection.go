package connection

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math/big"

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
	case LicenseMintedEventID:
		return h.HandleLicenseMinted(ctx, ev)
	case TransferEventID:
		return h.HandleTransfer(ctx, ev)
	default:
		return nil
	}
}

func (h *Handler) HandleLicenseMinted(ctx context.Context, ev *cmodels.ContractEventData) error {
	var lm LicenseMinted
	err := json.Unmarshal(ev.Arguments, &lm)
	if err != nil {
		return err
	}

	name, err := convertTokenIDToName(lm.LicenseId)
	if err != nil {
		return err
	}

	conn := dmodels.Connection{
		Name:     name,
		Address:  lm.LicenseAddr.Bytes(),
		Owner:    lm.Account.Bytes(),
		MintedAt: ev.Block.Time,
	}

	err = conn.Upsert(ctx, h.DBS.DBS().Writer, false, []string{dmodels.ConnectionColumns.Name}, boil.None(), boil.Infer())
	if err != nil {
		return err
	}

	h.Logger.Info().Msgf("New connection %q with address %s minted.", name, lm.Account.Hex())

	return nil
}

func (h *Handler) HandleTransfer(ctx context.Context, ev *cmodels.ContractEventData) error {
	var t Transfer
	err := json.Unmarshal(ev.Arguments, &t)
	if err != nil {
		return err
	}

	if t.From == zeroAddr {
		// Handle LicenseMinted instead.
		return nil
	}

	name, err := convertTokenIDToName(t.TokenId)
	if err != nil {
		return err
	}

	conn := dmodels.Connection{
		Name:  name,
		Owner: t.To.Bytes(),
	}

	_, err = conn.Update(ctx, h.DBS.DBS().Writer, boil.Whitelist(dmodels.ConnectionColumns.Owner))
	return err
}

var zeroAddr common.Address

func convertTokenIDToName(tokenID *big.Int) (string, error) {
	idBytes := tokenID.Bytes()
	if len(idBytes) > 32 {
		return "", errors.New("token id is more than 32 bytes")
	}

	// TODO(elffjs): What if this isn't valid UTF-8?
	idBytesTrimmed := bytes.TrimRight(idBytes, "\x00")

	return string(idBytesTrimmed), nil
}

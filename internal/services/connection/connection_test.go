package connection

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/services/models"
	dmodels "github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

const migrationsDir = "../../../migrations"

// You could say that this test is doing too much.
func TestHandleMintAndTransfer(t *testing.T) {
	ctx := context.Background()

	pdb, cont := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	defer cont.Terminate(t.Context()) //nolint

	log := zerolog.Nop()

	h := Handler{
		DBS:    pdb,
		Logger: &log,
	}

	// Case taken from
	// https://amoy.polygonscan.com/tx/0x344a769602df87e9c46f6f7f8752f6bb13ce6f9ae53e7598513af6c280c007a7
	err := h.Handle(t.Context(), &models.ContractEventData{
		EventSignature: common.HexToHash("0x16e7256a94e935dc419efa2b47bdb62ec5023b40947b172169eeb37b9b132686"),
		Block: models.Block{
			Time: time.Date(2025, 5, 4, 9, 0, 0, 0, time.UTC),
		},
		Arguments: []byte(`
{
	"account": "0xC008EF40B0b42AAD7e34879EB024385024f753ea",
	"connectionId": 37747592896913129884346430642309039154630403646040022073469559591247189901312,
	"connectionAddr": "0xb83DE952D389f9A6806819434450324197712FDA",
	"connectionName": "Staex",
	"connectionType": [53,60,0,6,41,31,227,82,7,42,41,217,143,3,3,229,190,100,74,224,37,28,180,32,125,55,108,225,133,19,113,107],
	"connectionCostInDimo": 2000
}`),
	})

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	c, err := dmodels.Connections().One(t.Context(), pdb.DBS().Reader)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if string(c.ID[:5]) != "Staex" {
		t.Errorf("Expected name Staex, got %q", string(bytes.TrimRight(c.ID, "\x00")))
	}

	if !bytes.Equal(c.Address, common.FromHex("0xb83DE952D389f9A6806819434450324197712FDA")) {
		t.Errorf("Unexpected address %q", c.Address)
	}

	if !c.MintedAt.Equal(time.Date(2025, 5, 4, 9, 0, 0, 0, time.UTC)) {
		t.Errorf("Unexpected mint time %q", c.MintedAt)
	}

	if !bytes.Equal(c.Owner, common.FromHex("0xC008EF40B0b42AAD7e34879EB024385024f753ea")) {
		t.Errorf("Unexpected owner %q", c.Owner)
	}

	err = h.Handle(t.Context(), &models.ContractEventData{
		EventSignature: common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
		Block: models.Block{
			Time: time.Date(2025, 5, 4, 11, 0, 0, 0, time.UTC),
		},
		Arguments: []byte(`
{
	"from": "0xc008ef40b0b42aad7e34879eb024385024f753ea",
	"to": "0x41799E9Dc893722844E771a1C1cAf3BBc2876132",
	"tokenId": 37747592896913129884346430642309039154630403646040022073469559591247189901312
}`),
	})

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	err = c.Reload(t.Context(), pdb.DBS().Reader.DB)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if !bytes.Equal(c.Owner, common.FromHex("0x41799E9Dc893722844E771a1C1cAf3BBc2876132")) {
		t.Errorf("Unexpected address %q", c.Address)
	}
}

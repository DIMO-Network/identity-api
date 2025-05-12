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

func TestHandleMintAndTransfer(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	log := zerolog.Nop()

	h := Handler{
		DBS:    pdb,
		Logger: &log,
	}

	// Case taken from
	// https://amoy.polygonscan.com/tx/0x344a769602df87e9c46f6f7f8752f6bb13ce6f9ae53e7598513af6c280c007a7
	err := h.Handle(t.Context(), &models.ContractEventData{
		EventSignature: common.HexToHash("0x4bce7eaeb5f9b0163fdd057deb2a52eefcf77f28f50703ef59a20ddcd4751067"),
		Block: models.Block{
			Time: time.Date(2025, 5, 4, 9, 0, 0, 0, time.UTC),
		},
		Arguments: []byte(`
{
	"account": "0xc008ef40b0b42aad7e34879eb024385024f753ea",
	"licenseId": 35025012972284307078409909351199105568725376826249000437555291297665349844992,
	"licenseAddr": "0x5879b43d88fa93ce8072d6612cbc8de93e98ce5d",
	"licenseCostInDimo": 90692000000000000000000
}`),
	})

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	c, err := dmodels.Connections().One(t.Context(), pdb.DBS().Reader)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if c.Name != "Motorq" {
		t.Errorf("Expected name Motorq, got %q", c.Name)
	}

	if !bytes.Equal(c.Address, common.FromHex("0x5879b43d88fa93ce8072d6612cbc8de93e98ce5d")) {
		t.Errorf("Unexpected address %q", c.Address)
	}

	if !c.MintedAt.Equal(time.Date(2025, 5, 4, 9, 0, 0, 0, time.UTC)) {
		t.Errorf("Unexpected mint time %q", c.MintedAt)
	}

	if !bytes.Equal(c.Owner, common.FromHex("0xc008ef40b0b42aad7e34879eb024385024f753ea")) {
		t.Errorf("Unexpected address %q", c.Address)
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
	"tokenId": 35025012972284307078409909351199105568725376826249000437555291297665349844992
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

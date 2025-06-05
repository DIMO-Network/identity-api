package loader

import (
	"context"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/repositories/connection"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const migrationsDir = "../../migrations"

func nameToConnID(name string) []byte {
	nameBytes := []byte(name)
	paddedBytes := make([]byte, 32)
	copy(paddedBytes, nameBytes)
	return paddedBytes
}

func TestBulk(t *testing.T) {
	ctx := context.Background()

	pdb, cont := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	defer cont.Terminate(t.Context()) //nolint

	log := zerolog.Nop()

	cl := ConnectionLoader{repo: connection.New(base.NewRepository(pdb, config.Settings{}, &log))}

	staexConnID := nameToConnID("Staex")
	teslaConnID := nameToConnID("Tesla")
	missingConnID := nameToConnID("Xdd")

	conn1 := models.Connection{
		Address:  common.FromHex("0xc008ef40b0b42aad7e34879eb024385024f753ea"),
		Owner:    common.FromHex("0xb83de952d389f9a6806819434450324197712fda"),
		MintedAt: time.Now(),
		ID:       staexConnID,
	}

	conn2 := models.Connection{
		Address:         common.FromHex("0x98308F9338841309E9F286e4053eA08d1963628B"),
		Owner:           common.FromHex("0xb83de952d389f9a6806819434450324197712fda"),
		MintedAt:        time.Now(),
		ID:              teslaConnID,
		IntegrationNode: null.IntFrom(2),
	}

	require.NoError(t, conn1.Insert(t.Context(), pdb.DBS().Writer, boil.Infer()))
	require.NoError(t, conn2.Insert(t.Context(), pdb.DBS().Writer, boil.Infer()))

	results := cl.BatchGetConnectionsByIDs(t.Context(),
		[]ConnectionQueryKey{
			{ConnectionID: [32]byte(staexConnID)},
			{ConnectionID: [32]byte(staexConnID)},
			{IntegrationNode: 2},
			{ConnectionID: [32]byte(missingConnID)},
		},
	)

	require.Len(t, results, 4)

	if assert.NoError(t, results[0].Error) {
		assert.Equal(t, "Staex", results[0].Data.Name)
	}

	if assert.NoError(t, results[1].Error) {
		assert.Equal(t, "Staex", results[1].Data.Name)
	}

	if assert.NoError(t, results[2].Error) {
		assert.Equal(t, "Tesla", results[2].Data.Name)
	}

	assert.Error(t, results[3].Error)
}

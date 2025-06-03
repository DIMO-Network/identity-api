package loader

import (
	"context"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const migrationsDir = "../../migrations"

func TestBulk(t *testing.T) {
	ctx := context.Background()

	pdb, cont := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	defer cont.Terminate(t.Context()) //nolint

	// log := zerolog.Nop()

	cl := ConnectionLoader{db: pdb}

	conn1 := models.Connection{
		Address:  common.FromHex("0xc008ef40b0b42aad7e34879eb024385024f753ea"),
		Owner:    common.FromHex("0xb83de952d389f9a6806819434450324197712fda"),
		MintedAt: time.Now(),
		ID:       common.FromHex("0x5374616578000000000000000000000000000000000000000000000000000000"),
	}

	conn2 := models.Connection{
		Address:  common.FromHex("0x98308F9338841309E9F286e4053eA08d1963628B"),
		Owner:    common.FromHex("0xb83de952d389f9a6806819434450324197712fda"),
		MintedAt: time.Now(),
		ID:       common.FromHex("5465736c61000000000000000000000000000000000000000000000000000000"),
	}

	require.NoError(t, conn1.Insert(t.Context(), pdb.DBS().Writer, boil.Infer()))
	require.NoError(t, conn2.Insert(t.Context(), pdb.DBS().Writer, boil.Infer()))

	results := cl.BatchGetConnectionsByIDs(t.Context(), [][32]byte{
		[32]byte(common.FromHex("0x5374616578000000000000000000000000000000000000000000000000000000")),
		[32]byte(common.FromHex("0x5374616578000000000000000000000000000000000000000000000000000000")),
		[32]byte(common.FromHex("0x5465736c61000000000000000000000000000000000000000000000000000000")),
		[32]byte(common.FromHex("0x5465736c61000000000000000000000000000000000000000000000000000001")),
	})

	require.Len(t, results, 4)

	assert.Equal(t, "Staex", results[0].Data.Name)
	assert.NoError(t, results[0].Error)

	assert.Equal(t, "Staex", results[1].Data.Name)
	assert.NoError(t, results[1].Error)

	assert.Equal(t, "Tesla", results[2].Data.Name)
	assert.NoError(t, results[2].Error)

	assert.Error(t, results[3].Error)
}

// \x5465736c61000000000000000000000000000000000000000000000000000000

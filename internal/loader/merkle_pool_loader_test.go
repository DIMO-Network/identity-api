package loader

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/dbtypes"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/repositories/merkle"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchGetMerklePoolsByID(t *testing.T) {
	ctx := context.Background()

	pdb, cont := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	defer cont.Terminate(t.Context()) //nolint

	log := zerolog.Nop()

	ml := NewMerklePoolLoader(&merkle.Repository{Repository: base.NewRepository(pdb, config.Settings{}, &log)})

	tokenAddr := common.HexToAddress("0xE261D618a959aFfFd53168Cd07D12E37B26761db")
	adminAddr := common.HexToAddress("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4")

	for _, poolID := range []int{0, 1} {
		pool := models.MerklePool{
			PoolID:    poolID,
			Token:     tokenAddr.Bytes(),
			Admin:     adminAddr.Bytes(),
			Balance:   dbtypes.IntToDecimal(big.NewInt(int64(1000 * (poolID + 1)))),
			CreatedAt: time.Now(),
		}
		require.NoError(t, pool.Insert(t.Context(), pdb.DBS().Writer, boil.Infer()))
	}

	results := ml.BatchGetMerklePoolsByID(t.Context(), []int{1, 0, 1, 99})

	require.Len(t, results, 4)

	if assert.NoError(t, results[0].Error) {
		require.NotNil(t, results[0].Data)
		assert.Equal(t, 1, results[0].Data.PoolID)
		assert.Equal(t, tokenAddr, results[0].Data.Token)
		assert.Equal(t, adminAddr, results[0].Data.Admin)
		assert.Equal(t, "2000", results[0].Data.Balance.String())
	}

	if assert.NoError(t, results[1].Error) {
		require.NotNil(t, results[1].Data)
		assert.Equal(t, 0, results[1].Data.PoolID)
		assert.Equal(t, "1000", results[1].Data.Balance.String())
	}

	// Duplicate keys in the batch each get the same pool.
	if assert.NoError(t, results[2].Error) {
		require.NotNil(t, results[2].Data)
		assert.Equal(t, 1, results[2].Data.PoolID)
	}

	// Unknown pools resolve to nil without error.
	if assert.NoError(t, results[3].Error) {
		assert.Nil(t, results[3].Data)
	}
}

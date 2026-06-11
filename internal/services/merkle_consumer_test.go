package services

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/services/merkle"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const merkleDistributorAddr = "0x4f5e9320b1c7cB3DE5ebDD760aD67375B66cF8a4"

// TestHandleMerklePoolCreatedEvent checks that the consumer routes
// MerkleDistributor events, addressed by the MERKLE_DISTRIBUTOR_ADDR setting,
// to the Merkle handler.
func TestHandleMerklePoolCreatedEvent(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()

	settings := config.Settings{
		DIMORegistryChainID:   contractEventData.ChainID,
		MerkleDistributorAddr: merkleDistributorAddr,
	}

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	consumer := NewContractsEventsConsumer(pdb, &logger, &settings)

	eventData := contractEventData
	eventData.EventName = merkle.PoolCreated
	eventData.Contract = common.HexToAddress(merkleDistributorAddr)

	args := merkle.PoolCreatedData{
		PoolId:      big.NewInt(0),
		Token:       common.HexToAddress("0xE261D618a959aFfFd53168Cd07D12E37B26761db"),
		Admin:       common.HexToAddress("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		WeeklyLimit: big.NewInt(5000),
	}

	e := prepareEvent(t, eventData, args)

	require.NoError(t, consumer.Process(ctx, &e))

	pool, err := models.FindMerklePool(ctx, pdb.DBS().Reader, 0)
	require.NoError(t, err)
	assert.Equal(t, args.Token.Bytes(), pool.Token)
	assert.Equal(t, args.Admin.Bytes(), pool.Admin)
	assert.Equal(t, "5000", pool.WeeklyLimit.String())
	assert.Equal(t, "0", pool.Balance.String())
}

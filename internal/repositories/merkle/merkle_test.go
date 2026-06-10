package merkle

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/dbtypes"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/big"
)

const migrationsDirRelPath = "../../../migrations"

var (
	tokenAddr = common.HexToAddress("0xE261D618a959aFfFd53168Cd07D12E37B26761db")
	adminAddr = common.HexToAddress("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4")
	account1  = common.HexToAddress("0x2222222222222222222222222222222222222222")
	account2  = common.HexToAddress("0x3333333333333333333333333333333333333333")
	claimTx   = common.HexToHash("0x811a85e24d0129a2018c9a6668652db63d73bc6d1c76f21b07da2162c6bfea7d")

	proof1 = []byte(`["0x1111111111111111111111111111111111111111111111111111111111111111","0x2222222222222222222222222222222222222222222222222222222222222222"]`)
	proof2 = []byte(`["0x3333333333333333333333333333333333333333333333333333333333333333"]`)
)

func setupRepo(t *testing.T) (*Repository, context.Context) {
	t.Helper()
	ctx := context.Background()

	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)

	repo := Repository{Repository: base.NewRepository(pdb, config.Settings{}, &logger)}

	writer := pdb.DBS().Writer

	pool := models.MerklePool{
		PoolID:      0,
		Token:       tokenAddr.Bytes(),
		Admin:       adminAddr.Bytes(),
		WeeklyLimit: dbtypes.NullIntToDecimal(big.NewInt(5000)),
		Balance:     dbtypes.IntToDecimal(big.NewInt(900)),
		CreatedAt:   time.Now(),
	}
	require.NoError(t, pool.Insert(ctx, writer, boil.Infer()))

	root213 := models.MerkleRoot{
		PoolID:         0,
		Epoch:          213,
		Root:           common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Bytes(),
		Allocation:     dbtypes.IntToDecimal(big.NewInt(300)),
		TotalClaimed:   dbtypes.IntToDecimal(big.NewInt(100)),
		ClaimCount:     1,
		RecipientCount: 2,
		ProofsURI:      "https://merkle.dimo.zone/pool-0/week-213.json",
		SetAt:          time.Now(),
	}
	require.NoError(t, root213.Insert(ctx, writer, boil.Infer()))

	root214 := models.MerkleRoot{
		PoolID:         0,
		Epoch:          214,
		Root:           common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb").Bytes(),
		Allocation:     dbtypes.IntToDecimal(big.NewInt(400)),
		TotalClaimed:   dbtypes.IntToDecimal(big.NewInt(0)),
		ClaimCount:     0,
		RecipientCount: 1,
		ProofsURI:      "https://merkle.dimo.zone/pool-0/week-214.json",
		SetAt:          time.Now(),
	}
	require.NoError(t, root214.Insert(ctx, writer, boil.Infer()))

	claimed := models.MerkleClaim{
		PoolID:    0,
		Epoch:     213,
		Account:   account1.Bytes(),
		Amount:    dbtypes.IntToDecimal(big.NewInt(100)),
		Proof:     proof1,
		ClaimedAt: null.TimeFrom(time.Now()),
		ClaimTX:   null.BytesFrom(claimTx.Bytes()),
	}
	require.NoError(t, claimed.Insert(ctx, writer, boil.Infer()))

	unclaimed := models.MerkleClaim{
		PoolID:  0,
		Epoch:   214,
		Account: account1.Bytes(),
		Amount:  dbtypes.IntToDecimal(big.NewInt(150)),
		Proof:   proof2,
	}
	require.NoError(t, unclaimed.Insert(ctx, writer, boil.Infer()))

	other := models.MerkleClaim{
		PoolID:  0,
		Epoch:   213,
		Account: account2.Bytes(),
		Amount:  dbtypes.IntToDecimal(big.NewInt(200)),
		Proof:   proof1,
	}
	require.NoError(t, other.Insert(ctx, writer, boil.Infer()))

	return &repo, ctx
}

func TestGetMerklePool(t *testing.T) {
	repo, ctx := setupRepo(t)

	pool, err := repo.GetMerklePool(ctx, 0)
	require.NoError(t, err)
	require.NotNil(t, pool)
	assert.Equal(t, 0, pool.PoolID)
	assert.Equal(t, tokenAddr, pool.Token)
	assert.Equal(t, adminAddr, pool.Admin)
	assert.Equal(t, "5000", pool.WeeklyLimit.String())
	assert.Equal(t, "900", pool.Balance.String())

	missing, err := repo.GetMerklePool(ctx, 7)
	require.NoError(t, err)
	assert.Nil(t, missing)
}

func TestGetMerklePools(t *testing.T) {
	repo, ctx := setupRepo(t)

	first := 10
	conn, err := repo.GetMerklePools(ctx, &first, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, conn.TotalCount)
	require.Len(t, conn.Nodes, 1)
	assert.Equal(t, 0, conn.Nodes[0].PoolID)
	assert.False(t, conn.PageInfo.HasNextPage)
}

func TestGetPoolEpochs(t *testing.T) {
	repo, ctx := setupRepo(t)

	pool, err := repo.GetMerklePool(ctx, 0)
	require.NoError(t, err)

	first := 10
	conn, err := repo.GetPoolEpochs(ctx, pool, &first, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 2, conn.TotalCount)
	require.Len(t, conn.Nodes, 2)

	// Descending by epoch.
	assert.Equal(t, 214, conn.Nodes[0].Epoch)
	assert.Equal(t, 213, conn.Nodes[1].Epoch)

	assert.Equal(t, "400", conn.Nodes[0].Allocation.String())
	assert.Equal(t, 1, conn.Nodes[0].RecipientCount)
	assert.Equal(t, "100", conn.Nodes[1].TotalClaimed.String())
	assert.Equal(t, 1, conn.Nodes[1].ClaimCount)
	assert.Equal(t, "https://merkle.dimo.zone/pool-0/week-213.json", conn.Nodes[1].ProofsURI)

	// Pagination with first = 1.
	one := 1
	page, err := repo.GetPoolEpochs(ctx, pool, &one, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, page.Nodes, 1)
	assert.Equal(t, 214, page.Nodes[0].Epoch)
	assert.True(t, page.PageInfo.HasNextPage)

	next, err := repo.GetPoolEpochs(ctx, pool, &one, page.PageInfo.EndCursor, nil, nil)
	require.NoError(t, err)
	require.Len(t, next.Nodes, 1)
	assert.Equal(t, 213, next.Nodes[0].Epoch)
}

func TestGetMerkleRewards(t *testing.T) {
	repo, ctx := setupRepo(t)

	first := 10

	conn, err := repo.GetMerkleRewards(ctx, account1, nil, nil, &first, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 2, conn.TotalCount)
	require.Len(t, conn.Nodes, 2)

	// Descending by (pool, epoch).
	unclaimed, claimed := conn.Nodes[0], conn.Nodes[1]

	assert.Equal(t, 214, unclaimed.Epoch)
	assert.Equal(t, account1, unclaimed.Account)
	assert.Equal(t, "150", unclaimed.Amount.String())
	assert.False(t, unclaimed.Claimed)
	assert.Nil(t, unclaimed.ClaimedAt)
	assert.Nil(t, unclaimed.ClaimTx)
	assert.Equal(t, []string{"0x3333333333333333333333333333333333333333333333333333333333333333"}, unclaimed.Proof)

	assert.Equal(t, 213, claimed.Epoch)
	assert.Equal(t, "100", claimed.Amount.String())
	assert.True(t, claimed.Claimed)
	require.NotNil(t, claimed.ClaimedAt)
	assert.Equal(t, claimTx.Bytes(), claimed.ClaimTx)
	assert.Len(t, claimed.Proof, 2)

	// Filter on claim status.
	claimedTrue := true
	conn, err = repo.GetMerkleRewards(ctx, account1, nil, &claimedTrue, &first, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, conn.TotalCount)
	require.Len(t, conn.Nodes, 1)
	assert.Equal(t, 213, conn.Nodes[0].Epoch)

	claimedFalse := false
	conn, err = repo.GetMerkleRewards(ctx, account1, nil, &claimedFalse, &first, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, conn.TotalCount)
	require.Len(t, conn.Nodes, 1)
	assert.Equal(t, 214, conn.Nodes[0].Epoch)

	// Filter on pool.
	poolID := 0
	conn, err = repo.GetMerkleRewards(ctx, account1, &poolID, nil, &first, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 2, conn.TotalCount)

	otherPool := 5
	conn, err = repo.GetMerkleRewards(ctx, account1, &otherPool, nil, &first, nil, nil, nil)
	require.NoError(t, err)
	assert.Zero(t, conn.TotalCount)
	assert.Empty(t, conn.Nodes)

	// Other account only sees its own rewards.
	conn, err = repo.GetMerkleRewards(ctx, account2, nil, nil, &first, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, conn.TotalCount)
	require.Len(t, conn.Nodes, 1)
	assert.Equal(t, "200", conn.Nodes[0].Amount.String())

	// Cursor pagination.
	one := 1
	page, err := repo.GetMerkleRewards(ctx, account1, nil, nil, &one, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, page.Nodes, 1)
	assert.Equal(t, 214, page.Nodes[0].Epoch)
	assert.True(t, page.PageInfo.HasNextPage)

	next, err := repo.GetMerkleRewards(ctx, account1, nil, nil, &one, page.PageInfo.EndCursor, nil, nil)
	require.NoError(t, err)
	require.Len(t, next.Nodes, 1)
	assert.Equal(t, 213, next.Nodes[0].Epoch)
}

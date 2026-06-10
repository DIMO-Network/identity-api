package merkle

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/helpers"
	cmodels "github.com/DIMO-Network/identity-api/internal/services/models"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/identity-api/pkg/merkletree"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const migrationsDirRelPath = "../../../migrations"

var (
	distributorAddr = common.HexToAddress("0x4f5e9320b1c7cB3DE5ebDD760aD67375B66cF8a4")
	tokenAddr       = common.HexToAddress("0xE261D618a959aFfFd53168Cd07D12E37B26761db")
	adminAddr       = common.HexToAddress("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4")
	account1        = common.HexToAddress("0x2222222222222222222222222222222222222222")
	account2        = common.HexToAddress("0x3333333333333333333333333333333333333333")
	account3        = common.HexToAddress("0x4444444444444444444444444444444444444444")
)

type fakeFetcher struct {
	data map[string][]byte
}

func (f *fakeFetcher) Fetch(_ context.Context, uri string) ([]byte, error) {
	d, ok := f.data[uri]
	if !ok {
		return nil, fmt.Errorf("no data for uri %q", uri)
	}
	return d, nil
}

func eventData(t *testing.T, name string, blockTime time.Time, args any) *cmodels.ContractEventData {
	t.Helper()

	argBytes, err := json.Marshal(args)
	require.NoError(t, err)

	return &cmodels.ContractEventData{
		EventName:       name,
		Contract:        distributorAddr,
		TransactionHash: common.HexToHash("0x811a85e24d0129a2018c9a6668652db63d73bc6d1c76f21b07da2162c6bfea7d"),
		Block:           cmodels.Block{Time: blockTime},
		Arguments:       argBytes,
	}
}

func newTestHandler(t *testing.T, fetcher TreeFetcher) (*Handler, context.Context) {
	t.Helper()
	ctx := context.Background()

	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", helpers.DBSettings.Name).Logger()
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDirRelPath)

	return &Handler{DBS: pdb, Logger: &logger, Fetcher: fetcher}, ctx
}

func createPool(t *testing.T, ctx context.Context, h *Handler, poolID int64) {
	t.Helper()

	err := h.Handle(ctx, eventData(t, PoolCreated, time.Now(), PoolCreatedData{
		PoolId:      big.NewInt(poolID),
		Token:       tokenAddr,
		Admin:       adminAddr,
		WeeklyLimit: big.NewInt(5000),
	}))
	require.NoError(t, err)
}

func buildTree(t *testing.T, poolID, week int64) (*merkletree.Tree, []byte) {
	t.Helper()

	tree, err := merkletree.New(distributorAddr, big.NewInt(poolID), big.NewInt(week), []merkletree.Leaf{
		{Account: account1, Amount: big.NewInt(100)},
		{Account: account2, Amount: big.NewInt(200)},
		{Account: account3, Amount: big.NewInt(300)},
	})
	require.NoError(t, err)

	data, err := tree.MarshalJSON()
	require.NoError(t, err)

	return tree, data
}

func TestHandlePoolCreatedAndBalanceEvents(t *testing.T) {
	h, ctx := newTestHandler(t, &fakeFetcher{})

	createdAt := time.Now()

	err := h.Handle(ctx, eventData(t, PoolCreated, createdAt, PoolCreatedData{
		PoolId:      big.NewInt(0),
		Token:       tokenAddr,
		Admin:       adminAddr,
		WeeklyLimit: big.NewInt(5000),
	}))
	require.NoError(t, err)

	pool, err := models.FindMerklePool(ctx, h.DBS.DBS().Reader, 0)
	require.NoError(t, err)
	assert.Equal(t, tokenAddr.Bytes(), pool.Token)
	assert.Equal(t, adminAddr.Bytes(), pool.Admin)
	assert.Equal(t, "5000", pool.WeeklyLimit.Big.String())
	assert.Equal(t, "0", pool.Balance.Big.String())
	assert.WithinDuration(t, createdAt, pool.CreatedAt, time.Second)

	// WeeklyLimitSet updates the limit.
	err = h.Handle(ctx, eventData(t, WeeklyLimitSet, time.Now(), WeeklyLimitSetData{
		PoolId: big.NewInt(0),
		Limit:  big.NewInt(7000),
	}))
	require.NoError(t, err)

	// Funded adds to the balance.
	err = h.Handle(ctx, eventData(t, Funded, time.Now(), FundedData{
		PoolId: big.NewInt(0),
		From:   adminAddr,
		Amount: big.NewInt(1000),
	}))
	require.NoError(t, err)

	require.NoError(t, pool.Reload(ctx, h.DBS.DBS().Reader))
	assert.Equal(t, "7000", pool.WeeklyLimit.Big.String())
	assert.Equal(t, "1000", pool.Balance.Big.String())

	// Swept sets the balance to the new absolute value.
	err = h.Handle(ctx, eventData(t, Swept, time.Now(), SweptData{
		PoolId:     big.NewInt(0),
		To:         adminAddr,
		Amount:     big.NewInt(600),
		NewBalance: big.NewInt(400),
	}))
	require.NoError(t, err)

	require.NoError(t, pool.Reload(ctx, h.DBS.DBS().Reader))
	assert.Equal(t, "400", pool.Balance.Big.String())

	// Balance events for unknown pools are errors.
	err = h.Handle(ctx, eventData(t, Funded, time.Now(), FundedData{
		PoolId: big.NewInt(99),
		From:   adminAddr,
		Amount: big.NewInt(1000),
	}))
	assert.Error(t, err)
}

func TestHandleRootSet(t *testing.T) {
	tree, treeData := buildTree(t, 0, 214)
	uri := "https://merkle.dimo.zone/pool-0/week-214.json"

	h, ctx := newTestHandler(t, &fakeFetcher{data: map[string][]byte{uri: treeData}})
	createPool(t, ctx, h, 0)

	setAt := time.Now()
	root := tree.Root()

	err := h.Handle(ctx, eventData(t, RootSet, setAt, RootSetData{
		PoolId:     big.NewInt(0),
		Week:       big.NewInt(214),
		Root:       root,
		Allocation: big.NewInt(600),
		ProofsURI:  uri,
	}))
	require.NoError(t, err)

	dbRoot, err := models.FindMerkleRoot(ctx, h.DBS.DBS().Reader, 0, 214)
	require.NoError(t, err)
	assert.Equal(t, root[:], dbRoot.Root)
	assert.Equal(t, "600", dbRoot.Allocation.Big.String())
	assert.Equal(t, 3, dbRoot.RecipientCount)
	assert.Equal(t, 0, dbRoot.ClaimCount)
	assert.Equal(t, "0", dbRoot.TotalClaimed.Big.String())
	assert.Equal(t, uri, dbRoot.ProofsURI)
	assert.WithinDuration(t, setAt, dbRoot.SetAt, time.Second)

	claims, err := models.MerkleClaims().All(ctx, h.DBS.DBS().Reader)
	require.NoError(t, err)
	assert.Len(t, claims, 3)

	claim, err := models.FindMerkleClaim(ctx, h.DBS.DBS().Reader, 0, 214, account1.Bytes())
	require.NoError(t, err)
	assert.Equal(t, "100", claim.Amount.Big.String())
	assert.False(t, claim.ClaimedAt.Valid)

	proof, err := tree.Proof(account1)
	require.NoError(t, err)
	expected := make([]string, len(proof))
	for i, p := range proof {
		expected[i] = p.Hex()
	}

	var stored []string
	require.NoError(t, json.Unmarshal(claim.Proof, &stored))
	assert.Equal(t, expected, stored)
}

func TestHandleRootSetTamperedFile(t *testing.T) {
	tree, treeData := buildTree(t, 0, 214)
	uri := "https://merkle.dimo.zone/pool-0/week-214.json"

	// Tamper with a leaf amount so root verification fails.
	var raw map[string]any
	require.NoError(t, json.Unmarshal(treeData, &raw))
	leaves := raw["leaves"].([]any)
	leaves[0].(map[string]any)["amount"] = "999"
	tampered, err := json.Marshal(raw)
	require.NoError(t, err)

	h, ctx := newTestHandler(t, &fakeFetcher{data: map[string][]byte{uri: tampered}})
	createPool(t, ctx, h, 0)

	err = h.Handle(ctx, eventData(t, RootSet, time.Now(), RootSetData{
		PoolId:     big.NewInt(0),
		Week:       big.NewInt(214),
		Root:       tree.Root(),
		Allocation: big.NewInt(600),
		ProofsURI:  uri,
	}))
	require.Error(t, err)

	// Nothing should have been written.
	count, err := models.MerkleRoots().Count(ctx, h.DBS.DBS().Reader)
	require.NoError(t, err)
	assert.Zero(t, count)

	count, err = models.MerkleClaims().Count(ctx, h.DBS.DBS().Reader)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestHandleRootSetEventRootMismatch(t *testing.T) {
	_, treeData := buildTree(t, 0, 214)
	uri := "https://merkle.dimo.zone/pool-0/week-214.json"

	h, ctx := newTestHandler(t, &fakeFetcher{data: map[string][]byte{uri: treeData}})
	createPool(t, ctx, h, 0)

	// The file is internally consistent, but the on-chain root is different.
	err := h.Handle(ctx, eventData(t, RootSet, time.Now(), RootSetData{
		PoolId:     big.NewInt(0),
		Week:       big.NewInt(214),
		Root:       common.HexToHash("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
		Allocation: big.NewInt(600),
		ProofsURI:  uri,
	}))
	require.Error(t, err)

	count, err := models.MerkleRoots().Count(ctx, h.DBS.DBS().Reader)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestHandleClaimedIdempotent(t *testing.T) {
	tree, treeData := buildTree(t, 0, 214)
	uri := "https://merkle.dimo.zone/pool-0/week-214.json"

	h, ctx := newTestHandler(t, &fakeFetcher{data: map[string][]byte{uri: treeData}})
	createPool(t, ctx, h, 0)

	err := h.Handle(ctx, eventData(t, Funded, time.Now(), FundedData{
		PoolId: big.NewInt(0),
		From:   adminAddr,
		Amount: big.NewInt(1000),
	}))
	require.NoError(t, err)

	err = h.Handle(ctx, eventData(t, RootSet, time.Now(), RootSetData{
		PoolId:     big.NewInt(0),
		Week:       big.NewInt(214),
		Root:       tree.Root(),
		Allocation: big.NewInt(600),
		ProofsURI:  uri,
	}))
	require.NoError(t, err)

	claimedAt := time.Now()
	claimedEvent := eventData(t, Claimed, claimedAt, ClaimedData{
		PoolId:  big.NewInt(0),
		Week:    big.NewInt(214),
		Account: account1,
		Amount:  big.NewInt(100),
	})

	require.NoError(t, h.Handle(ctx, claimedEvent))

	checkState := func() {
		claim, err := models.FindMerkleClaim(ctx, h.DBS.DBS().Reader, 0, 214, account1.Bytes())
		require.NoError(t, err)
		require.True(t, claim.ClaimedAt.Valid)
		assert.WithinDuration(t, claimedAt, claim.ClaimedAt.Time, time.Second)
		assert.Equal(t, claimedEvent.TransactionHash.Bytes(), claim.ClaimTX.Bytes)

		root, err := models.FindMerkleRoot(ctx, h.DBS.DBS().Reader, 0, 214)
		require.NoError(t, err)
		assert.Equal(t, 1, root.ClaimCount)
		assert.Equal(t, "100", root.TotalClaimed.Big.String())

		pool, err := models.FindMerklePool(ctx, h.DBS.DBS().Reader, 0)
		require.NoError(t, err)
		assert.Equal(t, "900", pool.Balance.Big.String())
	}

	checkState()

	// Redelivery of the same event must not double-count.
	require.NoError(t, h.Handle(ctx, claimedEvent))
	checkState()

	// Setting the root again must preserve the claim state.
	err = h.Handle(ctx, eventData(t, RootSet, time.Now(), RootSetData{
		PoolId:     big.NewInt(0),
		Week:       big.NewInt(214),
		Root:       tree.Root(),
		Allocation: big.NewInt(600),
		ProofsURI:  uri,
	}))
	require.NoError(t, err)
	checkState()

	// A claim for a leaf we don't know about is an error.
	err = h.Handle(ctx, eventData(t, Claimed, time.Now(), ClaimedData{
		PoolId:  big.NewInt(0),
		Week:    big.NewInt(214),
		Account: common.HexToAddress("0x5555555555555555555555555555555555555555"),
		Amount:  big.NewInt(50),
	}))
	assert.Error(t, err)
}

package merkletree

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTreeFileRoundTrip is a smoke test for the copied package: build a 3-leaf
// tree, marshal it to the dimo-merkle-v1 format, parse it back, and check that
// VerifyRoot passes. Then tamper with a leaf and check that VerifyRoot fails.
func TestTreeFileRoundTrip(t *testing.T) {
	distributor := common.HexToAddress("0x1111111111111111111111111111111111111111")
	leaves := []Leaf{
		{Account: common.HexToAddress("0x2222222222222222222222222222222222222222"), Amount: big.NewInt(100)},
		{Account: common.HexToAddress("0x3333333333333333333333333333333333333333"), Amount: big.NewInt(200)},
		{Account: common.HexToAddress("0x4444444444444444444444444444444444444444"), Amount: big.NewInt(300)},
	}

	tree, err := New(distributor, big.NewInt(0), big.NewInt(214), leaves)
	require.NoError(t, err)

	data, err := tree.MarshalJSON()
	require.NoError(t, err)

	file, err := UnmarshalTreeFile(data)
	require.NoError(t, err)

	assert.Equal(t, tree.Root(), file.Root)
	assert.Len(t, file.Leaves, 3)
	require.NoError(t, file.VerifyRoot())

	// Tampering with an amount must break root verification.
	file.Leaves[0].Amount = new(big.Int).Add(file.Leaves[0].Amount, big.NewInt(1))
	assert.Error(t, file.VerifyRoot())
}

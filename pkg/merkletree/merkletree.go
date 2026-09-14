// Copied from DIMO-Network/rewards-api pkg/merkletree (commit b61e90a); replace
// with module import once published.

// Package merkletree implements a Merkle tree byte-compatible with OpenZeppelin's
// StandardMerkleTree for the leaf encoding ["address", "uint256", "uint256",
// "address", "uint256"] = (distributor, poolId, week, account, amount), used for
// DIMO reward claims.
//
// Compatibility notes, ported from @openzeppelin/merkle-tree:
//   - Leaf hash: keccak256(keccak256(abi.encode(distributor, poolId, week, account, amount))).
//   - Leaves are sorted ascending by leaf hash, then stored at the tail of an
//     array-backed binary tree of size 2n-1, with sorted leaf i at index 2n-2-i.
//   - Parent hash: keccak256(concat(sort(a, b))) (commutative).
//   - A proof is the list of sibling hashes on the path from a leaf to the root.
package merkletree

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var leafABIArguments = func() abi.Arguments {
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		panic(fmt.Sprintf("merkletree: creating address ABI type: %v", err))
	}
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		panic(fmt.Sprintf("merkletree: creating uint256 ABI type: %v", err))
	}
	return abi.Arguments{
		{Type: addressType}, // distributor
		{Type: uint256Type}, // poolId
		{Type: uint256Type}, // week
		{Type: addressType}, // account
		{Type: uint256Type}, // amount
	}
}()

// Leaf is a single claim entry: an account and the amount it can claim.
type Leaf struct {
	Account common.Address
	Amount  *big.Int
}

// Tree is an OpenZeppelin StandardMerkleTree-compatible Merkle tree over reward
// claim leaves. Construct it with New.
type Tree struct {
	distributor common.Address
	poolID      *big.Int
	week        *big.Int
	// leaves in canonical order, i.e. ascending by leaf hash. The leaf at
	// position i sits at tree node index len(nodes)-1-i.
	leaves []Leaf
	// nodes is the array-backed binary tree of size 2n-1. nodes[0] is the root;
	// the children of node i are nodes 2i+1 and 2i+2.
	nodes []common.Hash
	// nodeIndexByAccount maps an account to the tree node index of its leaf.
	nodeIndexByAccount map[common.Address]int
}

// New builds a Merkle tree for the given distributor, pool, and week over the
// given leaves. It returns an error if leaves is empty, if any account appears
// more than once, or if any amount is nil, zero, or negative.
func New(distributor common.Address, poolID, week *big.Int, leaves []Leaf) (*Tree, error) {
	if len(leaves) == 0 {
		return nil, fmt.Errorf("expected non-zero number of leaves")
	}
	if poolID == nil || poolID.Sign() < 0 {
		return nil, fmt.Errorf("pool id must be a non-negative integer")
	}
	if week == nil || week.Sign() < 0 {
		return nil, fmt.Errorf("week must be a non-negative integer")
	}

	seen := make(map[common.Address]struct{}, len(leaves))
	for _, leaf := range leaves {
		if _, ok := seen[leaf.Account]; ok {
			return nil, fmt.Errorf("duplicate account %s", leaf.Account.Hex())
		}
		seen[leaf.Account] = struct{}{}
		if leaf.Amount == nil {
			return nil, fmt.Errorf("nil amount for account %s", leaf.Account.Hex())
		}
		if leaf.Amount.Sign() <= 0 {
			return nil, fmt.Errorf("amount for account %s must be positive, got %s", leaf.Account.Hex(), leaf.Amount)
		}
		if leaf.Amount.BitLen() > 256 {
			return nil, fmt.Errorf("amount for account %s overflows uint256", leaf.Account.Hex())
		}
	}

	type hashedLeaf struct {
		leaf Leaf
		hash common.Hash
	}
	hashed := make([]hashedLeaf, len(leaves))
	for i, leaf := range leaves {
		hash, err := leafHash(distributor, poolID, week, leaf)
		if err != nil {
			return nil, fmt.Errorf("hashing leaf for account %s: %w", leaf.Account.Hex(), err)
		}
		hashed[i] = hashedLeaf{leaf: leaf, hash: hash}
	}

	// OZ sorts leaves ascending by hash before building the tree. Leaf hashes
	// are collision-free in practice, so the order is total and sort.Slice
	// suffices; stability would be meaningless here.
	sort.Slice(hashed, func(i, j int) bool {
		return bytes.Compare(hashed[i].hash[:], hashed[j].hash[:]) < 0
	})

	n := len(hashed)
	nodes := make([]common.Hash, 2*n-1)
	sorted := make([]Leaf, n)
	nodeIndexByAccount := make(map[common.Address]int, n)
	for i, hl := range hashed {
		nodeIndex := len(nodes) - 1 - i
		nodes[nodeIndex] = hl.hash
		sorted[i] = hl.leaf
		nodeIndexByAccount[hl.leaf.Account] = nodeIndex
	}
	for i := len(nodes) - 1 - n; i >= 0; i-- {
		nodes[i] = hashPair(nodes[2*i+1], nodes[2*i+2])
	}

	return &Tree{
		distributor:        distributor,
		poolID:             new(big.Int).Set(poolID),
		week:               new(big.Int).Set(week),
		leaves:             sorted,
		nodes:              nodes,
		nodeIndexByAccount: nodeIndexByAccount,
	}, nil
}

// Root returns the Merkle root of the tree.
func (t *Tree) Root() common.Hash {
	return t.nodes[0]
}

// Proof returns the Merkle proof for the leaf belonging to the given account.
// It returns an error if the account is not in the tree.
func (t *Tree) Proof(account common.Address) ([]common.Hash, error) {
	index, ok := t.nodeIndexByAccount[account]
	if !ok {
		return nil, fmt.Errorf("account %s is not in the tree", account.Hex())
	}
	proof := []common.Hash{}
	for index > 0 {
		proof = append(proof, t.nodes[siblingIndex(index)])
		index = (index - 1) / 2
	}
	return proof, nil
}

// Leaves returns the leaves in canonical order, i.e. ascending by leaf hash.
func (t *Tree) Leaves() []Leaf {
	leaves := make([]Leaf, len(t.leaves))
	copy(leaves, t.leaves)
	return leaves
}

// MarshalJSON encodes the tree in the dimo-merkle-v1 file format. All numeric
// values are encoded as decimal strings, never as JSON numbers, and addresses
// are EIP-55 checksummed.
func (t *Tree) MarshalJSON() ([]byte, error) {
	file := TreeFile{
		Format:      TreeFileFormatV1,
		Distributor: t.distributor,
		PoolID:      new(big.Int).Set(t.poolID),
		Week:        new(big.Int).Set(t.week),
		Root:        t.Root(),
		Leaves:      make([]TreeFileLeaf, len(t.leaves)),
	}
	for i, leaf := range t.leaves {
		proof, err := t.Proof(leaf.Account)
		if err != nil {
			return nil, fmt.Errorf("computing proof for account %s: %w", leaf.Account.Hex(), err)
		}
		file.Leaves[i] = TreeFileLeaf{
			Account: leaf.Account,
			Amount:  new(big.Int).Set(leaf.Amount),
			Proof:   proof,
		}
	}
	return file.MarshalJSON()
}

// leafHash computes the OZ standard leaf hash: the keccak256 of the keccak256
// of the standard ABI encoding of (distributor, poolId, week, account, amount).
func leafHash(distributor common.Address, poolID, week *big.Int, leaf Leaf) (common.Hash, error) {
	encoded, err := leafABIArguments.Pack(distributor, poolID, week, leaf.Account, leaf.Amount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("abi-encoding leaf: %w", err)
	}
	return crypto.Keccak256Hash(crypto.Keccak256(encoded)), nil
}

// hashPair computes the OZ standard node hash: the keccak256 of the
// concatenation of a and b with the smaller of the two first.
func hashPair(a, b common.Hash) common.Hash {
	if bytes.Compare(a[:], b[:]) > 0 {
		a, b = b, a
	}
	return crypto.Keccak256Hash(a[:], b[:])
}

// siblingIndex returns the index of the sibling of node i in the array-backed
// tree. The root (i = 0) has no sibling.
func siblingIndex(i int) int {
	if i%2 == 1 {
		return i + 1
	}
	return i - 1
}

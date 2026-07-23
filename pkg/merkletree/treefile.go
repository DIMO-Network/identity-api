// Copied from DIMO-Network/rewards-api pkg/merkletree (commit b61e90a); replace
// with module import once published.

package merkletree

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// TreeFileFormatV1 is the format identifier for version 1 of the DIMO Merkle
// tree file format.
const TreeFileFormatV1 = "dimo-merkle-v1"

// TreeFile is the parsed form of a dimo-merkle-v1 tree file. It carries
// everything a consumer needs to verify the root and serve proofs.
type TreeFile struct {
	Format      string
	Distributor common.Address
	PoolID      *big.Int
	Week        *big.Int
	Root        common.Hash
	Leaves      []TreeFileLeaf
}

// TreeFileLeaf is a single claim entry in a tree file, together with its
// Merkle proof.
type TreeFileLeaf struct {
	Account common.Address
	Amount  *big.Int
	Proof   []common.Hash
}

type treeFileJSON struct {
	Format      string             `json:"format"`
	Distributor string             `json:"distributor"`
	PoolID      string             `json:"poolId"`
	Week        string             `json:"week"`
	Root        string             `json:"root"`
	Leaves      []treeFileLeafJSON `json:"leaves"`
}

type treeFileLeafJSON struct {
	Account string   `json:"account"`
	Amount  string   `json:"amount"`
	Proof   []string `json:"proof"`
}

// MarshalJSON encodes the tree file in the dimo-merkle-v1 format with all
// numeric values as decimal strings and addresses EIP-55 checksummed.
func (f *TreeFile) MarshalJSON() ([]byte, error) {
	out := treeFileJSON{
		Format:      f.Format,
		Distributor: f.Distributor.Hex(),
		PoolID:      f.PoolID.String(),
		Week:        f.Week.String(),
		Root:        f.Root.Hex(),
		Leaves:      make([]treeFileLeafJSON, len(f.Leaves)),
	}
	for i, leaf := range f.Leaves {
		proof := make([]string, len(leaf.Proof))
		for j, p := range leaf.Proof {
			proof[j] = p.Hex()
		}
		out.Leaves[i] = treeFileLeafJSON{
			Account: leaf.Account.Hex(),
			Amount:  leaf.Amount.String(),
			Proof:   proof,
		}
	}
	return json.Marshal(out)
}

// UnmarshalTreeFile parses and validates a dimo-merkle-v1 tree file. It does
// not verify the root; call VerifyRoot for that.
func UnmarshalTreeFile(data []byte) (*TreeFile, error) {
	var raw treeFileJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing tree file: %w", err)
	}
	if raw.Format != TreeFileFormatV1 {
		return nil, fmt.Errorf("unknown tree file format %q, expected %q", raw.Format, TreeFileFormatV1)
	}

	distributor, err := parseAddress(raw.Distributor)
	if err != nil {
		return nil, fmt.Errorf("parsing distributor: %w", err)
	}
	poolID, err := parseDecimal(raw.PoolID)
	if err != nil {
		return nil, fmt.Errorf("parsing pool id: %w", err)
	}
	if poolID.Sign() < 0 {
		return nil, fmt.Errorf("pool id must be non-negative, got %s", poolID)
	}
	week, err := parseDecimal(raw.Week)
	if err != nil {
		return nil, fmt.Errorf("parsing week: %w", err)
	}
	if week.Sign() < 0 {
		return nil, fmt.Errorf("week must be non-negative, got %s", week)
	}
	root, err := parseHash(raw.Root)
	if err != nil {
		return nil, fmt.Errorf("parsing root: %w", err)
	}

	file := &TreeFile{
		Format:      raw.Format,
		Distributor: distributor,
		PoolID:      poolID,
		Week:        week,
		Root:        root,
		Leaves:      make([]TreeFileLeaf, len(raw.Leaves)),
	}
	for i, rawLeaf := range raw.Leaves {
		account, err := parseAddress(rawLeaf.Account)
		if err != nil {
			return nil, fmt.Errorf("parsing account in leaf %d: %w", i, err)
		}
		amount, err := parseDecimal(rawLeaf.Amount)
		if err != nil {
			return nil, fmt.Errorf("parsing amount in leaf %d: %w", i, err)
		}
		if amount.Sign() <= 0 {
			return nil, fmt.Errorf("amount in leaf %d must be positive, got %s", i, amount)
		}
		proof := make([]common.Hash, len(rawLeaf.Proof))
		for j, p := range rawLeaf.Proof {
			proof[j], err = parseHash(p)
			if err != nil {
				return nil, fmt.Errorf("parsing proof element %d in leaf %d: %w", j, i, err)
			}
		}
		file.Leaves[i] = TreeFileLeaf{Account: account, Amount: amount, Proof: proof}
	}
	return file, nil
}

// VerifyRoot rebuilds the Merkle tree from the leaves in the file and returns
// an error if the recomputed root does not match the stored root.
func (f *TreeFile) VerifyRoot() error {
	leaves := make([]Leaf, len(f.Leaves))
	for i, leaf := range f.Leaves {
		leaves[i] = Leaf{Account: leaf.Account, Amount: leaf.Amount}
	}
	tree, err := New(f.Distributor, f.PoolID, f.Week, leaves)
	if err != nil {
		return fmt.Errorf("rebuilding tree from leaves: %w", err)
	}
	if root := tree.Root(); root != f.Root {
		return fmt.Errorf("recomputed root %s does not match stored root %s", root.Hex(), f.Root.Hex())
	}
	return nil
}

func parseAddress(s string) (common.Address, error) {
	if !common.IsHexAddress(s) {
		return common.Address{}, fmt.Errorf("invalid address %q", s)
	}
	return common.HexToAddress(s), nil
}

func parseDecimal(s string) (*big.Int, error) {
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("invalid decimal string %q", s)
	}
	return v, nil
}

func parseHash(s string) (common.Hash, error) {
	b, err := hexDecode32(s)
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(b), nil
}

func hexDecode32(s string) ([]byte, error) {
	if len(s) != 66 || s[:2] != "0x" {
		return nil, fmt.Errorf("expected 0x-prefixed 64-character hex string, got %q", s)
	}
	b, err := hex.DecodeString(s[2:])
	if err != nil {
		return nil, fmt.Errorf("decoding hex string %q: %w", s, err)
	}
	return b, nil
}

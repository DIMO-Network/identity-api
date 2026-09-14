package loader

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/merkle"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/graph-gophers/dataloader/v7"
)

// MerklePoolLoader batches lookups of Merkle pools by pool id, so resolving
// the pool field on a page of rewards costs one query instead of one per row.
type MerklePoolLoader struct {
	repo *merkle.Repository
}

func NewMerklePoolLoader(repo *merkle.Repository) *MerklePoolLoader {
	return &MerklePoolLoader{repo: repo}
}

// GetMerklePoolByID enqueues a pool lookup on the request's dataloader and
// returns the result. It returns nil without error if the pool does not
// exist.
func GetMerklePoolByID(ctx context.Context, poolID int) (*model.MerklePool, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	// invoke and get thunk
	thunk := loaders.MerklePoolByID.Load(ctx, poolID)
	// read value from thunk
	return thunk()
}

func (m *MerklePoolLoader) BatchGetMerklePoolsByID(ctx context.Context, poolIDs []int) []*dataloader.Result[*model.MerklePool] {
	results := make([]*dataloader.Result[*model.MerklePool], len(poolIDs))

	pools, err := models.MerklePools(models.MerklePoolWhere.PoolID.IN(poolIDs)).All(ctx, m.repo.PDB.DBS().Reader)
	if err != nil {
		for i := range poolIDs {
			results[i] = &dataloader.Result[*model.MerklePool]{Error: err}
		}
		return results
	}

	poolByID := make(map[int]*model.MerklePool, len(pools))
	for _, p := range pools {
		poolByID[p.PoolID] = merkle.PoolToAPI(p)
	}

	for i, id := range poolIDs {
		if pool, ok := poolByID[id]; ok {
			results[i] = &dataloader.Result[*model.MerklePool]{Data: pool}
		} else {
			results[i] = &dataloader.Result[*model.MerklePool]{}
		}
	}

	return results
}

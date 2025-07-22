package loader

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/storagenode"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/lib/pq"
)

type StorageNodeLoader struct {
	repo *storagenode.Repository
}

func NewStorageNodeLoader(repo *storagenode.Repository) *StorageNodeLoader {
	return &StorageNodeLoader{
		repo: repo,
	}
}

// GetStorageNodeByID uses the DataLoader pattern to retrieve information
// about a specific storage node. Here "id" is a bytes32, equal to
// keccak256(bytes(label)).
func GetStorageNodeByID(ctx context.Context, nodeID []byte) (*model.StorageNode, error) {
	if len(nodeID) != common.HashLength {
		return nil, fmt.Errorf("storage node id had unexpected length %d", len(nodeID))
	}

	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	thunk := loaders.StorageNodeByID.Load(ctx, [32]byte(nodeID))
	return thunk()
}

func (s *StorageNodeLoader) BatchGetStorageNodesByIDs(ctx context.Context, storageNodeIDs [][32]byte) []*dataloader.Result[*model.StorageNode] {
	results := make([]*dataloader.Result[*model.StorageNode], len(storageNodeIDs))

	nodeIDSlices := make([][]byte, len(storageNodeIDs))
	for i := range storageNodeIDs {
		nodeIDSlices[i] = storageNodeIDs[i][:]
	}

	nodes, err := models.StorageNodes(
		qm.Where(models.StorageNodeColumns.ID+" = ANY(?)", pq.ByteaArray(nodeIDSlices)),
	).All(ctx, s.repo.PDB.DBS().Reader)
	if err != nil {
		for i := range results {
			results[i] = &dataloader.Result[*model.StorageNode]{
				Error: err,
			}
		}
		return results
	}

	nodesByID := make(map[[32]byte]*models.StorageNode)
	for _, sn := range nodes {
		nodesByID[[32]byte(sn.ID)] = sn
	}

	for i, snID := range storageNodeIDs {
		sn, ok := nodesByID[snID]
		if ok {
			results[i] = &dataloader.Result[*model.StorageNode]{
				Data: s.repo.ToAPI(sn),
			}
		} else {
			// TODO(elffjs): This is what we're doing elsewhere, but
			// should it contain an error, instead?
			results[i] = &dataloader.Result[*model.StorageNode]{}
		}
	}

	return results
}

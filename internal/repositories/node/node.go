package node

import (
	"math/big"

	"github.com/DIMO-Network/cloudevent"
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
)

type Repository struct {
	*base.Repository
	chainID         uint64
	contractAddress common.Address
}

// New creates a new connection repository.
func New(baseRepo *base.Repository) *Repository {
	return &Repository{
		Repository:      baseRepo,
		chainID:         uint64(baseRepo.Settings.DIMORegistryChainID),
		contractAddress: common.HexToAddress(baseRepo.Settings.ConnectionAddr),
	}
}

func (r *Repository) ToAPI(v *models.StorageNode) *gmodel.StorageNode {
	tokenID := new(big.Int).SetBytes(v.ID)

	tokenDID := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.contractAddress,
		TokenID:         tokenID,
	}.String()

	return &gmodel.StorageNode{
		TokenID:  tokenID,
		TokenDID: tokenDID,
		Label:    v.Label,
		Owner:    common.BytesToAddress(v.Owner),
		Address:  common.BytesToAddress(v.Address),
		URI:      v.URI,
		MintedAt: v.MintedAt,
	}
}

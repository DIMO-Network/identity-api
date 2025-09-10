package template

import (
	"math/big"

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

// New creates a new template repository.
func New(db *base.Repository) *Repository {
	return &Repository{
		Repository:      db,
		chainID:         uint64(db.Settings.DIMORegistryChainID),
		contractAddress: common.HexToAddress(db.Settings.ConnectionAddr),
	}
}

func (r *Repository) ToAPI(template *models.Template) *gmodel.Template {
	tokenID := new(big.Int).SetBytes(template.ID)

	return &gmodel.Template{
		ID:          tokenID,
		Creator:     common.BytesToAddress(template.Creator),
		Asset:       common.BytesToAddress(template.Asset),
		Permissions: template.Permissions,
		Cid:         template.Cid,
		CreatedAt:   template.CreatedAt,
	}
}

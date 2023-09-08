package repositories

import (
	"context"
	"encoding/json"
	"log"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
)

func DCNToAPI(d *models.DCN) *gmodel.Dcn {
	return &gmodel.Dcn{
		Owner: common.BytesToAddress(d.OwnerAddress),
		Node:  d.Node,
	}
}

func (r *Repository) GetDCNByNode(ctx context.Context, node []byte) (*gmodel.Dcn, error) {
	var n []byte
	if err := json.Unmarshal(node, &n); err != nil {
		return nil, err
	}

	dcn, err := models.DCNS(
		models.DCNWhere.Node.EQ(n),
	).One(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	log.Println(dcn)

	return DCNToAPI(dcn), nil
}

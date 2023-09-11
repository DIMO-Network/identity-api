package repositories

import (
	"context"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
)

func DCNToAPI(d *models.DCN) *gmodel.Dcn {
	return &gmodel.Dcn{
		Owner:     common.BytesToAddress(d.OwnerAddress),
		Node:      d.Node,
		ExpiresAt: d.Expiration.Ptr(),
	}
}

func (r *Repository) GetDCNByNode(ctx context.Context, node []byte) (*gmodel.Dcn, error) {
	dcn, err := models.DCNS(models.DCNWhere.Node.EQ(node)).One(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return DCNToAPI(dcn), nil
}

package repositories

import (
	"context"
	"errors"

	"github.com/DIMO-Network/identity-api/graph/model"
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
)

func DCNToAPI(d *models.DCN) *gmodel.Dcn {
	return &gmodel.Dcn{
		Owner:     common.BytesToAddress(d.OwnerAddress),
		Node:      d.Node,
		ExpiresAt: d.Expiration.Ptr(),
		Name:      d.Name.Ptr(),
		VehicleID: d.VehicleID.Ptr(),
	}
}

func (r *Repository) getDCNByNode(ctx context.Context, node []byte) (*gmodel.Dcn, error) {
	dcn, err := models.DCNS(
		models.DCNWhere.Node.EQ(node),
	).One(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return DCNToAPI(dcn), nil
}

func (r *Repository) getDCNByName(ctx context.Context, name string) (*gmodel.Dcn, error) {
	dcn, err := models.DCNS(
		models.DCNWhere.Name.EQ(null.StringFrom(name)),
	).One(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return DCNToAPI(dcn), nil
}

func (r *Repository) GetDCN(ctx context.Context, params model.DCNBy) (*gmodel.Dcn, error) {
	if params.Name != nil && len(params.Node) > 0 {
		return nil, errors.New("provide one of Name or Node but not both")
	}

	if params.Name == nil && len(params.Node) == 0 {
		return nil, errors.New("provide either Name or Node")
	}

	if params.Name != nil {
		return r.getDCNByName(ctx, *params.Name)
	}

	return r.getDCNByNode(ctx, params.Node)
}

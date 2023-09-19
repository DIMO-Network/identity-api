package repositories

import (
	"context"
	"fmt"

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

func (r *Repository) getDCNByNode(ctx context.Context, dcnParams model.DCNBy) (*gmodel.Dcn, error) {
	dcn, err := models.DCNS(
		models.DCNWhere.Node.EQ(dcnParams.Node),
	).One(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return DCNToAPI(dcn), nil
}

func (r *Repository) getDCNByName(ctx context.Context, dcnParams model.DCNBy) (*gmodel.Dcn, error) {
	dcn, err := models.DCNS(
		models.DCNWhere.Name.EQ(null.StringFrom(*dcnParams.Name)),
	).One(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return DCNToAPI(dcn), nil
}

func (r *Repository) GetDCN(ctx context.Context, params model.DCNBy) (*gmodel.Dcn, error) {
	if params.Name != nil && len(params.Node) > 0 {
		return nil, fmt.Errorf("please provide one of Name or Node but not both")
	}

	if params.Name == nil && len(params.Node) == 0 {
		return nil, fmt.Errorf("please provide either Name or Node")
	}

	if params.Name != nil {
		return r.getDCNByName(ctx, params)
	}

	return r.getDCNByNode(ctx, params)
}

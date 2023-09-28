package repositories

import (
	"context"
	"errors"

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
		MintedAt:  d.MintedAt,
	}
}

func (r *Repository) GetDCN(ctx context.Context, by gmodel.DCNBy) (*gmodel.Dcn, error) {
	if countTrue(len(by.Node) != 0, by.Name != nil) != 1 {
		return nil, errors.New("Provide exactly one of `name` or `node`.")
	}

	if len(by.Node) != 0 {
		return r.GetDCNByNode(ctx, by.Node)
	}

	return r.GetDCNByName(ctx, *by.Name)
}

func (r *Repository) GetDCNByNode(ctx context.Context, node []byte) (*gmodel.Dcn, error) {
	if len(node) != common.HashLength {
		return nil, errors.New("`node` must be 32 bytes long")
	}

	dcn, err := models.DCNS(
		models.DCNWhere.Node.EQ(node),
	).One(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return DCNToAPI(dcn), nil
}

func (r *Repository) GetDCNByName(ctx context.Context, name string) (*gmodel.Dcn, error) {
	dcn, err := models.DCNS(
		models.DCNWhere.Name.EQ(null.StringFrom(name)),
	).One(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return DCNToAPI(dcn), nil
}

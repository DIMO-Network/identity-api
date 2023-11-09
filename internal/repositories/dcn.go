package repositories

import (
	"context"
	"errors"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
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
		return nil, gqlerror.Errorf("Provide exactly one of `name` or `node`.")
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

func (r *Repository) GetDCNs(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *gmodel.DCNFilter) (*gmodel.DCNConnection, error) {
	var limit int
	limit, err := helpers.ValidateFirstLast(first, last, maxPageSize)
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{}
	if filterBy != nil && filterBy.Owner != nil {
		queryMods = append(queryMods, models.DCNWhere.OwnerAddress.EQ(filterBy.Owner.Bytes()))
	}

	dcnCount, err := models.DCNS(queryMods...).Count(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	orderBy := " DESC"
	if last != nil {
		orderBy = " ASC"
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(models.DCNColumns.MintedAt+orderBy))

	pHelp := &helpers.PaginationHelper[time.Time]{}
	if after != nil {
		afterT, err := pHelp.DecodeCursor(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.DCNWhere.MintedAt.LT(*afterT))
	} else if before != nil {
		beforeT, err := pHelp.DecodeCursor(*before)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.DCNWhere.MintedAt.GT(*beforeT))
	}

	all, err := models.DCNS(queryMods...).All(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	hasNext := before != nil
	hasPrevious := after != nil

	if first != nil && len(all) == limit+1 {
		hasNext = true
		all = all[:limit]
	} else if last != nil && len(all) == limit+1 {
		hasPrevious = true
		all = all[:limit]
	}

	if last != nil {
		slices.Reverse(all)
	}

	edges := make([]*gmodel.DCNEdge, len(all))
	nodes := make([]*gmodel.Dcn, len(all))

	for i, dcn := range all {
		c, err := pHelp.EncodeCursor(dcn.MintedAt)
		if err != nil {
			return nil, err
		}
		edges[i] = &gmodel.DCNEdge{
			Node: &gmodel.Dcn{
				Node:      dcn.Node,
				Owner:     common.Address(dcn.OwnerAddress),
				ExpiresAt: &dcn.Expiration.Time,
				MintedAt:  dcn.MintedAt,
				Name:      &dcn.Name.String,
				VehicleID: &dcn.VehicleID.Int,
			},
			Cursor: c,
		}
		nodes[i] = edges[i].Node
	}

	var endCur, startCur *string

	if len(all) != 0 {
		ec, err := pHelp.EncodeCursor(all[len(all)-1].MintedAt)
		if err != nil {
			return nil, err
		}
		endCur = &ec

		sc, err := pHelp.EncodeCursor(all[0].MintedAt)
		if err != nil {
			return nil, err
		}
		startCur = &sc
	}

	res := &gmodel.DCNConnection{
		TotalCount: int(dcnCount),
		Edges:      edges,
		Nodes:      nodes,
		PageInfo: &gmodel.PageInfo{
			StartCursor:     startCur,
			EndCursor:       endCur,
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
		},
	}

	return res, nil
}

package dcn

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// TokenPrefix is the prefix for a global token id for DCNs.
const TokenPrefix = "D"

type Repository struct {
	*base.Repository
}

type DCNCursor struct {
	MintedAt time.Time
	Node     []byte
}

// ToAPI converts a DCN database row to a DCN API model.
func ToAPI(d *models.DCN) (*gmodel.Dcn, error) {
	tokenID := new(big.Int).SetBytes(d.Node)
	globalID, err := base.EncodeGlobalTokenID(TokenPrefix, int(tokenID.Int64()))
	if err != nil {
		return nil, fmt.Errorf("error encoding dcn id: %w", err)
	}
	return &gmodel.Dcn{
		ID:        globalID,
		Owner:     common.BytesToAddress(d.OwnerAddress),
		TokenID:   tokenID,
		Node:      d.Node,
		ExpiresAt: d.Expiration.Ptr(),
		Name:      d.Name.Ptr(),
		VehicleID: d.VehicleID.Ptr(),
		MintedAt:  d.MintedAt,
	}, nil
}

func (r *Repository) GetDCN(ctx context.Context, by gmodel.DCNBy) (*gmodel.Dcn, error) {
	if base.CountTrue(len(by.Node) != 0, by.Name != nil) != 1 {
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
	).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return ToAPI(dcn)
}

func (r *Repository) GetDCNByName(ctx context.Context, name string) (*gmodel.Dcn, error) {
	dcn, err := models.DCNS(
		models.DCNWhere.Name.EQ(null.StringFrom(name)),
	).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return ToAPI(dcn)
}

var dcnCursorColumnsTuple = "(" + models.DCNColumns.MintedAt + ", " + models.DCNColumns.Node + ")"

func (r *Repository) GetDCNs(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *gmodel.DCNFilter) (*gmodel.DCNConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{}
	if filterBy != nil && filterBy.Owner != nil {
		queryMods = append(queryMods, models.DCNWhere.OwnerAddress.EQ(filterBy.Owner.Bytes()))
	}

	dcnCount, err := models.DCNS(queryMods...).Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	orderBy := " DESC"
	if last != nil {
		orderBy = " ASC"
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(models.DCNColumns.MintedAt+orderBy+", "+models.DCNColumns.Node+orderBy),
	)

	pHelp := &helpers.PaginationHelper[DCNCursor]{}
	if after != nil {
		afterT, err := pHelp.DecodeCursor(*after)
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods,
			qm.Where(dcnCursorColumnsTuple+" < (?, ?)", afterT.MintedAt, afterT.Node),
		)
	} else if before != nil {
		beforeT, err := pHelp.DecodeCursor(*before)
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods,
			qm.Where(dcnCursorColumnsTuple+" < (?, ?)", beforeT.MintedAt, beforeT.Node),
		)
	}

	all, err := models.DCNS(queryMods...).All(ctx, r.PDB.DBS().Reader)
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
	var errList gqlerror.List
	for i, dcn := range all {
		c, err := pHelp.EncodeCursor(DCNCursor{MintedAt: dcn.MintedAt, Node: dcn.Node})
		if err != nil {
			return nil, err
		}
		apiDCN, err := ToAPI(dcn)
		if err != nil {
			errList = append(errList, gqlerror.Errorf("error converting dcn to api: %v", err))
			continue
		}
		edges[i] = &gmodel.DCNEdge{
			Node:   apiDCN,
			Cursor: c,
		}
		nodes[i] = edges[i].Node
	}

	var endCur, startCur *string

	if len(all) != 0 {
		startCur = &edges[0].Cursor
		endCur = &edges[len(edges)-1].Cursor
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
	if errList != nil {
		return res, errList
	}
	return res, nil
}

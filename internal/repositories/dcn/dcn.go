package dcn

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"time"

	"github.com/DIMO-Network/cloudevent"
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// TokenPrefix is the prefix for a global token id for DCNs.
const TokenPrefix = "D"

type Repository struct {
	*base.Repository
	chainID         uint64
	contractAddress common.Address
}

// New creates a new DCN repository.
func New(db *base.Repository) *Repository {
	return &Repository{
		Repository:      db,
		chainID:         uint64(db.Settings.DIMORegistryChainID),
		contractAddress: common.HexToAddress(db.Settings.DCNRegistryAddr),
	}
}

type DCNCursor struct {
	MintedAt time.Time
	Node     []byte
}

// ToAPI converts a DCN database row to a DCN API model.
func (r *Repository) ToAPI(d *models.DCN) (*gmodel.Dcn, error) {
	tokenID := new(big.Int).SetBytes(d.Node)
	globalID, err := base.EncodeGlobalTokenID(TokenPrefix, int(tokenID.Int64()))
	if err != nil {
		return nil, fmt.Errorf("error encoding dcn id: %w", err)
	}

	tokenDID := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.contractAddress,
		TokenID:         tokenID,
	}.String()

	return &gmodel.Dcn{
		ID:        globalID,
		Owner:     common.BytesToAddress(d.OwnerAddress),
		TokenID:   tokenID,
		TokenDID:  tokenDID,
		Node:      d.Node,
		ExpiresAt: d.Expiration.Ptr(),
		Name:      d.Name.Ptr(),
		VehicleID: d.VehicleID.Ptr(),
		MintedAt:  d.MintedAt,
	}, nil
}

func (r *Repository) GetDCN(ctx context.Context, by gmodel.DCNBy) (*gmodel.Dcn, error) {
	if base.CountTrue(len(by.Node) != 0, by.Name != nil, by.TokenDID != nil) != 1 {
		return nil, gqlerror.Errorf("Provide exactly one of `name`, `node`, or `tokenDID`.")
	}

	switch {
	case len(by.Node) != 0:
		return r.GetDCNByNode(ctx, by.Node)
	case by.Name != nil:
		return r.GetDCNByName(ctx, *by.Name)
	case by.TokenDID != nil:
		did, err := cloudevent.DecodeERC721DID(*by.TokenDID)
		if err != nil {
			return nil, fmt.Errorf("error decoding token did: %w", err)
		}
		if did.ChainID != r.chainID {
			return nil, fmt.Errorf("unknown chain id %d in token did", did.ChainID)
		}
		if did.ContractAddress != r.contractAddress {
			return nil, fmt.Errorf("invalid contract address '%s' in token did", did.ContractAddress.Hex())
		}
		id, err := helpers.ConvertTokenIDToID(did.TokenID)
		if err != nil {
			return nil, fmt.Errorf("error converting token id to id: %w", err)
		}
		return r.GetDCNByNode(ctx, id)
	default:
		return nil, fmt.Errorf("invalid filter")
	}
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

	return r.ToAPI(dcn)
}

func (r *Repository) GetDCNByName(ctx context.Context, name string) (*gmodel.Dcn, error) {
	dcn, err := models.DCNS(
		models.DCNWhere.Name.EQ(null.StringFrom(name)),
	).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return r.ToAPI(dcn)
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
		apiDCN, err := r.ToAPI(dcn)
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

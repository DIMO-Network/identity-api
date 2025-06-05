package connection

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"time"

	"github.com/DIMO-Network/cloudevent"
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Repository struct {
	*base.Repository
	chainID         uint64
	contractAddress common.Address
}

// New creates a new connection repository.
func New(db *base.Repository) *Repository {
	return &Repository{
		Repository:      db,
		chainID:         uint64(db.Settings.DIMORegistryChainID),
		contractAddress: common.HexToAddress(db.Settings.ConnectionAddr),
	}
}

type Cursor struct {
	MintedAt time.Time
	Address  []byte
}

var pageHelper = helpers.PaginationHelper[Cursor]{}

func (r *Repository) ToAPI(v *models.Connection) *gmodel.Connection {
	name := string(bytes.TrimRight(v.ID, "\x00"))

	tokenID := new(big.Int).SetBytes(v.ID)

	tokenDID := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.contractAddress,
		TokenID:         tokenID,
	}.String()

	return &gmodel.Connection{
		Name:     name,
		Address:  common.BytesToAddress(v.Address),
		Owner:    common.BytesToAddress(v.Owner),
		TokenID:  tokenID,
		TokenDID: tokenDID,
		MintedAt: v.MintedAt,
	}
}

func (r *Repository) GetConnections(ctx context.Context, first *int, after *string, last *int, before *string) (*gmodel.ConnectionConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	// None now, but making room.
	var queryMods []qm.QueryMod

	totalCount, err := models.Connections().Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if after != nil {
		afterCur, err := pageHelper.DecodeCursor(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, qm.Expr(
			models.ConnectionWhere.MintedAt.EQ(afterCur.MintedAt),
			models.ConnectionWhere.Address.GT(afterCur.Address),
			qm.Or2(models.ConnectionWhere.MintedAt.LT(afterCur.MintedAt)),
		))
	}

	if before != nil {
		beforeCur, err := pageHelper.DecodeCursor(*before)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, qm.Expr(
			models.ConnectionWhere.MintedAt.EQ(beforeCur.MintedAt),
			models.ConnectionWhere.Address.LT(beforeCur.Address),
			qm.Or2(models.ConnectionWhere.MintedAt.GT(beforeCur.MintedAt)),
		))
	}

	orderBy := fmt.Sprintf("%s DESC, %s ASC", models.ConnectionColumns.MintedAt, models.ConnectionColumns.Address)
	if last != nil {
		orderBy = fmt.Sprintf("%s ASC, %s DESC", models.ConnectionColumns.MintedAt, models.ConnectionColumns.Address)
	}

	queryMods = append(queryMods,
		// Use limit + 1 here to check if there's another page.
		qm.Limit(limit+1),
		qm.OrderBy(orderBy),
	)

	all, err := models.Connections(queryMods...).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	// We assume that cursors come from real elements.
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

	nodes := make([]*gmodel.Connection, len(all))
	edges := make([]*gmodel.ConnectionEdge, len(all))

	for i, dv := range all {
		dlv := r.ToAPI(dv)

		nodes[i] = dlv

		crsr, err := pageHelper.EncodeCursor(Cursor{
			MintedAt: dv.MintedAt,
			Address:  dv.Address,
		})
		if err != nil {
			return nil, err
		}

		edges[i] = &gmodel.ConnectionEdge{
			Node:   dlv,
			Cursor: crsr,
		}
	}

	var endCur, startCur *string

	if len(edges) != 0 {
		startCur = &edges[0].Cursor
		endCur = &edges[len(edges)-1].Cursor
	}

	res := &gmodel.ConnectionConnection{
		Edges: edges,
		Nodes: nodes,
		PageInfo: &gmodel.PageInfo{
			EndCursor:       endCur,
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCur,
		},
		TotalCount: int(totalCount),
	}

	return res, nil
}

func (r *Repository) GetConnection(ctx context.Context, by gmodel.ConnectionBy) (*gmodel.Connection, error) {
	if base.CountTrue(by.Name != nil, by.Address != nil, by.TokenID != nil, by.TokenDID != nil) != 1 {
		return nil, gqlerror.Errorf("must specify exactly one of `name`, `address`, `tokenId`, or `tokenDID`")
	}

	var mod qm.QueryMod

	switch {
	case by.Name != nil:
		nbs := []byte(*by.Name)
		if len(nbs) > 32 {
			return nil, repositories.ErrNotFound
		}

		nb32 := make([]byte, 32)
		copy(nb32, nbs)
		mod = models.ConnectionWhere.ID.EQ(nb32)
	case by.Address != nil:
		mod = models.ConnectionWhere.Address.EQ(by.Address.Bytes())
	case by.TokenID != nil:
		id, err := helpers.ConvertTokenIDToID(by.TokenID)
		if err != nil {
			return nil, err
		}

		mod = models.ConnectionWhere.ID.EQ(id)
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
			return nil, err
		}

		mod = models.ConnectionWhere.ID.EQ(id)
	}

	dl, err := models.Connections(mod).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}

	return r.ToAPI(dl), nil
}

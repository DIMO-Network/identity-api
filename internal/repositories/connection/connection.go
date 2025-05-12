package connection

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Repository struct {
	*base.Repository
}

type Cursor struct {
	MintedAt time.Time
	Address  []byte
}

var pageHelper = helpers.PaginationHelper[Cursor]{}

func (r *Repository) ToGQL(v *models.Connection) (*gmodel.Connection, error) {
	return &gmodel.Connection{
		Name:     v.Name,
		Address:  common.BytesToAddress(v.Address),
		Owner:    common.BytesToAddress(v.Owner),
		MintedAt: v.MintedAt,
	}, nil
}

func (r *Repository) ToCursor(v *models.Connection) (Cursor, error) {
	return Cursor{
		MintedAt: v.MintedAt,
		Address:  v.Address,
	}, nil
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

	nodes, cursors, pi, err := ConvertConnection(all, first, after, last, before, limit, r)
	if err != nil {
		return nil, err
	}

	edges := make([]*gmodel.ConnectionEdge, len(nodes))
	for i := range nodes {
		edges[i] = &gmodel.ConnectionEdge{
			Node:   nodes[i],
			Cursor: cursors[i],
		}
	}

	res := &gmodel.ConnectionConnection{
		Edges:      edges,
		Nodes:      nodes,
		PageInfo:   pi,
		TotalCount: int(totalCount),
	}

	return res, nil
}

type ConnectionConverter[DBModel, GQLModel, Cursor any] interface {
	ToGQL(DBModel) (GQLModel, error)
	ToCursor(DBModel) (Cursor, error)
}

func ConvertConnection[DBModel, GQLModel, Cursor any](dbs []DBModel, first *int, after *string, last *int, before *string, limit int, cc ConnectionConverter[DBModel, GQLModel, Cursor]) ([]GQLModel, []string, *gmodel.PageInfo, error) {
	hasNext := before != nil
	hasPrevious := after != nil

	if first != nil && len(dbs) == limit+1 {
		hasNext = true
		dbs = dbs[:limit]
	} else if last != nil && len(dbs) == limit+1 {
		hasPrevious = true
		dbs = dbs[:limit]
	}

	if last != nil {
		slices.Reverse(dbs)
	}

	out := make([]GQLModel, len(dbs))
	outCur := make([]string, len(dbs))

	for i, mod := range dbs {
		outMod, err := cc.ToGQL(mod)
		if err != nil {
			return nil, nil, nil, err
		}

		out[i] = outMod

		cur, err := cc.ToCursor(mod)
		if err != nil {
			return nil, nil, nil, err
		}

		curStr, err := EncodeCursor(cur)
		if err != nil {
			return nil, nil, nil, err
		}

		outCur[i] = curStr
	}

	var startCur, endCur *string
	if len(out) != 0 {
		startCur = &outCur[0]
		endCur = &outCur[len(out)-1]
	}

	pi := &gmodel.PageInfo{
		StartCursor:     startCur,
		EndCursor:       endCur,
		HasPreviousPage: hasPrevious,
		HasNextPage:     hasNext,
	}

	return out, outCur, pi, nil
}

func EncodeCursor[T any](cursor T) (string, error) {
	var b bytes.Buffer
	e := msgpack.NewEncoder(&b)
	e.UseArrayEncodedStructs(true)

	if err := e.Encode(cursor); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

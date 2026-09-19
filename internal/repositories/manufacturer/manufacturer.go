package manufacturer

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"

	"github.com/DIMO-Network/cloudevent"
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vmihailenco/msgpack/v5"
)

// TokenPrefix is the prfix for the global token id for manufacturers.
const TokenPrefix = "M"

type Repository struct {
	*base.Repository
	chainID         uint64
	contractAddress common.Address
}

// New creates a new manufacturer repository.
func New(db *base.Repository) *Repository {
	return &Repository{
		Repository:      db,
		chainID:         uint64(db.Settings.DIMORegistryChainID),
		contractAddress: common.HexToAddress(db.Settings.ManufacturerNFTAddr),
	}
}

type manufacturerPrimaryKey struct {
	TokenID int
}

// ToAPI converts a manufacturer to a corresponding graphql model.
func (r *Repository) ToAPI(m *models.Manufacturer) (*gmodel.Manufacturer, error) {
	globalID, err := base.EncodeGlobalTokenID(TokenPrefix, m.ID)
	if err != nil {
		return nil, fmt.Errorf("error encoding manufacturer id: %w", err)
	}

	tokenDID := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.contractAddress,
		TokenID:         new(big.Int).SetUint64(uint64(m.ID)),
	}.String()

	return &gmodel.Manufacturer{
		ID:       globalID,
		TokenID:  m.ID,
		TokenDID: tokenDID,
		Owner:    common.BytesToAddress(m.Owner),
		TableID:  m.TableID.Ptr(),
		MintedAt: m.MintedAt,
		Name:     m.Name,
	}, nil
}

// IDToToken converts token data to a token id.
func IDToToken(b []byte) (int, error) {
	var pk manufacturerPrimaryKey
	d := msgpack.NewDecoder(bytes.NewBuffer(b))
	if err := d.Decode(&pk); err != nil {
		return 0, fmt.Errorf("error decoding manufacturer id: %w", err)
	}

	return pk.TokenID, nil
}

func (r *Repository) GetManufacturer(ctx context.Context, by gmodel.ManufacturerBy) (*gmodel.Manufacturer, error) {
	if base.CountTrue(by.TokenID != nil, by.Name != nil, by.Slug != nil, by.TokenDID != nil) != 1 {
		return nil, gqlerror.Errorf("Provide exactly one of `name`, `tokenID`, `slug`, or `tokenDID`.")
	}

	var qm qm.QueryMod
	switch {
	case by.TokenID != nil:
		qm = models.ManufacturerWhere.ID.EQ(*by.TokenID)
	case by.Name != nil:
		qm = models.ManufacturerWhere.Name.EQ(*by.Name)
	case by.Slug != nil:
		qm = models.ManufacturerWhere.Slug.EQ(*by.Slug)
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
		if !did.TokenID.IsInt64() {
			return nil, fmt.Errorf("token id is too large")
		}
		qm = models.ManufacturerWhere.ID.EQ(int(did.TokenID.Int64()))
	default:
		return nil, fmt.Errorf("invalid filter")
	}

	m, err := manufacturerOrAbsent(models.Manufacturers(qm).One(ctx, r.PDB.DBS().Reader))
	if err != nil {
		// Whatever the database said stays in the log. The client gets
		// "Internal error", never a driver string it cannot act on.
		r.Log.Error().Err(err).Msg("Failed to look up manufacturer.")
		return nil, base.InternalError
	}
	if m == nil {
		// The contract for an unknown make: `data.manufacturer` is null and
		// there is no `errors` entry. The schema's
		// `manufacturer(by:): Manufacturer` is nullable, so absence is a
		// representable answer, and identity has no error presenter -- an
		// error here would reach clients as the raw `sql: no rows in result
		// set`, which a client that reads any errors entry as an outage
		// cannot tell from the service being down. Callers distinguish
		// "no such manufacturer" from "identity is broken" by the presence
		// of an errors entry, not by its text.
		return nil, nil
	}

	return r.ToAPI(m)
}

// manufacturerOrAbsent maps a single-row manufacturer lookup onto what the API
// serves for it: a row, or absence. The lookups here are all by a client-
// supplied identifier, so "nothing matched" is an ordinary answer and only a
// real failure is an error.
func manufacturerOrAbsent(m *models.Manufacturer, err error) (*models.Manufacturer, error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return m, nil
}

// SlugsByTokenID returns every manufacturer's slug keyed by token id. The
// device definitions catalog is checked against it: the catalog carries no
// chain marker, and the same slug has different token ids on different chains.
func (r *Repository) SlugsByTokenID(ctx context.Context) (map[int]string, error) {
	ms, err := models.Manufacturers(qm.Select(models.ManufacturerColumns.ID, models.ManufacturerColumns.Slug)).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, fmt.Errorf("error reading manufacturers: %w", err)
	}
	slugs := make(map[int]string, len(ms))
	for _, m := range ms {
		slugs[m.ID] = m.Slug
	}
	return slugs, nil
}

func (r *Repository) GetManufacturers(ctx context.Context) (*gmodel.ManufacturerConnection, error) {
	ms, err := models.Manufacturers().All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}
	res := &gmodel.ManufacturerConnection{}
	res.TotalCount = len(ms)
	res.PageInfo = &gmodel.PageInfo{
		StartCursor:     nil,
		EndCursor:       nil,
		HasPreviousPage: false,
		HasNextPage:     false,
	}
	res.Nodes = make([]*gmodel.Manufacturer, len(ms))
	res.Edges = make([]*gmodel.ManufacturerEdge, len(ms))
	for i, m := range ms {
		ma, err := r.ToAPI(m)
		if err != nil {
			return nil, err
		}
		res.Nodes[i] = ma

		res.Edges[i] = &gmodel.ManufacturerEdge{
			Node: ma,
		}
	}
	return res, nil
}

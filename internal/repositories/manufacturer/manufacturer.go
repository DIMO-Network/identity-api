package manufacturer

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/DIMO-Network/cloudevent"
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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

	m, err := models.Manufacturers(qm).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return r.ToAPI(m)
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

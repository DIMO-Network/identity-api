package manufacturer

import (
	"bytes"
	"context"
	"fmt"

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
}

type manufacturerPrimaryKey struct {
	TokenID int
}

// ToAPI converts a manufacturer to a corresponding graphql model.
func ToAPI(m *models.Manufacturer) (*gmodel.Manufacturer, error) {
	globalID, err := base.EncodeGlobalTokenID(TokenPrefix, m.ID)
	if err != nil {
		return nil, fmt.Errorf("error encoding manufacturer id: %w", err)
	}

	return &gmodel.Manufacturer{
		ID:       globalID,
		TokenID:  m.ID,
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
	if base.CountTrue(by.TokenID != nil, by.Name != nil) != 1 {
		return nil, gqlerror.Errorf("Provide exactly one of `name` or `tokenID`.")
	}

	var qm qm.QueryMod
	switch {
	case by.TokenID != nil:
		qm = models.ManufacturerWhere.ID.EQ(*by.TokenID)
	case by.Name != nil:
		qm = models.ManufacturerWhere.Name.EQ(*by.Name)
	case by.Slug != nil:
		qm = models.ManufacturerWhere.Slug.EQ(*by.Slug)
	}

	m, err := models.Manufacturers(qm).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return ToAPI(m)
}

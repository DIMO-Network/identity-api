package manufacturer

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Repository struct {
	*repositories.Repository
}

type manufacturerPrimaryKey struct {
	TokenID int
}

func ManufacturerToAPI(m *models.Manufacturer) *gmodel.Manufacturer {
	var b bytes.Buffer
	e := msgpack.NewEncoder(&b)
	e.UseArrayEncodedStructs(true)

	_ = e.Encode(manufacturerPrimaryKey{TokenID: m.ID})

	return &gmodel.Manufacturer{
		ID: "M_" + base64.StdEncoding.EncodeToString(b.Bytes()), TokenID: m.ID,
		Owner:    common.BytesToAddress(m.Owner),
		MintedAt: m.MintedAt,
		Name:     m.Name,
	}
}

func ManufacturerIDToToken(id string) (int, error) {
	if !strings.HasPrefix(id, "M_") {
		return 0, errors.New("id lacks the M_ prefix")
	}

	id = strings.TrimPrefix(id, "M_")

	b, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return 0, err
	}

	var pk manufacturerPrimaryKey
	d := msgpack.NewDecoder(bytes.NewBuffer(b))
	if err := d.Decode(&pk); err != nil {
		return 0, fmt.Errorf("error decoding manufacturer id: %w", err)
	}

	return pk.TokenID, nil
}

func (r *Repository) GetManufacturer(ctx context.Context, by gmodel.ManufacturerBy) (*gmodel.Manufacturer, error) {
	if repositories.CountTrue(by.TokenID != nil, by.Name != nil) != 1 {
		return nil, gqlerror.Errorf("Provide exactly one of `name` or `tokenID`.")
	}

	var qm qm.QueryMod
	switch {
	case by.TokenID != nil:
		qm = models.ManufacturerWhere.ID.EQ(*by.TokenID)
	case by.Name != nil:
		qm = models.ManufacturerWhere.Name.EQ(*by.Name)
	}

	m, err := models.Manufacturers(qm).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	res := ManufacturerToAPI(m)
	return res, nil
}

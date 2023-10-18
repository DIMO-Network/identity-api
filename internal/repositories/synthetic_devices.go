package repositories

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vmihailenco/msgpack/v5"
)

func SyntheticDeviceToAPI(sd *models.SyntheticDevice) *gmodel.SyntheticDevice {
	return &gmodel.SyntheticDevice{
		TokenID:       sd.ID,
		IntegrationID: sd.IntegrationID,
		Address:       common.BytesToAddress(sd.DeviceAddress),
		MintedAt:      sd.MintedAt,
	}
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

func (r *Repository) GetManufacturer(ctx context.Context, id int) (*gmodel.Manufacturer, error) {
	m, err := models.FindManufacturer(ctx, r.pdb.DBS().Reader, id)
	if err != nil {
		return nil, err
	}

	return ManufacturerToAPI(m), nil
}

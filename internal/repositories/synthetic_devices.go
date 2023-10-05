package repositories

import (
	"bytes"
	"encoding/base64"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vmihailenco/msgpack/v5"
)

func SyntheticDeviceToAPI(sd *models.SyntheticDevice) *gmodel.SyntheticDevice {
	return &gmodel.SyntheticDevice{
		ID:            sd.ID,
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
	}
}

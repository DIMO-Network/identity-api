package synthetic

import (
	"strings"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/mnemonic"
	"github.com/ethereum/go-ethereum/common"
)

type Repository struct {
	*base.Repository
}

func SyntheticDeviceToAPI(sd *models.SyntheticDevice) *gmodel.SyntheticDevice {
	nameList := mnemonic.FromInt32WithObfuscation(int32(sd.ID))
	name := strings.Join(nameList, " ")
	return &gmodel.SyntheticDevice{
		Name:          name,
		TokenID:       sd.ID,
		IntegrationID: sd.IntegrationID,
		Address:       common.BytesToAddress(sd.DeviceAddress),
		MintedAt:      sd.MintedAt,
	}
}

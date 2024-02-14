package synthetic

import (
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
)

type Repository struct {
	*repositories.Repository
}

func SyntheticDeviceToAPI(sd *models.SyntheticDevice) *gmodel.SyntheticDevice {
	return &gmodel.SyntheticDevice{
		TokenID:       sd.ID,
		IntegrationID: sd.IntegrationID,
		Address:       common.BytesToAddress(sd.DeviceAddress),
		MintedAt:      sd.MintedAt,
	}
}

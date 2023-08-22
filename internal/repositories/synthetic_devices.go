package repositories

import (
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
)

func SyntheticDeviceToAPI(sd *models.SyntheticDevice) *gmodel.SyntheticDevice {
	return &gmodel.SyntheticDevice{
		ID:            sd.ID,
		IntegrationID: sd.IntegrationID,
		Address:       common.BytesToAddress(sd.DeviceAddress),
		MintedAt:      sd.MintedAt,
	}
}
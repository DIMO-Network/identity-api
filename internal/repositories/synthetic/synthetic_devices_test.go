package synthetic

import (
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/stretchr/testify/assert"
)

func Test_SyntheticDeviceToAPI(t *testing.T) {
	_, wallet, err := helpers.GenerateWallet()
	assert.NoError(t, err)

	currTime := time.Now()
	sd := &models.SyntheticDevice{
		ID:            1,
		IntegrationID: 2,
		VehicleID:     1,
		DeviceAddress: wallet.Bytes(),
		MintedAt:      currTime,
	}

	res := SyntheticDeviceToAPI(sd)

	assert.Exactly(t, &model.SyntheticDevice{
		Name:          "learn island zoo",
		TokenID:       1,
		IntegrationID: 2,
		Address:       *wallet,
		MintedAt:      currTime,
	}, res)
}

package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func TestGetManufacturer(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)
	manufacturers := []string{"ford", "tesla", "kia", "acura", "honda", "jeep"}

	for i := 0; i < 6; i++ {
		m := models.Manufacturer{
			ID:       i,
			Name:     manufacturers[i],
			Owner:    common.FromHex("3232323232323232323232323232323232323232"),
			MintedAt: time.Now(),
		}

		err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	controller := New(pdb, config.Settings{})
	for i := 0; i < 6; i++ {
		tokenID := i
		res, err := controller.GetManufacturer(ctx, model.ManufacturerBy{TokenID: &tokenID})
		assert.NoError(t, err)
		assert.Equal(t, res.TokenID, i)
		assert.Equal(t, res.Name, manufacturers[i])

		manufacturerName := manufacturers[i]
		res, err = controller.GetManufacturer(ctx, model.ManufacturerBy{Name: &manufacturerName})
		assert.NoError(t, err)
		assert.Equal(t, res.TokenID, i)
		assert.Equal(t, res.Name, manufacturers[i])
	}

}

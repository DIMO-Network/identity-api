package manufacturer

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const migrationsDir = "../../../migrations"

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
			Slug:     manufacturers[i],
		}

		err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	logger := zerolog.Nop()
	controller := New(base.NewRepository(pdb, config.Settings{}, &logger))
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

func TestGetManufacturers(t *testing.T) {
	ctx := context.Background()
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)
	manufacturers := []string{"ford", "tesla", "kia", "acura", "honda", "jeep"}
	sort.Strings(manufacturers)

	for i := 0; i < 6; i++ {
		m := models.Manufacturer{
			ID:       i,
			Name:     manufacturers[i],
			Owner:    common.FromHex("3232323232323232323232323232323232323232"),
			MintedAt: time.Now(),
			Slug:     manufacturers[i],
		}

		err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	logger := zerolog.Nop()
	controller := New(base.NewRepository(pdb, config.Settings{}, &logger))

	res, err := controller.GetManufacturers(ctx)
	assert.NoError(t, err)
	assert.Len(t, res.Nodes, 6)
	assert.Equal(t, res.TotalCount, 6)
	assert.Equal(t, res.PageInfo.HasNextPage, false)
	assert.Equal(t, res.PageInfo.HasPreviousPage, false)
	assert.Len(t, res.Edges, 6)
	for i := 0; i < 6; i++ {
		assert.Equal(t, res.Nodes[i].TokenID, i)
		assert.Equal(t, res.Nodes[i].Name, manufacturers[i])
		assert.Equal(t, res.Edges[i].Node.TokenID, i)
		assert.Equal(t, res.Edges[i].Node.Name, manufacturers[i])
	}
}

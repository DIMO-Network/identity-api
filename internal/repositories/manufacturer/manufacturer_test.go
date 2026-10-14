package manufacturer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	"github.com/stretchr/testify/require"
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

// The schema's `manufacturer(by:): Manufacturer` is nullable, so an unknown
// make is an absent manufacturer, not an error: the answer is `data.manufacturer
// = null` with no `errors` entry. Returning the driver's sql.ErrNoRows made
// gqlgen attach `sql: no rows in result set`, which clients that treat any
// errors entry as an outage read as a permanent failure of the whole service.
func Test_manufacturerOrAbsent(t *testing.T) {
	ford := &models.Manufacturer{ID: 1, Name: "Ford", Slug: "ford"}

	got, err := manufacturerOrAbsent(ford, nil)
	require.NoError(t, err)
	assert.Equal(t, ford, got)

	got, err = manufacturerOrAbsent(nil, sql.ErrNoRows)
	require.NoError(t, err, "no such make is not an error")
	assert.Nil(t, got, "it is an absent manufacturer, which the nullable field serves as null")

	// sqlboiler wraps the driver error, so the check has to unwrap.
	got, err = manufacturerOrAbsent(nil, fmt.Errorf("bind failed: %w", sql.ErrNoRows))
	require.NoError(t, err)
	assert.Nil(t, got)

	boom := errors.New("read pool is down")
	got, err = manufacturerOrAbsent(nil, boom)
	assert.ErrorIs(t, err, boom, "a real failure is still a failure, and the caller masks it")
	assert.Nil(t, got)
}

func TestGetManufacturer_Missing(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	m := models.Manufacturer{
		ID:       1,
		Name:     "Ford",
		Owner:    common.FromHex("3232323232323232323232323232323232323232"),
		MintedAt: time.Now(),
		Slug:     "ford",
	}
	require.NoError(t, m.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

	logger := zerolog.Nop()
	controller := New(base.NewRepository(pdb, config.Settings{}, &logger))

	unknownToken := 4242
	unknownName := "Nothing Motors"
	unknownSlug := "nothing-motors"
	for _, by := range []model.ManufacturerBy{
		{TokenID: &unknownToken},
		{Name: &unknownName},
		{Slug: &unknownSlug},
	} {
		res, err := controller.GetManufacturer(ctx, by)
		require.NoError(t, err, "an unknown make is not an error")
		assert.Nil(t, res, "it is served as a null manufacturer")
	}
}

package devicedefinition

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const migrationsDir = "../../../migrations"

// manifest with three BMW definitions, one Alfa Romeo, and one definition
// whose id prefix is a manufacturer slug this database does not have; note
// metadata "" tolerance for backfilled legacy rows.
const manifestBody = `{
  "updatedAt": "2026-08-19T00:00:00.000Z",
  "count": 5,
  "definitions": [
    {
      "id": "bmw-m_z4_2021",
      "ksuid": "12G3iFH7Xc9Wvsw7pg6sD7uzoNN",
      "model": "Z4",
      "year": 2021,
      "devicetype": "vehicle",
      "imageuri": "https://image",
      "metadata": null,
      "manufacturer": {"tokenId": 13, "slug": "bmw-m", "name": "BMW M"}
    },
    {
      "id": "alfa-romeo_147_2007",
      "ksuid": "26G3iFH7Xc9Wvsw7pg6sD7uzoSS",
      "model": "147",
      "year": 2007,
      "devicetype": "vehicle",
      "imageuri": "https://image",
      "metadata": {"device_attributes": [{"name": "powertrain_type", "value": "ICE"}]},
      "manufacturer": {"tokenId": 137, "slug": "alfa-romeo", "name": "Alfa Romeo"}
    },
    {
      "id": "bmw_x5_2019",
      "ksuid": "12G3iFH7Xc9Wvsw7pg6sD7uzoKK",
      "model": "X5",
      "year": 2019,
      "devicetype": "vehicle",
      "imageuri": "https://image",
      "metadata": "",
      "manufacturer": {"tokenId": 13, "slug": "bmw", "name": "BMW"}
    },
    {
      "id": "bmw_x6_2019",
      "ksuid": "12G3iFH7Xc9Wvsw7pg6sD7uzoLL",
      "model": "X6",
      "year": 2019,
      "devicetype": "vehicle",
      "imageuri": "https://image",
      "metadata": null,
      "manufacturer": {"tokenId": 13, "slug": "bmw", "name": "BMW"}
    },
    {
      "id": "bmw_x7_2020",
      "ksuid": "12G3iFH7Xc9Wvsw7pg6sD7uzoMM",
      "model": "X7",
      "year": 2020,
      "devicetype": "vehicle",
      "imageuri": "https://image",
      "metadata": null,
      "manufacturer": {"tokenId": 13, "slug": "bmw", "name": "BMW"}
    }
  ]
}`

// catalogFixture serves the legacy manifest these tests are written against
// and records anything else that was asked for.
type catalogFixture struct {
	*services.DefinitionsCatalogService
	mu         sync.Mutex
	unexpected []string
}

func startCatalog(t *testing.T) *catalogFixture {
	t.Helper()
	f := &catalogFixture{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manifest.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(manifestBody))
		case "/idx/current.json":
			// No template index here, so the catalog reads the flat manifest
			// these tests are written against.
			http.NotFound(w, r)
		default:
			// This runs on a server goroutine, where require's FailNow would
			// call runtime.Goexit and kill the connection instead of failing
			// the test, leaving a confusing EOF for the test goroutine to trip
			// over. Record it and assert on the test goroutine instead.
			f.mu.Lock()
			f.unexpected = append(f.unexpected, r.URL.Path)
			f.mu.Unlock()
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	logger := zerolog.Nop()
	catalog, err := services.NewDefinitionsCatalogService(&logger, &config.Settings{DefinitionsCatalogURL: srv.URL})
	require.NoError(t, err)
	t.Cleanup(catalog.Close)
	f.DefinitionsCatalogService = catalog
	return f
}

// assertOnlyExpectedPaths runs on the test goroutine, after the queries.
func (f *catalogFixture) assertOnlyExpectedPaths(t *testing.T) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	assert.Empty(t, f.unexpected, "the catalog asked for paths this fixture does not serve")
}

// One query per page, not one per manufacturer slug: the ids are known before
// any of them is resolved.
func Test_manufacturerSlugs(t *testing.T) {
	defs := func(ids ...string) []*services.CatalogDefinition {
		out := make([]*services.CatalogDefinition, len(ids))
		for i, id := range ids {
			out[i] = &services.CatalogDefinition{ID: id}
		}
		return out
	}

	assert.Nil(t, manufacturerSlugs(nil))
	assert.Equal(t, []string{"bmw"}, manufacturerSlugs(defs("bmw_x5_2019", "bmw_x6_2019", "bmw_x7_2020")),
		"a page of one manufacturer is one slug, however many definitions it holds")
	assert.Equal(t, []string{"bmw", "alfa-romeo"}, manufacturerSlugs(defs("bmw_x5_2019", "alfa-romeo_147_2007", "bmw_x6_2019")),
		"distinct, in first-seen order")
	assert.Empty(t, manufacturerSlugs(defs("48682", "", "_orphan_2020")),
		"an id that names no manufacturer contributes no slug to look up")
	assert.Equal(t, []string{"dodge"}, manufacturerSlugs(defs("dodge_town-&-country_2012")))
}

// checkChain adopts a catalog in which up to 5% of the manufacturers shared
// with this chain carry a different slug, so an id prefix that matches no row
// is reachable by design. It used to fail the whole page with a raw
// sql.ErrNoRows.
func Test_indexManufacturers(t *testing.T) {
	bmw := &models.Manufacturer{ID: 13, Name: "BMW", Slug: "bmw"}
	alfa := &models.Manufacturer{ID: 137, Name: "Alfa Romeo", Slug: "alfa-romeo"}

	bySlug, missing := indexManufacturers([]string{"bmw", "alfa-romeo"}, models.ManufacturerSlice{bmw, alfa})
	assert.Equal(t, map[string]*models.Manufacturer{"bmw": bmw, "alfa-romeo": alfa}, bySlug)
	assert.Empty(t, missing)

	bySlug, missing = indexManufacturers([]string{"bmw", "renamed", "alfa-romeo"}, models.ManufacturerSlice{bmw, alfa})
	assert.Equal(t, bmw, bySlug["bmw"], "the definitions that do resolve still resolve")
	assert.Nil(t, bySlug["renamed"], "a definition with no manufacturer row has no manufacturer")
	assert.Equal(t, []string{"renamed"}, missing, "and the page says which, once, rather than failing")

	bySlug, missing = indexManufacturers([]string{"gone"}, nil)
	assert.Empty(t, bySlug)
	assert.Equal(t, []string{"gone"}, missing)
}

func Test_GetDeviceDefinitions_Query(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	mfr := models.Manufacturer{
		ID:    137,
		Name:  "Alfa Romeo",
		Owner: common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:  "alfa-romeo",
	}
	require.NoError(t, mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer()))
	mfr2 := models.Manufacturer{
		ID:    13,
		Name:  "BMW",
		Owner: common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:  "bmw",
	}
	require.NoError(t, mfr2.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

	logger := zerolog.Nop()
	repo := base.NewRepository(pdb, config.Settings{}, &logger)
	catalog := startCatalog(t)
	adController := New(repo, catalog.DefinitionsCatalogService)

	// Last 2 BMW definitions before the cursor for bmw_x7_2020.
	last := 2
	before := idToCursor("bmw_x7_2020")

	res, err := adController.GetDeviceDefinitions(ctx, 13, nil, nil, &last, &before, &model.DeviceDefinitionFilter{})
	require.NoError(t, err)

	assert.Len(t, res.Edges, 2)
	assert.Equal(t, 4, res.TotalCount)

	assert.Equal(t, "bmw_x5_2019", res.Edges[0].Node.DeviceDefinitionID)
	assert.Equal(t, "12G3iFH7Xc9Wvsw7pg6sD7uzoKK", *res.Edges[0].Node.LegacyID)
	assert.Equal(t, "vehicle", *res.Edges[0].Node.DeviceType)
	assert.Equal(t, "BMW", res.Edges[0].Node.Manufacturer.Name)
	assert.Equal(t, 13, res.Edges[0].Node.Manufacturer.TokenID)
	assert.Equal(t, "bmw_x6_2019", res.Edges[1].Node.DeviceDefinitionID)

	// Year filter.
	first := 10
	res, err = adController.GetDeviceDefinitions(ctx, 13, &first, nil, nil, nil, &model.DeviceDefinitionFilter{Year: &[]int{2020}[0]})
	require.NoError(t, err)
	assert.Equal(t, 1, res.TotalCount)
	assert.Equal(t, "bmw_x7_2020", res.Nodes[0].DeviceDefinitionID)

	// bmw-m_z4_2021's id prefix matches no manufacturer row. The whole page
	// used to fail with a raw sql.ErrNoRows; it is now one definition served
	// without a manufacturer.
	res, err = adController.GetDeviceDefinitions(ctx, 13, &first, nil, nil, nil, nil)
	require.NoError(t, err, "one unresolvable manufacturer must not fail the page")
	require.Len(t, res.Nodes, 4)
	assert.Equal(t, "bmw-m_z4_2021", res.Nodes[0].DeviceDefinitionID)
	assert.Nil(t, res.Nodes[0].Manufacturer)
	for _, node := range res.Nodes[1:] {
		require.NotNil(t, node.Manufacturer, node.DeviceDefinitionID)
		assert.Equal(t, 13, node.Manufacturer.TokenID)
	}

	catalog.assertOnlyExpectedPaths(t)
}

func Test_GetDeviceDefinition_Query(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	mfr := models.Manufacturer{
		ID:    137,
		Name:  "Alfa Romeo",
		Owner: common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:  "alfa-romeo",
	}
	require.NoError(t, mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

	logger := zerolog.Nop()
	repo := base.NewRepository(pdb, config.Settings{}, &logger)
	catalog := startCatalog(t)
	adController := New(repo, catalog.DefinitionsCatalogService)

	res, err := adController.GetDeviceDefinition(ctx, model.DeviceDefinitionBy{ID: "alfa-romeo_147_2007"})
	require.NoError(t, err)

	assert.Equal(t, "alfa-romeo_147_2007", res.DeviceDefinitionID)
	assert.Equal(t, "26G3iFH7Xc9Wvsw7pg6sD7uzoSS", *res.LegacyID)
	assert.Equal(t, "vehicle", *res.DeviceType)
	assert.Equal(t, "https://image", *res.ImageURI)
	assert.Equal(t, "Alfa Romeo", res.Manufacturer.Name)
	assert.Equal(t, 137, res.Manufacturer.TokenID)
	require.Len(t, res.Attributes, 1)
	assert.Equal(t, "powertrain_type", res.Attributes[0].Name)
	assert.Equal(t, "ICE", res.Attributes[0].Value)

	// bmw-m_z4_2021's id prefix matches no manufacturer row in this database.
	// The listing path serves it with a null manufacturer, so the by-id path
	// must too; it used to answer the raw driver string instead.
	res, err = adController.GetDeviceDefinition(ctx, model.DeviceDefinitionBy{ID: "bmw-m_z4_2021"})
	require.NoError(t, err, "a definition whose manufacturer slug has no row is still served")
	assert.Equal(t, "bmw-m_z4_2021", res.DeviceDefinitionID)
	assert.Nil(t, res.Manufacturer)

	// A definition the catalog genuinely lacks still gets the not-found
	// answer, whether or not its manufacturer resolves.
	_, err = adController.GetDeviceDefinition(ctx, model.DeviceDefinitionBy{ID: "alfa-romeo_nope_1999"})
	require.Error(t, err)
	assert.Equal(t, "no device definition found with that id", err.Error())

	_, err = adController.GetDeviceDefinition(ctx, model.DeviceDefinitionBy{ID: "no-such-make_nope_1999"})
	require.Error(t, err)
	assert.Equal(t, "no device definition found with that id", err.Error(),
		"an unknown make is not reported as a database error")

	catalog.assertOnlyExpectedPaths(t)
}

// The by-id path resolves its one manufacturer with a single-row query, so it
// has to make the same call the page-wide query makes through
// indexManufacturers: a slug with no row is an absent manufacturer, not an
// error, and certainly not the driver's own sql.ErrNoRows on the wire.
func Test_manufacturerOrAbsent(t *testing.T) {
	bmw := &models.Manufacturer{ID: 13, Name: "BMW", Slug: "bmw"}

	got, err := manufacturerOrAbsent(bmw, nil)
	require.NoError(t, err)
	assert.Equal(t, bmw, got)

	got, err = manufacturerOrAbsent(nil, sql.ErrNoRows)
	require.NoError(t, err, "a slug this database does not have is not an error")
	assert.Nil(t, got, "it is a definition with no manufacturer, which ToAPI leaves null")

	boom := errors.New("read pool is down")
	got, err = manufacturerOrAbsent(nil, boom)
	assert.ErrorIs(t, err, boom, "a real failure is still a failure")
	assert.Nil(t, got)

	// sqlboiler wraps the driver error, so the check has to unwrap.
	wrapped := fmt.Errorf("bind failed: %w", sql.ErrNoRows)
	got, err = manufacturerOrAbsent(nil, wrapped)
	require.NoError(t, err)
	assert.Nil(t, got)
}

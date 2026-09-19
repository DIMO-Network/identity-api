package devicedefinition

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"slices"
	"strings"

	"github.com/DIMO-Network/identity-api/internal/repositories/manufacturer"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// TokenPrefix is the prefix for a global token id for Device Definition.
const TokenPrefix = "DD"

type Repository struct {
	*base.Repository
	Catalog          *services.DefinitionsCatalogService
	ManufacturerRepo *manufacturer.Repository
}

// New creates a new device definition repository.
func New(db *base.Repository, catalog *services.DefinitionsCatalogService) *Repository {
	return &Repository{
		Repository:       db,
		Catalog:          catalog,
		ManufacturerRepo: manufacturer.New(db),
	}
}

func (r *Repository) ToAPI(v *services.CatalogDefinition, mfr *models.Manufacturer) (*gmodel.DeviceDefinition, error) {
	var result = gmodel.DeviceDefinition{
		DeviceDefinitionID: v.ID,
		Year:               v.Year,
		Model:              v.Model,
	}
	// A template carries no ksuid and legacyId is nullable, so leave it null
	// rather than answering with an empty legacy id that resolves to nothing.
	if v.KSUID != "" {
		ksuid := v.KSUID
		result.LegacyID = &ksuid
	}
	if mfr != nil {
		gmfr, err := r.ManufacturerRepo.ToAPI(mfr)
		if err != nil {
			return nil, err
		}
		result.Manufacturer = gmfr
	}

	if v.ImageURI != "" {
		imageURI := v.ImageURI
		result.ImageURI = &imageURI
	}

	if v.DeviceType != "" {
		deviceType := v.DeviceType
		result.DeviceType = &deviceType
	}

	for _, attr := range v.Metadata.DeviceAttributes {
		// No idea where this <nil> is coming from.
		if attr.Name == "" || attr.Value == "" || attr.Value == "<nil>" {
			continue
		}
		result.Attributes = append(result.Attributes, &gmodel.DeviceDefinitionAttribute{
			Name:  attr.Name,
			Value: attr.Value,
		})
	}

	return &result, nil
}

func (r *Repository) GetDeviceDefinition(ctx context.Context, by gmodel.DeviceDefinitionBy) (*gmodel.DeviceDefinition, error) {
	if len(by.ID) == 0 {
		return nil, gqlerror.Errorf("Provide an `id`.")
	}

	mfrSlug, _, found := strings.Cut(by.ID, "_")
	if !found {
		return nil, gqlerror.Errorf("The `ID` is incorrect.")
	}

	// The catalog decides whether this id names a definition at all; the
	// manufacturer is decoration on top of that answer, so it is resolved
	// second and never gets to turn a definition the snapshot holds into an
	// error.
	def, err := r.Catalog.GetDefinitionByID(ctx, by.ID)
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, errors.New("no device definition found with that id")
	}

	mfr, err := manufacturerOrAbsent(models.Manufacturers(models.ManufacturerWhere.Slug.EQ(mfrSlug)).One(ctx, r.PDB.DBS().Reader))
	if err != nil {
		return nil, err
	}
	if mfr == nil {
		r.Log.Warn().Str("slug", mfrSlug).Str("definition", by.ID).
			Msg("device definition names a manufacturer slug this database does not have; serving it without a manufacturer")
	}

	return r.ToAPI(def, mfr)
}

// manufacturerOrAbsent maps a single-row manufacturer lookup onto what the API
// serves for it. A slug with no row is an absent manufacturer, not an error:
// for the reasons manufacturersForPage gives, the catalog's idea of a
// manufacturer's slug and this database's can differ by design, the schema's
// `manufacturer: Manufacturer` is nullable, and ToAPI already leaves it null.
// Returning the driver's sql.ErrNoRows instead put `sql: no rows in result
// set` on the wire for a definition the listing path serves happily.
func manufacturerOrAbsent(mfr *models.Manufacturer, err error) (*models.Manufacturer, error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mfr, nil
}

func (r *Repository) GetDeviceDefinitions(ctx context.Context, manufacturerTokenID int, first *int, after *string, last *int, before *string, filterBy *gmodel.DeviceDefinitionFilter) (*gmodel.DeviceDefinitionConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	defs, err := r.Catalog.DefinitionsByManufacturer(ctx, manufacturerTokenID)
	if err != nil {
		return nil, err
	}

	// Filter; defs is already sorted by id ascending.
	all := make([]*services.CatalogDefinition, 0, len(defs))
	for _, d := range defs {
		if filterBy != nil {
			if filterBy.Year != nil && d.Year != *filterBy.Year {
				continue
			}
			if filterBy.Model != nil && !strings.EqualFold(d.Model, *filterBy.Model) {
				continue
			}
		}
		all = append(all, d)
	}

	totalCount := len(all)

	if after != nil {
		afterID, err := cursorToID(*after)
		if err != nil {
			return nil, err
		}
		all = slices.DeleteFunc(slices.Clone(all), func(d *services.CatalogDefinition) bool { return d.ID <= afterID })
	}

	if before != nil {
		beforeID, err := cursorToID(*before)
		if err != nil {
			return nil, err
		}
		all = slices.DeleteFunc(slices.Clone(all), func(d *services.CatalogDefinition) bool { return d.ID >= beforeID })
	}

	if last != nil {
		slices.Reverse(all)
	}

	// We assume that cursors come from real elements.
	hasNext := before != nil
	hasPrevious := after != nil

	if len(all) > limit {
		all = all[:limit]
		if last != nil {
			hasPrevious = true
		} else {
			hasNext = true
		}
	}

	if last != nil {
		slices.Reverse(all)
	}

	var errList gqlerror.List
	var endCur, startCur *string
	if len(all) != 0 {
		ec := idToCursor(all[len(all)-1].ID)
		endCur = &ec

		sc := idToCursor(all[0].ID)
		startCur = &sc
	}

	edges := make([]*gmodel.DeviceDefinitionEdge, len(all))
	nodes := make([]*gmodel.DeviceDefinition, len(all))

	mfrs, err := r.manufacturersForPage(ctx, all)
	if err != nil {
		return nil, err
	}

	for i, dv := range all {
		mfrSlug, _, _ := strings.Cut(dv.ID, "_")
		// A definition whose id prefix named no row is served without a
		// manufacturer, which ToAPI leaves null.
		gv, err := r.ToAPI(dv, mfrs[mfrSlug])
		if err != nil {
			errList = append(errList, gqlerror.Wrap(err))
			continue
		}

		edges[i] = &gmodel.DeviceDefinitionEdge{
			Node:   gv,
			Cursor: idToCursor(dv.ID),
		}

		nodes[i] = gv
	}

	res := &gmodel.DeviceDefinitionConnection{
		Edges: edges,
		Nodes: nodes,
		PageInfo: &gmodel.PageInfo{
			EndCursor:       endCur,
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCur,
		},
		TotalCount: totalCount,
	}

	if errList != nil {
		return res, errList
	}

	return res, nil
}

// manufacturersForPage fetches, in one query, the manufacturers the ids on
// this page name, keyed by slug.
//
// A slug with no row is not an error. The catalog's idea of a manufacturer's
// slug and this database's can differ: checkChain adopts a catalog in which up
// to 5% of the manufacturers shared with this chain carry a different slug, so
// an id prefix that resolves to nothing is reachable by design. Resolving each
// definition with its own .One() turned one such definition into a raw
// sql.ErrNoRows that failed the whole page -- every other definition included.
func (r *Repository) manufacturersForPage(ctx context.Context, defs []*services.CatalogDefinition) (map[string]*models.Manufacturer, error) {
	slugs := manufacturerSlugs(defs)
	if len(slugs) == 0 {
		return nil, nil
	}

	rows, err := models.Manufacturers(models.ManufacturerWhere.Slug.IN(slugs)).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	bySlug, missing := indexManufacturers(slugs, rows)
	if len(missing) > 0 {
		r.Log.Warn().Strs("slugs", missing).Int("definitions", len(defs)).
			Msg("device definitions name manufacturer slugs this database does not have; serving them without a manufacturer")
	}
	return bySlug, nil
}

// manufacturerSlugs returns the distinct manufacturer slugs this page's ids
// name, in the order they first appear. A definition id is
// <manufacturer slug>_<model>_<year>; an id not in that shape names no
// manufacturer to look up.
func manufacturerSlugs(defs []*services.CatalogDefinition) []string {
	var slugs []string
	seen := make(map[string]struct{}, len(defs))
	for _, d := range defs {
		slug, _, found := strings.Cut(d.ID, "_")
		if !found || slug == "" {
			continue
		}
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		slugs = append(slugs, slug)
	}
	return slugs
}

// indexManufacturers keys the rows the query returned by slug, and names the
// requested slugs it returned nothing for.
func indexManufacturers(slugs []string, rows models.ManufacturerSlice) (map[string]*models.Manufacturer, []string) {
	bySlug := make(map[string]*models.Manufacturer, len(rows))
	for _, m := range rows {
		bySlug[m.Slug] = m
	}

	var missing []string
	for _, slug := range slugs {
		if _, ok := bySlug[slug]; !ok {
			missing = append(missing, slug)
		}
	}
	return bySlug, missing
}

func idToCursor(id string) string {
	return base64.StdEncoding.EncodeToString([]byte(id))
}

func cursorToID(cursor string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

package devicedefinition

import (
	"context"
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
	ksuid := v.KSUID
	var result = gmodel.DeviceDefinition{
		DeviceDefinitionID: v.ID,
		LegacyID:           &ksuid,
		Year:               v.Year,
		Model:              v.Model,
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

	if v.Metadata != nil {
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

	mfr, err := models.Manufacturers(models.ManufacturerWhere.Slug.EQ(mfrSlug)).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	def, err := r.Catalog.GetDefinitionByID(ctx, by.ID)
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, errors.New("no device definition found with that id")
	}

	return r.ToAPI(def, mfr)
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

	mfrs := make(map[string]*models.Manufacturer)

	for i, dv := range all {
		// get the manufacturer and cache it in a map outside the loop, using the cached mfr when exists
		var mfr *models.Manufacturer
		mfrSlug, _, found := strings.Cut(dv.ID, "_")
		if found {
			ok := false
			if mfr, ok = mfrs[mfrSlug]; !ok {
				mfr, err = models.Manufacturers(models.ManufacturerWhere.Slug.EQ(mfrSlug)).One(ctx, r.PDB.DBS().Reader)
				if err != nil {
					return nil, err
				}
				mfrs[mfrSlug] = mfr
			}
		}

		gv, err := r.ToAPI(dv, mfr)
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

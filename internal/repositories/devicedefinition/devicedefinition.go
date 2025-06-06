package devicedefinition

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/DIMO-Network/identity-api/internal/repositories/manufacturer"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/doug-martin/goqu/v9"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// TokenPrefix is the prefix for a global token id for Device Definition.
const TokenPrefix = "DD"

type DeviceDefinitionTablelandCountModel struct {
	Count int `json:"count(*)"`
}

type DeviceDefinitionTablelandModel struct {
	ID         string                    `json:"id"`
	KSUID      string                    `json:"ksuid"`
	Model      string                    `json:"model"`
	Year       int                       `json:"year"`
	DeviceType string                    `json:"devicetype"`
	ImageURI   string                    `json:"imageuri"`
	Metadata   *DeviceDefinitionMetadata `json:"metadata"`
}

type DeviceDefinitionMetadata struct {
	DeviceAttributes []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"device_attributes"`
}

// UnmarshalJSON customizes the unmarshaling of DeviceDefinitionTablelandModel to handle cases where metadata is an empty string.
func (d *DeviceDefinitionTablelandModel) UnmarshalJSON(data []byte) error {
	type Alias DeviceDefinitionTablelandModel // Create an alias to avoid recursion

	aux := &struct {
		Metadata json.RawMessage `json:"metadata"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.Metadata) > 0 && string(aux.Metadata) != `""` {
		metadata := new(DeviceDefinitionMetadata)
		if err := json.Unmarshal(aux.Metadata, metadata); err != nil {
			return err
		}
		d.Metadata = metadata
	}

	return nil
}

type Repository struct {
	*base.Repository
	TablelandApiService *services.TablelandApiService
	ManufacturerRepo    *manufacturer.Repository
}

// New creates a new device definition repository.
func New(db *base.Repository, tablelandAPI *services.TablelandApiService) *Repository {
	return &Repository{
		Repository:          db,
		TablelandApiService: tablelandAPI,
		ManufacturerRepo:    manufacturer.New(db),
	}
}

func (r *Repository) ToAPI(v *DeviceDefinitionTablelandModel, mfr *models.Manufacturer) (*gmodel.DeviceDefinition, error) {
	var result = gmodel.DeviceDefinition{
		DeviceDefinitionID: v.ID,
		LegacyID:           &v.KSUID,
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
		result.ImageURI = &v.ImageURI
	}

	if v.DeviceType != "" {
		result.DeviceType = &v.DeviceType
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

	if !mfr.TableID.Valid {
		return nil, fmt.Errorf("manufacturer %d does not have a device definition table", mfr.ID)
	}

	table := fmt.Sprintf("_%d_%d", r.Settings.DIMORegistryChainID, mfr.TableID.Int)

	sql, _, err := goqu.Dialect("sqlite3").From(table).Where(goqu.Ex{"id": by.ID}).ToSQL()
	if err != nil {
		return nil, err
	}

	var modelTableland []DeviceDefinitionTablelandModel

	if err = r.TablelandApiService.Query(ctx, sql, &modelTableland); err != nil {
		return nil, err
	}

	if len(modelTableland) == 0 {
		return nil, errors.New("no device definition found with that id")
	}

	return r.ToAPI(&modelTableland[0], mfr)
}

func (r *Repository) GetDeviceDefinitions(ctx context.Context, tableID, first *int, after *string, last *int, before *string, filterBy *gmodel.DeviceDefinitionFilter) (*gmodel.DeviceDefinitionConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	if tableID == nil {
		return &gmodel.DeviceDefinitionConnection{
			Edges:      make([]*gmodel.DeviceDefinitionEdge, 0),
			Nodes:      make([]*gmodel.DeviceDefinition, 0),
			PageInfo:   &gmodel.PageInfo{},
			TotalCount: 0,
		}, nil
	}

	table := fmt.Sprintf("_%d_%d", r.Settings.DIMORegistryChainID, *tableID)

	sqlBuild := goqu.Dialect("sqlite3").From(table)

	if filterBy != nil {
		if filterBy.Year != nil {
			sqlBuild = sqlBuild.Where(goqu.Ex{"year": *filterBy.Year})
		}

		if filterBy.Model != nil {
			sqlBuild = sqlBuild.Where(goqu.L("model = ? COLLATE NOCASE", *filterBy.Model))
		}
	}

	countSQL, _, err := sqlBuild.Select(goqu.COUNT("*")).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error constructing count SQL: %w", err)
	}

	var modelCountTableland []DeviceDefinitionTablelandCountModel
	if err = r.TablelandApiService.Query(ctx, countSQL, &modelCountTableland); err != nil {
		return nil, err
	}

	if len(modelCountTableland) == 0 {
		return nil, errors.New("error from Tableland")
	}

	totalCount := modelCountTableland[0].Count

	if after != nil {
		afterID, err := cursorToID(*after)
		if err != nil {
			return nil, err
		}
		sqlBuild = sqlBuild.Where(goqu.C("id").Gt(afterID))
	}

	if before != nil {
		beforeID, err := cursorToID(*before)
		if err != nil {
			return nil, err
		}

		sqlBuild = sqlBuild.Where(goqu.C("id").Lt(beforeID))
	}

	if last != nil {
		sqlBuild = sqlBuild.Order(goqu.I("id").Desc())
	} else {
		sqlBuild = sqlBuild.Order(goqu.I("id").Asc())
	}

	sqlBuild = sqlBuild.Limit(uint(limit) + 1)

	allSQL, _, err := sqlBuild.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error constructing selection SQL: %w", err)
	}

	var all []DeviceDefinitionTablelandModel
	if err = r.TablelandApiService.Query(ctx, allSQL, &all); err != nil {
		return nil, err
	}

	// We assume that cursors come from real elements.
	hasNext := before != nil
	hasPrevious := after != nil

	if first != nil && len(all) == limit+1 {
		hasNext = true
		all = all[:limit]
	} else if last != nil && len(all) == limit+1 {
		hasPrevious = true
		all = all[:limit]
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
			}
		}

		gv, err := r.ToAPI(&dv, mfr)
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
		TotalCount: int(totalCount),
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

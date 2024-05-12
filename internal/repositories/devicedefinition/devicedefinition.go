package devicedefinition

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/doug-martin/goqu/v9"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"golang.org/x/exp/slices"
)

var dialect = goqu.Dialect("sqlite3")

// TokenPrefix is the prefix for a global token id for Device Definition.
const TokenPrefix = "DD"

type DeviceDefinitionTablelandCountModel struct {
	Count int `json:"count(*)"`
}

type DeviceDefinitionTablelandModel struct {
	ID         string `json:"id"`
	KSUID      string `json:"ksuid"`
	Model      string `json:"model"`
	Year       int    `json:"year"`
	DeviceType string `json:"deviceType"`
	ImageURI   string `json:"imageURI"`
	Metadata   struct {
		DeviceAttributes []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"device_attributes"`
	} `json:"metadata"`
}

type Repository struct {
	*base.Repository
	TablelandApiService *services.TablelandApiService
}

func ToAPI(v *DeviceDefinitionTablelandModel) (*gmodel.DeviceDefinition, error) {
	var result = gmodel.DeviceDefinition{
		ID:                 definitionIDToGlobalID(v.ID),
		DeviceDefinitionID: v.ID,
		LegacyID:           &v.KSUID,
		Year:               v.Year,
		Model:              v.Model,
		DeviceType:         &v.DeviceType,
		ImageURI:           &v.ImageURI,
	}

	for _, attr := range v.Metadata.DeviceAttributes {
		result.Attributes = append(result.Attributes, &gmodel.DeviceDefinitionAttribute{
			Name:  &attr.Name,
			Value: &attr.Value,
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

	mfr, err := models.Manufacturers(models.ManufacturerWhere.Slug.EQ(mfrSlug)).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if !mfr.TableID.Valid {
		return nil, fmt.Errorf("manufacturer %d does not have a device definition table", mfr.ID)
	}

	table := fmt.Sprintf("_%d_%d", r.Settings.DIMORegistryChainID, mfr.TableID.Int)

	sql, _, err := dialect.From(table).Where(goqu.Ex{"id": by.ID}).ToSQL()
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

	return ToAPI(&modelTableland[0])
}

func (r *Repository) GetDeviceDefinitions(ctx context.Context, tableID, first *int, after *string, last *int, before *string, filterBy *gmodel.DeviceDefinitionFilter) (*gmodel.DeviceDefinitionConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	if tableID == nil {
		return nil, gqlerror.Errorf("Manufacturer does not have a device definitions table.")
	}

	table := fmt.Sprintf("_%d_%d", r.Settings.DIMORegistryChainID, *tableID)

	sqlBuild := dialect.From(table)

	if filterBy != nil {
		if filterBy.Year != nil {
			sqlBuild = sqlBuild.Where(goqu.Ex{"year": *filterBy.Year})
		}

		if filterBy.Model != nil {
			sqlBuild = sqlBuild.Where(goqu.Ex{"model": *filterBy.Model})
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
		afterID, err := cursorToDefinitionID(*after)
		if err != nil {
			return nil, err
		}

		sqlBuild = sqlBuild.Where(goqu.C("id").Gt(afterID))
	}

	if before != nil {
		beforeID, err := cursorToDefinitionID(*before)
		if err != nil {
			return nil, err
		}

		sqlBuild = sqlBuild.Where(goqu.C("id").Lt(beforeID))
	}

	orderBy := goqu.C("id").Asc()
	if last != nil {
		orderBy = goqu.C("id").Desc()
	}

	allSQL, _, err := sqlBuild.Limit(uint(limit) + 1).Order(orderBy).ToSQL()
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
		ec := definitionIDToCursor(all[len(all)-1].ID)
		endCur = &ec

		sc := definitionIDToCursor(all[0].ID)
		startCur = &sc
	}

	edges := make([]*gmodel.DeviceDefinitionEdge, len(all))
	nodes := make([]*gmodel.DeviceDefinition, len(all))

	for i, dv := range all {
		gv, err := ToAPI(&dv)
		if err != nil {
			errList = append(errList, gqlerror.Wrap(err))
			continue
		}

		edges[i] = &gmodel.DeviceDefinitionEdge{
			Node:   gv,
			Cursor: definitionIDToCursor(dv.ID),
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

func definitionIDToCursor(id string) string {
	return base64.StdEncoding.EncodeToString([]byte(id))
}

func cursorToDefinitionID(cur string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(cur)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func definitionIDToGlobalID(id string) string {
	return fmt.Sprintf("%s_%s", TokenPrefix, base64.StdEncoding.EncodeToString([]byte(id)))
}
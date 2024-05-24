package devicedefinition

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/DIMO-Network/identity-api/graph/model"
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/doug-martin/goqu/v9"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"golang.org/x/exp/slices"
)

// TokenPrefix is the prefix for a global token id for Device Definition.
const TokenPrefix = "DD"

type DeviceDefinitionTablelandCountModel struct {
	Count int `json:"count(*)"`
}

type DeviceDefinitionTablelandModel struct {
	Index      int    `json:"index"`
	ID         string `json:"id"`
	KSUID      string `json:"ksuid"`
	Model      string `json:"model"`
	Year       int    `json:"year"`
	DeviceType string `json:"devicetype"`
	ImageURI   string `json:"imageuri"`
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
	var result = gmodel.DeviceDefinition{}

	globalID, err := base.EncodeGlobalTokenID(TokenPrefix, v.Index)
	if err != nil {
		return nil, fmt.Errorf("error encoding device definition id: %w", err)
	}

	result.ID = globalID
	result.DeviceDefinitionID = &v.ID
	result.LegacyID = &v.KSUID
	result.Year = &v.Year
	result.Model = &v.Model
	result.DeviceType = &v.DeviceType

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

	return ToAPI(&modelTableland[0])
}

func (r *Repository) GetDeviceDefinitions(ctx context.Context, tableID, first *int, after *string, last *int, before *string, filterBy *model.DeviceDefinitionFilter) (*gmodel.DeviceDefinitionConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	if tableID == nil {
		return nil, gqlerror.Errorf("Provide exactly one `manufacturer`.")
	}

	mfr, err := models.Manufacturers(models.ManufacturerWhere.Slug.EQ(filterBy.Manufacturer)).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if !mfr.TableID.Valid {
		return nil, fmt.Errorf("manufacturer %d does not have a device definition table", mfr.ID)
	}

	table := fmt.Sprintf("_%d_%d", r.Settings.DIMORegistryChainID, mfr.TableID.Int)

	sqlBuild := goqu.Dialect("sqlite3").From(table)

	if filterBy.Year != nil {
		sqlBuild = sqlBuild.Where(goqu.Ex{"year": *filterBy.Year})
	}

	if filterBy.Model != nil {
		sqlBuild = sqlBuild.Where(goqu.Ex{"model": *filterBy.Model})
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

	allSQL, _, err := sqlBuild.Order(goqu.I("id").Asc()).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error constructing selection SQL: %w", err)
	}

	var all []DeviceDefinitionTablelandModel
	if err = r.TablelandApiService.Query(ctx, allSQL, &all); err != nil {
		return nil, err
	}

	for i, model := range all {
		model.Index = i
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
		ec := helpers.IDToCursor(all[len(all)-1].Index)
		endCur = &ec

		sc := helpers.IDToCursor(all[0].Index)
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
			Cursor: helpers.IDToCursor(dv.Index),
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

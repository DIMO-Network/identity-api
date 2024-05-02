package devicedefinition

import (
	"context"
	"fmt"
	"strings"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
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
	ManufacturerContractService *services.ManufacturerContractService
	TablelandApiService         *services.TablelandApiService
	ManufacturerCacheService    *services.ManufacturerCacheService
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
	result.ImageURI = &v.ImageURI

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
		return nil, gqlerror.Errorf("Provide exactly one `ID`.")
	}
	splitID := strings.Split(by.ID, "_")

	if len(splitID) < 2 {
		return nil, gqlerror.Errorf("The `ID` is incorrect.")
	}

	manufacturerSlug := splitID[0]

	manufactures, err := r.ManufacturerCacheService.GetAllManufacturers(ctx)
	if err != nil {
		return nil, gqlerror.Errorf("failed load manufactures: %s", err)
	}

	var manufacturer *services.ManufacturerCacheModel
	for _, model := range manufactures {
		if strings.EqualFold(model.Slug, manufacturerSlug) {
			manufacturer = &model
			break
		}
	}

	if manufacturer == nil {
		return nil, gqlerror.Errorf("Manufacturer %s doesn't exist.", manufacturerSlug)
	}

	tableName, err := r.ManufacturerContractService.GetTableName(ctx, manufacturer.Name)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	statement := fmt.Sprintf("SELECT * FROM %s WHERE id = '%s'", *tableName, strings.ToLower(by.ID))
	queryParams := map[string]string{
		"statement": statement,
	}
	var modelTableland []DeviceDefinitionTablelandModel

	if err = r.TablelandApiService.Query(ctx, queryParams, &modelTableland); err != nil {
		return nil, err
	}

	fmt.Print(statement)

	var cont = 1
	for _, item := range modelTableland {
		item.Index = cont
		return ToAPI(&item)
	}

	return nil, nil
}

func (r *Repository) GetDeviceDefinitions(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *gmodel.DeviceDefinitionFilter) (*gmodel.DeviceDefinitionConnection, error) {

	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	if len(filterBy.Manufacturer) == 0 {
		return nil, gqlerror.Errorf("Provide exactly one `manufacturer`.")
	}

	var conditions []string
	if filterBy.Year != nil && (*filterBy.Year > 1980 && *filterBy.Year < 2999) {
		conditions = append(conditions, fmt.Sprintf("year = %d", *filterBy.Year))
	}
	if filterBy.Model != nil && len(*filterBy.Model) > 0 {
		conditions = append(conditions, fmt.Sprintf("model = '%s'", *filterBy.Model))
	}
	if filterBy.ID != nil && len(*filterBy.ID) > 0 {
		conditions = append(conditions, fmt.Sprintf("id = '%s'", *filterBy.ID))
	}

	whereClause := strings.Join(conditions, " AND ")
	if whereClause != "" {
		whereClause = " WHERE " + whereClause
	}

	manufactures, err := r.ManufacturerCacheService.GetAllManufacturers(ctx)
	if err != nil {
		return nil, gqlerror.Errorf("failed load manufactures: %s", err)
	}

	var manufacturer *services.ManufacturerCacheModel
	for _, model := range manufactures {
		if strings.EqualFold(model.Slug, filterBy.Manufacturer) {
			manufacturer = &model
			break
		}
	}

	tableName, err := r.ManufacturerContractService.GetTableName(ctx, manufacturer.Name)
	if err != nil {
		return nil, err
	}

	countParams := map[string]string{
		"statement": fmt.Sprintf("SELECT count(*) FROM %s%s", *tableName, whereClause),
	}
	var modelCountTableland []DeviceDefinitionTablelandCountModel
	if err = r.TablelandApiService.Query(ctx, countParams, &modelCountTableland); err != nil {
		return nil, err
	}

	totalCount := modelCountTableland[0].Count

	queryParams := map[string]string{
		"statement": fmt.Sprintf("SELECT * FROM %s%s", *tableName, whereClause),
	}

	var all []DeviceDefinitionTablelandModel
	if err = r.TablelandApiService.Query(ctx, queryParams, &all); err != nil {
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
		TotalCount: totalCount,
	}

	if errList != nil {
		return res, errList
	}

	return res, nil
}

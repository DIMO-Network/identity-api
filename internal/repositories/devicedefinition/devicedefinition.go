package devicedefinition

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"net/http"
	"net/url"
	"strings"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/patrickmn/go-cache"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"golang.org/x/exp/slices"

	"github.com/DIMO-Network/identity-api/internal/contracts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TokenPrefix is the prefix for a global token id for Device Definition.
const TokenPrefix = "DD"

type DeviceDefinitionTablelandCountModel struct {
	Count int `json:"count(*)"`
}

type DeviceDefinitionTablelandModel struct {
	ID       string `json:"id"`
	KSUID    string `json:"ksuid"`
	Model    string `json:"model"`
	Year     int    `json:"year"`
	Metadata struct {
		DeviceAttributes []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"device_attributes"`
	} `json:"metadata"`
}

type ManufacturerCacheModel struct {
	ID   int
	Name string
	Slug string
}

type DeviceDefinitionRepository struct {
	*base.Repository
	client *ethclient.Client
	Cache  *cache.Cache
}

func ToAPI(v *DeviceDefinitionTablelandModel) *gmodel.DeviceDefinition {
	var result = gmodel.DeviceDefinition{}

	result.ID = v.ID
	result.Ksuid = &v.KSUID
	result.Year = &v.Year
	result.Model = &v.Model

	for _, attr := range v.Metadata.DeviceAttributes {
		result.Attributes = append(result.Attributes, &gmodel.DeviceDefinitionAttribute{
			Name:  &attr.Name,
			Value: &attr.Value,
		})
	}

	return &result
}

func (r *DeviceDefinitionRepository) GetDeviceDefinition(ctx context.Context, by gmodel.DevicedefinitionBy) (*gmodel.DeviceDefinition, error) {

	if len(by.ID) == 0 {
		return nil, gqlerror.Errorf("Provide exactly one `ID`.")
	}

	splitID := strings.Split(by.ID, "_")

	if len(splitID) < 2 {
		return nil, gqlerror.Errorf("The `ID` is incorrect.")
	}

	manufacturerSlug := splitID[0]

	tableName, err := r.ResolveTableName(ctx, manufacturerSlug)
	if err != nil {
		return nil, err
	}

	statement := fmt.Sprintf("SELECT * FROM %s WHERE id = '%s'", tableName, by.ID)
	queryParams := map[string]string{
		"statement": statement,
	}
	var modelTableland []DeviceDefinitionTablelandModel

	if err = r.QueryTableland(ctx, queryParams, &modelTableland); err != nil {
		return nil, err
	}

	for _, item := range modelTableland {
		return ToAPI(&item), nil
	}

	return nil, nil
}

func (r *DeviceDefinitionRepository) GetDeviceDefinitions(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *gmodel.DevicedefinitionFilter) (*gmodel.DeviceDefinitionConnection, error) {
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

	tableName, err := r.ResolveTableName(ctx, filterBy.Manufacturer)
	if err != nil {
		return nil, err
	}

	countParams := map[string]string{
		"statement": fmt.Sprintf("SELECT count(*) FROM %s%s", tableName, whereClause),
	}
	var modelCountTableland []DeviceDefinitionTablelandCountModel
	if err = r.QueryTableland(ctx, countParams, &modelCountTableland); err != nil {
		return nil, err
	}

	totalCount := modelCountTableland[0].Count

	queryParams := map[string]string{
		"statement": fmt.Sprintf("SELECT * FROM %s%s LIMIT %d OFFSET %d", tableName, whereClause, limit, 1),
	}

	var modelTableland []DeviceDefinitionTablelandModel
	if err = r.QueryTableland(ctx, queryParams, &modelTableland); err != nil {
		return nil, err
	}

	var all []*gmodel.DeviceDefinition
	for _, item := range modelTableland {
		all = append(all, ToAPI(&item))
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
		//ec := helpers.IDToCursor(all[len(all)-1].ID)
		//endCur = &ec
		//
		//sc := helpers.IDToCursor(all[0].ID)
		//startCur = &sc
	}

	edges := make([]*gmodel.DeviceDefinitionEdge, len(all))
	nodes := make([]*gmodel.DeviceDefinition, len(all))

	for i, dv := range all {

		edges[i] = &gmodel.DeviceDefinitionEdge{
			Node: dv,
			//Cursor: helpers.IDToCursor(dv.ID),
		}

		nodes[i] = dv
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

func (r *DeviceDefinitionRepository) ResolveTableName(ctx context.Context, manufacturerSlug string) (*string, error) {

	manufactures, err := r.GetManufacturers(ctx)
	if err != nil {
		return nil, gqlerror.Errorf("failed load manufactures: %w", err)
	}

	var manufacturer *ManufacturerCacheModel
	for _, model := range manufactures {
		if model.Slug == manufacturerSlug {
			manufacturer = &model
			break
		}
	}

	if manufacturer == nil {
		return nil, gqlerror.Errorf("Manufacturer %s doesn't exist.", manufacturerSlug)
	}

	contractAddress := common.HexToAddress(r.Settings.DIMORegistryAddr)
	queryInstance, err := contracts.NewRegistry(contractAddress, r.client)
	tableName, err := queryInstance.GetDeviceDefinitionTableName(&bind.CallOpts{Context: ctx, Pending: true}, big.NewInt(int64(manufacturer.ID)))

	if err != nil {
		return nil, gqlerror.Errorf("failed get GetDeviceDefinitionTableName: %w", err)
	}

	return &tableName, nil
}

func (r *DeviceDefinitionRepository) QueryTableland(ctx context.Context, queryParams map[string]string, result interface{}) error {
	fullURL, err := url.Parse(r.Settings.TablelandAPIGateway)
	if err != nil {
		return err
	}

	fullURL = fullURL.JoinPath(fullURL.Path, "api/v1/query")

	if queryParams != nil {
		values := fullURL.Query()
		for key, value := range queryParams {
			values.Set(key, value)
		}
		fullURL.RawQuery = values.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to complete request: %w", err)
	}

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}

	return nil
}

func (r *DeviceDefinitionRepository) GetManufacturers(ctx context.Context) ([]ManufacturerCacheModel, error) {

	if manufacturers, ok := r.Cache.Get("manufacturers"); ok {
		return manufacturers.([]ManufacturerCacheModel), nil
	}

	manufacturers, err := models.Manufacturers().All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, gqlerror.Errorf("could not all manufacturers")
		}

		return nil, err
	}

	var all []ManufacturerCacheModel
	for _, manufacturer := range manufacturers {
		all = append(all, ManufacturerCacheModel{
			ID:   manufacturer.ID,
			Name: manufacturer.Name,
			Slug: helpers.SlugString(manufacturer.Name),
		})
	}

	r.Cache.Set("manufacturers", all, time.Hour*24*7)

	return all, nil
}

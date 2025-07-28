package vehicle

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/cloudevent"
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/mnemonic"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// TokenPrefix is the prefix for a global token id for vehicles.
const TokenPrefix = "V"

type Repository struct {
	*base.Repository
	chainID         uint64
	contractAddress common.Address
}

// New creates a new vehicle repository.
func New(db *base.Repository) *Repository {
	return &Repository{
		Repository:      db,
		chainID:         uint64(db.Settings.DIMORegistryChainID),
		contractAddress: common.HexToAddress(db.Settings.VehicleNFTAddr),
	}
}

func (r *Repository) createVehiclesResponse(totalCount int64, vehicles models.VehicleSlice, hasNext bool, hasPrevious bool) (*gmodel.VehicleConnection, error) {
	var errList gqlerror.List
	var endCur, startCur *string
	if len(vehicles) != 0 {
		ec := helpers.IDToCursor(vehicles[len(vehicles)-1].ID)
		endCur = &ec

		sc := helpers.IDToCursor(vehicles[0].ID)
		startCur = &sc
	}

	edges := make([]*gmodel.VehicleEdge, len(vehicles))
	nodes := make([]*gmodel.Vehicle, len(vehicles))

	for i, dv := range vehicles {
		var imageURI string

		if dv.ImageURI.Valid {
			imageURI = dv.ImageURI.String
		} else {
			var err error
			imageURI, err = DefaultImageURI(r.Settings.BaseImageURL, dv.ID)
			if err != nil {
				wErr := fmt.Errorf("error getting vehicle image url: %w", err)
				errList = append(errList, gqlerror.Wrap(wErr))
				continue
			}
		}

		dataURI, err := GetVehicleDataURI(r.Settings.BaseVehicleDataURI, dv.ID)
		if err != nil {
			wErr := fmt.Errorf("error getting vehicle data uri: %w", err)
			errList = append(errList, gqlerror.Wrap(wErr))
			continue
		}
		gv, err := r.ToAPI(dv, imageURI, dataURI, nil)
		if err != nil {
			wErr := fmt.Errorf("error converting vehicle to API: %w", err)
			errList = append(errList, gqlerror.Wrap(wErr))
			continue
		}

		edges[i] = &gmodel.VehicleEdge{
			Node:   gv,
			Cursor: helpers.IDToCursor(dv.ID),
		}

		nodes[i] = gv
	}

	res := &gmodel.VehicleConnection{
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

// GetVehicles godoc
// @Description gets devices for an owner address
// @Param addr [common.Address] "eth address of owner"
// @Param first [*int] "the number of devices to return per page"
// @Param after [*string] "base64 string representing a device tokenID. This is a pointer to where we start fetching devices from on each page"
// @Param last [*int] "the number of devices to return from previous pages"
// @Param before [*string] "base64 string representing a device tokenID. Pointer to where we start fetching devices from previous pages"
func (r *Repository) GetVehicles(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *gmodel.VehiclesFilter) (*gmodel.VehicleConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	var totalCount int64
	queryMods, err := r.queryModsFromFilters(filterBy)
	if err != nil {
		return nil, err
	}
	if filterBy != nil && filterBy.Privileged != nil {
		totalCount, err = models.Vehicles(
			// We're performing this because SQLBoiler doesn't understand DISTINCT ON. If we use
			// the original version of queryMods the entire SELECT clause will be replaced by
			// SELECT COUNT(*), and we'll probably over-count the number of vehicles.
			append([]qm.QueryMod{qm.Distinct(models.VehicleTableColumns.ID)}, queryMods[1:]...)...,
		).Count(ctx, r.PDB.DBS().Reader)
		if err != nil {
			return nil, err
		}
	} else {
		totalCount, err = models.Vehicles(queryMods...).Count(ctx, r.PDB.DBS().Reader)
		if err != nil {
			return nil, err
		}
	}

	if after != nil {
		afterID, err := helpers.CursorToID(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.VehicleWhere.ID.LT(afterID))
	}

	if before != nil {
		beforeID, err := helpers.CursorToID(*before)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.VehicleWhere.ID.GT(beforeID))
	}

	orderBy := "DESC"
	if last != nil {
		orderBy = "ASC"
	}

	queryMods = append(queryMods,
		// Use limit + 1 here to check if there's another page.
		qm.Limit(limit+1),
		qm.OrderBy(models.VehicleColumns.ID+" "+orderBy),
	)

	all, err := models.Vehicles(queryMods...).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
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

	return r.createVehiclesResponse(totalCount, all, hasNext, hasPrevious)
}

func (r *Repository) GetVehicle(ctx context.Context, tokenID *int, tokenDID *string) (*gmodel.Vehicle, error) {
	if base.CountTrue(tokenID != nil, tokenDID != nil) != 1 {
		return nil, fmt.Errorf("provide exactly one of `tokenID` or `tokenDID`")
	}

	var id int
	switch {
	case tokenID != nil:
		id = *tokenID
	case tokenDID != nil:
		did, err := cloudevent.DecodeERC721DID(*tokenDID)
		if err != nil {
			return nil, fmt.Errorf("error decoding tokenDID: %w", err)
		}
		if !did.TokenID.IsInt64() {
			return nil, fmt.Errorf("token id is too large")
		}
		id = int(did.TokenID.Int64())
	default:
		return nil, fmt.Errorf("invalid filter")
	}
	v, err := models.FindVehicle(ctx, r.PDB.DBS().Reader, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}
	var imageURI string

	if v.ImageURI.Valid {
		imageURI = v.ImageURI.String
	} else {
		var err error
		imageURI, err = DefaultImageURI(r.Settings.BaseImageURL, v.ID)
		if err != nil {
			wErr := fmt.Errorf("error getting vehicle image url: %w", err)
			return nil, gqlerror.Wrap(wErr)
		}
	}

	dataURI, err := GetVehicleDataURI(r.Settings.BaseVehicleDataURI, v.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting vehicle data uri: %w", err)
	}

	return r.ToAPI(v, imageURI, dataURI, nil)
}

// queryModsFromFilters returns a slice of query mods from the given filters.
func (r *Repository) queryModsFromFilters(filter *gmodel.VehiclesFilter) ([]qm.QueryMod, error) {
	var queryMods []qm.QueryMod
	if filter == nil {
		return queryMods, nil
	}

	// To maintain correct count behavior the privilege filter must be the first filter added to the query.
	if filter.Privileged != nil {
		addr := *filter.Privileged
		queryMods = append(queryMods,
			// SELECT DISTINCT ON (vehicles.id) identity_api.vehicles.*
			// LEFT OUTER JOIN identity_api.privileges ON vehicles.id = privileges.token_id
			// WHERE vehicles.owner_address = <filter.Privileged> OR (privileges.user_address = <filter.Privileged> AND privileges.expires_at >= <time.Now()> )
			qm.Select("DISTINCT ON ("+models.VehicleTableColumns.ID+") "+helpers.WithSchema(models.TableNames.Vehicles)+".*"),
			qm.LeftOuterJoin(
				helpers.WithSchema(models.TableNames.Privileges)+" ON "+models.VehicleTableColumns.ID+" = "+models.PrivilegeTableColumns.TokenID,
			),
			qm.LeftOuterJoin(
				helpers.WithSchema(models.TableNames.VehicleSacds)+" ON "+models.VehicleTableColumns.ID+" = "+models.VehicleSacdColumns.VehicleID,
			),
			qm.Expr(
				models.VehicleWhere.OwnerAddress.EQ(addr.Bytes()),
				qm.Or2(
					qm.Expr(
						models.PrivilegeWhere.UserAddress.EQ(addr.Bytes()),
						models.PrivilegeWhere.ExpiresAt.GT(time.Now()),
					),
				),
				qm.Or2(
					qm.Expr(
						models.VehicleSacdWhere.Grantee.EQ(addr.Bytes()),
						models.VehicleSacdWhere.ExpiresAt.GT(time.Now()),
					),
				),
			),
		)
	}

	if filter.Make != nil {
		queryMods = append(queryMods, models.VehicleWhere.Make.EQ(null.StringFrom(*filter.Make)))
	}
	if filter.Model != nil {
		queryMods = append(queryMods, models.VehicleWhere.Model.EQ(null.StringFrom(*filter.Model)))
	}
	if filter.Year != nil {
		queryMods = append(queryMods, models.VehicleWhere.Year.EQ(null.IntFrom(*filter.Year)))
	}
	if filter.Owner != nil {
		queryMods = append(queryMods, models.VehicleWhere.OwnerAddress.EQ(filter.Owner.Bytes()))
	}
	if filter.ManufacturerTokenID != nil {
		queryMods = append(queryMods, models.VehicleWhere.ManufacturerID.EQ(*filter.ManufacturerTokenID))
	}
	if filter.DeviceDefinitionID != nil {
		queryMods = append(queryMods, models.VehicleWhere.DeviceDefinitionID.EQ(null.StringFrom(*filter.DeviceDefinitionID)))
	}

	return queryMods, nil
}

// ToAPI converts a vehicle to a corresponding graphql model.
func (r *Repository) ToAPI(v *models.Vehicle, imageURI string, dataURI string, definition *gmodel.DeviceDefinition) (*gmodel.Vehicle, error) {
	nameList := mnemonic.FromInt32WithObfuscation(int32(v.ID))
	name := strings.Join(nameList, " ")

	globalID, err := base.EncodeGlobalTokenID(TokenPrefix, v.ID)
	if err != nil {
		return nil, fmt.Errorf("error encoding vehicle id: %w", err)
	}

	tokenDID := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.contractAddress,
		TokenID:         new(big.Int).SetInt64(int64(v.ID)),
	}.String()

	out := &gmodel.Vehicle{
		ID:       globalID,
		TokenID:  v.ID,
		TokenDID: tokenDID,
		Owner:    common.BytesToAddress(v.OwnerAddress),
		MintedAt: v.MintedAt,
		Definition: &gmodel.Definition{
			ID:    v.DeviceDefinitionID.Ptr(),
			Make:  v.Make.Ptr(),
			Model: v.Model.Ptr(),
			Year:  v.Year.Ptr(),
		},
		ManufacturerID: v.ManufacturerID,
		Name:           name,
		ImageURI:       imageURI,
		Image:          imageURI,
		DataURI:        dataURI,
	}
	if definition != nil {
		out.Definition.Make = &definition.Manufacturer.Name
		out.Definition.Model = &definition.Model
		out.Definition.Year = &definition.Year
	}

	if v.StorageNodeID.Valid {
		out.StorageNodeID = v.StorageNodeID.Bytes
	}

	return out, nil
}

// DefaultImageURI craates a URL for the vehicle image.
func DefaultImageURI(baseURL string, tokenID int) (string, error) {
	tokenStr := strconv.Itoa(tokenID)
	return url.JoinPath(baseURL, "vehicle", tokenStr, "image")
}

func GetVehicleDataURI(baseURL string, tokenID int) (string, error) {
	tokenStr := strconv.Itoa(tokenID)
	return url.JoinPath(baseURL, tokenStr)
}

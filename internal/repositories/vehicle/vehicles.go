package vehicle

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/KevinJoiner/mnemonic"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
)

type Repository struct {
	*repositories.Repository
}

type vehiclePrimaryKey struct {
	TokenID int
}

func VehicleToAPI(v *models.Vehicle, imageUrl string, dataURI string) *gmodel.Vehicle {
	var b bytes.Buffer
	e := msgpack.NewEncoder(&b)
	e.UseArrayEncodedStructs(true)

	_ = e.Encode(vehiclePrimaryKey{TokenID: v.ID})
	nameList := mnemonic.FromInt64WithObfuscation(int64(v.ID))
	name := strings.Join(nameList, " ")

	return &gmodel.Vehicle{
		ID:       "V_" + base64.StdEncoding.EncodeToString(b.Bytes()),
		TokenID:  v.ID,
		Owner:    common.BytesToAddress(v.OwnerAddress),
		MintedAt: v.MintedAt,
		Definition: &gmodel.Definition{
			URI:   v.DefinitionURI.Ptr(),
			Make:  v.Make.Ptr(),
			Model: v.Model.Ptr(),
			Year:  v.Year.Ptr(),
		},
		ManufacturerID: v.ManufacturerID.Ptr(),
		Name:           name,
		Image:          imageUrl,
		DataURI:        dataURI,
	}
}

func VehicleIDToToken(id string) (int, error) {
	if !strings.HasPrefix(id, "V_") {
		return 0, errors.New("id lacks the V_ prefix")
	}

	id = strings.TrimPrefix(id, "V_")

	b, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return 0, err
	}

	var pk vehiclePrimaryKey
	d := msgpack.NewDecoder(bytes.NewBuffer(b))
	if err := d.Decode(&pk); err != nil {
		return 0, fmt.Errorf("error decoding vehicle id: %w", err)
	}

	return pk.TokenID, nil
}

func (v *Repository) createVehiclesResponse(totalCount int64, vehicles models.VehicleSlice, hasNext bool, hasPrevious bool) *gmodel.VehicleConnection {
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
		imageUrl := helpers.GetVehicleImageUrl(v.Settings.BaseImageURL, dv.ID)
		dataURI := helpers.GetVehicleDataURI(v.Settings.BaseVehicleDataURI, dv.ID)
		gv := VehicleToAPI(dv, imageUrl, dataURI)

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

	return res
}

// GetVehicles godoc
// @Description gets devices for an owner address
// @Param addr [common.Address] "eth address of owner"
// @Param first [*int] "the number of devices to return per page"
// @Param after [*string] "base64 string representing a device tokenID. This is a pointer to where we start fetching devices from on each page"
// @Param last [*int] "the number of devices to return from previous pages"
// @Param before [*string] "base64 string representing a device tokenID. Pointer to where we start fetching devices from previous pages"
func (v *Repository) GetVehicles(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *gmodel.VehiclesFilter) (*gmodel.VehicleConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, repositories.MaxPageSize)
	if err != nil {
		return nil, err
	}

	var totalCount int64
	queryMods := queryModsFromFilters(filterBy)
	if filterBy != nil && filterBy.Privileged != nil {
		totalCount, err = models.Vehicles(
			// We're performing this because SQLBoiler doesn't understand DISTINCT ON. If we use
			// the original version of queryMods the entire SELECT clause will be replaced by
			// SELECT COUNT(*), and we'll probably over-count the number of vehicles.
			append([]qm.QueryMod{qm.Distinct(models.VehicleTableColumns.ID)}, queryMods[1:]...)...,
		).Count(ctx, v.PDB.DBS().Reader)
		if err != nil {
			return nil, err
		}
	} else {
		totalCount, err = models.Vehicles(queryMods...).Count(ctx, v.PDB.DBS().Reader)
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

	all, err := models.Vehicles(queryMods...).All(ctx, v.PDB.DBS().Reader)
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

	return v.createVehiclesResponse(totalCount, all, hasNext, hasPrevious), nil
}

func (r *Repository) GetVehicle(ctx context.Context, id int) (*gmodel.Vehicle, error) {
	v, err := models.FindVehicle(ctx, r.PDB.DBS().Reader, id)
	if err != nil {
		return nil, err
	}
	imageUrl := helpers.GetVehicleImageUrl(r.Settings.BaseImageURL, v.ID)
	dataURI := helpers.GetVehicleDataURI(r.Settings.BaseVehicleDataURI, v.ID)
	return VehicleToAPI(v, imageUrl, dataURI), nil
}

// queryModsFromFilters returns a slice of query mods from the given filters.
func queryModsFromFilters(filter *gmodel.VehiclesFilter) []qm.QueryMod {
	var queryMods []qm.QueryMod
	if filter == nil {
		return queryMods
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
			qm.Expr(
				models.VehicleWhere.OwnerAddress.EQ(addr.Bytes()),
				qm.Or2(
					qm.Expr(
						models.PrivilegeWhere.UserAddress.EQ(addr.Bytes()),
						models.PrivilegeWhere.ExpiresAt.GTE(time.Now()),
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

	return queryMods
}

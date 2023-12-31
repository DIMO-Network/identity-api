package repositories

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
)

const (
	defaultPageSize = 20
)

type Repository struct {
	pdb      db.Store
	settings config.Settings
}

func New(pdb db.Store, settings config.Settings) *Repository {
	return &Repository{
		pdb:      pdb,
		settings: settings,
	}
}

type vehiclePrimaryKey struct {
	TokenID int
}

func VehicleToAPI(v *models.Vehicle, imageUrl string) *gmodel.Vehicle {
	var b bytes.Buffer
	e := msgpack.NewEncoder(&b)
	e.UseArrayEncodedStructs(true)

	_ = e.Encode(vehiclePrimaryKey{TokenID: v.ID})

	bid := helpers.IntToBytes(v.ID)
	name, _ := helpers.CreateMnemonic(bid)

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
		imageUrl := helpers.GetVehicleImageUrl(v.settings.BaseImageURL, dv.ID)
		gv := VehicleToAPI(dv, imageUrl)

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

const maxPageSize = 100

// GetVehicles godoc
// @Description gets devices for an owner address
// @Param addr [common.Address] "eth address of owner"
// @Param first [*int] "the number of devices to return per page"
// @Param after [*string] "base64 string representing a device tokenID. This is a pointer to where we start fetching devices from on each page"
// @Param last [*int] "the number of devices to return from previous pages"
// @Param before [*string] "base64 string representing a device tokenID. Pointer to where we start fetching devices from previous pages"
func (v *Repository) GetVehicles(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *gmodel.VehiclesFilter) (*gmodel.VehicleConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, maxPageSize)
	if err != nil {
		return nil, err
	}

	var totalCount int64
	var queryMods []qm.QueryMod
	if filterBy != nil && filterBy.Privileged != nil {
		addr := *filterBy.Privileged
		queryMods = []qm.QueryMod{
			qm.Select("DISTINCT ON (" + models.VehicleTableColumns.ID + ") " + helpers.WithSchema(models.TableNames.Vehicles) + ".*"),
			qm.LeftOuterJoin(
				helpers.WithSchema(models.TableNames.Privileges) + " ON " + models.VehicleTableColumns.ID + " = " + models.PrivilegeTableColumns.TokenID,
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
		}

		if filterBy.Owner != nil {
			queryMods = append(queryMods, models.VehicleWhere.OwnerAddress.EQ(filterBy.Owner.Bytes()))
		}

		totalCount, err = models.Vehicles(
			// We're performing this because SQLBoiler doesn't understand DISTINCT ON. If we use
			// the original version of queryMods the entire SELECT clause will be replaced by
			// SELECT COUNT(*), and we'll probably over-count the number of vehicles.
			append([]qm.QueryMod{qm.Distinct(models.VehicleTableColumns.ID)}, queryMods[1:]...)...,
		).Count(ctx, v.pdb.DBS().Reader)
		if err != nil {
			return nil, err
		}
	} else {
		if filterBy != nil && filterBy.Owner != nil {
			queryMods = append(queryMods,
				models.VehicleWhere.OwnerAddress.EQ(filterBy.Owner.Bytes()),
			)
		}

		totalCount, err = models.Vehicles(queryMods...).Count(ctx, v.pdb.DBS().Reader)
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

	all, err := models.Vehicles(queryMods...).All(ctx, v.pdb.DBS().Reader)
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
	v, err := models.FindVehicle(ctx, r.pdb.DBS().Reader, id)
	if err != nil {
		return nil, err
	}
	imageUrl := helpers.GetVehicleImageUrl(r.settings.BaseImageURL, v.ID)
	return VehicleToAPI(v, imageUrl), nil
}

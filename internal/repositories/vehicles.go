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
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	defaultPageSize = 20
)

type Repository struct {
	pdb db.Store
}

func New(pdb db.Store) *Repository {
	return &Repository{
		pdb: pdb,
	}
}

type vehiclePrimaryKey struct {
	TokenID int
}

func VehicleToAPI(v *models.Vehicle) *gmodel.Vehicle {
	var b bytes.Buffer
	e := msgpack.NewEncoder(&b)
	e.UseArrayEncodedStructs(true)

	_ = e.Encode(vehiclePrimaryKey{TokenID: v.ID})

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
	}
}

func VehicleIDToToken(id string) (int, error) {
	if !strings.HasPrefix(id, "V_") {
		return 0, errors.New("id lacks the V_ prefix")
	}

	id = id[2:]

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

func (v *Repository) createVehiclesResponse(totalCount int64, vehicles models.VehicleSlice, hasNext bool) *gmodel.VehicleConnection {
	endCursr := helpers.IDToCursor(vehicles[len(vehicles)-1].ID)

	var vEdges []*gmodel.VehicleEdge
	for _, v := range vehicles {
		edge := &gmodel.VehicleEdge{
			Node:   VehicleToAPI(v),
			Cursor: helpers.IDToCursor(v.ID),
		}

		vEdges = append(vEdges, edge)
	}

	res := &gmodel.VehicleConnection{
		TotalCount: int(totalCount),
		PageInfo: &gmodel.PageInfo{
			HasNextPage: hasNext,
			EndCursor:   &endCursr,
		},
		Edges: vEdges,
	}

	return res
}

// GetAccessibleVehicles godoc
// @Description gets devices for an owner address
// @Param addr [common.Address] "eth address of owner"
// @Param first [*int] "the number of devices to return per page"
// @Param after [*string] "base64 string representing a device tokenID. This is a pointer to where we start fetching devices from on each page"
// @Param last [*int] "the number of devices to return from previous pages"
// @Param before [*string] "base64 string representing a device tokenID. Pointer to where we start fetching devices from previous pages"
func (v *Repository) GetAccessibleVehicles(ctx context.Context, addr common.Address, first *int, after *string, last *int, before *string) (*gmodel.VehicleConnection, error) {
	limit := defaultPageSize

	if first != nil {
		limit = *first
	} else if last != nil {
		limit = *last
	}

	queryMods := []qm.QueryMod{
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
		// Use limit + 1 here to check if there's a next page.
	}

	totalCount, err := models.Vehicles(
		// We're performing this because SQLBoiler doesn't understand DISTINCT ON. If we use
		// the original version of queryMods the entire SELECT clause will be replaced by
		// SELECT COUNT(*), and we'll probably over-count the number of vehicles.
		append([]qm.QueryMod{qm.Distinct(models.VehicleTableColumns.ID)}, queryMods[1:]...)...,
	).Count(ctx, v.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if totalCount == 0 {
		return &gmodel.VehicleConnection{
			TotalCount: 0,
			Edges:      []*gmodel.VehicleEdge{},
			PageInfo:   &gmodel.PageInfo{},
		}, nil
	}

	if after != nil {
		afterID, err := helpers.CursorToID(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.VehicleWhere.ID.LT(afterID))
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(models.VehicleColumns.ID+" DESC"),
	)

	all, err := models.Vehicles(queryMods...).All(ctx, v.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if len(all) == 0 {
		return &gmodel.VehicleConnection{
			TotalCount: int(totalCount),
			Edges:      []*gmodel.VehicleEdge{},
		}, nil
	}

	hasNext := len(all) > limit
	if hasNext {
		all = all[:limit]
	}

	return v.createVehiclesResponse(totalCount, all, hasNext), nil
}

func (r *Repository) GetVehicle(ctx context.Context, id int) (*gmodel.Vehicle, error) {
	v, err := models.FindVehicle(ctx, r.pdb.DBS().Reader, id)
	if err != nil {
		return nil, err
	}

	return VehicleToAPI(v), nil
}

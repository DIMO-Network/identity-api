package repositories

import (
	"context"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
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

func VehicleToAPI(v *models.Vehicle) *gmodel.Vehicle {
	return &gmodel.Vehicle{
		ID:       v.ID,
		Owner:    common.BytesToAddress(v.OwnerAddress),
		Make:     v.Make.Ptr(),
		Model:    v.Model.Ptr(),
		Year:     v.Year.Ptr(),
		MintedAt: v.MintedAt,
	}
}

func (v *Repository) createVehiclesResponse(totalCount int64, vehicles models.VehicleSlice, hasNext bool, hasPrevious bool) *gmodel.VehicleConnection {
	endCursr, startCursr := "", ""
	if hasNext {
		endCursr = helpers.IDToCursor(vehicles[len(vehicles)-1].ID)
	}
	if hasPrevious {
		startCursr = helpers.IDToCursor(vehicles[0].ID)
	}

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
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     &startCursr,
			EndCursor:       &endCursr,
		},
		Edges: vEdges,
	}

	return res
}

func applyPaginationDirectionFromCursor(cursor *string, isAfter bool, queryMods *[]qm.QueryMod) error {
	id, err := helpers.CursorToID(*cursor)
	if err != nil {
		return err
	}

	if isAfter {
		*queryMods = append(*queryMods, models.VehicleWhere.ID.LT(id))
	} else {
		*queryMods = append(*queryMods, models.VehicleWhere.ID.GT(id))
	}

	return nil
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

	if totalCount < int64(limit) {
		limit = int(totalCount)
	}

	if after != nil {
		if err := applyPaginationDirectionFromCursor(after, true, &queryMods); err != nil {
			return nil, err
		}
	} else if before != nil {
		if err := applyPaginationDirectionFromCursor(before, false, &queryMods); err != nil {
			return nil, err
		}
	}

	orderBy := "DESC"
	if before != nil {
		orderBy = "ASC"
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(models.VehicleColumns.ID+" "+orderBy),
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

	hasNext, hasPrevious := before != nil, after != nil
	if first != nil && *first+1 == len(all) {
		hasNext = true
		all = all[:limit]
	} else if last != nil && *last+1 == len(all) {
		hasPrevious = true
		all = all[:limit]
	}

	return v.createVehiclesResponse(totalCount, all, hasNext, hasPrevious), nil
}

func (r *Repository) GetVehicle(ctx context.Context, id int) (*gmodel.Vehicle, error) {
	v, err := models.FindVehicle(ctx, r.pdb.DBS().Reader, id)
	if err != nil {
		return nil, err
	}

	return VehicleToAPI(v), nil
}

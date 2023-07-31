package repositories

import (
	"context"
	"errors"
	"strconv"

	"encoding/base64"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	defaultPageSize = 20
)

type VehiclesRepo struct {
	pdb db.Store
}

func NewVehiclesRepo(pdb db.Store) VehiclesRepo {
	return VehiclesRepo{
		pdb: pdb,
	}
}

func (v *VehiclesRepo) createVehiclesResponse(totalCount int64, vehicles models.VehicleSlice, hasNext bool) *gmodel.VehicleConnection {
	lastItmID := vehicles[len(vehicles)-1].ID
	endCursr := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(lastItmID)))

	var vEdges []*gmodel.VehicleEdge
	for _, v := range vehicles {
		crs := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(v.ID)))
		cursor := crs

		owner := common.BytesToAddress(v.OwnerAddress)
		privs := []*gmodel.Privilege{}

		for _, p := range v.R.TokenPrivileges {
			privs = append(privs, &gmodel.Privilege{
				ID:        p.PrivilegeID,
				User:      common.BytesToAddress(p.UserAddress),
				SetAt:     p.SetAt,
				ExpiresAt: p.ExpiresAt,
			})
		}

		edge := &gmodel.VehicleEdge{
			Node: &gmodel.Vehicle{
				ID:         v.ID,
				Owner:      owner,
				Make:       v.Make.Ptr(),
				Model:      v.Model.Ptr(),
				Year:       v.Year.Ptr(),
				MintedAt:   v.MintedAt,
				Privileges: privs,
			},
			Cursor: cursor,
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

func (v *VehiclesRepo) GetAccessibleVehicles(ctx context.Context, addr common.Address, first *int, after *string) (*gmodel.VehicleConnection, error) {
	limit := defaultPageSize
	if first != nil {
		limit = *first
		if limit <= 0 {
			return nil, errors.New("invalid value provided for number of vehicles to retrieve")
		}
	}

	vAlias := "identity_api." + models.TableNames.Vehicles
	pAlias := "identity_api." + models.TableNames.Privileges
	queryMods := []qm.QueryMod{
		qm.Select("DISTINCT ON (" + vAlias + ".id) " + vAlias + ".*"),
		qm.LeftOuterJoin(
			pAlias + " ON " + models.VehicleTableColumns.ID + " = " + models.PrivilegeTableColumns.TokenID,
		),
		qm.Or2(models.VehicleWhere.OwnerAddress.EQ(addr.Bytes())),
		qm.Or2(models.PrivilegeWhere.UserAddress.EQ(addr.Bytes())),
		// Use limit + 1 here to check if there's a next page.
	}

	if after != nil {
		lastCursor, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return nil, err
		}

		lastCursorVal, err := strconv.Atoi(string(lastCursor))
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods, models.VehicleWhere.ID.LT(lastCursorVal))
	}

	totalCount, err := models.Vehicles(queryMods...).Count(ctx, v.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if totalCount == 0 {
		return &gmodel.VehicleConnection{
			TotalCount: 0,
			Edges:      []*gmodel.VehicleEdge{},
		}, nil
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(models.VehicleColumns.ID+" DESC"),
		qm.Load(models.VehicleRels.TokenPrivileges))

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

	vIterate := all
	if len(all) > limit {
		vIterate = all[:limit]
	}

	return v.createVehiclesResponse(totalCount, vIterate, len(all) > limit), nil
}

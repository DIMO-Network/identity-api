package repositories

import (
	"context"
	"errors"
	"fmt"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func (r *Repository) GetOwnedAftermarketDevices(ctx context.Context, addr common.Address, first *int, after *string) (*gmodel.AftermarketDeviceConnection, error) {
	ownedADCount, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.Owner.EQ(addr.Bytes()),
	).Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	limit := r.PageSize
	if first != nil {
		if *first < 1 {
			return nil, errors.New("invalid pagination parameter provided")
		}
		limit = *first
	}

	if ownedADCount == 0 {
		return &gmodel.AftermarketDeviceConnection{
			TotalCount: int(ownedADCount),
			Edges:      []*gmodel.AftermarketDeviceEdge{},
			PageInfo:   &gmodel.PageInfo{},
		}, nil
	}

	queryMods := []qm.QueryMod{
		models.AftermarketDeviceWhere.Owner.EQ(addr.Bytes()),
		// Use limit + 1 here to check if there's a next page.
		qm.Limit(limit + 1),
		qm.OrderBy(models.AftermarketDeviceColumns.ID + " DESC"),
	}

	if after != nil {
		afterID, err := helpers.CursorToID(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor %q", *after)
		}

		queryMods = append(queryMods, models.AftermarketDeviceWhere.ID.LT(afterID))
	}

	ads, err := models.AftermarketDevices(queryMods...).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	hasNextPage := len(ads) > limit
	if hasNextPage {
		ads = ads[:limit]
	}

	var adEdges []*gmodel.AftermarketDeviceEdge
	for _, d := range ads {
		adEdges = append(adEdges,
			&gmodel.AftermarketDeviceEdge{
				Node: &gmodel.AftermarketDevice{
					ID:        d.ID,
					Address:   helpers.BytesToAddr(d.Address),
					Owner:     common.BytesToAddress(d.Owner),
					Serial:    d.Serial.Ptr(),
					IMEI:      d.Imei.Ptr(),
					VehicleID: d.VehicleID.Ptr(),
					MintedAt:  d.MintedAt,
				},
				Cursor: helpers.IDToCursor(d.ID),
			},
		)
	}

	res := &gmodel.AftermarketDeviceConnection{
		TotalCount: int(ownedADCount),
		Edges:      adEdges,
		PageInfo: &gmodel.PageInfo{
			HasNextPage: hasNextPage,
		},
	}

	if len(ads) == 0 {
		return res, nil
	}

	res.PageInfo.EndCursor = &adEdges[len(adEdges)-1].Cursor
	return res, nil
}

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

func (r *Repository) GetOwnedAftermarketDevices(ctx context.Context, addr common.Address, first *int, after *string, last *int, before *string) (*gmodel.AftermarketDeviceConnection, error) {
	where := []qm.QueryMod{
		models.AftermarketDeviceWhere.Owner.EQ(addr.Bytes()),
	}

	ownedADCount, err := models.AftermarketDevices(where...).Count(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	limit := defaultPageSize
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

	queryMods := append(where,
		// Use limit + 1 here to check if there's a next page.
		qm.Limit(limit+1),
		qm.OrderBy(models.AftermarketDeviceColumns.ID+" DESC"),
	)

	if after != nil {
		afterID, err := helpers.CursorToID(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor %q", *after)
		}

		queryMods = append(queryMods, models.AftermarketDeviceWhere.ID.LT(afterID))
	}

	ads, err := models.AftermarketDevices(queryMods...).All(ctx, r.pdb.DBS().Reader)
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
				Node:   AftermarketDeviceToAPI(d),
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

func AftermarketDeviceToAPI(d *models.AftermarketDevice) *gmodel.AftermarketDevice {
	return &gmodel.AftermarketDevice{
		ID:          d.ID,
		Address:     helpers.BytesToAddr(d.Address),
		Owner:       common.BytesToAddress(d.Owner),
		Serial:      d.Serial.Ptr(),
		IMEI:        d.Imei.Ptr(),
		Beneficiary: common.BytesToAddress(d.Beneficiary),
		VehicleID:   d.VehicleID.Ptr(),
		MintedAt:    d.MintedAt,
	}
}

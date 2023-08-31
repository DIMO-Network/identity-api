package repositories

import (
	"context"
	"errors"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
)

// GetOwnedAftermarketDevices godoc
// @Description gets aftermarket devices for an owner address
// @Param addr [common.Address] "eth address of owner"
// @Param first [*int] "the number of devices to return per page"
// @Param after [*string] "base64 string representing a device tokenID. This is a pointer to where we start fetching devices from on each page"
// @Param last [*int] "the number of devices to return from previous pages"
// @Param before [*string] "base64 string representing a device tokenID. Pointer to where we start fetching devices from previous pages"
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
		limit = *first
	} else if last != nil {
		limit = *last
	}
	if limit < 1 {
		return nil, errors.New("invalid pagination parameter provided")
	}

	if ownedADCount == 0 {
		return &gmodel.AftermarketDeviceConnection{
			TotalCount: int(ownedADCount),
			Edges:      []*gmodel.AftermarketDeviceEdge{},
			PageInfo:   &gmodel.PageInfo{},
		}, nil
	}

	orderBy := " DESC"
	if before != nil {
		orderBy = " ASC"
	}

	queryMods := append(where,
		// Use limit + 1 here to check if there's a next page.
		qm.Limit(limit+1),
		qm.OrderBy(models.AftermarketDeviceColumns.ID+orderBy),
	)

	if after != nil {
		if err := applyPaginationDirectionFromCursor(*after, true, &queryMods); err != nil {
			return nil, err
		}
	} else if before != nil {
		if err := applyPaginationDirectionFromCursor(*before, false, &queryMods); err != nil {
			return nil, err
		}
	}

	ads, err := models.AftermarketDevices(queryMods...).All(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	hasNextPage, hasPreviousPage := before != nil, after != nil
	if first != nil && *first+1 == len(ads) {
		hasNextPage = true
		ads = ads[:limit]
	} else if last != nil && *last+1 == len(ads) {
		hasPreviousPage = true
		ads = ads[:limit]
	}

	if before != nil {
		slices.Reverse(ads)
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

	endCursr, startCursr := helpers.IDToCursor(ads[len(ads)-1].ID), helpers.IDToCursor(ads[0].ID)

	res := &gmodel.AftermarketDeviceConnection{
		TotalCount: int(ownedADCount),
		Edges:      adEdges,
		PageInfo: &gmodel.PageInfo{
			StartCursor:     &startCursr,
			EndCursor:       &endCursr,
			HasNextPage:     hasNextPage,
			HasPreviousPage: hasPreviousPage,
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

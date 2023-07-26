package repositories

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func BytesToAddr(addrB null.Bytes) *common.Address {
	var addr *common.Address
	if addrB.Valid {
		addr = (*common.Address)(*addrB.Ptr())
	}
	return addr
}

func (v *VehiclesRepo) GetOwnedAftermarketDevices(ctx context.Context, addr common.Address, first *int, after *string) (*gmodel.AftermarketDeviceConnection, error) {
	ownedADCount, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.Owner.EQ(null.BytesFrom(addr.Bytes())),
	).Count(ctx, v.pdb.DBS().Reader)
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

	queryMods := []qm.QueryMod{
		models.AftermarketDeviceWhere.Owner.EQ(null.BytesFrom(addr.Bytes())),
		// Use limit + 1 here to check if there's a next page.
		qm.Limit(limit + 1),
		qm.OrderBy(models.AftermarketDeviceColumns.ID + " DESC"),
	}

	if after != nil {
		sB, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return nil, err
		}

		searchAfter, err := strconv.Atoi(string(sB))
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.AftermarketDeviceWhere.ID.LT(searchAfter))
	}

	ads, err := models.AftermarketDevices(queryMods...).All(ctx, v.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	hasNextPage := len(ads) > limit
	if hasNextPage {
		ads = ads[:len(ads)-1]
	}

	var adEdges []*gmodel.AftermarketDeviceEdge
	for _, d := range ads {
		adEdges = append(adEdges,
			&gmodel.AftermarketDeviceEdge{
				Node: &gmodel.AftermarketDevice{
					ID:       strconv.Itoa(d.ID),
					Address:  BytesToAddr(d.Address),
					Owner:    common.BytesToAddress(d.Owner.Bytes),
					Serial:   d.Serial.Ptr(),
					Imei:     d.Imei.Ptr(),
					MintedAt: d.MintedAt.Time,
				},
				Cursor: base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(d.ID))),
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

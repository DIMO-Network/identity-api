package controllers

import (
	"context"
	"errors"
	"strconv"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	defaultPageSize = 20
)

type AftermarketDevicesCtrl struct {
	ctx context.Context
	pdb db.Store
}

func NewADRepo(ctx context.Context, pdb db.Store) AftermarketDevicesCtrl {
	return AftermarketDevicesCtrl{
		ctx: ctx,
		pdb: pdb,
	}
}

func (ad *AftermarketDevicesCtrl) GetOwnedAftermarketDevices(addr common.Address, first *int, after *string) (*gmodel.AftermarketDeviceConnection, error) {
	ownedADCount, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.Owner.EQ(addr.Bytes()),
	).Count(context.Background(), ad.pdb.DBS().Reader)
	limit := defaultPageSize
	if first != nil {
		if *first == 0 {
			return nil, errors.New("invalid pagination parameter provided")
		}
		limit = *first
	}

	if ownedADCount == 0 {
		return &gmodel.AftermarketDeviceConnection{
			TotalCount: 0,
			Edges:      []*gmodel.AftermarketDeviceEdge{},
		}, nil
	}

	queryMods := []qm.QueryMod{
		models.AftermarketDeviceWhere.Owner.EQ(addr.Bytes()),
		qm.Limit(limit + 1),
		qm.OrderBy(models.AftermarketDeviceColumns.ID + " ASC"),
	}

	if after != nil {
		searchAfter, err := strconv.Atoi(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.AftermarketDeviceWhere.ID.GT(searchAfter))
	}

	ads, err := models.AftermarketDevices(queryMods...).All(ad.ctx, ad.pdb.DBS().Reader)
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
					ID:        strconv.Itoa(d.ID),
					Address:   common.BytesToAddress(d.Owner),
					Owner:     addr,
					Serial:    &d.Serial.String,
					Imei:      &d.Imei.String,
					MintedAt:  d.MintedAt,
					VehicleID: strconv.Itoa(d.VehicleID.Int),
				},
				Cursor: strconv.Itoa(d.ID),
			},
		)
	}

	res := &gmodel.AftermarketDeviceConnection{
		TotalCount: int(ownedADCount),
		PageInfo: &gmodel.PageInfo{
			HasNextPage: hasNextPage,
			EndCursor:   adEdges[len(adEdges)-1].Node.ID,
			StartCursor: adEdges[0].Node.ID,
		},
		Edges: adEdges,
	}

	return res, nil
}

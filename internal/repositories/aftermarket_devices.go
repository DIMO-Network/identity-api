package repositories

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"strconv"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func (r *Repository) GetOwnedAftermarketDevices(ctx context.Context, addr common.Address, first *int, after *string) (*gmodel.AftermarketDeviceConnection, error) {
	ownedADCount, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.Owner.EQ(null.BytesFrom(addr.Bytes())),
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
					Owner:     helpers.BytesToAddr(d.Owner),
					Serial:    d.Serial.Ptr(),
					IMEI:      d.Imei.Ptr(),
					VehicleID: d.VehicleID.Ptr(),
					MintedAt:  d.MintedAt.Ptr(),
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

func (r *Repository) GetLinkedAftermarketDeviceByVehicleID(ctx context.Context, vehicleID string) (*gmodel.AftermarketDevice, error) {
	vID, err := strconv.Atoi(vehicleID)
	if err != nil {
		return nil, err
	}

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.VehicleID.EQ(null.IntFrom(vID)),
	).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	res := &gmodel.AftermarketDevice{
		ID:       ad.ID,
		Address:  helpers.BytesToAddr(ad.Address),
		Owner:    helpers.BytesToAddr(ad.Address),
		Serial:   ad.Serial.Ptr(),
		IMEI:     ad.Imei.Ptr(),
		MintedAt: ad.MintedAt.Ptr(),
	}

	return res, nil
}

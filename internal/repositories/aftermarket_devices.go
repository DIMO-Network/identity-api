package repositories

import (
	"context"
	"database/sql"
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
		qm.Load(models.AftermarketDeviceRels.Vehicle),
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
		ads = ads[:limit]
	}

	var adEdges []*gmodel.AftermarketDeviceEdge
	for _, d := range ads {
		var vehicle gmodel.Vehicle
		if d.R.Vehicle != nil {
			vehicle.ID = strconv.Itoa(d.R.Vehicle.ID)
			vehicle.Owner = *BytesToAddr(d.R.Vehicle.OwnerAddress)
			vehicle.Make = d.R.Vehicle.Make.Ptr()
			vehicle.Model = d.R.Vehicle.Model.Ptr()
			vehicle.Year = d.R.Vehicle.Year.Ptr()
			vehicle.MintedAt = d.R.Vehicle.MintedAt.Time
		}

		adEdges = append(adEdges,
			&gmodel.AftermarketDeviceEdge{
				Node: &gmodel.AftermarketDevice{
					ID:          strconv.Itoa(d.ID),
					Address:     BytesToAddr(d.Address),
					Owner:       BytesToAddr(d.Owner),
					Serial:      d.Serial.Ptr(),
					Imei:        d.Imei.Ptr(),
					MintedAt:    d.MintedAt.Ptr(),
					Vehicle:     &vehicle,
					Beneficiary: BytesToAddr(d.Beneficiary),
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

func (v *VehiclesRepo) GetLinkedAftermarketDeviceByVehicleID(ctx context.Context, vehicleID string) (*gmodel.AftermarketDevice, error) {
	vID, err := strconv.Atoi(vehicleID)
	if err != nil {
		return nil, err
	}

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.VehicleID.EQ(null.IntFrom(vID)),
	).One(ctx, v.pdb.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	res := &gmodel.AftermarketDevice{
		ID:          strconv.Itoa(ad.ID),
		Address:     BytesToAddr(ad.Address),
		Owner:       BytesToAddr(ad.Address),
		Serial:      ad.Serial.Ptr(),
		Imei:        ad.Imei.Ptr(),
		MintedAt:    ad.MintedAt.Ptr(),
		Beneficiary: BytesToAddr(ad.Beneficiary),
	}

	return res, nil
}

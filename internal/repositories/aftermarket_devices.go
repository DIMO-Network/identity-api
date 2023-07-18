package repositories

import (
	"context"
	"errors"
	"strconv"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func (v *VehiclesRepo) GetOwnedAftermarketDevices(ctx context.Context, addr common.Address, first *int, after *string) (*gmodel.AftermarketDeviceConnection, error) {
	ownedADCount, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.Owner.EQ(null.BytesFrom(addr.Bytes())),
	).Count(context.Background(), v.pdb.DBS().Reader)
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
		models.AftermarketDeviceWhere.Owner.EQ(null.BytesFrom(addr.Bytes())),
		qm.Load(models.AftermarketDeviceRels.Vehicle),
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

		vConn := gmodel.Vehicle{}
		if d.R.Vehicle != nil {
			vConn.ID = strconv.Itoa(d.R.Vehicle.ID)
			vConn.Owner = &addr
			vConn.Make = d.R.Vehicle.Make.Ptr()
			vConn.Model = d.R.Vehicle.Model.Ptr()
			vConn.Year = d.R.Vehicle.Year.Ptr()
			vConn.MintedAt = d.R.Vehicle.MintedAt.Ptr()
		}

		adEdges = append(adEdges,
			&gmodel.AftermarketDeviceEdge{
				Node: &gmodel.AftermarketDevice{
					ID:                strconv.Itoa(d.ID),
					Address:           common.BytesToAddress(d.Owner.Bytes),
					Owner:             addr,
					Serial:            &d.Serial.String,
					Imei:              &d.Imei.String,
					MintedAt:          d.MintedAt.Time,
					VehicleID:         strconv.Itoa(d.VehicleID.Int),
					VehicleConnection: &vConn,
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

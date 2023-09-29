package repositories

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/volatiletech/null/v8"
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
func (r *Repository) GetAftermarketDevices(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *gmodel.AftermarketDevicesFilter) (*gmodel.AftermarketDeviceConnection, error) {
	var limit int

	if first != nil {
		if last != nil {
			return nil, errors.New("Pass `first` or `last`, but not both.")
		}
		if *first < 0 {
			return nil, errors.New("The value for `first` cannot be negative.")
		}
		if *first > maxPageSize {
			return nil, fmt.Errorf("The value %d for `first` exceeds the limit %d.", *last, maxPageSize)
		}
		limit = *first
	} else {
		if last == nil {
			return nil, errors.New("Provide `first` or `last`.")
		}
		if *last < 0 {
			return nil, errors.New("The value for `last` cannot be negative.")
		}
		if *last > maxPageSize {
			return nil, fmt.Errorf("The value %d for `last` exceeds the limit %d.", *last, maxPageSize)
		}
		limit = *last
	}

	where := []qm.QueryMod{}

	if filterBy != nil && filterBy.Owner != nil {
		where = append(where, models.AftermarketDeviceWhere.Owner.EQ(filterBy.Owner.Bytes()))
	}

	adCount, err := models.AftermarketDevices(where...).Count(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
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
		afterID, err := helpers.CursorToID(*after)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.AftermarketDeviceWhere.ID.LT(afterID))
	} else if before != nil {
		beforeID, err := helpers.CursorToID(*before)
		if err != nil {
			return nil, err
		}

		queryMods = append(queryMods, models.AftermarketDeviceWhere.ID.GT(beforeID))
	}

	all, err := models.AftermarketDevices(queryMods...).All(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	hasNext := before != nil
	hasPrevious := after != nil

	if first != nil && len(all) == limit+1 {
		hasNext = true
		all = all[:limit]
	} else if last != nil && len(all) == limit+1 {
		hasPrevious = true
		all = all[:limit]
	}

	if before != nil {
		slices.Reverse(all)
	}

	var adEdges []*gmodel.AftermarketDeviceEdge
	for _, d := range all {
		adEdges = append(adEdges,
			&gmodel.AftermarketDeviceEdge{
				Node:   AftermarketDeviceToAPI(d),
				Cursor: helpers.IDToCursor(d.ID),
			},
		)
	}

	var endCur, startCur *string

	if len(all) != 0 {
		ec := helpers.IDToCursor(all[len(all)-1].ID)
		endCur = &ec

		sc := helpers.IDToCursor(all[0].ID)
		startCur = &sc
	}

	res := &gmodel.AftermarketDeviceConnection{
		TotalCount: int(adCount),
		Edges:      adEdges,
		PageInfo: &gmodel.PageInfo{
			StartCursor:     startCur,
			EndCursor:       endCur,
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
		},
	}

	return res, nil
}

func (r *Repository) GetAftermarketDevice(ctx context.Context, by gmodel.AftermarketDeviceBy) (*gmodel.AftermarketDevice, error) {
	if countTrue(by.Address != nil, by.ID != nil, by.Serial != nil) != 1 {
		return nil, errors.New("Pass in exactly one of `address`, `id`, or `serial`.")
	}

	var qm qm.QueryMod

	switch {
	case by.Address != nil:
		qm = models.AftermarketDeviceWhere.Address.EQ(by.Address.Bytes())
	case by.ID != nil:
		qm = models.AftermarketDeviceWhere.ID.EQ(*by.ID)
	case by.Serial != nil:
		qm = models.AftermarketDeviceWhere.Serial.EQ(null.StringFrom(*by.Serial))
	}

	ad, err := models.AftermarketDevices(qm).One(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return AftermarketDeviceToAPI(ad), nil
}

type aftermarketDevicePrimaryKey struct {
	TokenID int
}

func AftermarketDeviceToAPI(d *models.AftermarketDevice) *gmodel.AftermarketDevice {
	var b bytes.Buffer
	e := msgpack.NewEncoder(&b)
	e.UseArrayEncodedStructs(true)

	_ = e.Encode(aftermarketDevicePrimaryKey{TokenID: d.ID})

	return &gmodel.AftermarketDevice{
		ID:          "AD_" + base64.StdEncoding.EncodeToString(b.Bytes()),
		TokenID:     d.ID,
		Address:     common.BytesToAddress(d.Address),
		Owner:       common.BytesToAddress(d.Owner),
		Serial:      d.Serial.Ptr(),
		Imei:        d.Imei.Ptr(),
		Beneficiary: common.BytesToAddress(d.Beneficiary),
		VehicleID:   d.VehicleID.Ptr(),
		MintedAt:    d.MintedAt,
	}
}

func countTrue(ps ...bool) int {
	out := 0

	for _, p := range ps {
		if p {
			out++
		}
	}

	return out
}

package repositories

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
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
			return nil, gqlerror.Errorf("Pass `first` or `last`, but not both.")
		}
		if *first < 0 {
			return nil, gqlerror.Errorf("The value for `first` cannot be negative.")
		}
		if *first > maxPageSize {
			return nil, gqlerror.Errorf("The value %d for `first` exceeds the limit %d.", *last, maxPageSize)
		}
		limit = *first
	} else {
		if last == nil {
			return nil, gqlerror.Errorf("Provide `first` or `last`.")
		}
		if *last < 0 {
			return nil, gqlerror.Errorf("The value for `last` cannot be negative.")
		}
		if *last > maxPageSize {
			return nil, gqlerror.Errorf("The value %d for `last` exceeds the limit %d.", *last, maxPageSize)
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
	if last != nil {
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

	if last != nil {
		slices.Reverse(all)
	}

	edges := make([]*gmodel.AftermarketDeviceEdge, len(all))
	nodes := make([]*gmodel.AftermarketDevice, len(all))

	for i, da := range all {
		imageUrl := helpers.GetAfterMarketDeviceImageUrl(r.settings.BaseImageURL, da.ID)
		ga := AftermarketDeviceToAPI(da, imageUrl)

		edges[i] = &gmodel.AftermarketDeviceEdge{
			Node:   ga,
			Cursor: helpers.IDToCursor(da.ID),
		}

		nodes[i] = ga
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
		Edges:      edges,
		Nodes:      nodes,
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
	if countTrue(by.Address != nil, by.TokenID != nil, by.Serial != nil) != 1 {
		return nil, gqlerror.Errorf("Pass in exactly one of `address`, `id`, or `serial`.")
	}

	var qm qm.QueryMod

	switch {
	case by.Address != nil:
		qm = models.AftermarketDeviceWhere.Address.EQ(by.Address.Bytes())
	case by.TokenID != nil:
		qm = models.AftermarketDeviceWhere.ID.EQ(*by.TokenID)
	case by.Serial != nil:
		qm = models.AftermarketDeviceWhere.Serial.EQ(null.StringFrom(*by.Serial))
	}

	ad, err := models.AftermarketDevices(qm).One(ctx, r.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	imageUrl := helpers.GetAfterMarketDeviceImageUrl(r.settings.BaseImageURL, ad.ID)
	return AftermarketDeviceToAPI(ad, imageUrl), nil
}

type aftermarketDevicePrimaryKey struct {
	TokenID int
}

func AftermarketDeviceToAPI(d *models.AftermarketDevice, imageUrl string) *gmodel.AftermarketDevice {
	var b bytes.Buffer
	e := msgpack.NewEncoder(&b)
	e.UseArrayEncodedStructs(true)

	_ = e.Encode(aftermarketDevicePrimaryKey{TokenID: d.ID})

	name, _ := helpers.CreateMnemonic(d.Address)

	return &gmodel.AftermarketDevice{
		ID:             "AD_" + base64.StdEncoding.EncodeToString(b.Bytes()),
		TokenID:        d.ID,
		Address:        common.BytesToAddress(d.Address),
		Owner:          common.BytesToAddress(d.Owner),
		Serial:         d.Serial.Ptr(),
		Imei:           d.Imei.Ptr(),
		Beneficiary:    common.BytesToAddress(d.Beneficiary),
		VehicleID:      d.VehicleID.Ptr(),
		MintedAt:       d.MintedAt,
		ClaimedAt:      d.ClaimedAt.Ptr(),
		ManufacturerID: d.ManufacturerID.Ptr(),
		Name:           name,
		Image:          imageUrl,
	}
}

func AftermarketDeviceIDToToken(id string) (int, error) {
	if !strings.HasPrefix(id, "AD_") {
		return 0, errors.New("id lacks the AD_ prefix")
	}

	id = strings.TrimPrefix(id, "AD_")

	b, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return 0, err
	}

	var pk aftermarketDevicePrimaryKey
	d := msgpack.NewDecoder(bytes.NewBuffer(b))
	if err := d.Decode(&pk); err != nil {
		return 0, fmt.Errorf("error decoding vehicle id: %w", err)
	}

	return pk.TokenID, nil
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

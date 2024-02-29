package aftermarket

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/DIMO-Network/identity-api/graph/model"
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/mnemonic"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
)

type Repository struct {
	*repositories.Repository
}

// GetOwnedAftermarketDevices godoc
// @Description gets aftermarket devices for an owner address
// @Param addr [common.Address] "eth address of owner"
// @Param first [*int] "the number of devices to return per page"
// @Param after [*string] "base64 string representing a device tokenID. This is a pointer to where we start fetching devices from on each page"
// @Param last [*int] "the number of devices to return from previous pages"
// @Param before [*string] "base64 string representing a device tokenID. Pointer to where we start fetching devices from previous pages"
func (r *Repository) GetAftermarketDevices(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *gmodel.AftermarketDevicesFilter) (*gmodel.AftermarketDeviceConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, repositories.MaxPageSize)
	if err != nil {
		return nil, err
	}

	where := []qm.QueryMod{}

	if filterBy != nil {
		if filterBy.Owner != nil {
			where = append(where, models.AftermarketDeviceWhere.Owner.EQ(filterBy.Owner.Bytes()))
		}
		if filterBy.Beneficiary != nil {
			where = append(where, models.AftermarketDeviceWhere.Beneficiary.EQ(filterBy.Beneficiary.Bytes()))
		}
		if filterBy.ManufacturerID != nil {
			where = append(where, models.AftermarketDeviceWhere.ManufacturerID.EQ(null.IntFrom(*filterBy.ManufacturerID)))
		}
	}

	adCount, err := models.AftermarketDevices(where...).Count(ctx, r.PDB.DBS().Reader)
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

	all, err := models.AftermarketDevices(queryMods...).All(ctx, r.PDB.DBS().Reader)
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
		imageUrl := helpers.GetAftermarketDeviceImageUrl(r.Settings.BaseImageURL, da.ID)
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
	if repositories.CountTrue(by.Address != nil, by.TokenID != nil, by.Serial != nil) != 1 {
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

	ad, err := models.AftermarketDevices(qm).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	imageUrl := helpers.GetAftermarketDeviceImageUrl(r.Settings.BaseImageURL, ad.ID)
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

	nameList := mnemonic.FromInt32WithObfuscation(int32(d.ID))
	name := strings.Join(nameList, " ")

	return &gmodel.AftermarketDevice{
		ID:             "AD_" + base64.StdEncoding.EncodeToString(b.Bytes()),
		TokenID:        d.ID,
		Address:        common.BytesToAddress(d.Address),
		Owner:          common.BytesToAddress(d.Owner),
		Serial:         d.Serial.Ptr(),
		Imei:           d.Imei.Ptr(),
		DevEui:         d.DevEui.Ptr(),
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

func (r *Repository) GetAftermarketDevicesForManufacturer(ctx context.Context, obj *model.Manufacturer, first *int, after *string, last *int, before *string, filterBy *model.AftermarketDevicesFilter) (*gmodel.AftermarketDeviceConnection, error) {
	if filterBy != nil {
		if filterBy.ManufacturerID != nil {
			if filterBy.ManufacturerID != &obj.TokenID {
				return nil, gqlerror.Errorf("Aftermarket device filter must be consistent with manufacturer query.")
			}
		}
		filterBy.ManufacturerID = &obj.TokenID
		return r.GetAftermarketDevices(ctx, first, after, last, before, filterBy)
	}

	filterBy = &model.AftermarketDevicesFilter{
		ManufacturerID: &obj.TokenID,
	}
	return r.GetAftermarketDevices(ctx, first, after, last, before, filterBy)
}

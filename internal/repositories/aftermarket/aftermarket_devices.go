package aftermarket

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/DIMO-Network/cloudevent"
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/mnemonic"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// TokenPrefix is the prefix for a global token id for aftermarket devices.
const TokenPrefix = "AD"

type Repository struct {
	*base.Repository
	chainID         uint64
	contractAddress common.Address
}

// New creates a new aftermarket device repository.
func New(db *base.Repository) *Repository {
	return &Repository{
		Repository:      db,
		chainID:         uint64(db.Settings.DIMORegistryChainID),
		contractAddress: common.HexToAddress(db.Settings.AftermarketDeviceAddr),
	}
}

// GetOwnedAftermarketDevices godoc
// @Description gets aftermarket devices for an owner address
// @Param addr [common.Address] "eth address of owner"
// @Param first [*int] "the number of devices to return per page"
// @Param after [*string] "base64 string representing a device tokenID. This is a pointer to where we start fetching devices from on each page"
// @Param last [*int] "the number of devices to return from previous pages"
// @Param before [*string] "base64 string representing a device tokenID. Pointer to where we start fetching devices from previous pages"
func (r *Repository) GetAftermarketDevices(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *gmodel.AftermarketDevicesFilter) (*gmodel.AftermarketDeviceConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
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
			where = append(where, models.AftermarketDeviceWhere.ManufacturerID.EQ(*filterBy.ManufacturerID))
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
	var errList gqlerror.List
	for i, da := range all {
		imageURL, err := GetAftermarketDeviceImageURL(r.Settings.BaseImageURL, da.ID)
		if err != nil {
			errList = append(errList, gqlerror.Errorf("error getting aftermarket device image url: %v", err))
			continue
		}
		ga, err := r.ToAPI(da, imageURL)
		if err != nil {
			errList = append(errList, gqlerror.Errorf("error converting aftermarket device to API: %v", err))
			continue
		}

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
	if errList != nil {
		return res, errList
	}
	return res, nil
}

func (r *Repository) GetAftermarketDevice(ctx context.Context, by gmodel.AftermarketDeviceBy) (*gmodel.AftermarketDevice, error) {
	if base.CountTrue(by.Address != nil, by.TokenID != nil, by.TokenDID != nil, by.Serial != nil, by.Imei != nil, by.DevEui != nil) != 1 {
		return nil, gqlerror.Errorf("Pass in exactly one of `address`, `tokenId`, `tokenDid`, `serial`, `imei` or `devEUI`.")
	}

	var qm qm.QueryMod

	switch {
	case by.Address != nil:
		qm = models.AftermarketDeviceWhere.Address.EQ(by.Address.Bytes())
	case by.TokenID != nil:
		qm = models.AftermarketDeviceWhere.ID.EQ(*by.TokenID)
	case by.TokenDID != nil:
		did, err := cloudevent.DecodeERC721DID(*by.TokenDID)
		if err != nil {
			return nil, fmt.Errorf("error decoding token did: %w", err)
		}
		if did.ChainID != r.chainID {
			return nil, fmt.Errorf("unknown chain id %d in token did", did.ChainID)
		}
		if did.ContractAddress != r.contractAddress {
			return nil, fmt.Errorf("invalid contract address '%s' in token did", did.ContractAddress.Hex())
		}
		qm = models.AftermarketDeviceWhere.ID.EQ(int(did.TokenID.Int64()))
	case by.Serial != nil:
		qm = models.AftermarketDeviceWhere.Serial.EQ(null.StringFrom(*by.Serial))
	case by.Imei != nil:
		qm = models.AftermarketDeviceWhere.Imei.EQ(null.StringFrom(*by.Imei))
	case by.DevEui != nil:
		qm = models.AftermarketDeviceWhere.DevEui.EQ(null.StringFrom(*by.DevEui))
	default:
		return nil, fmt.Errorf("invalid filter")
	}

	ad, err := models.AftermarketDevices(qm).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	imageURL, err := GetAftermarketDeviceImageURL(r.Settings.BaseImageURL, ad.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image url: %w", err)
	}

	return r.ToAPI(ad, imageURL)
}

type aftermarketDevicePrimaryKey struct {
	TokenID int
}

// ToAPI converts an aftermarket device to a corresponding API aftermarket device.
func (r *Repository) ToAPI(d *models.AftermarketDevice, imageURL string) (*gmodel.AftermarketDevice, error) {
	globalID, err := base.EncodeGlobalTokenID(TokenPrefix, d.ID)
	if err != nil {
		return nil, fmt.Errorf("error encoding vehicle id: %w", err)
	}

	nameList := mnemonic.FromInt32WithObfuscation(int32(d.ID))
	name := strings.Join(nameList, " ")

	tokenDid := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.contractAddress,
		TokenID:         new(big.Int).SetUint64(uint64(d.ID)),
	}.String()

	addressDid := cloudevent.EthrDID{
		ChainID:         r.chainID,
		ContractAddress: common.BytesToAddress(d.Address),
	}.String()

	ownerDid := cloudevent.EthrDID{
		ChainID:         r.chainID,
		ContractAddress: common.BytesToAddress(d.Owner),
	}.String()

	return &gmodel.AftermarketDevice{
		ID:               globalID,
		TokenID:          d.ID,
		TokenDid:         tokenDid,
		Address:          common.BytesToAddress(d.Address),
		AddressDid:       addressDid,
		Owner:            common.BytesToAddress(d.Owner),
		OwnerDid:         ownerDid,
		Serial:           d.Serial.Ptr(),
		Imei:             d.Imei.Ptr(),
		DevEui:           d.DevEui.Ptr(),
		Beneficiary:      common.BytesToAddress(d.Beneficiary),
		VehicleID:        d.VehicleID.Ptr(),
		MintedAt:         d.MintedAt,
		ClaimedAt:        d.ClaimedAt.Ptr(),
		ManufacturerID:   d.ManufacturerID,
		Name:             name,
		Image:            imageURL,
		HardwareRevision: d.HardwareRevision.Ptr(),
		PairedAt:         d.PairedAt.Ptr(),
	}, nil
}

func GetAftermarketDeviceImageURL(baseURL string, tokenID int) (string, error) {
	tokenStr := strconv.Itoa(tokenID)
	return url.JoinPath(baseURL, "aftermarket", "device", tokenStr, "image")
}

// IDToToken converts token data to a token id.
func IDToToken(b []byte) (int, error) {
	var pk aftermarketDevicePrimaryKey
	d := msgpack.NewDecoder(bytes.NewBuffer(b))
	if err := d.Decode(&pk); err != nil {
		return 0, fmt.Errorf("error decoding vehicle id: %w", err)
	}

	return pk.TokenID, nil
}

func (r *Repository) GetAftermarketDevicesForManufacturer(ctx context.Context, obj *gmodel.Manufacturer, first *int, after *string, last *int, before *string, filterBy *gmodel.AftermarketDevicesFilter) (*gmodel.AftermarketDeviceConnection, error) {
	if filterBy != nil {
		if filterBy.ManufacturerID != nil {
			if filterBy.ManufacturerID != &obj.TokenID {
				return nil, gqlerror.Errorf("Aftermarket device filter must be consistent with manufacturer query.")
			}
		}
		filterBy.ManufacturerID = &obj.TokenID
		return r.GetAftermarketDevices(ctx, first, after, last, before, filterBy)
	}

	filterBy = &gmodel.AftermarketDevicesFilter{
		ManufacturerID: &obj.TokenID,
	}
	return r.GetAftermarketDevices(ctx, first, after, last, before, filterBy)
}

// Package synthetic contains the repository for synthetic devices.
package synthetic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"

	"github.com/DIMO-Network/cloudevent"
	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/mnemonic"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// TokenPrefix is the prefix for a global token id for synthetic devices.
const TokenPrefix = "S"

var vehicleJoin = fmt.Sprintf("%s ON %s = %s", helpers.WithSchema(models.TableNames.Vehicles), models.VehicleTableColumns.ID, models.SyntheticDeviceTableColumns.VehicleID)

// Repository is the repository for synthetic devices.
type Repository struct {
	*base.Repository
	chainID         uint64
	contractAddress common.Address
}

// New creates a new synthetic device repository.
func New(db *base.Repository) *Repository {
	return &Repository{
		Repository:      db,
		chainID:         uint64(db.Settings.DIMORegistryChainID),
		contractAddress: common.HexToAddress(db.Settings.SyntheticDeviceAddr),
	}
}

// ToAPI converts a synthetic device from the database to a GraphQL API model.
func (r *Repository) ToAPI(sd *models.SyntheticDevice) (*gmodel.SyntheticDevice, error) {
	globalID, err := base.EncodeGlobalTokenID(TokenPrefix, sd.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to encode synthetic device primary key: %w", err)
	}

	nameList := mnemonic.FromInt32WithObfuscation(int32(sd.ID))
	name := strings.Join(nameList, " ")

	tokenDID := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.contractAddress,
		TokenID:         new(big.Int).SetUint64(uint64(sd.ID)),
	}.String()

	out := &gmodel.SyntheticDevice{
		ID:            globalID,
		Name:          name,
		TokenID:       sd.ID,
		TokenDID:      tokenDID,
		IntegrationID: sd.IntegrationID,
		Address:       common.BytesToAddress(sd.DeviceAddress),
		MintedAt:      sd.MintedAt,
		VehicleID:     sd.VehicleID,
	}

	if sd.ConnectionID.Valid {
		out.ConnectionID = sd.ConnectionID.Bytes
	}

	return out, err
}

// GetSyntheticDevice Device retrieves a synthetic device by either its address or tokenID from the database.
func (r *Repository) GetSyntheticDevice(ctx context.Context, by gmodel.SyntheticDeviceBy) (*gmodel.SyntheticDevice, error) {
	if base.CountTrue(by.Address != nil, by.TokenID != nil, by.TokenDID != nil) != 1 {
		return nil, gqlerror.Errorf("Pass in exactly one of `address`, `tokenId`, or `tokenDID`.")
	}

	var mod qm.QueryMod

	switch {
	case by.Address != nil:
		mod = models.SyntheticDeviceWhere.DeviceAddress.EQ(by.Address.Bytes())
	case by.TokenID != nil:
		mod = models.SyntheticDeviceWhere.ID.EQ(*by.TokenID)
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
		if !did.TokenID.IsInt64() {
			return nil, fmt.Errorf("token id is too large")
		}
		mod = models.SyntheticDeviceWhere.ID.EQ(int(did.TokenID.Int64()))
	}

	synth, err := models.SyntheticDevices(mod).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		r.Log.Error().Err(err).Msg("Failed to find synthetic device.")
		return nil, base.InternalError
	}

	return r.ToAPI(synth)
}

// GetSyntheticDevices retrieves a list of synthetic devices from the database.
func (r *Repository) GetSyntheticDevices(ctx context.Context, first *int, last *int, after *string, before *string, filterBy *gmodel.SyntheticDevicesFilter) (*gmodel.SyntheticDeviceConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, fmt.Errorf("invalid first/last argument: %w", err)
	}

	var totalCount int64
	queryMods := queryModsFromFilters(filterBy)
	totalCount, err = models.SyntheticDevices(queryMods...).Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		r.Log.Err(err).Msg("failed to get synthetic device count")
		return nil, base.InternalError
	}

	if after != nil {
		afterID, err := helpers.CursorToID(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid after cursor: %w", err)
		}
		queryMods = append(queryMods, models.SyntheticDeviceWhere.ID.LT(afterID))
	}

	if before != nil {
		beforeID, err := helpers.CursorToID(*before)
		if err != nil {
			return nil, fmt.Errorf("invalid before cursor: %w", err)
		}
		queryMods = append(queryMods, models.SyntheticDeviceWhere.ID.GT(beforeID))
	}

	orderBy := " DESC"
	if last != nil {
		orderBy = " ASC"
	}

	queryMods = append(queryMods,
		// Use limit + 1 here to check if there's another page.
		qm.Limit(limit+1),
		qm.OrderBy(models.SyntheticDeviceColumns.ID+orderBy),
	)

	all, err := models.SyntheticDevices(queryMods...).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		r.Log.Err(err).Msg("failed to get synthetic devices")
		return nil, base.InternalError
	}

	// We assume that cursors come from real elements.
	hasNext := before != nil
	hasPrevious := after != nil

	if first != nil && len(all) == limit+1 {
		hasNext = true
		all = all[:limit]
	} else if last != nil && len(all) == limit+1 {
		hasPrevious = true
		all = all[:limit]
	}

	// reverse the order if we're using last so tokenID is in descending order
	if last != nil {
		slices.Reverse(all)
	}

	return r.createSyntheticDevicesResponse(totalCount, all, hasNext, hasPrevious)
}

func queryModsFromFilters(filterBy *gmodel.SyntheticDevicesFilter) []qm.QueryMod {
	if filterBy == nil {
		return nil
	}

	var where []qm.QueryMod
	if filterBy.IntegrationID != nil {
		where = append(where, models.SyntheticDeviceWhere.IntegrationID.EQ(*filterBy.IntegrationID))
	}
	if filterBy.Owner != nil {
		// join with the SyntheticDevice table to get the owner
		vehicleJoin := qm.InnerJoin(vehicleJoin)
		ownerWhere := models.VehicleWhere.OwnerAddress.EQ(filterBy.Owner.Bytes())
		where = append(where, vehicleJoin, ownerWhere)
	}

	return where
}

func (r *Repository) createSyntheticDevicesResponse(totalCount int64, syntheticDevices models.SyntheticDeviceSlice, hasNext bool, hasPrevious bool) (*gmodel.SyntheticDeviceConnection, error) {
	var errList gqlerror.List
	var endCur, startCur *string
	if len(syntheticDevices) != 0 {
		ec := helpers.IDToCursor(syntheticDevices[len(syntheticDevices)-1].ID)
		endCur = &ec

		sc := helpers.IDToCursor(syntheticDevices[0].ID)
		startCur = &sc
	}

	edges := make([]*gmodel.SyntheticDeviceEdge, len(syntheticDevices))
	nodes := make([]*gmodel.SyntheticDevice, len(syntheticDevices))

	for i, synth := range syntheticDevices {
		synthAPI, err := r.ToAPI(synth)
		if err != nil {
			errList = append(errList, gqlerror.Errorf("failed to convert synthetic device to API: %v", err))
			continue
		}
		edges[i] = &gmodel.SyntheticDeviceEdge{
			Node:   synthAPI,
			Cursor: helpers.IDToCursor(synth.ID),
		}
		nodes[i] = synthAPI
	}

	res := &gmodel.SyntheticDeviceConnection{
		Edges: edges,
		Nodes: nodes,
		PageInfo: &gmodel.PageInfo{
			EndCursor:       endCur,
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCur,
		},
		TotalCount: int(totalCount),
	}

	if errList != nil {
		return res, errList
	}
	return res, nil
}

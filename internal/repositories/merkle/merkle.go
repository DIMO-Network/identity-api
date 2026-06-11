// Package merkle contains the repository for MerkleDistributor pools, roots,
// and claims.
package merkle

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"slices"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
)

// Repository is the data access layer for MerkleDistributor pools, epochs,
// and rewards.
type Repository struct {
	*base.Repository
}

// RewardCursor is the pagination cursor for the merkleRewards query.
type RewardCursor struct {
	PoolID int
	Epoch  int
}

var rewardCursorColumns = "(" + models.MerkleClaimColumns.PoolID + ", " + models.MerkleClaimColumns.Epoch + ")"

func decimalToBig(d types.Decimal) *big.Int {
	return d.Int(nil)
}

// PoolToAPI converts a database Merkle pool row to its GraphQL form.
func PoolToAPI(pool *models.MerklePool) *gmodel.MerklePool {
	out := &gmodel.MerklePool{
		PoolID:    pool.PoolID,
		Token:     common.BytesToAddress(pool.Token),
		Admin:     common.BytesToAddress(pool.Admin),
		Balance:   decimalToBig(pool.Balance),
		CreatedAt: pool.CreatedAt,
	}

	if pool.WeeklyLimit.Big != nil {
		out.WeeklyLimit = pool.WeeklyLimit.Int(nil)
	}

	return out
}

// EpochToAPI converts a database Merkle root row to its GraphQL form.
func EpochToAPI(root *models.MerkleRoot) *gmodel.MerkleEpoch {
	return &gmodel.MerkleEpoch{
		Epoch:          root.Epoch,
		Root:           root.Root,
		Allocation:     decimalToBig(root.Allocation),
		TotalClaimed:   decimalToBig(root.TotalClaimed),
		ClaimCount:     root.ClaimCount,
		RecipientCount: root.RecipientCount,
		ProofsURI:      root.ProofsURI,
		SetAt:          root.SetAt,
	}
}

// RewardToAPI converts a database Merkle claim row to its GraphQL form.
func RewardToAPI(claim *models.MerkleClaim) (*gmodel.MerkleReward, error) {
	var proof []string
	if err := json.Unmarshal(claim.Proof, &proof); err != nil {
		return nil, fmt.Errorf("parsing proof for pool %d, epoch %d: %w", claim.PoolID, claim.Epoch, err)
	}
	if proof == nil {
		proof = []string{}
	}

	out := &gmodel.MerkleReward{
		PoolID:    claim.PoolID,
		Epoch:     claim.Epoch,
		Account:   common.BytesToAddress(claim.Account),
		Amount:    decimalToBig(claim.Amount),
		Proof:     proof,
		Claimed:   claim.ClaimedAt.Valid,
		ClaimedAt: claim.ClaimedAt.Ptr(),
	}

	if claim.ClaimTX.Valid {
		out.ClaimTx = claim.ClaimTX.Bytes
	}

	return out, nil
}

// GetMerklePool retrieves a single pool by its id. It returns nil if the pool
// does not exist.
func (r *Repository) GetMerklePool(ctx context.Context, poolID int) (*gmodel.MerklePool, error) {
	pool, err := models.FindMerklePool(ctx, r.PDB.DBS().Reader, poolID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return PoolToAPI(pool), nil
}

// GetMerklePools retrieves a page of pools, sorted by pool id, descending.
func (r *Repository) GetMerklePools(ctx context.Context, first *int, after *string, last *int, before *string) (*gmodel.MerklePoolConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	totalCount, err := models.MerklePools().Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	var queryMods []qm.QueryMod

	if after != nil {
		afterID, err := helpers.CursorToID(*after)
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods, models.MerklePoolWhere.PoolID.LT(afterID))
	}

	if before != nil {
		beforeID, err := helpers.CursorToID(*before)
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods, models.MerklePoolWhere.PoolID.GT(beforeID))
	}

	orderBy := "DESC"
	if last != nil {
		orderBy = "ASC"
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(models.MerklePoolColumns.PoolID+" "+orderBy),
	)

	all, err := models.MerklePools(queryMods...).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
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

	if last != nil {
		slices.Reverse(all)
	}

	edges := make([]*gmodel.MerklePoolEdge, len(all))
	nodes := make([]*gmodel.MerklePool, len(all))

	for i, pool := range all {
		node := PoolToAPI(pool)

		edges[i] = &gmodel.MerklePoolEdge{
			Node:   node,
			Cursor: helpers.IDToCursor(pool.PoolID),
		}
		nodes[i] = node
	}

	var endCur, startCur *string
	if len(all) != 0 {
		endCur = &edges[len(edges)-1].Cursor
		startCur = &edges[0].Cursor
	}

	return &gmodel.MerklePoolConnection{
		TotalCount: int(totalCount),
		Edges:      edges,
		Nodes:      nodes,
		PageInfo: &gmodel.PageInfo{
			StartCursor:     startCur,
			EndCursor:       endCur,
			HasPreviousPage: hasPrevious,
			HasNextPage:     hasNext,
		},
	}, nil
}

// GetPoolEpochs retrieves a page of epochs for a pool, sorted by epoch,
// descending.
func (r *Repository) GetPoolEpochs(ctx context.Context, obj *gmodel.MerklePool, first *int, after *string, last *int, before *string) (*gmodel.MerkleEpochConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{
		models.MerkleRootWhere.PoolID.EQ(obj.PoolID),
	}

	totalCount, err := models.MerkleRoots(queryMods...).Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	if after != nil {
		afterEpoch, err := helpers.CursorToID(*after)
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods, models.MerkleRootWhere.Epoch.LT(afterEpoch))
	}

	if before != nil {
		beforeEpoch, err := helpers.CursorToID(*before)
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods, models.MerkleRootWhere.Epoch.GT(beforeEpoch))
	}

	orderBy := "DESC"
	if last != nil {
		orderBy = "ASC"
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(models.MerkleRootColumns.Epoch+" "+orderBy),
	)

	all, err := models.MerkleRoots(queryMods...).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
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

	if last != nil {
		slices.Reverse(all)
	}

	edges := make([]*gmodel.MerkleEpochEdge, len(all))
	nodes := make([]*gmodel.MerkleEpoch, len(all))

	for i, root := range all {
		node := EpochToAPI(root)

		edges[i] = &gmodel.MerkleEpochEdge{
			Node:   node,
			Cursor: helpers.IDToCursor(root.Epoch),
		}
		nodes[i] = node
	}

	var endCur, startCur *string
	if len(all) != 0 {
		endCur = &edges[len(edges)-1].Cursor
		startCur = &edges[0].Cursor
	}

	return &gmodel.MerkleEpochConnection{
		TotalCount: int(totalCount),
		Edges:      edges,
		Nodes:      nodes,
		PageInfo: &gmodel.PageInfo{
			StartCursor:     startCur,
			EndCursor:       endCur,
			HasPreviousPage: hasPrevious,
			HasNextPage:     hasNext,
		},
	}, nil
}

// GetMerkleRewards retrieves a page of rewards for an account, optionally
// filtered by pool and claim status. Sorts by pool id and then epoch,
// descending.
func (r *Repository) GetMerkleRewards(ctx context.Context, account common.Address, poolID *int, claimed *bool, first *int, after *string, last *int, before *string) (*gmodel.MerkleRewardConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{
		models.MerkleClaimWhere.Account.EQ(account.Bytes()),
	}

	if poolID != nil {
		queryMods = append(queryMods, models.MerkleClaimWhere.PoolID.EQ(*poolID))
	}

	if claimed != nil {
		if *claimed {
			queryMods = append(queryMods, models.MerkleClaimWhere.ClaimedAt.IsNotNull())
		} else {
			queryMods = append(queryMods, models.MerkleClaimWhere.ClaimedAt.IsNull())
		}
	}

	totalCount, err := models.MerkleClaims(queryMods...).Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	cursorHelper := &helpers.PaginationHelper[RewardCursor]{}

	if after != nil {
		afterT, err := cursorHelper.DecodeCursor(*after)
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods, qm.Where(rewardCursorColumns+" < (?, ?)", afterT.PoolID, afterT.Epoch))
	}

	if before != nil {
		beforeT, err := cursorHelper.DecodeCursor(*before)
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods, qm.Where(rewardCursorColumns+" > (?, ?)", beforeT.PoolID, beforeT.Epoch))
	}

	orderBy := " DESC"
	if last != nil {
		orderBy = " ASC"
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(models.MerkleClaimColumns.PoolID+orderBy+", "+models.MerkleClaimColumns.Epoch+orderBy),
	)

	all, err := models.MerkleClaims(queryMods...).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
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

	if last != nil {
		slices.Reverse(all)
	}

	edges := make([]*gmodel.MerkleRewardEdge, len(all))
	nodes := make([]*gmodel.MerkleReward, len(all))

	for i, claim := range all {
		node, err := RewardToAPI(claim)
		if err != nil {
			return nil, err
		}

		cursor, err := cursorHelper.EncodeCursor(RewardCursor{PoolID: claim.PoolID, Epoch: claim.Epoch})
		if err != nil {
			return nil, err
		}

		edges[i] = &gmodel.MerkleRewardEdge{
			Node:   node,
			Cursor: cursor,
		}
		nodes[i] = node
	}

	var endCur, startCur *string
	if len(all) != 0 {
		endCur = &edges[len(edges)-1].Cursor
		startCur = &edges[0].Cursor
	}

	return &gmodel.MerkleRewardConnection{
		TotalCount: int(totalCount),
		Edges:      edges,
		Nodes:      nodes,
		PageInfo: &gmodel.PageInfo{
			StartCursor:     startCur,
			EndCursor:       endCur,
			HasPreviousPage: hasPrevious,
			HasNextPage:     hasNext,
		},
	}, nil
}

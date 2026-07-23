// Package merkle indexes events emitted by the MerkleDistributor contract:
// reward pools, weekly Merkle roots, and individual claims.
package merkle

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/DIMO-Network/identity-api/internal/dbtypes"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	cmodels "github.com/DIMO-Network/identity-api/internal/services/models"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/identity-api/pkg/merkletree"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
)

// Handler processes MerkleDistributor contract events.
type Handler struct {
	DBS     db.Store
	Logger  *zerolog.Logger
	Fetcher TreeFetcher
}

// toInt converts a big.Int event argument to an int, returning an error if
// the value does not fit in an int64.
func toInt(name string, x *big.Int) (int, error) {
	if !x.IsInt64() {
		return 0, fmt.Errorf("event argument %s value %s does not fit in an int64", name, x)
	}
	return int(x.Int64()), nil
}

// Handle routes a MerkleDistributor contract event to the matching handler.
func (h *Handler) Handle(ctx context.Context, event *cmodels.ContractEventData) error {
	switch event.EventName {
	case PoolCreated:
		var args PoolCreatedData
		if err := json.Unmarshal(event.Arguments, &args); err != nil {
			return err
		}
		return h.handlePoolCreated(ctx, event, &args)
	case RootSet:
		var args RootSetData
		if err := json.Unmarshal(event.Arguments, &args); err != nil {
			return err
		}
		return h.handleRootSet(ctx, event, &args)
	case Claimed:
		var args ClaimedData
		if err := json.Unmarshal(event.Arguments, &args); err != nil {
			return err
		}
		return h.handleClaimed(ctx, event, &args)
	case Funded:
		var args FundedData
		if err := json.Unmarshal(event.Arguments, &args); err != nil {
			return err
		}
		return h.handleFunded(ctx, event, &args)
	case Swept:
		var args SweptData
		if err := json.Unmarshal(event.Arguments, &args); err != nil {
			return err
		}
		return h.handleSwept(ctx, event, &args)
	case WeeklyLimitSet:
		var args WeeklyLimitSetData
		if err := json.Unmarshal(event.Arguments, &args); err != nil {
			return err
		}
		return h.handleWeeklyLimitSet(ctx, event, &args)
	}

	return nil
}

func (h *Handler) handlePoolCreated(ctx context.Context, e *cmodels.ContractEventData, args *PoolCreatedData) error {
	poolID, err := toInt("poolId", args.PoolId)
	if err != nil {
		return err
	}

	pool := models.MerklePool{
		PoolID:      poolID,
		Token:       args.Token.Bytes(),
		Admin:       args.Admin.Bytes(),
		WeeklyLimit: dbtypes.NullIntToDecimal(args.WeeklyLimit),
		CreatedAt:   e.Block.Time,
	}

	cols := models.MerklePoolColumns

	// Don't touch the balance on conflict.
	return pool.Upsert(ctx, h.DBS.DBS().Writer, true,
		[]string{cols.PoolID},
		boil.Whitelist(cols.Token, cols.Admin, cols.WeeklyLimit, cols.CreatedAt),
		boil.Infer(),
	)
}

func (h *Handler) handleWeeklyLimitSet(ctx context.Context, e *cmodels.ContractEventData, args *WeeklyLimitSetData) error {
	poolID, err := toInt("poolId", args.PoolId)
	if err != nil {
		return err
	}

	pool := models.MerklePool{
		PoolID:      poolID,
		WeeklyLimit: dbtypes.NullIntToDecimal(args.Limit),
	}

	rowsAff, err := pool.Update(ctx, h.DBS.DBS().Writer, boil.Whitelist(models.MerklePoolColumns.WeeklyLimit))
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return fmt.Errorf("WeeklyLimitSet for unknown pool %d", pool.PoolID)
	}

	return nil
}

func (h *Handler) handleFunded(ctx context.Context, e *cmodels.ContractEventData, args *FundedData) error {
	poolID, err := toInt("poolId", args.PoolId)
	if err != nil {
		return err
	}

	res, err := h.DBS.DBS().Writer.ExecContext(ctx,
		fmt.Sprintf("UPDATE %s SET balance = balance + $1 WHERE pool_id = $2", helpers.WithSchema(models.TableNames.MerklePools)),
		dbtypes.IntToDecimal(args.Amount), poolID,
	)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return fmt.Errorf("funded event for unknown pool %d", args.PoolId)
	}

	return nil
}

func (h *Handler) handleSwept(ctx context.Context, e *cmodels.ContractEventData, args *SweptData) error {
	poolID, err := toInt("poolId", args.PoolId)
	if err != nil {
		return err
	}

	pool := models.MerklePool{
		PoolID:  poolID,
		Balance: dbtypes.IntToDecimal(args.NewBalance),
	}

	rowsAff, err := pool.Update(ctx, h.DBS.DBS().Writer, boil.Whitelist(models.MerklePoolColumns.Balance))
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return fmt.Errorf("swept event for unknown pool %d", pool.PoolID)
	}

	return nil
}

func (h *Handler) handleRootSet(ctx context.Context, e *cmodels.ContractEventData, args *RootSetData) error {
	poolID, err := toInt("poolId", args.PoolId)
	if err != nil {
		return err
	}
	epoch, err := toInt("week", args.Week)
	if err != nil {
		return err
	}

	logger := h.Logger.With().
		Str("EventName", RootSet).
		Str("poolId", args.PoolId.String()).
		Str("week", args.Week.String()).
		Logger()

	fetchStart := time.Now()
	data, err := h.Fetcher.Fetch(ctx, args.ProofsURI)
	if err != nil {
		logger.Err(err).Str("proofsURI", args.ProofsURI).Msg("Failed to fetch Merkle tree file.")
		return fmt.Errorf("fetching tree file: %w", err)
	}
	fetchDuration := time.Since(fetchStart)

	file, err := merkletree.UnmarshalTreeFile(data)
	if err != nil {
		logger.Err(err).Str("proofsURI", args.ProofsURI).Msg("Failed to parse Merkle tree file.")
		return fmt.Errorf("parsing tree file: %w", err)
	}

	if err := file.VerifyRoot(); err != nil {
		logger.Err(err).Str("proofsURI", args.ProofsURI).Msg("Merkle tree file failed root verification.")
		return fmt.Errorf("verifying tree file root: %w", err)
	}

	if file.Root != common.Hash(args.Root) {
		// Returning an error here means the Kafka consumer never commits the
		// offset for this event, so it will be redelivered and fail forever
		// until the file at proofsURI is replaced with one matching the
		// on-chain root. That is deliberate: a published tree file that
		// disagrees with the root the contract accepted must page an
		// operator, and the endless redelivery errors are the alerting
		// mechanism. Skipping the event instead would silently leave the
		// epoch's claims unindexed.
		err := fmt.Errorf("tree file root %s does not match event root %s", file.Root, common.Hash(args.Root))
		logger.Err(err).Str("proofsURI", args.ProofsURI).Msg("Merkle tree file root mismatch.")
		return err
	}

	if file.PoolID.Cmp(args.PoolId) != 0 || file.Week.Cmp(args.Week) != 0 {
		err := fmt.Errorf("tree file is for pool %s, week %s; event is for pool %s, week %s", file.PoolID, file.Week, args.PoolId, args.Week)
		logger.Err(err).Str("proofsURI", args.ProofsURI).Msg("Merkle tree file pool or week mismatch.")
		return err
	}

	if file.Distributor != e.Contract {
		logger.Warn().Msgf("Tree file distributor %s does not match emitting contract %s.", file.Distributor, e.Contract)
	}

	dbWriteStart := time.Now()

	tx, err := h.DBS.DBS().Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	root := models.MerkleRoot{
		PoolID:         poolID,
		Epoch:          epoch,
		Root:           args.Root[:],
		Allocation:     dbtypes.IntToDecimal(args.Allocation),
		RecipientCount: len(file.Leaves),
		ProofsURI:      args.ProofsURI,
		SetAt:          e.Block.Time,
	}

	rootCols := models.MerkleRootColumns

	// Preserve total_claimed and claim_count if the root is set again.
	err = root.Upsert(ctx, tx, true,
		[]string{rootCols.PoolID, rootCols.Epoch},
		boil.Whitelist(rootCols.Root, rootCols.Allocation, rootCols.RecipientCount, rootCols.ProofsURI, rootCols.SetAt),
		boil.Infer(),
	)
	if err != nil {
		return fmt.Errorf("upserting Merkle root: %w", err)
	}

	accounts := make([][]byte, len(file.Leaves))
	amounts := make([]string, len(file.Leaves))
	proofs := make([]string, len(file.Leaves))

	for i, leaf := range file.Leaves {
		proof := make([]string, len(leaf.Proof))
		for j, p := range leaf.Proof {
			proof[j] = p.Hex()
		}
		proofJSON, err := json.Marshal(proof)
		if err != nil {
			return fmt.Errorf("marshaling proof for account %s: %w", leaf.Account, err)
		}

		accounts[i] = leaf.Account.Bytes()
		amounts[i] = leaf.Amount.String()
		proofs[i] = string(proofJSON)
	}

	// Upsert all leaves in a single statement. Only amount and proof are in
	// the SET list, so claimed_at and claim_tx are preserved if the root is
	// set again after some leaves have already been claimed.
	_, err = tx.ExecContext(ctx,
		fmt.Sprintf(`
			INSERT INTO %s (pool_id, epoch, account, amount, proof)
			SELECT $1, $2, unnest($3::bytea[]), unnest($4::numeric[]), unnest($5::jsonb[])
			ON CONFLICT (pool_id, epoch, account)
			DO UPDATE SET amount = EXCLUDED.amount, proof = EXCLUDED.proof`,
			helpers.WithSchema(models.TableNames.MerkleClaims)),
		root.PoolID, root.Epoch, pq.ByteaArray(accounts), pq.Array(amounts), pq.Array(proofs),
	)
	if err != nil {
		return fmt.Errorf("upserting Merkle claims: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Info().
		Int("recipientCount", len(file.Leaves)).
		Int64("fetchDurationMs", fetchDuration.Milliseconds()).
		Int64("dbWriteDurationMs", time.Since(dbWriteStart).Milliseconds()).
		Msg("Merkle root set.")

	return nil
}

func (h *Handler) handleClaimed(ctx context.Context, e *cmodels.ContractEventData, args *ClaimedData) error {
	poolID, err := toInt("poolId", args.PoolId)
	if err != nil {
		return err
	}
	epoch, err := toInt("week", args.Week)
	if err != nil {
		return err
	}

	tx, err := h.DBS.DBS().Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	claim, err := models.MerkleClaims(
		models.MerkleClaimWhere.PoolID.EQ(poolID),
		models.MerkleClaimWhere.Epoch.EQ(epoch),
		models.MerkleClaimWhere.Account.EQ(args.Account.Bytes()),
		qm.For("UPDATE"),
	).One(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("claimed event for unknown leaf: pool %d, epoch %d, account %s", poolID, epoch, args.Account)
		}
		return err
	}

	if claim.ClaimedAt.Valid {
		// Already processed; don't double-count.
		return nil
	}

	claim.ClaimedAt = null.TimeFrom(e.Block.Time)
	claim.ClaimTX = null.BytesFrom(e.TransactionHash.Bytes())

	if _, err := claim.Update(ctx, tx, boil.Whitelist(models.MerkleClaimColumns.ClaimedAt, models.MerkleClaimColumns.ClaimTX)); err != nil {
		return fmt.Errorf("updating Merkle claim: %w", err)
	}

	amount := dbtypes.IntToDecimal(args.Amount)

	if _, err := tx.ExecContext(ctx,
		fmt.Sprintf("UPDATE %s SET total_claimed = total_claimed + $1, claim_count = claim_count + 1 WHERE pool_id = $2 AND epoch = $3", helpers.WithSchema(models.TableNames.MerkleRoots)),
		amount, poolID, epoch,
	); err != nil {
		return fmt.Errorf("updating Merkle root claim totals: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		fmt.Sprintf("UPDATE %s SET balance = balance - $1 WHERE pool_id = $2", helpers.WithSchema(models.TableNames.MerklePools)),
		amount, poolID,
	); err != nil {
		return fmt.Errorf("updating Merkle pool balance: %w", err)
	}

	return tx.Commit()
}

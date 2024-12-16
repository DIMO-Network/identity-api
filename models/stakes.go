// Code generated by SQLBoiler 4.17.1 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/strmangle"
)

// Stake is an object representing the database table.
type Stake struct {
	ID          int           `boil:"id" json:"id" toml:"id" yaml:"id"`
	Owner       []byte        `boil:"owner" json:"owner" toml:"owner" yaml:"owner"`
	Level       int           `boil:"level" json:"level" toml:"level" yaml:"level"`
	Points      int           `boil:"points" json:"points" toml:"points" yaml:"points"`
	Amount      types.Decimal `boil:"amount" json:"amount" toml:"amount" yaml:"amount"`
	VehicleID   null.Int      `boil:"vehicle_id" json:"vehicle_id,omitempty" toml:"vehicle_id" yaml:"vehicle_id,omitempty"`
	StakedAt    time.Time     `boil:"staked_at" json:"staked_at" toml:"staked_at" yaml:"staked_at"`
	EndsAt      time.Time     `boil:"ends_at" json:"ends_at" toml:"ends_at" yaml:"ends_at"`
	WithdrawnAt null.Time     `boil:"withdrawn_at" json:"withdrawn_at,omitempty" toml:"withdrawn_at" yaml:"withdrawn_at,omitempty"`

	R *stakeR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L stakeL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var StakeColumns = struct {
	ID          string
	Owner       string
	Level       string
	Points      string
	Amount      string
	VehicleID   string
	StakedAt    string
	EndsAt      string
	WithdrawnAt string
}{
	ID:          "id",
	Owner:       "owner",
	Level:       "level",
	Points:      "points",
	Amount:      "amount",
	VehicleID:   "vehicle_id",
	StakedAt:    "staked_at",
	EndsAt:      "ends_at",
	WithdrawnAt: "withdrawn_at",
}

var StakeTableColumns = struct {
	ID          string
	Owner       string
	Level       string
	Points      string
	Amount      string
	VehicleID   string
	StakedAt    string
	EndsAt      string
	WithdrawnAt string
}{
	ID:          "stakes.id",
	Owner:       "stakes.owner",
	Level:       "stakes.level",
	Points:      "stakes.points",
	Amount:      "stakes.amount",
	VehicleID:   "stakes.vehicle_id",
	StakedAt:    "stakes.staked_at",
	EndsAt:      "stakes.ends_at",
	WithdrawnAt: "stakes.withdrawn_at",
}

// Generated where

var StakeWhere = struct {
	ID          whereHelperint
	Owner       whereHelper__byte
	Level       whereHelperint
	Points      whereHelperint
	Amount      whereHelpertypes_Decimal
	VehicleID   whereHelpernull_Int
	StakedAt    whereHelpertime_Time
	EndsAt      whereHelpertime_Time
	WithdrawnAt whereHelpernull_Time
}{
	ID:          whereHelperint{field: "\"identity_api\".\"stakes\".\"id\""},
	Owner:       whereHelper__byte{field: "\"identity_api\".\"stakes\".\"owner\""},
	Level:       whereHelperint{field: "\"identity_api\".\"stakes\".\"level\""},
	Points:      whereHelperint{field: "\"identity_api\".\"stakes\".\"points\""},
	Amount:      whereHelpertypes_Decimal{field: "\"identity_api\".\"stakes\".\"amount\""},
	VehicleID:   whereHelpernull_Int{field: "\"identity_api\".\"stakes\".\"vehicle_id\""},
	StakedAt:    whereHelpertime_Time{field: "\"identity_api\".\"stakes\".\"staked_at\""},
	EndsAt:      whereHelpertime_Time{field: "\"identity_api\".\"stakes\".\"ends_at\""},
	WithdrawnAt: whereHelpernull_Time{field: "\"identity_api\".\"stakes\".\"withdrawn_at\""},
}

// StakeRels is where relationship names are stored.
var StakeRels = struct {
	Vehicle string
}{
	Vehicle: "Vehicle",
}

// stakeR is where relationships are stored.
type stakeR struct {
	Vehicle *Vehicle `boil:"Vehicle" json:"Vehicle" toml:"Vehicle" yaml:"Vehicle"`
}

// NewStruct creates a new relationship struct
func (*stakeR) NewStruct() *stakeR {
	return &stakeR{}
}

func (r *stakeR) GetVehicle() *Vehicle {
	if r == nil {
		return nil
	}
	return r.Vehicle
}

// stakeL is where Load methods for each relationship are stored.
type stakeL struct{}

var (
	stakeAllColumns            = []string{"id", "owner", "level", "points", "amount", "vehicle_id", "staked_at", "ends_at", "withdrawn_at"}
	stakeColumnsWithoutDefault = []string{"id", "owner", "level", "points", "amount", "staked_at", "ends_at"}
	stakeColumnsWithDefault    = []string{"vehicle_id", "withdrawn_at"}
	stakePrimaryKeyColumns     = []string{"id"}
	stakeGeneratedColumns      = []string{}
)

type (
	// StakeSlice is an alias for a slice of pointers to Stake.
	// This should almost always be used instead of []Stake.
	StakeSlice []*Stake
	// StakeHook is the signature for custom Stake hook methods
	StakeHook func(context.Context, boil.ContextExecutor, *Stake) error

	stakeQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	stakeType                 = reflect.TypeOf(&Stake{})
	stakeMapping              = queries.MakeStructMapping(stakeType)
	stakePrimaryKeyMapping, _ = queries.BindMapping(stakeType, stakeMapping, stakePrimaryKeyColumns)
	stakeInsertCacheMut       sync.RWMutex
	stakeInsertCache          = make(map[string]insertCache)
	stakeUpdateCacheMut       sync.RWMutex
	stakeUpdateCache          = make(map[string]updateCache)
	stakeUpsertCacheMut       sync.RWMutex
	stakeUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var stakeAfterSelectMu sync.Mutex
var stakeAfterSelectHooks []StakeHook

var stakeBeforeInsertMu sync.Mutex
var stakeBeforeInsertHooks []StakeHook
var stakeAfterInsertMu sync.Mutex
var stakeAfterInsertHooks []StakeHook

var stakeBeforeUpdateMu sync.Mutex
var stakeBeforeUpdateHooks []StakeHook
var stakeAfterUpdateMu sync.Mutex
var stakeAfterUpdateHooks []StakeHook

var stakeBeforeDeleteMu sync.Mutex
var stakeBeforeDeleteHooks []StakeHook
var stakeAfterDeleteMu sync.Mutex
var stakeAfterDeleteHooks []StakeHook

var stakeBeforeUpsertMu sync.Mutex
var stakeBeforeUpsertHooks []StakeHook
var stakeAfterUpsertMu sync.Mutex
var stakeAfterUpsertHooks []StakeHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Stake) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range stakeAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Stake) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range stakeBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Stake) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range stakeAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Stake) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range stakeBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Stake) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range stakeAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Stake) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range stakeBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Stake) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range stakeAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Stake) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range stakeBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Stake) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range stakeAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddStakeHook registers your hook function for all future operations.
func AddStakeHook(hookPoint boil.HookPoint, stakeHook StakeHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		stakeAfterSelectMu.Lock()
		stakeAfterSelectHooks = append(stakeAfterSelectHooks, stakeHook)
		stakeAfterSelectMu.Unlock()
	case boil.BeforeInsertHook:
		stakeBeforeInsertMu.Lock()
		stakeBeforeInsertHooks = append(stakeBeforeInsertHooks, stakeHook)
		stakeBeforeInsertMu.Unlock()
	case boil.AfterInsertHook:
		stakeAfterInsertMu.Lock()
		stakeAfterInsertHooks = append(stakeAfterInsertHooks, stakeHook)
		stakeAfterInsertMu.Unlock()
	case boil.BeforeUpdateHook:
		stakeBeforeUpdateMu.Lock()
		stakeBeforeUpdateHooks = append(stakeBeforeUpdateHooks, stakeHook)
		stakeBeforeUpdateMu.Unlock()
	case boil.AfterUpdateHook:
		stakeAfterUpdateMu.Lock()
		stakeAfterUpdateHooks = append(stakeAfterUpdateHooks, stakeHook)
		stakeAfterUpdateMu.Unlock()
	case boil.BeforeDeleteHook:
		stakeBeforeDeleteMu.Lock()
		stakeBeforeDeleteHooks = append(stakeBeforeDeleteHooks, stakeHook)
		stakeBeforeDeleteMu.Unlock()
	case boil.AfterDeleteHook:
		stakeAfterDeleteMu.Lock()
		stakeAfterDeleteHooks = append(stakeAfterDeleteHooks, stakeHook)
		stakeAfterDeleteMu.Unlock()
	case boil.BeforeUpsertHook:
		stakeBeforeUpsertMu.Lock()
		stakeBeforeUpsertHooks = append(stakeBeforeUpsertHooks, stakeHook)
		stakeBeforeUpsertMu.Unlock()
	case boil.AfterUpsertHook:
		stakeAfterUpsertMu.Lock()
		stakeAfterUpsertHooks = append(stakeAfterUpsertHooks, stakeHook)
		stakeAfterUpsertMu.Unlock()
	}
}

// One returns a single stake record from the query.
func (q stakeQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Stake, error) {
	o := &Stake{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for stakes")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Stake records from the query.
func (q stakeQuery) All(ctx context.Context, exec boil.ContextExecutor) (StakeSlice, error) {
	var o []*Stake

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Stake slice")
	}

	if len(stakeAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Stake records in the query.
func (q stakeQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count stakes rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q stakeQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if stakes exists")
	}

	return count > 0, nil
}

// Vehicle pointed to by the foreign key.
func (o *Stake) Vehicle(mods ...qm.QueryMod) vehicleQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.VehicleID),
	}

	queryMods = append(queryMods, mods...)

	return Vehicles(queryMods...)
}

// LoadVehicle allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (stakeL) LoadVehicle(ctx context.Context, e boil.ContextExecutor, singular bool, maybeStake interface{}, mods queries.Applicator) error {
	var slice []*Stake
	var object *Stake

	if singular {
		var ok bool
		object, ok = maybeStake.(*Stake)
		if !ok {
			object = new(Stake)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeStake)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeStake))
			}
		}
	} else {
		s, ok := maybeStake.(*[]*Stake)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeStake)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeStake))
			}
		}
	}

	args := make(map[interface{}]struct{})
	if singular {
		if object.R == nil {
			object.R = &stakeR{}
		}
		if !queries.IsNil(object.VehicleID) {
			args[object.VehicleID] = struct{}{}
		}

	} else {
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &stakeR{}
			}

			if !queries.IsNil(obj.VehicleID) {
				args[obj.VehicleID] = struct{}{}
			}

		}
	}

	if len(args) == 0 {
		return nil
	}

	argsSlice := make([]interface{}, len(args))
	i := 0
	for arg := range args {
		argsSlice[i] = arg
		i++
	}

	query := NewQuery(
		qm.From(`identity_api.vehicles`),
		qm.WhereIn(`identity_api.vehicles.id in ?`, argsSlice...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Vehicle")
	}

	var resultSlice []*Vehicle
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Vehicle")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for vehicles")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for vehicles")
	}

	if len(vehicleAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Vehicle = foreign
		if foreign.R == nil {
			foreign.R = &vehicleR{}
		}
		foreign.R.Stakes = append(foreign.R.Stakes, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if queries.Equal(local.VehicleID, foreign.ID) {
				local.R.Vehicle = foreign
				if foreign.R == nil {
					foreign.R = &vehicleR{}
				}
				foreign.R.Stakes = append(foreign.R.Stakes, local)
				break
			}
		}
	}

	return nil
}

// SetVehicle of the stake to the related item.
// Sets o.R.Vehicle to related.
// Adds o to related.R.Stakes.
func (o *Stake) SetVehicle(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Vehicle) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"identity_api\".\"stakes\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"vehicle_id"}),
		strmangle.WhereClause("\"", "\"", 2, stakePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	queries.Assign(&o.VehicleID, related.ID)
	if o.R == nil {
		o.R = &stakeR{
			Vehicle: related,
		}
	} else {
		o.R.Vehicle = related
	}

	if related.R == nil {
		related.R = &vehicleR{
			Stakes: StakeSlice{o},
		}
	} else {
		related.R.Stakes = append(related.R.Stakes, o)
	}

	return nil
}

// RemoveVehicle relationship.
// Sets o.R.Vehicle to nil.
// Removes o from all passed in related items' relationships struct.
func (o *Stake) RemoveVehicle(ctx context.Context, exec boil.ContextExecutor, related *Vehicle) error {
	var err error

	queries.SetScanner(&o.VehicleID, nil)
	if _, err = o.Update(ctx, exec, boil.Whitelist("vehicle_id")); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	if o.R != nil {
		o.R.Vehicle = nil
	}
	if related == nil || related.R == nil {
		return nil
	}

	for i, ri := range related.R.Stakes {
		if queries.Equal(o.VehicleID, ri.VehicleID) {
			continue
		}

		ln := len(related.R.Stakes)
		if ln > 1 && i < ln-1 {
			related.R.Stakes[i] = related.R.Stakes[ln-1]
		}
		related.R.Stakes = related.R.Stakes[:ln-1]
		break
	}
	return nil
}

// Stakes retrieves all the records using an executor.
func Stakes(mods ...qm.QueryMod) stakeQuery {
	mods = append(mods, qm.From("\"identity_api\".\"stakes\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"identity_api\".\"stakes\".*"})
	}

	return stakeQuery{q}
}

// FindStake retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindStake(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*Stake, error) {
	stakeObj := &Stake{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"identity_api\".\"stakes\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, stakeObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from stakes")
	}

	if err = stakeObj.doAfterSelectHooks(ctx, exec); err != nil {
		return stakeObj, err
	}

	return stakeObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Stake) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no stakes provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(stakeColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	stakeInsertCacheMut.RLock()
	cache, cached := stakeInsertCache[key]
	stakeInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			stakeAllColumns,
			stakeColumnsWithDefault,
			stakeColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(stakeType, stakeMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(stakeType, stakeMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"identity_api\".\"stakes\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"identity_api\".\"stakes\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into stakes")
	}

	if !cached {
		stakeInsertCacheMut.Lock()
		stakeInsertCache[key] = cache
		stakeInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Stake.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Stake) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	stakeUpdateCacheMut.RLock()
	cache, cached := stakeUpdateCache[key]
	stakeUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			stakeAllColumns,
			stakePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update stakes, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"identity_api\".\"stakes\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, stakePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(stakeType, stakeMapping, append(wl, stakePrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update stakes row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for stakes")
	}

	if !cached {
		stakeUpdateCacheMut.Lock()
		stakeUpdateCache[key] = cache
		stakeUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q stakeQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for stakes")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for stakes")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o StakeSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), stakePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"identity_api\".\"stakes\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, stakePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in stake slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all stake")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Stake) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns, opts ...UpsertOptionFunc) error {
	if o == nil {
		return errors.New("models: no stakes provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(stakeColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	stakeUpsertCacheMut.RLock()
	cache, cached := stakeUpsertCache[key]
	stakeUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, _ := insertColumns.InsertColumnSet(
			stakeAllColumns,
			stakeColumnsWithDefault,
			stakeColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			stakeAllColumns,
			stakePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert stakes, could not build update column list")
		}

		ret := strmangle.SetComplement(stakeAllColumns, strmangle.SetIntersect(insert, update))

		conflict := conflictColumns
		if len(conflict) == 0 && updateOnConflict && len(update) != 0 {
			if len(stakePrimaryKeyColumns) == 0 {
				return errors.New("models: unable to upsert stakes, could not build conflict column list")
			}

			conflict = make([]string, len(stakePrimaryKeyColumns))
			copy(conflict, stakePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"identity_api\".\"stakes\"", updateOnConflict, ret, update, conflict, insert, opts...)

		cache.valueMapping, err = queries.BindMapping(stakeType, stakeMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(stakeType, stakeMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert stakes")
	}

	if !cached {
		stakeUpsertCacheMut.Lock()
		stakeUpsertCache[key] = cache
		stakeUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Stake record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Stake) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Stake provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), stakePrimaryKeyMapping)
	sql := "DELETE FROM \"identity_api\".\"stakes\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from stakes")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for stakes")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q stakeQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no stakeQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from stakes")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for stakes")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o StakeSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(stakeBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), stakePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"identity_api\".\"stakes\" WHERE " +
		strmangle.WhereInClause(string(dialect.LQ), string(dialect.RQ), 1, stakePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from stake slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for stakes")
	}

	if len(stakeAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *Stake) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindStake(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *StakeSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := StakeSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), stakePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"identity_api\".\"stakes\".* FROM \"identity_api\".\"stakes\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, stakePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in StakeSlice")
	}

	*o = slice

	return nil
}

// StakeExists checks if the Stake row exists.
func StakeExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"identity_api\".\"stakes\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if stakes exists")
	}

	return exists, nil
}

// Exists checks if the Stake row exists.
func (o *Stake) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return StakeExists(ctx, exec, o.ID)
}

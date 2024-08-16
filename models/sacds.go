// Code generated by SQLBoiler 4.16.2 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
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
	"github.com/volatiletech/strmangle"
)

// Sacd is an object representing the database table.
type Sacd struct {
	TokenID     int         `boil:"token_id" json:"token_id" toml:"token_id" yaml:"token_id"`
	Grantee     []byte      `boil:"grantee" json:"grantee" toml:"grantee" yaml:"grantee"`
	Permissions string      `boil:"permissions" json:"permissions" toml:"permissions" yaml:"permissions"`
	Source      null.String `boil:"source" json:"source,omitempty" toml:"source" yaml:"source,omitempty"`
	CreatedAt   time.Time   `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	ExpiresAt   time.Time   `boil:"expires_at" json:"expires_at" toml:"expires_at" yaml:"expires_at"`

	R *sacdR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L sacdL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var SacdColumns = struct {
	TokenID     string
	Grantee     string
	Permissions string
	Source      string
	CreatedAt   string
	ExpiresAt   string
}{
	TokenID:     "token_id",
	Grantee:     "grantee",
	Permissions: "permissions",
	Source:      "source",
	CreatedAt:   "created_at",
	ExpiresAt:   "expires_at",
}

var SacdTableColumns = struct {
	TokenID     string
	Grantee     string
	Permissions string
	Source      string
	CreatedAt   string
	ExpiresAt   string
}{
	TokenID:     "sacds.token_id",
	Grantee:     "sacds.grantee",
	Permissions: "sacds.permissions",
	Source:      "sacds.source",
	CreatedAt:   "sacds.created_at",
	ExpiresAt:   "sacds.expires_at",
}

// Generated where

var SacdWhere = struct {
	TokenID     whereHelperint
	Grantee     whereHelper__byte
	Permissions whereHelperstring
	Source      whereHelpernull_String
	CreatedAt   whereHelpertime_Time
	ExpiresAt   whereHelpertime_Time
}{
	TokenID:     whereHelperint{field: "\"identity_api\".\"sacds\".\"token_id\""},
	Grantee:     whereHelper__byte{field: "\"identity_api\".\"sacds\".\"grantee\""},
	Permissions: whereHelperstring{field: "\"identity_api\".\"sacds\".\"permissions\""},
	Source:      whereHelpernull_String{field: "\"identity_api\".\"sacds\".\"source\""},
	CreatedAt:   whereHelpertime_Time{field: "\"identity_api\".\"sacds\".\"created_at\""},
	ExpiresAt:   whereHelpertime_Time{field: "\"identity_api\".\"sacds\".\"expires_at\""},
}

// SacdRels is where relationship names are stored.
var SacdRels = struct {
	Token string
}{
	Token: "Token",
}

// sacdR is where relationships are stored.
type sacdR struct {
	Token *Vehicle `boil:"Token" json:"Token" toml:"Token" yaml:"Token"`
}

// NewStruct creates a new relationship struct
func (*sacdR) NewStruct() *sacdR {
	return &sacdR{}
}

func (r *sacdR) GetToken() *Vehicle {
	if r == nil {
		return nil
	}
	return r.Token
}

// sacdL is where Load methods for each relationship are stored.
type sacdL struct{}

var (
	sacdAllColumns            = []string{"token_id", "grantee", "permissions", "source", "created_at", "expires_at"}
	sacdColumnsWithoutDefault = []string{"token_id", "grantee", "permissions", "created_at", "expires_at"}
	sacdColumnsWithDefault    = []string{"source"}
	sacdPrimaryKeyColumns     = []string{"token_id", "grantee", "permissions"}
	sacdGeneratedColumns      = []string{}
)

type (
	// SacdSlice is an alias for a slice of pointers to Sacd.
	// This should almost always be used instead of []Sacd.
	SacdSlice []*Sacd
	// SacdHook is the signature for custom Sacd hook methods
	SacdHook func(context.Context, boil.ContextExecutor, *Sacd) error

	sacdQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	sacdType                 = reflect.TypeOf(&Sacd{})
	sacdMapping              = queries.MakeStructMapping(sacdType)
	sacdPrimaryKeyMapping, _ = queries.BindMapping(sacdType, sacdMapping, sacdPrimaryKeyColumns)
	sacdInsertCacheMut       sync.RWMutex
	sacdInsertCache          = make(map[string]insertCache)
	sacdUpdateCacheMut       sync.RWMutex
	sacdUpdateCache          = make(map[string]updateCache)
	sacdUpsertCacheMut       sync.RWMutex
	sacdUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var sacdAfterSelectMu sync.Mutex
var sacdAfterSelectHooks []SacdHook

var sacdBeforeInsertMu sync.Mutex
var sacdBeforeInsertHooks []SacdHook
var sacdAfterInsertMu sync.Mutex
var sacdAfterInsertHooks []SacdHook

var sacdBeforeUpdateMu sync.Mutex
var sacdBeforeUpdateHooks []SacdHook
var sacdAfterUpdateMu sync.Mutex
var sacdAfterUpdateHooks []SacdHook

var sacdBeforeDeleteMu sync.Mutex
var sacdBeforeDeleteHooks []SacdHook
var sacdAfterDeleteMu sync.Mutex
var sacdAfterDeleteHooks []SacdHook

var sacdBeforeUpsertMu sync.Mutex
var sacdBeforeUpsertHooks []SacdHook
var sacdAfterUpsertMu sync.Mutex
var sacdAfterUpsertHooks []SacdHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Sacd) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sacdAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Sacd) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sacdBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Sacd) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sacdAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Sacd) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sacdBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Sacd) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sacdAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Sacd) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sacdBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Sacd) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sacdAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Sacd) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sacdBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Sacd) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range sacdAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddSacdHook registers your hook function for all future operations.
func AddSacdHook(hookPoint boil.HookPoint, sacdHook SacdHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		sacdAfterSelectMu.Lock()
		sacdAfterSelectHooks = append(sacdAfterSelectHooks, sacdHook)
		sacdAfterSelectMu.Unlock()
	case boil.BeforeInsertHook:
		sacdBeforeInsertMu.Lock()
		sacdBeforeInsertHooks = append(sacdBeforeInsertHooks, sacdHook)
		sacdBeforeInsertMu.Unlock()
	case boil.AfterInsertHook:
		sacdAfterInsertMu.Lock()
		sacdAfterInsertHooks = append(sacdAfterInsertHooks, sacdHook)
		sacdAfterInsertMu.Unlock()
	case boil.BeforeUpdateHook:
		sacdBeforeUpdateMu.Lock()
		sacdBeforeUpdateHooks = append(sacdBeforeUpdateHooks, sacdHook)
		sacdBeforeUpdateMu.Unlock()
	case boil.AfterUpdateHook:
		sacdAfterUpdateMu.Lock()
		sacdAfterUpdateHooks = append(sacdAfterUpdateHooks, sacdHook)
		sacdAfterUpdateMu.Unlock()
	case boil.BeforeDeleteHook:
		sacdBeforeDeleteMu.Lock()
		sacdBeforeDeleteHooks = append(sacdBeforeDeleteHooks, sacdHook)
		sacdBeforeDeleteMu.Unlock()
	case boil.AfterDeleteHook:
		sacdAfterDeleteMu.Lock()
		sacdAfterDeleteHooks = append(sacdAfterDeleteHooks, sacdHook)
		sacdAfterDeleteMu.Unlock()
	case boil.BeforeUpsertHook:
		sacdBeforeUpsertMu.Lock()
		sacdBeforeUpsertHooks = append(sacdBeforeUpsertHooks, sacdHook)
		sacdBeforeUpsertMu.Unlock()
	case boil.AfterUpsertHook:
		sacdAfterUpsertMu.Lock()
		sacdAfterUpsertHooks = append(sacdAfterUpsertHooks, sacdHook)
		sacdAfterUpsertMu.Unlock()
	}
}

// One returns a single sacd record from the query.
func (q sacdQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Sacd, error) {
	o := &Sacd{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for sacds")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Sacd records from the query.
func (q sacdQuery) All(ctx context.Context, exec boil.ContextExecutor) (SacdSlice, error) {
	var o []*Sacd

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Sacd slice")
	}

	if len(sacdAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Sacd records in the query.
func (q sacdQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count sacds rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q sacdQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if sacds exists")
	}

	return count > 0, nil
}

// Token pointed to by the foreign key.
func (o *Sacd) Token(mods ...qm.QueryMod) vehicleQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.TokenID),
	}

	queryMods = append(queryMods, mods...)

	return Vehicles(queryMods...)
}

// LoadToken allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (sacdL) LoadToken(ctx context.Context, e boil.ContextExecutor, singular bool, maybeSacd interface{}, mods queries.Applicator) error {
	var slice []*Sacd
	var object *Sacd

	if singular {
		var ok bool
		object, ok = maybeSacd.(*Sacd)
		if !ok {
			object = new(Sacd)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeSacd)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeSacd))
			}
		}
	} else {
		s, ok := maybeSacd.(*[]*Sacd)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeSacd)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeSacd))
			}
		}
	}

	args := make(map[interface{}]struct{})
	if singular {
		if object.R == nil {
			object.R = &sacdR{}
		}
		args[object.TokenID] = struct{}{}

	} else {
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &sacdR{}
			}

			args[obj.TokenID] = struct{}{}

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
		object.R.Token = foreign
		if foreign.R == nil {
			foreign.R = &vehicleR{}
		}
		foreign.R.TokenSacds = append(foreign.R.TokenSacds, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.TokenID == foreign.ID {
				local.R.Token = foreign
				if foreign.R == nil {
					foreign.R = &vehicleR{}
				}
				foreign.R.TokenSacds = append(foreign.R.TokenSacds, local)
				break
			}
		}
	}

	return nil
}

// SetToken of the sacd to the related item.
// Sets o.R.Token to related.
// Adds o to related.R.TokenSacds.
func (o *Sacd) SetToken(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Vehicle) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"identity_api\".\"sacds\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"token_id"}),
		strmangle.WhereClause("\"", "\"", 2, sacdPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.TokenID, o.Grantee, o.Permissions}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.TokenID = related.ID
	if o.R == nil {
		o.R = &sacdR{
			Token: related,
		}
	} else {
		o.R.Token = related
	}

	if related.R == nil {
		related.R = &vehicleR{
			TokenSacds: SacdSlice{o},
		}
	} else {
		related.R.TokenSacds = append(related.R.TokenSacds, o)
	}

	return nil
}

// Sacds retrieves all the records using an executor.
func Sacds(mods ...qm.QueryMod) sacdQuery {
	mods = append(mods, qm.From("\"identity_api\".\"sacds\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"identity_api\".\"sacds\".*"})
	}

	return sacdQuery{q}
}

// FindSacd retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindSacd(ctx context.Context, exec boil.ContextExecutor, tokenID int, grantee []byte, permissions string, selectCols ...string) (*Sacd, error) {
	sacdObj := &Sacd{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"identity_api\".\"sacds\" where \"token_id\"=$1 AND \"grantee\"=$2 AND \"permissions\"=$3", sel,
	)

	q := queries.Raw(query, tokenID, grantee, permissions)

	err := q.Bind(ctx, exec, sacdObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from sacds")
	}

	if err = sacdObj.doAfterSelectHooks(ctx, exec); err != nil {
		return sacdObj, err
	}

	return sacdObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Sacd) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no sacds provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(sacdColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	sacdInsertCacheMut.RLock()
	cache, cached := sacdInsertCache[key]
	sacdInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			sacdAllColumns,
			sacdColumnsWithDefault,
			sacdColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(sacdType, sacdMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(sacdType, sacdMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"identity_api\".\"sacds\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"identity_api\".\"sacds\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into sacds")
	}

	if !cached {
		sacdInsertCacheMut.Lock()
		sacdInsertCache[key] = cache
		sacdInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Sacd.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Sacd) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	sacdUpdateCacheMut.RLock()
	cache, cached := sacdUpdateCache[key]
	sacdUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			sacdAllColumns,
			sacdPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update sacds, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"identity_api\".\"sacds\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, sacdPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(sacdType, sacdMapping, append(wl, sacdPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update sacds row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for sacds")
	}

	if !cached {
		sacdUpdateCacheMut.Lock()
		sacdUpdateCache[key] = cache
		sacdUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q sacdQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for sacds")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for sacds")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o SacdSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), sacdPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"identity_api\".\"sacds\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, sacdPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in sacd slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all sacd")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Sacd) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns, opts ...UpsertOptionFunc) error {
	if o == nil {
		return errors.New("models: no sacds provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(sacdColumnsWithDefault, o)

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

	sacdUpsertCacheMut.RLock()
	cache, cached := sacdUpsertCache[key]
	sacdUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, _ := insertColumns.InsertColumnSet(
			sacdAllColumns,
			sacdColumnsWithDefault,
			sacdColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			sacdAllColumns,
			sacdPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert sacds, could not build update column list")
		}

		ret := strmangle.SetComplement(sacdAllColumns, strmangle.SetIntersect(insert, update))

		conflict := conflictColumns
		if len(conflict) == 0 && updateOnConflict && len(update) != 0 {
			if len(sacdPrimaryKeyColumns) == 0 {
				return errors.New("models: unable to upsert sacds, could not build conflict column list")
			}

			conflict = make([]string, len(sacdPrimaryKeyColumns))
			copy(conflict, sacdPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"identity_api\".\"sacds\"", updateOnConflict, ret, update, conflict, insert, opts...)

		cache.valueMapping, err = queries.BindMapping(sacdType, sacdMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(sacdType, sacdMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert sacds")
	}

	if !cached {
		sacdUpsertCacheMut.Lock()
		sacdUpsertCache[key] = cache
		sacdUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Sacd record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Sacd) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Sacd provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), sacdPrimaryKeyMapping)
	sql := "DELETE FROM \"identity_api\".\"sacds\" WHERE \"token_id\"=$1 AND \"grantee\"=$2 AND \"permissions\"=$3"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from sacds")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for sacds")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q sacdQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no sacdQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from sacds")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for sacds")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o SacdSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(sacdBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), sacdPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"identity_api\".\"sacds\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, sacdPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from sacd slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for sacds")
	}

	if len(sacdAfterDeleteHooks) != 0 {
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
func (o *Sacd) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindSacd(ctx, exec, o.TokenID, o.Grantee, o.Permissions)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *SacdSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := SacdSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), sacdPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"identity_api\".\"sacds\".* FROM \"identity_api\".\"sacds\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, sacdPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in SacdSlice")
	}

	*o = slice

	return nil
}

// SacdExists checks if the Sacd row exists.
func SacdExists(ctx context.Context, exec boil.ContextExecutor, tokenID int, grantee []byte, permissions string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"identity_api\".\"sacds\" where \"token_id\"=$1 AND \"grantee\"=$2 AND \"permissions\"=$3 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, tokenID, grantee, permissions)
	}
	row := exec.QueryRowContext(ctx, sql, tokenID, grantee, permissions)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if sacds exists")
	}

	return exists, nil
}

// Exists checks if the Sacd row exists.
func (o *Sacd) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return SacdExists(ctx, exec, o.TokenID, o.Grantee, o.Permissions)
}
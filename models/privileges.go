// Code generated by SQLBoiler 4.14.2 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
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
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// Privilege is an object representing the database table.
type Privilege struct {
	ID               string    `boil:"id" json:"id" toml:"id" yaml:"id"`
	TokenID          int       `boil:"token_id" json:"token_id" toml:"token_id" yaml:"token_id"`
	PrivilegeID      int       `boil:"privilege_id" json:"privilege_id" toml:"privilege_id" yaml:"privilege_id"`
	GrantedToAddress []byte    `boil:"granted_to_address" json:"granted_to_address" toml:"granted_to_address" yaml:"granted_to_address"`
	GrantedAt        time.Time `boil:"granted_at" json:"granted_at" toml:"granted_at" yaml:"granted_at"`
	ExpiresAt        time.Time `boil:"expires_at" json:"expires_at" toml:"expires_at" yaml:"expires_at"`

	R *privilegeR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L privilegeL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var PrivilegeColumns = struct {
	ID               string
	TokenID          string
	PrivilegeID      string
	GrantedToAddress string
	GrantedAt        string
	ExpiresAt        string
}{
	ID:               "id",
	TokenID:          "token_id",
	PrivilegeID:      "privilege_id",
	GrantedToAddress: "granted_to_address",
	GrantedAt:        "granted_at",
	ExpiresAt:        "expires_at",
}

var PrivilegeTableColumns = struct {
	ID               string
	TokenID          string
	PrivilegeID      string
	GrantedToAddress string
	GrantedAt        string
	ExpiresAt        string
}{
	ID:               "privileges.id",
	TokenID:          "privileges.token_id",
	PrivilegeID:      "privileges.privilege_id",
	GrantedToAddress: "privileges.granted_to_address",
	GrantedAt:        "privileges.granted_at",
	ExpiresAt:        "privileges.expires_at",
}

// Generated where

type whereHelperstring struct{ field string }

func (w whereHelperstring) EQ(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperstring) NEQ(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperstring) LT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperstring) LTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperstring) GT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperstring) GTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperstring) IN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperstring) NIN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

type whereHelper__byte struct{ field string }

func (w whereHelper__byte) EQ(x []byte) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelper__byte) NEQ(x []byte) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelper__byte) LT(x []byte) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelper__byte) LTE(x []byte) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelper__byte) GT(x []byte) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelper__byte) GTE(x []byte) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }

type whereHelpertime_Time struct{ field string }

func (w whereHelpertime_Time) EQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpertime_Time) NEQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpertime_Time) LT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertime_Time) LTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertime_Time) GT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertime_Time) GTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var PrivilegeWhere = struct {
	ID               whereHelperstring
	TokenID          whereHelperint
	PrivilegeID      whereHelperint
	GrantedToAddress whereHelper__byte
	GrantedAt        whereHelpertime_Time
	ExpiresAt        whereHelpertime_Time
}{
	ID:               whereHelperstring{field: "\"identity_api\".\"privileges\".\"id\""},
	TokenID:          whereHelperint{field: "\"identity_api\".\"privileges\".\"token_id\""},
	PrivilegeID:      whereHelperint{field: "\"identity_api\".\"privileges\".\"privilege_id\""},
	GrantedToAddress: whereHelper__byte{field: "\"identity_api\".\"privileges\".\"granted_to_address\""},
	GrantedAt:        whereHelpertime_Time{field: "\"identity_api\".\"privileges\".\"granted_at\""},
	ExpiresAt:        whereHelpertime_Time{field: "\"identity_api\".\"privileges\".\"expires_at\""},
}

// PrivilegeRels is where relationship names are stored.
var PrivilegeRels = struct {
	Token string
}{
	Token: "Token",
}

// privilegeR is where relationships are stored.
type privilegeR struct {
	Token *Vehicle `boil:"Token" json:"Token" toml:"Token" yaml:"Token"`
}

// NewStruct creates a new relationship struct
func (*privilegeR) NewStruct() *privilegeR {
	return &privilegeR{}
}

func (r *privilegeR) GetToken() *Vehicle {
	if r == nil {
		return nil
	}
	return r.Token
}

// privilegeL is where Load methods for each relationship are stored.
type privilegeL struct{}

var (
	privilegeAllColumns            = []string{"id", "token_id", "privilege_id", "granted_to_address", "granted_at", "expires_at"}
	privilegeColumnsWithoutDefault = []string{"id", "token_id", "privilege_id", "granted_to_address", "granted_at", "expires_at"}
	privilegeColumnsWithDefault    = []string{}
	privilegePrimaryKeyColumns     = []string{"id"}
	privilegeGeneratedColumns      = []string{}
)

type (
	// PrivilegeSlice is an alias for a slice of pointers to Privilege.
	// This should almost always be used instead of []Privilege.
	PrivilegeSlice []*Privilege
	// PrivilegeHook is the signature for custom Privilege hook methods
	PrivilegeHook func(context.Context, boil.ContextExecutor, *Privilege) error

	privilegeQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	privilegeType                 = reflect.TypeOf(&Privilege{})
	privilegeMapping              = queries.MakeStructMapping(privilegeType)
	privilegePrimaryKeyMapping, _ = queries.BindMapping(privilegeType, privilegeMapping, privilegePrimaryKeyColumns)
	privilegeInsertCacheMut       sync.RWMutex
	privilegeInsertCache          = make(map[string]insertCache)
	privilegeUpdateCacheMut       sync.RWMutex
	privilegeUpdateCache          = make(map[string]updateCache)
	privilegeUpsertCacheMut       sync.RWMutex
	privilegeUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var privilegeAfterSelectHooks []PrivilegeHook

var privilegeBeforeInsertHooks []PrivilegeHook
var privilegeAfterInsertHooks []PrivilegeHook

var privilegeBeforeUpdateHooks []PrivilegeHook
var privilegeAfterUpdateHooks []PrivilegeHook

var privilegeBeforeDeleteHooks []PrivilegeHook
var privilegeAfterDeleteHooks []PrivilegeHook

var privilegeBeforeUpsertHooks []PrivilegeHook
var privilegeAfterUpsertHooks []PrivilegeHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Privilege) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range privilegeAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Privilege) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range privilegeBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Privilege) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range privilegeAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Privilege) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range privilegeBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Privilege) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range privilegeAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Privilege) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range privilegeBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Privilege) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range privilegeAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Privilege) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range privilegeBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Privilege) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range privilegeAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddPrivilegeHook registers your hook function for all future operations.
func AddPrivilegeHook(hookPoint boil.HookPoint, privilegeHook PrivilegeHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		privilegeAfterSelectHooks = append(privilegeAfterSelectHooks, privilegeHook)
	case boil.BeforeInsertHook:
		privilegeBeforeInsertHooks = append(privilegeBeforeInsertHooks, privilegeHook)
	case boil.AfterInsertHook:
		privilegeAfterInsertHooks = append(privilegeAfterInsertHooks, privilegeHook)
	case boil.BeforeUpdateHook:
		privilegeBeforeUpdateHooks = append(privilegeBeforeUpdateHooks, privilegeHook)
	case boil.AfterUpdateHook:
		privilegeAfterUpdateHooks = append(privilegeAfterUpdateHooks, privilegeHook)
	case boil.BeforeDeleteHook:
		privilegeBeforeDeleteHooks = append(privilegeBeforeDeleteHooks, privilegeHook)
	case boil.AfterDeleteHook:
		privilegeAfterDeleteHooks = append(privilegeAfterDeleteHooks, privilegeHook)
	case boil.BeforeUpsertHook:
		privilegeBeforeUpsertHooks = append(privilegeBeforeUpsertHooks, privilegeHook)
	case boil.AfterUpsertHook:
		privilegeAfterUpsertHooks = append(privilegeAfterUpsertHooks, privilegeHook)
	}
}

// One returns a single privilege record from the query.
func (q privilegeQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Privilege, error) {
	o := &Privilege{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for privileges")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Privilege records from the query.
func (q privilegeQuery) All(ctx context.Context, exec boil.ContextExecutor) (PrivilegeSlice, error) {
	var o []*Privilege

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Privilege slice")
	}

	if len(privilegeAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Privilege records in the query.
func (q privilegeQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count privileges rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q privilegeQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if privileges exists")
	}

	return count > 0, nil
}

// Token pointed to by the foreign key.
func (o *Privilege) Token(mods ...qm.QueryMod) vehicleQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.TokenID),
	}

	queryMods = append(queryMods, mods...)

	return Vehicles(queryMods...)
}

// LoadToken allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (privilegeL) LoadToken(ctx context.Context, e boil.ContextExecutor, singular bool, maybePrivilege interface{}, mods queries.Applicator) error {
	var slice []*Privilege
	var object *Privilege

	if singular {
		var ok bool
		object, ok = maybePrivilege.(*Privilege)
		if !ok {
			object = new(Privilege)
			ok = queries.SetFromEmbeddedStruct(&object, &maybePrivilege)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybePrivilege))
			}
		}
	} else {
		s, ok := maybePrivilege.(*[]*Privilege)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybePrivilege)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybePrivilege))
			}
		}
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &privilegeR{}
		}
		args = append(args, object.TokenID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &privilegeR{}
			}

			for _, a := range args {
				if a == obj.TokenID {
					continue Outer
				}
			}

			args = append(args, obj.TokenID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`identity_api.vehicles`),
		qm.WhereIn(`identity_api.vehicles.id in ?`, args...),
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
		foreign.R.TokenPrivileges = append(foreign.R.TokenPrivileges, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.TokenID == foreign.ID {
				local.R.Token = foreign
				if foreign.R == nil {
					foreign.R = &vehicleR{}
				}
				foreign.R.TokenPrivileges = append(foreign.R.TokenPrivileges, local)
				break
			}
		}
	}

	return nil
}

// SetToken of the privilege to the related item.
// Sets o.R.Token to related.
// Adds o to related.R.TokenPrivileges.
func (o *Privilege) SetToken(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Vehicle) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"identity_api\".\"privileges\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"token_id"}),
		strmangle.WhereClause("\"", "\"", 2, privilegePrimaryKeyColumns),
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

	o.TokenID = related.ID
	if o.R == nil {
		o.R = &privilegeR{
			Token: related,
		}
	} else {
		o.R.Token = related
	}

	if related.R == nil {
		related.R = &vehicleR{
			TokenPrivileges: PrivilegeSlice{o},
		}
	} else {
		related.R.TokenPrivileges = append(related.R.TokenPrivileges, o)
	}

	return nil
}

// Privileges retrieves all the records using an executor.
func Privileges(mods ...qm.QueryMod) privilegeQuery {
	mods = append(mods, qm.From("\"identity_api\".\"privileges\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"identity_api\".\"privileges\".*"})
	}

	return privilegeQuery{q}
}

// FindPrivilege retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindPrivilege(ctx context.Context, exec boil.ContextExecutor, iD string, selectCols ...string) (*Privilege, error) {
	privilegeObj := &Privilege{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"identity_api\".\"privileges\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, privilegeObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from privileges")
	}

	if err = privilegeObj.doAfterSelectHooks(ctx, exec); err != nil {
		return privilegeObj, err
	}

	return privilegeObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Privilege) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no privileges provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(privilegeColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	privilegeInsertCacheMut.RLock()
	cache, cached := privilegeInsertCache[key]
	privilegeInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			privilegeAllColumns,
			privilegeColumnsWithDefault,
			privilegeColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(privilegeType, privilegeMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(privilegeType, privilegeMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"identity_api\".\"privileges\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"identity_api\".\"privileges\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into privileges")
	}

	if !cached {
		privilegeInsertCacheMut.Lock()
		privilegeInsertCache[key] = cache
		privilegeInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Privilege.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Privilege) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	privilegeUpdateCacheMut.RLock()
	cache, cached := privilegeUpdateCache[key]
	privilegeUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			privilegeAllColumns,
			privilegePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update privileges, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"identity_api\".\"privileges\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, privilegePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(privilegeType, privilegeMapping, append(wl, privilegePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update privileges row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for privileges")
	}

	if !cached {
		privilegeUpdateCacheMut.Lock()
		privilegeUpdateCache[key] = cache
		privilegeUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q privilegeQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for privileges")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for privileges")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o PrivilegeSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), privilegePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"identity_api\".\"privileges\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, privilegePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in privilege slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all privilege")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Privilege) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no privileges provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(privilegeColumnsWithDefault, o)

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

	privilegeUpsertCacheMut.RLock()
	cache, cached := privilegeUpsertCache[key]
	privilegeUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			privilegeAllColumns,
			privilegeColumnsWithDefault,
			privilegeColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			privilegeAllColumns,
			privilegePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert privileges, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(privilegePrimaryKeyColumns))
			copy(conflict, privilegePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"identity_api\".\"privileges\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(privilegeType, privilegeMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(privilegeType, privilegeMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert privileges")
	}

	if !cached {
		privilegeUpsertCacheMut.Lock()
		privilegeUpsertCache[key] = cache
		privilegeUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Privilege record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Privilege) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Privilege provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), privilegePrimaryKeyMapping)
	sql := "DELETE FROM \"identity_api\".\"privileges\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from privileges")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for privileges")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q privilegeQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no privilegeQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from privileges")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for privileges")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o PrivilegeSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(privilegeBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), privilegePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"identity_api\".\"privileges\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, privilegePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from privilege slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for privileges")
	}

	if len(privilegeAfterDeleteHooks) != 0 {
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
func (o *Privilege) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindPrivilege(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *PrivilegeSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := PrivilegeSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), privilegePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"identity_api\".\"privileges\".* FROM \"identity_api\".\"privileges\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, privilegePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in PrivilegeSlice")
	}

	*o = slice

	return nil
}

// PrivilegeExists checks if the Privilege row exists.
func PrivilegeExists(ctx context.Context, exec boil.ContextExecutor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"identity_api\".\"privileges\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if privileges exists")
	}

	return exists, nil
}

// Exists checks if the Privilege row exists.
func (o *Privilege) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return PrivilegeExists(ctx, exec, o.ID)
}

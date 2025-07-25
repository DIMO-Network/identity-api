// Code generated by SQLBoiler 4.19.5 (https://github.com/aarondl/sqlboiler). DO NOT EDIT.
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

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/queries/qmhelper"
	"github.com/aarondl/strmangle"
	"github.com/friendsofgo/errors"
)

// VehicleSacd is an object representing the database table.
type VehicleSacd struct {
	VehicleID   int       `boil:"vehicle_id" json:"vehicle_id" toml:"vehicle_id" yaml:"vehicle_id"`
	Grantee     []byte    `boil:"grantee" json:"grantee" toml:"grantee" yaml:"grantee"`
	Permissions string    `boil:"permissions" json:"permissions" toml:"permissions" yaml:"permissions"`
	Source      string    `boil:"source" json:"source" toml:"source" yaml:"source"`
	CreatedAt   time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	ExpiresAt   time.Time `boil:"expires_at" json:"expires_at" toml:"expires_at" yaml:"expires_at"`

	R *vehicleSacdR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L vehicleSacdL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var VehicleSacdColumns = struct {
	VehicleID   string
	Grantee     string
	Permissions string
	Source      string
	CreatedAt   string
	ExpiresAt   string
}{
	VehicleID:   "vehicle_id",
	Grantee:     "grantee",
	Permissions: "permissions",
	Source:      "source",
	CreatedAt:   "created_at",
	ExpiresAt:   "expires_at",
}

var VehicleSacdTableColumns = struct {
	VehicleID   string
	Grantee     string
	Permissions string
	Source      string
	CreatedAt   string
	ExpiresAt   string
}{
	VehicleID:   "vehicle_sacds.vehicle_id",
	Grantee:     "vehicle_sacds.grantee",
	Permissions: "vehicle_sacds.permissions",
	Source:      "vehicle_sacds.source",
	CreatedAt:   "vehicle_sacds.created_at",
	ExpiresAt:   "vehicle_sacds.expires_at",
}

// Generated where

var VehicleSacdWhere = struct {
	VehicleID   whereHelperint
	Grantee     whereHelper__byte
	Permissions whereHelperstring
	Source      whereHelperstring
	CreatedAt   whereHelpertime_Time
	ExpiresAt   whereHelpertime_Time
}{
	VehicleID:   whereHelperint{field: "\"identity_api\".\"vehicle_sacds\".\"vehicle_id\""},
	Grantee:     whereHelper__byte{field: "\"identity_api\".\"vehicle_sacds\".\"grantee\""},
	Permissions: whereHelperstring{field: "\"identity_api\".\"vehicle_sacds\".\"permissions\""},
	Source:      whereHelperstring{field: "\"identity_api\".\"vehicle_sacds\".\"source\""},
	CreatedAt:   whereHelpertime_Time{field: "\"identity_api\".\"vehicle_sacds\".\"created_at\""},
	ExpiresAt:   whereHelpertime_Time{field: "\"identity_api\".\"vehicle_sacds\".\"expires_at\""},
}

// VehicleSacdRels is where relationship names are stored.
var VehicleSacdRels = struct {
	Vehicle string
}{
	Vehicle: "Vehicle",
}

// vehicleSacdR is where relationships are stored.
type vehicleSacdR struct {
	Vehicle *Vehicle `boil:"Vehicle" json:"Vehicle" toml:"Vehicle" yaml:"Vehicle"`
}

// NewStruct creates a new relationship struct
func (*vehicleSacdR) NewStruct() *vehicleSacdR {
	return &vehicleSacdR{}
}

func (o *VehicleSacd) GetVehicle() *Vehicle {
	if o == nil {
		return nil
	}

	return o.R.GetVehicle()
}

func (r *vehicleSacdR) GetVehicle() *Vehicle {
	if r == nil {
		return nil
	}

	return r.Vehicle
}

// vehicleSacdL is where Load methods for each relationship are stored.
type vehicleSacdL struct{}

var (
	vehicleSacdAllColumns            = []string{"vehicle_id", "grantee", "permissions", "source", "created_at", "expires_at"}
	vehicleSacdColumnsWithoutDefault = []string{"vehicle_id", "grantee", "permissions", "source", "created_at", "expires_at"}
	vehicleSacdColumnsWithDefault    = []string{}
	vehicleSacdPrimaryKeyColumns     = []string{"vehicle_id", "grantee"}
	vehicleSacdGeneratedColumns      = []string{}
)

type (
	// VehicleSacdSlice is an alias for a slice of pointers to VehicleSacd.
	// This should almost always be used instead of []VehicleSacd.
	VehicleSacdSlice []*VehicleSacd
	// VehicleSacdHook is the signature for custom VehicleSacd hook methods
	VehicleSacdHook func(context.Context, boil.ContextExecutor, *VehicleSacd) error

	vehicleSacdQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	vehicleSacdType                 = reflect.TypeOf(&VehicleSacd{})
	vehicleSacdMapping              = queries.MakeStructMapping(vehicleSacdType)
	vehicleSacdPrimaryKeyMapping, _ = queries.BindMapping(vehicleSacdType, vehicleSacdMapping, vehicleSacdPrimaryKeyColumns)
	vehicleSacdInsertCacheMut       sync.RWMutex
	vehicleSacdInsertCache          = make(map[string]insertCache)
	vehicleSacdUpdateCacheMut       sync.RWMutex
	vehicleSacdUpdateCache          = make(map[string]updateCache)
	vehicleSacdUpsertCacheMut       sync.RWMutex
	vehicleSacdUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var vehicleSacdAfterSelectMu sync.Mutex
var vehicleSacdAfterSelectHooks []VehicleSacdHook

var vehicleSacdBeforeInsertMu sync.Mutex
var vehicleSacdBeforeInsertHooks []VehicleSacdHook
var vehicleSacdAfterInsertMu sync.Mutex
var vehicleSacdAfterInsertHooks []VehicleSacdHook

var vehicleSacdBeforeUpdateMu sync.Mutex
var vehicleSacdBeforeUpdateHooks []VehicleSacdHook
var vehicleSacdAfterUpdateMu sync.Mutex
var vehicleSacdAfterUpdateHooks []VehicleSacdHook

var vehicleSacdBeforeDeleteMu sync.Mutex
var vehicleSacdBeforeDeleteHooks []VehicleSacdHook
var vehicleSacdAfterDeleteMu sync.Mutex
var vehicleSacdAfterDeleteHooks []VehicleSacdHook

var vehicleSacdBeforeUpsertMu sync.Mutex
var vehicleSacdBeforeUpsertHooks []VehicleSacdHook
var vehicleSacdAfterUpsertMu sync.Mutex
var vehicleSacdAfterUpsertHooks []VehicleSacdHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *VehicleSacd) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleSacdAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *VehicleSacd) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleSacdBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *VehicleSacd) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleSacdAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *VehicleSacd) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleSacdBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *VehicleSacd) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleSacdAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *VehicleSacd) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleSacdBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *VehicleSacd) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleSacdAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *VehicleSacd) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleSacdBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *VehicleSacd) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleSacdAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddVehicleSacdHook registers your hook function for all future operations.
func AddVehicleSacdHook(hookPoint boil.HookPoint, vehicleSacdHook VehicleSacdHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		vehicleSacdAfterSelectMu.Lock()
		vehicleSacdAfterSelectHooks = append(vehicleSacdAfterSelectHooks, vehicleSacdHook)
		vehicleSacdAfterSelectMu.Unlock()
	case boil.BeforeInsertHook:
		vehicleSacdBeforeInsertMu.Lock()
		vehicleSacdBeforeInsertHooks = append(vehicleSacdBeforeInsertHooks, vehicleSacdHook)
		vehicleSacdBeforeInsertMu.Unlock()
	case boil.AfterInsertHook:
		vehicleSacdAfterInsertMu.Lock()
		vehicleSacdAfterInsertHooks = append(vehicleSacdAfterInsertHooks, vehicleSacdHook)
		vehicleSacdAfterInsertMu.Unlock()
	case boil.BeforeUpdateHook:
		vehicleSacdBeforeUpdateMu.Lock()
		vehicleSacdBeforeUpdateHooks = append(vehicleSacdBeforeUpdateHooks, vehicleSacdHook)
		vehicleSacdBeforeUpdateMu.Unlock()
	case boil.AfterUpdateHook:
		vehicleSacdAfterUpdateMu.Lock()
		vehicleSacdAfterUpdateHooks = append(vehicleSacdAfterUpdateHooks, vehicleSacdHook)
		vehicleSacdAfterUpdateMu.Unlock()
	case boil.BeforeDeleteHook:
		vehicleSacdBeforeDeleteMu.Lock()
		vehicleSacdBeforeDeleteHooks = append(vehicleSacdBeforeDeleteHooks, vehicleSacdHook)
		vehicleSacdBeforeDeleteMu.Unlock()
	case boil.AfterDeleteHook:
		vehicleSacdAfterDeleteMu.Lock()
		vehicleSacdAfterDeleteHooks = append(vehicleSacdAfterDeleteHooks, vehicleSacdHook)
		vehicleSacdAfterDeleteMu.Unlock()
	case boil.BeforeUpsertHook:
		vehicleSacdBeforeUpsertMu.Lock()
		vehicleSacdBeforeUpsertHooks = append(vehicleSacdBeforeUpsertHooks, vehicleSacdHook)
		vehicleSacdBeforeUpsertMu.Unlock()
	case boil.AfterUpsertHook:
		vehicleSacdAfterUpsertMu.Lock()
		vehicleSacdAfterUpsertHooks = append(vehicleSacdAfterUpsertHooks, vehicleSacdHook)
		vehicleSacdAfterUpsertMu.Unlock()
	}
}

// One returns a single vehicleSacd record from the query.
func (q vehicleSacdQuery) One(ctx context.Context, exec boil.ContextExecutor) (*VehicleSacd, error) {
	o := &VehicleSacd{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for vehicle_sacds")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all VehicleSacd records from the query.
func (q vehicleSacdQuery) All(ctx context.Context, exec boil.ContextExecutor) (VehicleSacdSlice, error) {
	var o []*VehicleSacd

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to VehicleSacd slice")
	}

	if len(vehicleSacdAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all VehicleSacd records in the query.
func (q vehicleSacdQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count vehicle_sacds rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q vehicleSacdQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if vehicle_sacds exists")
	}

	return count > 0, nil
}

// Vehicle pointed to by the foreign key.
func (o *VehicleSacd) Vehicle(mods ...qm.QueryMod) vehicleQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.VehicleID),
	}

	queryMods = append(queryMods, mods...)

	return Vehicles(queryMods...)
}

// LoadVehicle allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (vehicleSacdL) LoadVehicle(ctx context.Context, e boil.ContextExecutor, singular bool, maybeVehicleSacd interface{}, mods queries.Applicator) error {
	var slice []*VehicleSacd
	var object *VehicleSacd

	if singular {
		var ok bool
		object, ok = maybeVehicleSacd.(*VehicleSacd)
		if !ok {
			object = new(VehicleSacd)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeVehicleSacd)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeVehicleSacd))
			}
		}
	} else {
		s, ok := maybeVehicleSacd.(*[]*VehicleSacd)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeVehicleSacd)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeVehicleSacd))
			}
		}
	}

	args := make(map[interface{}]struct{})
	if singular {
		if object.R == nil {
			object.R = &vehicleSacdR{}
		}
		args[object.VehicleID] = struct{}{}

	} else {
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &vehicleSacdR{}
			}

			args[obj.VehicleID] = struct{}{}

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
		foreign.R.VehicleSacds = append(foreign.R.VehicleSacds, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.VehicleID == foreign.ID {
				local.R.Vehicle = foreign
				if foreign.R == nil {
					foreign.R = &vehicleR{}
				}
				foreign.R.VehicleSacds = append(foreign.R.VehicleSacds, local)
				break
			}
		}
	}

	return nil
}

// SetVehicle of the vehicleSacd to the related item.
// Sets o.R.Vehicle to related.
// Adds o to related.R.VehicleSacds.
func (o *VehicleSacd) SetVehicle(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Vehicle) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"identity_api\".\"vehicle_sacds\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"vehicle_id"}),
		strmangle.WhereClause("\"", "\"", 2, vehicleSacdPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.VehicleID, o.Grantee}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.VehicleID = related.ID
	if o.R == nil {
		o.R = &vehicleSacdR{
			Vehicle: related,
		}
	} else {
		o.R.Vehicle = related
	}

	if related.R == nil {
		related.R = &vehicleR{
			VehicleSacds: VehicleSacdSlice{o},
		}
	} else {
		related.R.VehicleSacds = append(related.R.VehicleSacds, o)
	}

	return nil
}

// VehicleSacds retrieves all the records using an executor.
func VehicleSacds(mods ...qm.QueryMod) vehicleSacdQuery {
	mods = append(mods, qm.From("\"identity_api\".\"vehicle_sacds\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"identity_api\".\"vehicle_sacds\".*"})
	}

	return vehicleSacdQuery{q}
}

// FindVehicleSacd retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindVehicleSacd(ctx context.Context, exec boil.ContextExecutor, vehicleID int, grantee []byte, selectCols ...string) (*VehicleSacd, error) {
	vehicleSacdObj := &VehicleSacd{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"identity_api\".\"vehicle_sacds\" where \"vehicle_id\"=$1 AND \"grantee\"=$2", sel,
	)

	q := queries.Raw(query, vehicleID, grantee)

	err := q.Bind(ctx, exec, vehicleSacdObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from vehicle_sacds")
	}

	if err = vehicleSacdObj.doAfterSelectHooks(ctx, exec); err != nil {
		return vehicleSacdObj, err
	}

	return vehicleSacdObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *VehicleSacd) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no vehicle_sacds provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(vehicleSacdColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	vehicleSacdInsertCacheMut.RLock()
	cache, cached := vehicleSacdInsertCache[key]
	vehicleSacdInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			vehicleSacdAllColumns,
			vehicleSacdColumnsWithDefault,
			vehicleSacdColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(vehicleSacdType, vehicleSacdMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(vehicleSacdType, vehicleSacdMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"identity_api\".\"vehicle_sacds\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"identity_api\".\"vehicle_sacds\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into vehicle_sacds")
	}

	if !cached {
		vehicleSacdInsertCacheMut.Lock()
		vehicleSacdInsertCache[key] = cache
		vehicleSacdInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the VehicleSacd.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *VehicleSacd) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	vehicleSacdUpdateCacheMut.RLock()
	cache, cached := vehicleSacdUpdateCache[key]
	vehicleSacdUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			vehicleSacdAllColumns,
			vehicleSacdPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update vehicle_sacds, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"identity_api\".\"vehicle_sacds\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, vehicleSacdPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(vehicleSacdType, vehicleSacdMapping, append(wl, vehicleSacdPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update vehicle_sacds row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for vehicle_sacds")
	}

	if !cached {
		vehicleSacdUpdateCacheMut.Lock()
		vehicleSacdUpdateCache[key] = cache
		vehicleSacdUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q vehicleSacdQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for vehicle_sacds")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for vehicle_sacds")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o VehicleSacdSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), vehicleSacdPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"identity_api\".\"vehicle_sacds\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, vehicleSacdPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in vehicleSacd slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all vehicleSacd")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *VehicleSacd) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns, opts ...UpsertOptionFunc) error {
	if o == nil {
		return errors.New("models: no vehicle_sacds provided for upsert")
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

	nzDefaults := queries.NonZeroDefaultSet(vehicleSacdColumnsWithDefault, o)

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

	vehicleSacdUpsertCacheMut.RLock()
	cache, cached := vehicleSacdUpsertCache[key]
	vehicleSacdUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, _ := insertColumns.InsertColumnSet(
			vehicleSacdAllColumns,
			vehicleSacdColumnsWithDefault,
			vehicleSacdColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			vehicleSacdAllColumns,
			vehicleSacdPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert vehicle_sacds, could not build update column list")
		}

		ret := strmangle.SetComplement(vehicleSacdAllColumns, strmangle.SetIntersect(insert, update))

		conflict := conflictColumns
		if len(conflict) == 0 && updateOnConflict && len(update) != 0 {
			if len(vehicleSacdPrimaryKeyColumns) == 0 {
				return errors.New("models: unable to upsert vehicle_sacds, could not build conflict column list")
			}

			conflict = make([]string, len(vehicleSacdPrimaryKeyColumns))
			copy(conflict, vehicleSacdPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"identity_api\".\"vehicle_sacds\"", updateOnConflict, ret, update, conflict, insert, opts...)

		cache.valueMapping, err = queries.BindMapping(vehicleSacdType, vehicleSacdMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(vehicleSacdType, vehicleSacdMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert vehicle_sacds")
	}

	if !cached {
		vehicleSacdUpsertCacheMut.Lock()
		vehicleSacdUpsertCache[key] = cache
		vehicleSacdUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single VehicleSacd record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *VehicleSacd) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no VehicleSacd provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), vehicleSacdPrimaryKeyMapping)
	sql := "DELETE FROM \"identity_api\".\"vehicle_sacds\" WHERE \"vehicle_id\"=$1 AND \"grantee\"=$2"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from vehicle_sacds")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for vehicle_sacds")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q vehicleSacdQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no vehicleSacdQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from vehicle_sacds")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for vehicle_sacds")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o VehicleSacdSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(vehicleSacdBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), vehicleSacdPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"identity_api\".\"vehicle_sacds\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, vehicleSacdPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from vehicleSacd slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for vehicle_sacds")
	}

	if len(vehicleSacdAfterDeleteHooks) != 0 {
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
func (o *VehicleSacd) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindVehicleSacd(ctx, exec, o.VehicleID, o.Grantee)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *VehicleSacdSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := VehicleSacdSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), vehicleSacdPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"identity_api\".\"vehicle_sacds\".* FROM \"identity_api\".\"vehicle_sacds\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, vehicleSacdPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in VehicleSacdSlice")
	}

	*o = slice

	return nil
}

// VehicleSacdExists checks if the VehicleSacd row exists.
func VehicleSacdExists(ctx context.Context, exec boil.ContextExecutor, vehicleID int, grantee []byte) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"identity_api\".\"vehicle_sacds\" where \"vehicle_id\"=$1 AND \"grantee\"=$2 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, vehicleID, grantee)
	}
	row := exec.QueryRowContext(ctx, sql, vehicleID, grantee)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if vehicle_sacds exists")
	}

	return exists, nil
}

// Exists checks if the VehicleSacd row exists.
func (o *VehicleSacd) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return VehicleSacdExists(ctx, exec, o.VehicleID, o.Grantee)
}

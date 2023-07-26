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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// Vehicle is an object representing the database table.
type Vehicle struct {
	ID           int         `boil:"id" json:"id" toml:"id" yaml:"id"`
	OwnerAddress []byte      `boil:"owner_address" json:"owner_address" toml:"owner_address" yaml:"owner_address"`
	Make         null.String `boil:"make" json:"make,omitempty" toml:"make" yaml:"make,omitempty"`
	Model        null.String `boil:"model" json:"model,omitempty" toml:"model" yaml:"model,omitempty"`
	Year         null.Int    `boil:"year" json:"year,omitempty" toml:"year" yaml:"year,omitempty"`
	MintedAt     time.Time   `boil:"minted_at" json:"minted_at" toml:"minted_at" yaml:"minted_at"`

	R *vehicleR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L vehicleL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var VehicleColumns = struct {
	ID           string
	OwnerAddress string
	Make         string
	Model        string
	Year         string
	MintedAt     string
}{
	ID:           "id",
	OwnerAddress: "owner_address",
	Make:         "make",
	Model:        "model",
	Year:         "year",
	MintedAt:     "minted_at",
}

var VehicleTableColumns = struct {
	ID           string
	OwnerAddress string
	Make         string
	Model        string
	Year         string
	MintedAt     string
}{
	ID:           "vehicles.id",
	OwnerAddress: "vehicles.owner_address",
	Make:         "vehicles.make",
	Model:        "vehicles.model",
	Year:         "vehicles.year",
	MintedAt:     "vehicles.minted_at",
}

// Generated where

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

var VehicleWhere = struct {
	ID           whereHelperint
	OwnerAddress whereHelper__byte
	Make         whereHelpernull_String
	Model        whereHelpernull_String
	Year         whereHelpernull_Int
	MintedAt     whereHelpertime_Time
}{
	ID:           whereHelperint{field: "\"identity_api\".\"vehicles\".\"id\""},
	OwnerAddress: whereHelper__byte{field: "\"identity_api\".\"vehicles\".\"owner_address\""},
	Make:         whereHelpernull_String{field: "\"identity_api\".\"vehicles\".\"make\""},
	Model:        whereHelpernull_String{field: "\"identity_api\".\"vehicles\".\"model\""},
	Year:         whereHelpernull_Int{field: "\"identity_api\".\"vehicles\".\"year\""},
	MintedAt:     whereHelpertime_Time{field: "\"identity_api\".\"vehicles\".\"minted_at\""},
}

// VehicleRels is where relationship names are stored.
var VehicleRels = struct {
	AftermarketDevice string
}{
	AftermarketDevice: "AftermarketDevice",
}

// vehicleR is where relationships are stored.
type vehicleR struct {
	AftermarketDevice *AftermarketDevice `boil:"AftermarketDevice" json:"AftermarketDevice" toml:"AftermarketDevice" yaml:"AftermarketDevice"`
}

// NewStruct creates a new relationship struct
func (*vehicleR) NewStruct() *vehicleR {
	return &vehicleR{}
}

func (r *vehicleR) GetAftermarketDevice() *AftermarketDevice {
	if r == nil {
		return nil
	}
	return r.AftermarketDevice
}

// vehicleL is where Load methods for each relationship are stored.
type vehicleL struct{}

var (
	vehicleAllColumns            = []string{"id", "owner_address", "make", "model", "year", "minted_at"}
	vehicleColumnsWithoutDefault = []string{"id", "owner_address", "minted_at"}
	vehicleColumnsWithDefault    = []string{"make", "model", "year"}
	vehiclePrimaryKeyColumns     = []string{"id"}
	vehicleGeneratedColumns      = []string{}
)

type (
	// VehicleSlice is an alias for a slice of pointers to Vehicle.
	// This should almost always be used instead of []Vehicle.
	VehicleSlice []*Vehicle
	// VehicleHook is the signature for custom Vehicle hook methods
	VehicleHook func(context.Context, boil.ContextExecutor, *Vehicle) error

	vehicleQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	vehicleType                 = reflect.TypeOf(&Vehicle{})
	vehicleMapping              = queries.MakeStructMapping(vehicleType)
	vehiclePrimaryKeyMapping, _ = queries.BindMapping(vehicleType, vehicleMapping, vehiclePrimaryKeyColumns)
	vehicleInsertCacheMut       sync.RWMutex
	vehicleInsertCache          = make(map[string]insertCache)
	vehicleUpdateCacheMut       sync.RWMutex
	vehicleUpdateCache          = make(map[string]updateCache)
	vehicleUpsertCacheMut       sync.RWMutex
	vehicleUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var vehicleAfterSelectHooks []VehicleHook

var vehicleBeforeInsertHooks []VehicleHook
var vehicleAfterInsertHooks []VehicleHook

var vehicleBeforeUpdateHooks []VehicleHook
var vehicleAfterUpdateHooks []VehicleHook

var vehicleBeforeDeleteHooks []VehicleHook
var vehicleAfterDeleteHooks []VehicleHook

var vehicleBeforeUpsertHooks []VehicleHook
var vehicleAfterUpsertHooks []VehicleHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Vehicle) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Vehicle) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Vehicle) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Vehicle) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Vehicle) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Vehicle) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Vehicle) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Vehicle) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Vehicle) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range vehicleAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddVehicleHook registers your hook function for all future operations.
func AddVehicleHook(hookPoint boil.HookPoint, vehicleHook VehicleHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		vehicleAfterSelectHooks = append(vehicleAfterSelectHooks, vehicleHook)
	case boil.BeforeInsertHook:
		vehicleBeforeInsertHooks = append(vehicleBeforeInsertHooks, vehicleHook)
	case boil.AfterInsertHook:
		vehicleAfterInsertHooks = append(vehicleAfterInsertHooks, vehicleHook)
	case boil.BeforeUpdateHook:
		vehicleBeforeUpdateHooks = append(vehicleBeforeUpdateHooks, vehicleHook)
	case boil.AfterUpdateHook:
		vehicleAfterUpdateHooks = append(vehicleAfterUpdateHooks, vehicleHook)
	case boil.BeforeDeleteHook:
		vehicleBeforeDeleteHooks = append(vehicleBeforeDeleteHooks, vehicleHook)
	case boil.AfterDeleteHook:
		vehicleAfterDeleteHooks = append(vehicleAfterDeleteHooks, vehicleHook)
	case boil.BeforeUpsertHook:
		vehicleBeforeUpsertHooks = append(vehicleBeforeUpsertHooks, vehicleHook)
	case boil.AfterUpsertHook:
		vehicleAfterUpsertHooks = append(vehicleAfterUpsertHooks, vehicleHook)
	}
}

// One returns a single vehicle record from the query.
func (q vehicleQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Vehicle, error) {
	o := &Vehicle{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for vehicles")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Vehicle records from the query.
func (q vehicleQuery) All(ctx context.Context, exec boil.ContextExecutor) (VehicleSlice, error) {
	var o []*Vehicle

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Vehicle slice")
	}

	if len(vehicleAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Vehicle records in the query.
func (q vehicleQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count vehicles rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q vehicleQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if vehicles exists")
	}

	return count > 0, nil
}

// AftermarketDevice pointed to by the foreign key.
func (o *Vehicle) AftermarketDevice(mods ...qm.QueryMod) aftermarketDeviceQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"vehicle_id\" = ?", o.ID),
	}

	queryMods = append(queryMods, mods...)

	return AftermarketDevices(queryMods...)
}

// LoadAftermarketDevice allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-1 relationship.
func (vehicleL) LoadAftermarketDevice(ctx context.Context, e boil.ContextExecutor, singular bool, maybeVehicle interface{}, mods queries.Applicator) error {
	var slice []*Vehicle
	var object *Vehicle

	if singular {
		var ok bool
		object, ok = maybeVehicle.(*Vehicle)
		if !ok {
			object = new(Vehicle)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeVehicle)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeVehicle))
			}
		}
	} else {
		s, ok := maybeVehicle.(*[]*Vehicle)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeVehicle)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeVehicle))
			}
		}
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &vehicleR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &vehicleR{}
			}

			for _, a := range args {
				if queries.Equal(a, obj.ID) {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`identity_api.aftermarket_devices`),
		qm.WhereIn(`identity_api.aftermarket_devices.vehicle_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load AftermarketDevice")
	}

	var resultSlice []*AftermarketDevice
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice AftermarketDevice")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for aftermarket_devices")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for aftermarket_devices")
	}

	if len(aftermarketDeviceAfterSelectHooks) != 0 {
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
		object.R.AftermarketDevice = foreign
		if foreign.R == nil {
			foreign.R = &aftermarketDeviceR{}
		}
		foreign.R.Vehicle = object
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if queries.Equal(local.ID, foreign.VehicleID) {
				local.R.AftermarketDevice = foreign
				if foreign.R == nil {
					foreign.R = &aftermarketDeviceR{}
				}
				foreign.R.Vehicle = local
				break
			}
		}
	}

	return nil
}

// SetAftermarketDevice of the vehicle to the related item.
// Sets o.R.AftermarketDevice to related.
// Adds o to related.R.Vehicle.
func (o *Vehicle) SetAftermarketDevice(ctx context.Context, exec boil.ContextExecutor, insert bool, related *AftermarketDevice) error {
	var err error

	if insert {
		queries.Assign(&related.VehicleID, o.ID)

		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	} else {
		updateQuery := fmt.Sprintf(
			"UPDATE \"identity_api\".\"aftermarket_devices\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, []string{"vehicle_id"}),
			strmangle.WhereClause("\"", "\"", 2, aftermarketDevicePrimaryKeyColumns),
		)
		values := []interface{}{o.ID, related.ID}

		if boil.IsDebug(ctx) {
			writer := boil.DebugWriterFrom(ctx)
			fmt.Fprintln(writer, updateQuery)
			fmt.Fprintln(writer, values)
		}
		if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
			return errors.Wrap(err, "failed to update foreign table")
		}

		queries.Assign(&related.VehicleID, o.ID)
	}

	if o.R == nil {
		o.R = &vehicleR{
			AftermarketDevice: related,
		}
	} else {
		o.R.AftermarketDevice = related
	}

	if related.R == nil {
		related.R = &aftermarketDeviceR{
			Vehicle: o,
		}
	} else {
		related.R.Vehicle = o
	}
	return nil
}

// RemoveAftermarketDevice relationship.
// Sets o.R.AftermarketDevice to nil.
// Removes o from all passed in related items' relationships struct.
func (o *Vehicle) RemoveAftermarketDevice(ctx context.Context, exec boil.ContextExecutor, related *AftermarketDevice) error {
	var err error

	queries.SetScanner(&related.VehicleID, nil)
	if _, err = related.Update(ctx, exec, boil.Whitelist("vehicle_id")); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	if o.R != nil {
		o.R.AftermarketDevice = nil
	}

	if related == nil || related.R == nil {
		return nil
	}

	related.R.Vehicle = nil

	return nil
}

// Vehicles retrieves all the records using an executor.
func Vehicles(mods ...qm.QueryMod) vehicleQuery {
	mods = append(mods, qm.From("\"identity_api\".\"vehicles\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"identity_api\".\"vehicles\".*"})
	}

	return vehicleQuery{q}
}

// FindVehicle retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindVehicle(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*Vehicle, error) {
	vehicleObj := &Vehicle{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"identity_api\".\"vehicles\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, vehicleObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from vehicles")
	}

	if err = vehicleObj.doAfterSelectHooks(ctx, exec); err != nil {
		return vehicleObj, err
	}

	return vehicleObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Vehicle) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no vehicles provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(vehicleColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	vehicleInsertCacheMut.RLock()
	cache, cached := vehicleInsertCache[key]
	vehicleInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			vehicleAllColumns,
			vehicleColumnsWithDefault,
			vehicleColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(vehicleType, vehicleMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(vehicleType, vehicleMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"identity_api\".\"vehicles\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"identity_api\".\"vehicles\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into vehicles")
	}

	if !cached {
		vehicleInsertCacheMut.Lock()
		vehicleInsertCache[key] = cache
		vehicleInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Vehicle.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Vehicle) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	vehicleUpdateCacheMut.RLock()
	cache, cached := vehicleUpdateCache[key]
	vehicleUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			vehicleAllColumns,
			vehiclePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update vehicles, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"identity_api\".\"vehicles\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, vehiclePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(vehicleType, vehicleMapping, append(wl, vehiclePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update vehicles row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for vehicles")
	}

	if !cached {
		vehicleUpdateCacheMut.Lock()
		vehicleUpdateCache[key] = cache
		vehicleUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q vehicleQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for vehicles")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for vehicles")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o VehicleSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), vehiclePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"identity_api\".\"vehicles\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, vehiclePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in vehicle slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all vehicle")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Vehicle) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no vehicles provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(vehicleColumnsWithDefault, o)

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

	vehicleUpsertCacheMut.RLock()
	cache, cached := vehicleUpsertCache[key]
	vehicleUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			vehicleAllColumns,
			vehicleColumnsWithDefault,
			vehicleColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			vehicleAllColumns,
			vehiclePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert vehicles, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(vehiclePrimaryKeyColumns))
			copy(conflict, vehiclePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"identity_api\".\"vehicles\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(vehicleType, vehicleMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(vehicleType, vehicleMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert vehicles")
	}

	if !cached {
		vehicleUpsertCacheMut.Lock()
		vehicleUpsertCache[key] = cache
		vehicleUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Vehicle record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Vehicle) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Vehicle provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), vehiclePrimaryKeyMapping)
	sql := "DELETE FROM \"identity_api\".\"vehicles\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from vehicles")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for vehicles")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q vehicleQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no vehicleQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from vehicles")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for vehicles")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o VehicleSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(vehicleBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), vehiclePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"identity_api\".\"vehicles\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, vehiclePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from vehicle slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for vehicles")
	}

	if len(vehicleAfterDeleteHooks) != 0 {
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
func (o *Vehicle) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindVehicle(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *VehicleSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := VehicleSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), vehiclePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"identity_api\".\"vehicles\".* FROM \"identity_api\".\"vehicles\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, vehiclePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in VehicleSlice")
	}

	*o = slice

	return nil
}

// VehicleExists checks if the Vehicle row exists.
func VehicleExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"identity_api\".\"vehicles\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if vehicles exists")
	}

	return exists, nil
}

// Exists checks if the Vehicle row exists.
func (o *Vehicle) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return VehicleExists(ctx, exec, o.ID)
}

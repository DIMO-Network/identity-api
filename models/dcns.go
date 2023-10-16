// Code generated by SQLBoiler 4.15.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
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

// DCN is an object representing the database table.
type DCN struct {
	Node         []byte      `boil:"node" json:"node" toml:"node" yaml:"node"`
	OwnerAddress []byte      `boil:"owner_address" json:"owner_address" toml:"owner_address" yaml:"owner_address"`
	Expiration   null.Time   `boil:"expiration" json:"expiration,omitempty" toml:"expiration" yaml:"expiration,omitempty"`
	Name         null.String `boil:"name" json:"name,omitempty" toml:"name" yaml:"name,omitempty"`
	VehicleID    null.Int    `boil:"vehicle_id" json:"vehicle_id,omitempty" toml:"vehicle_id" yaml:"vehicle_id,omitempty"`
	MintedAt     time.Time   `boil:"minted_at" json:"minted_at" toml:"minted_at" yaml:"minted_at"`

	R *dcnR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L dcnL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var DCNColumns = struct {
	Node         string
	OwnerAddress string
	Expiration   string
	Name         string
	VehicleID    string
	MintedAt     string
}{
	Node:         "node",
	OwnerAddress: "owner_address",
	Expiration:   "expiration",
	Name:         "name",
	VehicleID:    "vehicle_id",
	MintedAt:     "minted_at",
}

var DCNTableColumns = struct {
	Node         string
	OwnerAddress string
	Expiration   string
	Name         string
	VehicleID    string
	MintedAt     string
}{
	Node:         "dcns.node",
	OwnerAddress: "dcns.owner_address",
	Expiration:   "dcns.expiration",
	Name:         "dcns.name",
	VehicleID:    "dcns.vehicle_id",
	MintedAt:     "dcns.minted_at",
}

// Generated where

var DCNWhere = struct {
	Node         whereHelper__byte
	OwnerAddress whereHelper__byte
	Expiration   whereHelpernull_Time
	Name         whereHelpernull_String
	VehicleID    whereHelpernull_Int
	MintedAt     whereHelpertime_Time
}{
	Node:         whereHelper__byte{field: "\"identity_api\".\"dcns\".\"node\""},
	OwnerAddress: whereHelper__byte{field: "\"identity_api\".\"dcns\".\"owner_address\""},
	Expiration:   whereHelpernull_Time{field: "\"identity_api\".\"dcns\".\"expiration\""},
	Name:         whereHelpernull_String{field: "\"identity_api\".\"dcns\".\"name\""},
	VehicleID:    whereHelpernull_Int{field: "\"identity_api\".\"dcns\".\"vehicle_id\""},
	MintedAt:     whereHelpertime_Time{field: "\"identity_api\".\"dcns\".\"minted_at\""},
}

// DCNRels is where relationship names are stored.
var DCNRels = struct {
	Vehicle string
}{
	Vehicle: "Vehicle",
}

// dcnR is where relationships are stored.
type dcnR struct {
	Vehicle *Vehicle `boil:"Vehicle" json:"Vehicle" toml:"Vehicle" yaml:"Vehicle"`
}

// NewStruct creates a new relationship struct
func (*dcnR) NewStruct() *dcnR {
	return &dcnR{}
}

func (r *dcnR) GetVehicle() *Vehicle {
	if r == nil {
		return nil
	}
	return r.Vehicle
}

// dcnL is where Load methods for each relationship are stored.
type dcnL struct{}

var (
	dcnAllColumns            = []string{"node", "owner_address", "expiration", "name", "vehicle_id", "minted_at"}
	dcnColumnsWithoutDefault = []string{"node", "owner_address", "minted_at"}
	dcnColumnsWithDefault    = []string{"expiration", "name", "vehicle_id"}
	dcnPrimaryKeyColumns     = []string{"node"}
	dcnGeneratedColumns      = []string{}
)

type (
	// DCNSlice is an alias for a slice of pointers to DCN.
	// This should almost always be used instead of []DCN.
	DCNSlice []*DCN
	// DCNHook is the signature for custom DCN hook methods
	DCNHook func(context.Context, boil.ContextExecutor, *DCN) error

	dcnQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	dcnType                 = reflect.TypeOf(&DCN{})
	dcnMapping              = queries.MakeStructMapping(dcnType)
	dcnPrimaryKeyMapping, _ = queries.BindMapping(dcnType, dcnMapping, dcnPrimaryKeyColumns)
	dcnInsertCacheMut       sync.RWMutex
	dcnInsertCache          = make(map[string]insertCache)
	dcnUpdateCacheMut       sync.RWMutex
	dcnUpdateCache          = make(map[string]updateCache)
	dcnUpsertCacheMut       sync.RWMutex
	dcnUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var dcnAfterSelectHooks []DCNHook

var dcnBeforeInsertHooks []DCNHook
var dcnAfterInsertHooks []DCNHook

var dcnBeforeUpdateHooks []DCNHook
var dcnAfterUpdateHooks []DCNHook

var dcnBeforeDeleteHooks []DCNHook
var dcnAfterDeleteHooks []DCNHook

var dcnBeforeUpsertHooks []DCNHook
var dcnAfterUpsertHooks []DCNHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *DCN) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dcnAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *DCN) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dcnBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *DCN) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dcnAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *DCN) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dcnBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *DCN) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dcnAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *DCN) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dcnBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *DCN) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dcnAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *DCN) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dcnBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *DCN) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dcnAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddDCNHook registers your hook function for all future operations.
func AddDCNHook(hookPoint boil.HookPoint, dcnHook DCNHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		dcnAfterSelectHooks = append(dcnAfterSelectHooks, dcnHook)
	case boil.BeforeInsertHook:
		dcnBeforeInsertHooks = append(dcnBeforeInsertHooks, dcnHook)
	case boil.AfterInsertHook:
		dcnAfterInsertHooks = append(dcnAfterInsertHooks, dcnHook)
	case boil.BeforeUpdateHook:
		dcnBeforeUpdateHooks = append(dcnBeforeUpdateHooks, dcnHook)
	case boil.AfterUpdateHook:
		dcnAfterUpdateHooks = append(dcnAfterUpdateHooks, dcnHook)
	case boil.BeforeDeleteHook:
		dcnBeforeDeleteHooks = append(dcnBeforeDeleteHooks, dcnHook)
	case boil.AfterDeleteHook:
		dcnAfterDeleteHooks = append(dcnAfterDeleteHooks, dcnHook)
	case boil.BeforeUpsertHook:
		dcnBeforeUpsertHooks = append(dcnBeforeUpsertHooks, dcnHook)
	case boil.AfterUpsertHook:
		dcnAfterUpsertHooks = append(dcnAfterUpsertHooks, dcnHook)
	}
}

// One returns a single dcn record from the query.
func (q dcnQuery) One(ctx context.Context, exec boil.ContextExecutor) (*DCN, error) {
	o := &DCN{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for dcns")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all DCN records from the query.
func (q dcnQuery) All(ctx context.Context, exec boil.ContextExecutor) (DCNSlice, error) {
	var o []*DCN

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to DCN slice")
	}

	if len(dcnAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all DCN records in the query.
func (q dcnQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count dcns rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q dcnQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if dcns exists")
	}

	return count > 0, nil
}

// Vehicle pointed to by the foreign key.
func (o *DCN) Vehicle(mods ...qm.QueryMod) vehicleQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.VehicleID),
	}

	queryMods = append(queryMods, mods...)

	return Vehicles(queryMods...)
}

// LoadVehicle allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (dcnL) LoadVehicle(ctx context.Context, e boil.ContextExecutor, singular bool, maybeDCN interface{}, mods queries.Applicator) error {
	var slice []*DCN
	var object *DCN

	if singular {
		var ok bool
		object, ok = maybeDCN.(*DCN)
		if !ok {
			object = new(DCN)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeDCN)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeDCN))
			}
		}
	} else {
		s, ok := maybeDCN.(*[]*DCN)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeDCN)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeDCN))
			}
		}
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &dcnR{}
		}
		if !queries.IsNil(object.VehicleID) {
			args = append(args, object.VehicleID)
		}

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &dcnR{}
			}

			for _, a := range args {
				if queries.Equal(a, obj.VehicleID) {
					continue Outer
				}
			}

			if !queries.IsNil(obj.VehicleID) {
				args = append(args, obj.VehicleID)
			}

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
		object.R.Vehicle = foreign
		if foreign.R == nil {
			foreign.R = &vehicleR{}
		}
		foreign.R.DCNS = append(foreign.R.DCNS, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if queries.Equal(local.VehicleID, foreign.ID) {
				local.R.Vehicle = foreign
				if foreign.R == nil {
					foreign.R = &vehicleR{}
				}
				foreign.R.DCNS = append(foreign.R.DCNS, local)
				break
			}
		}
	}

	return nil
}

// SetVehicle of the dcn to the related item.
// Sets o.R.Vehicle to related.
// Adds o to related.R.DCNS.
func (o *DCN) SetVehicle(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Vehicle) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"identity_api\".\"dcns\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"vehicle_id"}),
		strmangle.WhereClause("\"", "\"", 2, dcnPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.Node}

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
		o.R = &dcnR{
			Vehicle: related,
		}
	} else {
		o.R.Vehicle = related
	}

	if related.R == nil {
		related.R = &vehicleR{
			DCNS: DCNSlice{o},
		}
	} else {
		related.R.DCNS = append(related.R.DCNS, o)
	}

	return nil
}

// RemoveVehicle relationship.
// Sets o.R.Vehicle to nil.
// Removes o from all passed in related items' relationships struct.
func (o *DCN) RemoveVehicle(ctx context.Context, exec boil.ContextExecutor, related *Vehicle) error {
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

	for i, ri := range related.R.DCNS {
		if queries.Equal(o.VehicleID, ri.VehicleID) {
			continue
		}

		ln := len(related.R.DCNS)
		if ln > 1 && i < ln-1 {
			related.R.DCNS[i] = related.R.DCNS[ln-1]
		}
		related.R.DCNS = related.R.DCNS[:ln-1]
		break
	}
	return nil
}

// DCNS retrieves all the records using an executor.
func DCNS(mods ...qm.QueryMod) dcnQuery {
	mods = append(mods, qm.From("\"identity_api\".\"dcns\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"identity_api\".\"dcns\".*"})
	}

	return dcnQuery{q}
}

// FindDCN retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindDCN(ctx context.Context, exec boil.ContextExecutor, node []byte, selectCols ...string) (*DCN, error) {
	dcnObj := &DCN{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"identity_api\".\"dcns\" where \"node\"=$1", sel,
	)

	q := queries.Raw(query, node)

	err := q.Bind(ctx, exec, dcnObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from dcns")
	}

	if err = dcnObj.doAfterSelectHooks(ctx, exec); err != nil {
		return dcnObj, err
	}

	return dcnObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *DCN) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no dcns provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(dcnColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	dcnInsertCacheMut.RLock()
	cache, cached := dcnInsertCache[key]
	dcnInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			dcnAllColumns,
			dcnColumnsWithDefault,
			dcnColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(dcnType, dcnMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(dcnType, dcnMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"identity_api\".\"dcns\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"identity_api\".\"dcns\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into dcns")
	}

	if !cached {
		dcnInsertCacheMut.Lock()
		dcnInsertCache[key] = cache
		dcnInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the DCN.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *DCN) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	dcnUpdateCacheMut.RLock()
	cache, cached := dcnUpdateCache[key]
	dcnUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			dcnAllColumns,
			dcnPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update dcns, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"identity_api\".\"dcns\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, dcnPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(dcnType, dcnMapping, append(wl, dcnPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update dcns row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for dcns")
	}

	if !cached {
		dcnUpdateCacheMut.Lock()
		dcnUpdateCache[key] = cache
		dcnUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q dcnQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for dcns")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for dcns")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o DCNSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), dcnPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"identity_api\".\"dcns\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, dcnPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in dcn slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all dcn")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *DCN) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no dcns provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(dcnColumnsWithDefault, o)

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

	dcnUpsertCacheMut.RLock()
	cache, cached := dcnUpsertCache[key]
	dcnUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			dcnAllColumns,
			dcnColumnsWithDefault,
			dcnColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			dcnAllColumns,
			dcnPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert dcns, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(dcnPrimaryKeyColumns))
			copy(conflict, dcnPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"identity_api\".\"dcns\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(dcnType, dcnMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(dcnType, dcnMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert dcns")
	}

	if !cached {
		dcnUpsertCacheMut.Lock()
		dcnUpsertCache[key] = cache
		dcnUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single DCN record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *DCN) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no DCN provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), dcnPrimaryKeyMapping)
	sql := "DELETE FROM \"identity_api\".\"dcns\" WHERE \"node\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from dcns")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for dcns")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q dcnQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no dcnQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from dcns")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for dcns")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o DCNSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(dcnBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), dcnPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"identity_api\".\"dcns\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, dcnPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from dcn slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for dcns")
	}

	if len(dcnAfterDeleteHooks) != 0 {
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
func (o *DCN) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindDCN(ctx, exec, o.Node)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *DCNSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := DCNSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), dcnPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"identity_api\".\"dcns\".* FROM \"identity_api\".\"dcns\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, dcnPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in DCNSlice")
	}

	*o = slice

	return nil
}

// DCNExists checks if the DCN row exists.
func DCNExists(ctx context.Context, exec boil.ContextExecutor, node []byte) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"identity_api\".\"dcns\" where \"node\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, node)
	}
	row := exec.QueryRowContext(ctx, sql, node)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if dcns exists")
	}

	return exists, nil
}

// Exists checks if the DCN row exists.
func (o *DCN) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return DCNExists(ctx, exec, o.Node)
}

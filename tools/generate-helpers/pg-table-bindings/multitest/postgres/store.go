// Code generated by pg-bindings generator. DO NOT EDIT.

package postgres

import (
	"context"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stackrox/rox/central/metrics"
	pkgSchema "github.com/stackrox/rox/central/postgres/schema"
	v1 "github.com/stackrox/rox/generated/api/v1"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/logging"
	ops "github.com/stackrox/rox/pkg/metrics"
	"github.com/stackrox/rox/pkg/postgres/pgutils"
	"github.com/stackrox/rox/pkg/search"
	"github.com/stackrox/rox/pkg/search/postgres"
)

const (
	baseTable = "test_multi_key_structs"

	existsStmt = "SELECT EXISTS(SELECT 1 FROM test_multi_key_structs WHERE Key1 = $1 AND Key2 = $2)"
	getStmt    = "SELECT serialized FROM test_multi_key_structs WHERE Key1 = $1 AND Key2 = $2"
	deleteStmt = "DELETE FROM test_multi_key_structs WHERE Key1 = $1 AND Key2 = $2"

	walkStmt    = "SELECT serialized FROM test_multi_key_structs"
	getManyStmt = "SELECT serialized FROM test_multi_key_structs WHERE Key1 = ANY($1::text[])"

	batchAfter = 100

	// using copyFrom, we may not even want to batch.  It would probably be simpler
	// to deal with failures if we just sent it all.  Something to think about as we
	// proceed and move into more e2e and larger performance testing
	batchSize = 10000
)

var (
	log    = logging.LoggerForModule()
	schema = pkgSchema.TestMultiKeyStructsSchema
)

type Store interface {
	Count(ctx context.Context) (int, error)
	Exists(ctx context.Context, key1 string, key2 string) (bool, error)
	Get(ctx context.Context, key1 string, key2 string) (*storage.TestMultiKeyStruct, bool, error)
	Upsert(ctx context.Context, obj *storage.TestMultiKeyStruct) error
	UpsertMany(ctx context.Context, objs []*storage.TestMultiKeyStruct) error
	Delete(ctx context.Context, key1 string, key2 string) error
	GetIDs(ctx context.Context) ([]string, error)
	GetMany(ctx context.Context, ids []string) ([]*storage.TestMultiKeyStruct, []int, error)
	DeleteMany(ctx context.Context, ids []string) error

	Walk(ctx context.Context, fn func(obj *storage.TestMultiKeyStruct) error) error

	AckKeysIndexed(ctx context.Context, keys ...string) error
	GetKeysToIndex(ctx context.Context) ([]string, error)
}

type storeImpl struct {
	db *pgxpool.Pool
}

// New returns a new Store instance using the provided sql instance.
func New(ctx context.Context, db *pgxpool.Pool) Store {
	pgutils.CreateTable(ctx, db, pkgSchema.CreateTableTestMultiKeyStructsStmt)

	return &storeImpl{
		db: db,
	}
}

func insertIntoTestMultiKeyStructs(ctx context.Context, tx pgx.Tx, obj *storage.TestMultiKeyStruct) error {

	serialized, marshalErr := obj.Marshal()
	if marshalErr != nil {
		return marshalErr
	}

	values := []interface{}{
		// parent primary keys start
		obj.GetKey1(),
		obj.GetKey2(),
		obj.GetStringSlice(),
		obj.GetBool(),
		obj.GetUint64(),
		obj.GetInt64(),
		obj.GetFloat(),
		obj.GetLabels(),
		pgutils.NilOrTime(obj.GetTimestamp()),
		obj.GetEnum(),
		obj.GetEnums(),
		obj.GetString_(),
		obj.GetIntSlice(),
		obj.GetOneofnested().GetNested(),
		serialized,
	}

	finalStr := "INSERT INTO test_multi_key_structs (Key1, Key2, StringSlice, Bool, Uint64, Int64, Float, Labels, Timestamp, Enum, Enums, String_, IntSlice, Oneofnested_Nested, serialized) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) ON CONFLICT(Key1, Key2) DO UPDATE SET Key1 = EXCLUDED.Key1, Key2 = EXCLUDED.Key2, StringSlice = EXCLUDED.StringSlice, Bool = EXCLUDED.Bool, Uint64 = EXCLUDED.Uint64, Int64 = EXCLUDED.Int64, Float = EXCLUDED.Float, Labels = EXCLUDED.Labels, Timestamp = EXCLUDED.Timestamp, Enum = EXCLUDED.Enum, Enums = EXCLUDED.Enums, String_ = EXCLUDED.String_, IntSlice = EXCLUDED.IntSlice, Oneofnested_Nested = EXCLUDED.Oneofnested_Nested, serialized = EXCLUDED.serialized"
	_, err := tx.Exec(ctx, finalStr, values...)
	if err != nil {
		return err
	}

	var query string

	for childIdx, child := range obj.GetNested() {
		if err := insertIntoTestMultiKeyStructsNesteds(ctx, tx, child, obj.GetKey1(), obj.GetKey2(), childIdx); err != nil {
			return err
		}
	}

	query = "delete from test_multi_key_structs_nesteds where test_multi_key_structs_Key1 = $1 AND test_multi_key_structs_Key2 = $2 AND idx >= $3"
	_, err = tx.Exec(ctx, query, obj.GetKey1(), obj.GetKey2(), len(obj.GetNested()))
	if err != nil {
		return err
	}
	return nil
}

func insertIntoTestMultiKeyStructsNesteds(ctx context.Context, tx pgx.Tx, obj *storage.TestMultiKeyStruct_Nested, test_multi_key_structs_Key1 string, test_multi_key_structs_Key2 string, idx int) error {

	values := []interface{}{
		// parent primary keys start
		test_multi_key_structs_Key1,
		test_multi_key_structs_Key2,
		idx,
		obj.GetNested(),
		obj.GetIsNested(),
		obj.GetInt64(),
		obj.GetNested2().GetNested2(),
		obj.GetNested2().GetIsNested(),
		obj.GetNested2().GetInt64(),
	}

	finalStr := "INSERT INTO test_multi_key_structs_nesteds (test_multi_key_structs_Key1, test_multi_key_structs_Key2, idx, Nested, IsNested, Int64, Nested2_Nested2, Nested2_IsNested, Nested2_Int64) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT(test_multi_key_structs_Key1, test_multi_key_structs_Key2, idx) DO UPDATE SET test_multi_key_structs_Key1 = EXCLUDED.test_multi_key_structs_Key1, test_multi_key_structs_Key2 = EXCLUDED.test_multi_key_structs_Key2, idx = EXCLUDED.idx, Nested = EXCLUDED.Nested, IsNested = EXCLUDED.IsNested, Int64 = EXCLUDED.Int64, Nested2_Nested2 = EXCLUDED.Nested2_Nested2, Nested2_IsNested = EXCLUDED.Nested2_IsNested, Nested2_Int64 = EXCLUDED.Nested2_Int64"
	_, err := tx.Exec(ctx, finalStr, values...)
	if err != nil {
		return err
	}

	return nil
}

func (s *storeImpl) copyFromTestMultiKeyStructs(ctx context.Context, tx pgx.Tx, objs ...*storage.TestMultiKeyStruct) error {

	inputRows := [][]interface{}{}

	var err error

	copyCols := []string{

		"key1",

		"key2",

		"stringslice",

		"bool",

		"uint64",

		"int64",

		"float",

		"labels",

		"timestamp",

		"enum",

		"enums",

		"string_",

		"intslice",

		"oneofnested_nested",

		"serialized",
	}

	for idx, obj := range objs {
		// Todo: ROX-9499 Figure out how to more cleanly template around this issue.
		log.Debugf("This is here for now because there is an issue with pods_TerminatedInstances where the obj in the loop is not used as it only consists of the parent id and the idx.  Putting this here as a stop gap to simply use the object.  %s", obj)

		serialized, marshalErr := obj.Marshal()
		if marshalErr != nil {
			return marshalErr
		}

		inputRows = append(inputRows, []interface{}{

			obj.GetKey1(),

			obj.GetKey2(),

			obj.GetStringSlice(),

			obj.GetBool(),

			obj.GetUint64(),

			obj.GetInt64(),

			obj.GetFloat(),

			obj.GetLabels(),

			pgutils.NilOrTime(obj.GetTimestamp()),

			obj.GetEnum(),

			obj.GetEnums(),

			obj.GetString_(),

			obj.GetIntSlice(),

			obj.GetOneofnested().GetNested(),

			serialized,
		})

		if err := s.Delete(ctx, obj.GetKey1(), obj.GetKey2()); err != nil {
			return err
		}

		// if we hit our batch size we need to push the data
		if (idx+1)%batchSize == 0 || idx == len(objs)-1 {
			// copy does not upsert so have to delete first.  parent deletion cascades so only need to
			// delete for the top level parent

			_, err = tx.CopyFrom(ctx, pgx.Identifier{"test_multi_key_structs"}, copyCols, pgx.CopyFromRows(inputRows))

			if err != nil {
				return err
			}

			// clear the input rows for the next batch
			inputRows = inputRows[:0]
		}
	}

	for idx, obj := range objs {
		_ = idx // idx may or may not be used depending on how nested we are, so avoid compile-time errors.

		if err = s.copyFromTestMultiKeyStructsNesteds(ctx, tx, obj.GetKey1(), obj.GetKey2(), obj.GetNested()...); err != nil {
			return err
		}
	}

	return err
}

func (s *storeImpl) copyFromTestMultiKeyStructsNesteds(ctx context.Context, tx pgx.Tx, test_multi_key_structs_Key1 string, test_multi_key_structs_Key2 string, objs ...*storage.TestMultiKeyStruct_Nested) error {

	inputRows := [][]interface{}{}

	var err error

	copyCols := []string{

		"test_multi_key_structs_key1",

		"test_multi_key_structs_key2",

		"idx",

		"nested",

		"isnested",

		"int64",

		"nested2_nested2",

		"nested2_isnested",

		"nested2_int64",
	}

	for idx, obj := range objs {
		// Todo: ROX-9499 Figure out how to more cleanly template around this issue.
		log.Debugf("This is here for now because there is an issue with pods_TerminatedInstances where the obj in the loop is not used as it only consists of the parent id and the idx.  Putting this here as a stop gap to simply use the object.  %s", obj)

		inputRows = append(inputRows, []interface{}{

			test_multi_key_structs_Key1,

			test_multi_key_structs_Key2,

			idx,

			obj.GetNested(),

			obj.GetIsNested(),

			obj.GetInt64(),

			obj.GetNested2().GetNested2(),

			obj.GetNested2().GetIsNested(),

			obj.GetNested2().GetInt64(),
		})

		// if we hit our batch size we need to push the data
		if (idx+1)%batchSize == 0 || idx == len(objs)-1 {
			// copy does not upsert so have to delete first.  parent deletion cascades so only need to
			// delete for the top level parent

			_, err = tx.CopyFrom(ctx, pgx.Identifier{"test_multi_key_structs_nesteds"}, copyCols, pgx.CopyFromRows(inputRows))

			if err != nil {
				return err
			}

			// clear the input rows for the next batch
			inputRows = inputRows[:0]
		}
	}

	return err
}

func (s *storeImpl) copyFrom(ctx context.Context, objs ...*storage.TestMultiKeyStruct) error {
	conn, release, err := s.acquireConn(ctx, ops.Get, "TestMultiKeyStruct")
	if err != nil {
		return err
	}
	defer release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	if err := s.copyFromTestMultiKeyStructs(ctx, tx, objs...); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return err
		}
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (s *storeImpl) upsert(ctx context.Context, objs ...*storage.TestMultiKeyStruct) error {
	conn, release, err := s.acquireConn(ctx, ops.Get, "TestMultiKeyStruct")
	if err != nil {
		return err
	}
	defer release()

	for _, obj := range objs {
		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}

		if err := insertIntoTestMultiKeyStructs(ctx, tx, obj); err != nil {
			if err := tx.Rollback(ctx); err != nil {
				return err
			}
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *storeImpl) Upsert(ctx context.Context, obj *storage.TestMultiKeyStruct) error {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.Upsert, "TestMultiKeyStruct")

	return s.upsert(ctx, obj)
}

func (s *storeImpl) UpsertMany(ctx context.Context, objs []*storage.TestMultiKeyStruct) error {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.UpdateMany, "TestMultiKeyStruct")

	if len(objs) < batchAfter {
		return s.upsert(ctx, objs...)
	} else {
		return s.copyFrom(ctx, objs...)
	}
}

// Count returns the number of objects in the store
func (s *storeImpl) Count(ctx context.Context) (int, error) {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.Count, "TestMultiKeyStruct")

	var sacQueryFilter *v1.Query

	return postgres.RunCountRequestForSchema(schema, sacQueryFilter, s.db)
}

// Exists returns if the id exists in the store
func (s *storeImpl) Exists(ctx context.Context, key1 string, key2 string) (bool, error) {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.Exists, "TestMultiKeyStruct")

	row := s.db.QueryRow(ctx, existsStmt, key1, key2)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, pgutils.ErrNilIfNoRows(err)
	}
	return exists, nil
}

// Get returns the object, if it exists from the store
func (s *storeImpl) Get(ctx context.Context, key1 string, key2 string) (*storage.TestMultiKeyStruct, bool, error) {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.Get, "TestMultiKeyStruct")

	conn, release, err := s.acquireConn(ctx, ops.Get, "TestMultiKeyStruct")
	if err != nil {
		return nil, false, err
	}
	defer release()

	row := conn.QueryRow(ctx, getStmt, key1, key2)
	var data []byte
	if err := row.Scan(&data); err != nil {
		return nil, false, pgutils.ErrNilIfNoRows(err)
	}

	var msg storage.TestMultiKeyStruct
	if err := proto.Unmarshal(data, &msg); err != nil {
		return nil, false, err
	}
	return &msg, true, nil
}

func (s *storeImpl) acquireConn(ctx context.Context, op ops.Op, typ string) (*pgxpool.Conn, func(), error) {
	defer metrics.SetAcquireDBConnDuration(time.Now(), op, typ)
	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return nil, nil, err
	}
	return conn, conn.Release, nil
}

// Delete removes the specified ID from the store
func (s *storeImpl) Delete(ctx context.Context, key1 string, key2 string) error {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.Remove, "TestMultiKeyStruct")

	conn, release, err := s.acquireConn(ctx, ops.Remove, "TestMultiKeyStruct")
	if err != nil {
		return err
	}
	defer release()

	if _, err := conn.Exec(ctx, deleteStmt, key1, key2); err != nil {
		return err
	}
	return nil
}

// GetIDs returns all the IDs for the store
func (s *storeImpl) GetIDs(ctx context.Context) ([]string, error) {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.GetAll, "storage.TestMultiKeyStructIDs")
	var sacQueryFilter *v1.Query

	result, err := postgres.RunSearchRequestForSchema(schema, sacQueryFilter, s.db)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(result))
	for _, entry := range result {
		ids = append(ids, entry.ID)
	}

	return ids, nil
}

// GetMany returns the objects specified by the IDs or the index in the missing indices slice
func (s *storeImpl) GetMany(ctx context.Context, ids []string) ([]*storage.TestMultiKeyStruct, []int, error) {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.GetMany, "TestMultiKeyStruct")

	conn, release, err := s.acquireConn(ctx, ops.GetMany, "TestMultiKeyStruct")
	if err != nil {
		return nil, nil, err
	}
	defer release()

	rows, err := conn.Query(ctx, getManyStmt, ids)
	if err != nil {
		if err == pgx.ErrNoRows {
			missingIndices := make([]int, 0, len(ids))
			for i := range ids {
				missingIndices = append(missingIndices, i)
			}
			return nil, missingIndices, nil
		}
		return nil, nil, err
	}
	defer rows.Close()
	resultsByID := make(map[string]*storage.TestMultiKeyStruct)
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, nil, err
		}
		msg := &storage.TestMultiKeyStruct{}
		if err := proto.Unmarshal(data, msg); err != nil {
			return nil, nil, err
		}
		resultsByID[msg.GetKey1()] = msg
	}
	missingIndices := make([]int, 0, len(ids)-len(resultsByID))
	// It is important that the elems are populated in the same order as the input ids
	// slice, since some calling code relies on that to maintain order.
	elems := make([]*storage.TestMultiKeyStruct, 0, len(resultsByID))
	for i, id := range ids {
		if result, ok := resultsByID[id]; !ok {
			missingIndices = append(missingIndices, i)
		} else {
			elems = append(elems, result)
		}
	}
	return elems, missingIndices, nil
}

// Delete removes the specified IDs from the store
func (s *storeImpl) DeleteMany(ctx context.Context, ids []string) error {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.RemoveMany, "TestMultiKeyStruct")

	var sacQueryFilter *v1.Query

	q := search.ConjunctionQuery(
		sacQueryFilter,
		search.NewQueryBuilder().AddDocIDs(ids...).ProtoQuery(),
	)

	return postgres.RunDeleteRequestForSchema(schema, q, s.db)
}

// Walk iterates over all of the objects in the store and applies the closure
func (s *storeImpl) Walk(ctx context.Context, fn func(obj *storage.TestMultiKeyStruct) error) error {
	rows, err := s.db.Query(ctx, walkStmt)
	if err != nil {
		return pgutils.ErrNilIfNoRows(err)
	}
	defer rows.Close()
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return err
		}
		var msg storage.TestMultiKeyStruct
		if err := proto.Unmarshal(data, &msg); err != nil {
			return err
		}
		if err := fn(&msg); err != nil {
			return err
		}
	}
	return nil
}

//// Used for testing

func dropTableTestMultiKeyStructs(ctx context.Context, db *pgxpool.Pool) {
	_, _ = db.Exec(ctx, "DROP TABLE IF EXISTS test_multi_key_structs CASCADE")
	dropTableTestMultiKeyStructsNesteds(ctx, db)

}

func dropTableTestMultiKeyStructsNesteds(ctx context.Context, db *pgxpool.Pool) {
	_, _ = db.Exec(ctx, "DROP TABLE IF EXISTS test_multi_key_structs_nesteds CASCADE")

}

func Destroy(ctx context.Context, db *pgxpool.Pool) {
	dropTableTestMultiKeyStructs(ctx, db)
}

//// Stubs for satisfying legacy interfaces

// AckKeysIndexed acknowledges the passed keys were indexed
func (s *storeImpl) AckKeysIndexed(ctx context.Context, keys ...string) error {
	return nil
}

// GetKeysToIndex returns the keys that need to be indexed
func (s *storeImpl) GetKeysToIndex(ctx context.Context) ([]string, error) {
	return nil, nil
}

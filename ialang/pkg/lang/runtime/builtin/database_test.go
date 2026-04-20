package builtin

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	rt "ialang/pkg/lang/runtime"
)

var errTestDatabaseCallback = errors.New("test database callback error")

func openTestSQLiteDB(t *testing.T) Object {
	t.Helper()
	mod := mustModuleObject(t, DefaultModules(rt.NewGoroutineRuntime()), "db")
	path := filepath.Join(t.TempDir(), "test.sqlite")
	return callNative(t, mod, "sqlite", path).(Object)
}

func TestDatabaseSQLiteQueryExecPrepareAndClose(t *testing.T) {
	db := openTestSQLiteDB(t)

	if ping := callNative(t, db, "ping"); ping != true {
		t.Fatalf("expected ping true, got %#v", ping)
	}

	createResult := callNative(t, db, "exec", "CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, age INTEGER)").(Object)
	if createResult["affectedRows"].(float64) != 0 {
		t.Fatalf("expected create affectedRows 0, got %#v", createResult)
	}

	insertResult := callNative(t, db, "execute", "INSERT INTO users (name, age) VALUES (?, ?)", Array{"alice", float64(30)}).(Object)
	if insertResult["affectedRows"].(float64) != 1 {
		t.Fatalf("expected insert affectedRows 1, got %#v", insertResult)
	}

	stmt := callNative(t, db, "prepare", "INSERT INTO users (name, age) VALUES (?, ?)").(Object)
	stmtResult := callNative(t, stmt, "exec", Array{"bob", float64(25)}).(Object)
	if stmtResult["affectedRows"].(float64) != 1 {
		t.Fatalf("expected statement insert affectedRows 1, got %#v", stmtResult)
	}
	if closed := callNative(t, stmt, "close"); closed != true {
		t.Fatalf("expected statement close true, got %#v", closed)
	}

	rows := callNative(t, db, "query", "SELECT name, age FROM users ORDER BY id").(Array)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %#v", rows)
	}
	first := rows[0].(Object)
	if first["name"] != "alice" || first["age"] != float64(30) {
		t.Fatalf("unexpected first row: %#v", first)
	}

	one := callNative(t, db, "queryOne", "SELECT name, age FROM users WHERE name = ?", Array{"bob"}).(Object)
	if one["name"] != "bob" || one["age"] != float64(25) {
		t.Fatalf("unexpected queryOne row: %#v", one)
	}

	empty := callNative(t, db, "queryOne", "SELECT name FROM users WHERE name = ?", Array{"missing"}).(Object)
	if len(empty) != 0 {
		t.Fatalf("expected empty object for missing row, got %#v", empty)
	}

	selectStmt := callNative(t, db, "prepare", "SELECT name FROM users WHERE age > ? ORDER BY age").(Object)
	stmtRows := callNative(t, selectStmt, "query", Array{float64(20)}).(Array)
	if len(stmtRows) != 2 {
		t.Fatalf("expected statement query rows 2, got %#v", stmtRows)
	}
	_ = callNative(t, selectStmt, "close")

	if closed := callNative(t, db, "close"); closed != true {
		t.Fatalf("expected close true, got %#v", closed)
	}
}

func TestDatabaseTransactionCommitAndRollback(t *testing.T) {
	db := openTestSQLiteDB(t)
	defer callNative(t, db, "close")
	_ = callNative(t, db, "exec", "CREATE TABLE events (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")

	commitResult := callNative(t, db, "transaction", NativeFunction(func(args []Value) (Value, error) {
		tx := args[0].(Object)
		_ = callNative(t, tx, "exec", "INSERT INTO events (name) VALUES (?)", Array{"commit"})
		rows := callNative(t, tx, "query", "SELECT name FROM events").(Array)
		if len(rows) != 1 {
			t.Fatalf("expected tx query to see inserted row, got %#v", rows)
		}
		return true, nil
	})).(Object)
	if commitResult["committed"] != true {
		t.Fatalf("expected committed transaction, got %#v", commitResult)
	}

	rollbackResult := callNative(t, db, "transaction", NativeFunction(func(args []Value) (Value, error) {
		tx := args[0].(Object)
		_ = callNative(t, tx, "exec", "INSERT INTO events (name) VALUES (?)", Array{"rollback"})
		return false, nil
	})).(Object)
	if rollbackResult["committed"] != false {
		t.Fatalf("expected rolled back transaction, got %#v", rollbackResult)
	}

	rows := callNative(t, db, "query", "SELECT name FROM events ORDER BY id").(Array)
	if len(rows) != 1 || rows[0].(Object)["name"] != "commit" {
		t.Fatalf("expected only committed row, got %#v", rows)
	}
}

func TestDatabaseAsyncFunctions(t *testing.T) {
	db := openTestSQLiteDB(t)
	defer callNative(t, db, "close")
	_ = callNative(t, db, "exec", "CREATE TABLE items (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")

	execPromise := callNative(t, db, "execAsync", "INSERT INTO items (name) VALUES (?)", Array{"async"})
	execResult := awaitValue(t, execPromise).(Object)
	if execResult["affectedRows"].(float64) != 1 {
		t.Fatalf("expected async affectedRows 1, got %#v", execResult)
	}

	queryPromise := callNative(t, db, "queryAsync", "SELECT name FROM items")
	rows := awaitValue(t, queryPromise).(Array)
	if len(rows) != 1 || rows[0].(Object)["name"] != "async" {
		t.Fatalf("unexpected async query rows: %#v", rows)
	}

	onePromise := callNative(t, db, "queryOneAsync", "SELECT name FROM items WHERE name = ?", Array{"async"})
	one := awaitValue(t, onePromise).(Object)
	if one["name"] != "async" {
		t.Fatalf("unexpected async queryOne row: %#v", one)
	}
}

func TestDatabaseCreateTransactionObjectDirectMethods(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tx.sqlite")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open() error: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("CREATE TABLE events (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)"); err != nil {
		t.Fatalf("create table error: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("db.Begin() error: %v", err)
	}
	txObj := createTransactionObject(tx)

	execResult := callNative(t, txObj, "exec", "INSERT INTO events (name) VALUES (?)", Array{"commit-direct"}).(Object)
	if execResult["affectedRows"] != float64(1) {
		t.Fatalf("tx.exec affectedRows = %#v, want 1", execResult)
	}
	rows := callNative(t, txObj, "query", "SELECT name FROM events ORDER BY id").(Array)
	if len(rows) != 1 || rows[0].(Object)["name"] != "commit-direct" {
		t.Fatalf("tx.query rows = %#v, want committed candidate row", rows)
	}
	if committed := callNative(t, txObj, "commit"); committed != true {
		t.Fatalf("tx.commit = %#v, want true", committed)
	}

	persistedRows, err := db.Query("SELECT name FROM events ORDER BY id")
	if err != nil {
		t.Fatalf("db.Query() error: %v", err)
	}
	defer persistedRows.Close()
	if !persistedRows.Next() {
		t.Fatal("expected committed row to persist")
	}
	var name string
	if err := persistedRows.Scan(&name); err != nil {
		t.Fatalf("rows.Scan() error: %v", err)
	}
	if name != "commit-direct" {
		t.Fatalf("persisted row name = %q, want commit-direct", name)
	}

	tx2, err := db.Begin()
	if err != nil {
		t.Fatalf("db.Begin() second tx error: %v", err)
	}
	txObj2 := createTransactionObject(tx2)
	_ = callNative(t, txObj2, "exec", "INSERT INTO events (name) VALUES (?)", Array{"rollback-direct"}).(Object)
	if rolledBack := callNative(t, txObj2, "rollback"); rolledBack != true {
		t.Fatalf("tx.rollback = %#v, want true", rolledBack)
	}

	checkRows := callNative(t, createDatabaseObject(db, rt.NewGoroutineRuntime()), "query", "SELECT name FROM events ORDER BY id").(Array)
	if len(checkRows) != 1 || checkRows[0].(Object)["name"] != "commit-direct" {
		t.Fatalf("rows after rollback = %#v, want only committed row", checkRows)
	}
}

func TestDatabaseErrors(t *testing.T) {
	mod := mustModuleObject(t, DefaultModules(rt.NewGoroutineRuntime()), "db")

	if _, err := callNativeWithError(mod, "sqlite"); err == nil {
		t.Fatal("expected sqlite arity error")
	}
	if _, err := callNativeWithError(mod, "sqlite", Object{}); err == nil {
		t.Fatal("expected sqlite path type error")
	}
	if _, err := callNativeWithError(mod, "connect"); err == nil {
		t.Fatal("expected connect arity error")
	}
	if _, err := callNativeWithError(mod, "connect", Object{}); err == nil {
		t.Fatal("expected connect driver type error")
	}
	if _, err := callNativeWithError(mod, "connect", "sqlite", Object{}); err == nil {
		t.Fatal("expected connect dsn type error")
	}
	if _, err := callNativeWithError(mod, "connect", "missing-driver"); err == nil {
		t.Fatal("expected connect open error")
	}
	if _, err := callNativeWithError(mod, "mysql"); err == nil {
		t.Fatal("expected mysql arity error")
	}
	if _, err := callNativeWithError(mod, "mysql", Object{}); err == nil {
		t.Fatal("expected mysql dsn type error")
	}
	if _, err := callNativeWithError(mod, "postgres"); err == nil {
		t.Fatal("expected postgres arity error")
	}
	if _, err := callNativeWithError(mod, "postgres", Object{}); err == nil {
		t.Fatal("expected postgres dsn type error")
	}
	if _, err := callNativeWithError(mod, "sqlserver"); err == nil {
		t.Fatal("expected sqlserver arity error")
	}
	if _, err := callNativeWithError(mod, "sqlserver", Object{}); err == nil {
		t.Fatal("expected sqlserver dsn type error")
	}

	db := openTestSQLiteDB(t)
	defer callNative(t, db, "close")
	if _, err := callNativeWithError(db, "query"); err == nil {
		t.Fatal("expected query arity error")
	}
	if _, err := callNativeWithError(db, "query", Object{}); err == nil {
		t.Fatal("expected query sql type error")
	}
	if _, err := callNativeWithError(db, "query", "SELECT * FROM missing"); err == nil {
		t.Fatal("expected query missing table error")
	}
	if _, err := callNativeWithError(db, "exec", Object{}); err == nil {
		t.Fatal("expected exec sql type error")
	}
	if _, err := callNativeWithError(db, "exec", "INSERT INTO missing VALUES (1)"); err == nil {
		t.Fatal("expected exec missing table error")
	}
	if _, err := callNativeWithError(db, "prepare"); err == nil {
		t.Fatal("expected prepare arity error")
	}
	if _, err := callNativeWithError(db, "prepare", Object{}); err == nil {
		t.Fatal("expected prepare sql type error")
	}
	if _, err := callNativeWithError(db, "transaction"); err == nil {
		t.Fatal("expected transaction arity error")
	}
	if _, err := callNativeWithError(db, "transaction", "not-a-function"); err == nil {
		t.Fatal("expected transaction callback type error")
	}
	if _, err := callNativeWithError(db, "transaction", NativeFunction(func(args []Value) (Value, error) {
		return nil, errTestDatabaseCallback
	})); err == nil {
		t.Fatal("expected transaction callback error")
	}
}

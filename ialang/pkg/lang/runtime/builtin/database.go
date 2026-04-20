package builtin

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"

	rtvm "ialang/pkg/lang/runtime/vm"
)

// databaseModule 创建数据库模块
func newDatabaseModule(asyncRuntime AsyncRuntime) Value {
	// connect 方法
	connectFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("db.connect expects at least 1 arg, got %d", len(args))
		}

		driverName, err := asStringValue("db.connect.driver", args[0])
		if err != nil {
			return nil, err
		}

		dsn := ""
		if len(args) > 1 {
			dsn, err = asStringValue("db.connect.dsn", args[1])
			if err != nil {
				return nil, err
			}
		}

		// 连接数据库
		db, err := sql.Open(driverName, dsn)
		if err != nil {
			return nil, fmt.Errorf("db.connect failed: %w", err)
		}

		// 测试连接
		if err := db.Ping(); err != nil {
			db.Close()
			return nil, fmt.Errorf("db.ping failed: %w", err)
		}

		return createDatabaseObject(db, asyncRuntime), nil
	})

	// sqlite 便捷方法
	sqliteFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("db.sqlite expects 1 arg (path), got %d", len(args))
		}

		path, err := asStringValue("db.sqlite.path", args[0])
		if err != nil {
			return nil, err
		}

		db, err := sql.Open("sqlite", path)
		if err != nil {
			return nil, fmt.Errorf("db.sqlite open failed: %w", err)
		}

		if err := db.Ping(); err != nil {
			db.Close()
			return nil, fmt.Errorf("db.sqlite ping failed: %w", err)
		}

		return createDatabaseObject(db, asyncRuntime), nil
	})

	// mysql 便捷方法
	mysqlFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("db.mysql expects 1 arg (dsn), got %d", len(args))
		}

		dsn, err := asStringValue("db.mysql.dsn", args[0])
		if err != nil {
			return nil, err
		}

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("db.mysql open failed: %w", err)
		}

		if err := db.Ping(); err != nil {
			db.Close()
			return nil, fmt.Errorf("db.mysql ping failed: %w", err)
		}

		return createDatabaseObject(db, asyncRuntime), nil
	})

	// postgres 便捷方法
	postgresFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("db.postgres expects 1 arg (dsn), got %d", len(args))
		}

		dsn, err := asStringValue("db.postgres.dsn", args[0])
		if err != nil {
			return nil, err
		}

		db, err := sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("db.postgres open failed: %w", err)
		}

		if err := db.Ping(); err != nil {
			db.Close()
			return nil, fmt.Errorf("db.postgres ping failed: %w", err)
		}

		return createDatabaseObject(db, asyncRuntime), nil
	})

	// sqlserver 便捷方法
	sqlserverFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("db.sqlserver expects 1 arg (dsn), got %d", len(args))
		}

		dsn, err := asStringValue("db.sqlserver.dsn", args[0])
		if err != nil {
			return nil, err
		}

		db, err := sql.Open("sqlserver", dsn)
		if err != nil {
			return nil, fmt.Errorf("db.sqlserver open failed: %w", err)
		}

		if err := db.Ping(); err != nil {
			db.Close()
			return nil, fmt.Errorf("db.sqlserver ping failed: %w", err)
		}

		return createDatabaseObject(db, asyncRuntime), nil
	})

	// Create namespace with functions
	namespace := Object{
		"connect":    connectFn,
		"sqlite":     sqliteFn,
		"mysql":      mysqlFn,
		"postgres":   postgresFn,
		"sqlserver":  sqlserverFn,
		"postgresql": postgresFn,
		"pg":         postgresFn,
	}

	// Create module with self-reference to namespace (non-circular)
	module := cloneObject(namespace)
	module["db"] = namespace
	return module
}

type dbQuerier interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type dbPreparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

func parseDBArgs(args []Value, prefix string) (string, []interface{}, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("%s expects at least 1 arg", prefix)
	}
	sqlQuery, err := asStringValue(prefix+".sql", args[0])
	if err != nil {
		return "", nil, err
	}
	var queryArgs []interface{}
	if len(args) > 1 {
		if params, ok := args[1].(Array); ok {
			queryArgs = make([]interface{}, len(params))
			for i, param := range params {
				queryArgs[i] = param
			}
		}
	}
	return sqlQuery, queryArgs, nil
}

func scanRows(rows *sql.Rows) (Array, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var results Array
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		rowObj := make(Object, len(columns))
		for i, col := range columns {
			if values[i] == nil {
				rowObj[col] = nil
			} else {
				rowObj[col] = convertToValue(values[i])
			}
		}
		results = append(results, rowObj)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func dbExecResult(result sql.Result) Object {
	rowsAffected, _ := result.RowsAffected()
	lastInsertId, _ := result.LastInsertId()
	return Object{
		"affectedRows": float64(rowsAffected),
		"insertId":     float64(lastInsertId),
	}
}

func makeDBQueryFn(q dbQuerier, mu *sync.Mutex, prefix string) NativeFunction {
	return NativeFunction(func(args []Value) (Value, error) {
		sqlQuery, queryArgs, err := parseDBArgs(args, prefix)
		if err != nil {
			return nil, err
		}
		mu.Lock()
		defer mu.Unlock()
		rows, err := q.Query(sqlQuery, queryArgs...)
		if err != nil {
			return nil, fmt.Errorf("%s failed: %w", prefix, err)
		}
		defer rows.Close()
		results, err := scanRows(rows)
		if err != nil {
			return nil, fmt.Errorf("%s scan error: %w", prefix, err)
		}
		return results, nil
	})
}

func makeDBQueryOneFn(q dbQuerier, mu *sync.Mutex, prefix string) NativeFunction {
	return NativeFunction(func(args []Value) (Value, error) {
		sqlQuery, queryArgs, err := parseDBArgs(args, prefix)
		if err != nil {
			return nil, err
		}
		mu.Lock()
		defer mu.Unlock()
		rows, err := q.Query(sqlQuery, queryArgs...)
		if err != nil {
			return nil, fmt.Errorf("%s failed: %w", prefix, err)
		}
		defer rows.Close()
		results, err := scanRows(rows)
		if err != nil {
			return nil, fmt.Errorf("%s scan error: %w", prefix, err)
		}
		if len(results) == 0 {
			return Object{}, nil
		}
		return results[0], nil
	})
}

func makeDBExecFn(q dbQuerier, mu *sync.Mutex, prefix string) NativeFunction {
	return NativeFunction(func(args []Value) (Value, error) {
		sqlQuery, queryArgs, err := parseDBArgs(args, prefix)
		if err != nil {
			return nil, err
		}
		mu.Lock()
		defer mu.Unlock()
		result, err := q.Exec(sqlQuery, queryArgs...)
		if err != nil {
			return nil, fmt.Errorf("%s failed: %w", prefix, err)
		}
		return dbExecResult(result), nil
	})
}

// createDatabaseObject 创建数据库对象
func createDatabaseObject(db *sql.DB, asyncRuntime AsyncRuntime) Object {
	var mu sync.Mutex

	queryFn := makeDBQueryFn(db, &mu, "db.query")
	queryOneFn := makeDBQueryOneFn(db, &mu, "db.queryOne")
	execFn := makeDBExecFn(db, &mu, "db.exec")
	executeFn := execFn

	queryAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) { return queryFn(args) }), nil
	})
	queryOneAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) { return queryOneFn(args) }), nil
	})
	execAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) { return execFn(args) }), nil
	})

	transactionFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("db.transaction expects 1 arg (callback), got %d", len(args))
		}
		callback := args[0]
		var isFn bool
		switch callback.(type) {
		case NativeFunction, *UserFunction:
			isFn = true
		}
		if !isFn {
			return nil, fmt.Errorf("db.transaction expects a function as callback")
		}
		mu.Lock()
		defer mu.Unlock()
		tx, err := db.Begin()
		if err != nil {
			return nil, fmt.Errorf("db.transaction begin failed: %w", err)
		}
		txObj := createTransactionObject(tx)
		var result Value
		switch cb := callback.(type) {
		case NativeFunction:
			result, err = cb([]Value{txObj})
		case *UserFunction:
			result, err = rtvm.CallUserFunctionSync(cb, []Value{txObj})
		}
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("db.transaction callback failed: %w", err)
		}
		shouldCommit := true
		if b, ok := result.(bool); ok && !b {
			shouldCommit = false
		}
		if shouldCommit {
			if err := tx.Commit(); err != nil {
				return nil, fmt.Errorf("db.transaction commit failed: %w", err)
			}
			return Object{"committed": true, "result": result}, nil
		}
		tx.Rollback()
		return Object{"committed": false, "result": result}, nil
	})

	pingFn := NativeFunction(func(args []Value) (Value, error) {
		mu.Lock()
		defer mu.Unlock()
		if err := db.Ping(); err != nil {
			return false, err
		}
		return true, nil
	})

	closeFn := NativeFunction(func(args []Value) (Value, error) {
		mu.Lock()
		defer mu.Unlock()
		if err := db.Close(); err != nil {
			return false, err
		}
		return true, nil
	})

	prepareFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("db.prepare expects 1 arg (sql), got %d", len(args))
		}
		sqlQuery, err := asStringValue("db.prepare.sql", args[0])
		if err != nil {
			return nil, err
		}
		mu.Lock()
		defer mu.Unlock()
		stmt, err := db.Prepare(sqlQuery)
		if err != nil {
			return nil, fmt.Errorf("db.prepare failed: %w", err)
		}
		return createStatementObject(stmt, &mu), nil
	})

	return Object{
		"query":          queryFn,
		"queryOne":       queryOneFn,
		"exec":           execFn,
		"execute":        executeFn,
		"queryAsync":     queryAsyncFn,
		"queryOneAsync":  queryOneAsyncFn,
		"execAsync":      execAsyncFn,
		"transaction":    transactionFn,
		"ping":           pingFn,
		"close":          closeFn,
		"prepare":        prepareFn,
	}
}

// createTransactionObject 创建事务对象
func createTransactionObject(tx *sql.Tx) Object {
	var mu sync.Mutex
	return Object{
		"query": makeDBQueryFn(tx, &mu, "tx.query"),
		"exec":  makeDBExecFn(tx, &mu, "tx.exec"),
		"commit": NativeFunction(func(args []Value) (Value, error) {
			mu.Lock()
			defer mu.Unlock()
			if err := tx.Commit(); err != nil {
				return false, err
			}
			return true, nil
		}),
		"rollback": NativeFunction(func(args []Value) (Value, error) {
			mu.Lock()
			defer mu.Unlock()
			if err := tx.Rollback(); err != nil {
				return false, err
			}
			return true, nil
		}),
	}
}

// createStatementObject 创建预处理语句对象
func createStatementObject(stmt *sql.Stmt, mu *sync.Mutex) Object {
	return Object{
		"query": makeStmtQueryFn(stmt, mu),
		"exec":  makeStmtExecFn(stmt, mu),
		"close": NativeFunction(func(args []Value) (Value, error) {
			mu.Lock()
			defer mu.Unlock()
			if err := stmt.Close(); err != nil {
				return false, err
			}
			return true, nil
		}),
	}
}

func makeStmtQueryFn(stmt *sql.Stmt, mu *sync.Mutex) NativeFunction {
	return NativeFunction(func(args []Value) (Value, error) {
		var queryArgs []interface{}
		if len(args) > 0 {
			if params, ok := args[0].(Array); ok {
				queryArgs = make([]interface{}, len(params))
				for i, param := range params {
					queryArgs[i] = param
				}
			}
		}
		mu.Lock()
		defer mu.Unlock()
		rows, err := stmt.Query(queryArgs...)
		if err != nil {
			return nil, fmt.Errorf("stmt.query failed: %w", err)
		}
		defer rows.Close()
		results, err := scanRows(rows)
		if err != nil {
			return nil, fmt.Errorf("stmt.query scan error: %w", err)
		}
		return results, nil
	})
}

func makeStmtExecFn(stmt *sql.Stmt, mu *sync.Mutex) NativeFunction {
	return NativeFunction(func(args []Value) (Value, error) {
		var execArgs []interface{}
		if len(args) > 0 {
			if params, ok := args[0].(Array); ok {
				execArgs = make([]interface{}, len(params))
				for i, param := range params {
					execArgs[i] = param
				}
			}
		}
		mu.Lock()
		defer mu.Unlock()
		result, err := stmt.Exec(execArgs...)
		if err != nil {
			return nil, fmt.Errorf("stmt.exec failed: %w", err)
		}
		return dbExecResult(result), nil
	})
}

func convertToValue(v interface{}) Value {
	return yamlToValue(v)
}

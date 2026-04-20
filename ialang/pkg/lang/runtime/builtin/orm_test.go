package builtin

import (
	"strings"
	"testing"

	rt "ialang/pkg/lang/runtime"
)

func assertParamsEqual(t *testing.T, got Array, want Array) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("params length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("params[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

// TestORMInit 测试 ORM 初始化
func TestORMInit(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ormMod := mustModuleObject(t, modules, "orm")

	// 测试 init
	db := Object{} // 模拟数据库对象
	result := callNative(t, ormMod, "init", db)
	resultObj := mustRuntimeObject(t, result, "init result")

	// 验证返回的对象包含 DataTypes
	if _, ok := resultObj["DataTypes"]; !ok {
		t.Fatal("ORM init result should contain DataTypes")
	}
}

// TestORMDefineModel 测试模型定义
func TestORMDefineModel(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ormMod := mustModuleObject(t, modules, "orm")

	// 测试定义简单模型
	fields := Object{
		"name": Object{
			"type":    "TEXT",
			"notNull": true,
		},
		"email": Object{
			"type":    "TEXT",
			"unique":  true,
			"notNull": true,
		},
		"age": Object{
			"type":         "INTEGER",
			"defaultValue": "0",
		},
	}

	result := callNative(t, ormMod, "defineModel", "User", fields)
	model := mustRuntimeObject(t, result, "defineModel result")

	// 验证模型属性
	if name, ok := model["name"].(string); !ok || name != "User" {
		t.Fatalf("Model name = %v, want User", model["name"])
	}
	if tableName, ok := model["tableName"].(string); !ok || tableName != "users" {
		t.Fatalf("Table name = %v, want users", tableName)
	}
	if fieldList, ok := model["fieldList"].(Array); !ok || len(fieldList) < 3 {
		t.Fatalf("Field list length = %d, want at least 3", len(fieldList))
	}
}

// TestORMDefineModelWithOptions 测试带选项的模型定义
func TestORMDefineModelWithOptions(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ormMod := mustModuleObject(t, modules, "orm")

	fields := Object{
		"id": Object{
			"primaryKey":    true,
			"autoIncrement": true,
		},
		"title": Object{
			"type":    "TEXT",
			"notNull": true,
		},
	}

	options := Object{
		"tableName":  "my_posts",
		"timestamps": false,
	}

	result := callNative(t, ormMod, "defineModel", "Post", fields, options)
	model := mustRuntimeObject(t, result, "defineModel result")

	// 验证自定义表名
	if tableName, ok := model["tableName"].(string); !ok || tableName != "my_posts" {
		t.Fatalf("Table name = %v, want my_posts", tableName)
	}

	// 验证时间戳禁用
	if timestamps, ok := model["timestamps"].(bool); ok && timestamps {
		t.Fatal("Timestamps should be false when disabled")
	}
}

// TestORMCreateModel 测试创建记录
func TestORMCreateModel(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ormMod := mustModuleObject(t, modules, "orm")

	// 定义模型
	fields := Object{
		"id": Object{
			"primaryKey":    true,
			"autoIncrement": true,
		},
		"name": Object{
			"type":    "TEXT",
			"notNull": true,
		},
		"email": Object{
			"type":    "TEXT",
			"notNull": true,
		},
	}

	modelResult := callNative(t, ormMod, "defineModel", "User", fields, Object{
		"timestamps": false,
	})
	model := mustRuntimeObject(t, modelResult, "defineModel result")

	// 测试创建查询
	data := Object{
		"name":  "Alice",
		"email": "alice@example.com",
	}

	result := callNative(t, ormMod, "createModel", model, data)
	query := mustRuntimeObject(t, result, "createModel result")

	// 验证 SQL
	sql, ok := query["sql"].(string)
	if !ok {
		t.Fatal("Query should have sql field")
	}

	// 验证包含 INSERT INTO
	if len(sql) == 0 || sql[:11] != "INSERT INTO" {
		t.Fatalf("SQL should start with INSERT INTO, got: %s", sql)
	}

	// 验证参数
	params, ok := query["params"].(Array)
	if !ok {
		t.Fatal("Query should have params field")
	}

	if len(params) != 2 {
		t.Fatalf("Params length = %d, want 2", len(params))
	}

	if params[0] != "alice@example.com" {
		t.Fatalf("First param = %v, want alice@example.com", params[0])
	}

	if params[1] != "Alice" {
		t.Fatalf("Second param = %v, want Alice", params[1])
	}
}

func TestORMQueryDeterministicOrder(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ormMod := mustModuleObject(t, modules, "orm")

	fields := Object{
		"id": Object{
			"primaryKey":    true,
			"autoIncrement": true,
		},
		"name": Object{
			"type": "TEXT",
		},
		"email": Object{
			"type": "TEXT",
		},
		"age": Object{
			"type": "INTEGER",
		},
	}
	modelResult := callNative(t, ormMod, "defineModel", "User", fields, Object{
		"timestamps": false,
	})
	model := mustRuntimeObject(t, modelResult, "defineModel result")

	createData := Object{
		"email": "alice@example.com",
		"name":  "Alice",
		"age":   18.0,
	}
	wantCreateSQL := "INSERT INTO users (age, email, name) VALUES (?, ?, ?)"
	wantCreateParams := Array{18.0, "alice@example.com", "Alice"}

	updateData := Object{
		"name": "Bob",
		"age":  30.0,
	}
	updateCond := Object{
		"id":    1.0,
		"email": "alice@example.com",
	}
	wantUpdateSQL := "UPDATE users SET age = ?, name = ? WHERE email = ? AND id = ?"
	wantUpdateParams := Array{30.0, "Bob", "alice@example.com", 1.0}

	deleteCond := Object{
		"id":    1.0,
		"email": "alice@example.com",
	}
	wantDeleteSQL := "DELETE FROM users WHERE email = ? AND id = ?"
	wantDeleteParams := Array{"alice@example.com", 1.0}

	qbResult := callNative(t, ormMod, "QueryBuilder", model)
	qb := mustRuntimeObject(t, qbResult, "query builder")
	qb["state"] = Object{
		"conditions": Object{
			"email": "alice@example.com",
			"name":  "Alice",
		},
		"orderBy": Array{},
		"limit":   nil,
		"offset":  nil,
	}
	wantBuildSQL := "SELECT * FROM users WHERE email = ? AND name = ?"
	wantBuildParams := Array{"alice@example.com", "Alice"}

	for i := 0; i < 100; i++ {
		createResult := callNative(t, ormMod, "createModel", model, createData)
		createQuery := mustRuntimeObject(t, createResult, "create query")
		if createQuery["sql"] != wantCreateSQL {
			t.Fatalf("create sql = %v, want %q", createQuery["sql"], wantCreateSQL)
		}
		assertParamsEqual(t, createQuery["params"].(Array), wantCreateParams)

		updateResult := callNative(t, ormMod, "updateModel", model, updateData, updateCond)
		updateQuery := mustRuntimeObject(t, updateResult, "update query")
		if updateQuery["sql"] != wantUpdateSQL {
			t.Fatalf("update sql = %v, want %q", updateQuery["sql"], wantUpdateSQL)
		}
		assertParamsEqual(t, updateQuery["params"].(Array), wantUpdateParams)

		deleteResult := callNative(t, ormMod, "deleteModel", model, deleteCond)
		deleteQuery := mustRuntimeObject(t, deleteResult, "delete query")
		if deleteQuery["sql"] != wantDeleteSQL {
			t.Fatalf("delete sql = %v, want %q", deleteQuery["sql"], wantDeleteSQL)
		}
		assertParamsEqual(t, deleteQuery["params"].(Array), wantDeleteParams)

		buildResult := callNative(t, ormMod, "buildQuery", qb)
		buildQuery := mustRuntimeObject(t, buildResult, "build query")
		if buildQuery["sql"] != wantBuildSQL {
			t.Fatalf("build sql = %v, want %q", buildQuery["sql"], wantBuildSQL)
		}
		assertParamsEqual(t, buildQuery["params"].(Array), wantBuildParams)
	}
}

// TestORMUpdateModel 测试更新记录
func TestORMUpdateModel(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ormMod := mustModuleObject(t, modules, "orm")

	// 定义模型
	fields := Object{
		"id": Object{
			"primaryKey":    true,
			"autoIncrement": true,
		},
		"name": Object{
			"type": "TEXT",
		},
		"age": Object{
			"type": "INTEGER",
		},
	}

	modelResult := callNative(t, ormMod, "defineModel", "User", fields, Object{
		"timestamps": false,
	})
	model := mustRuntimeObject(t, modelResult, "defineModel result")

	// 测试更新查询
	data := Object{
		"name": "Bob",
		"age":  30.0,
	}
	conditions := Object{
		"id": 1.0,
	}

	result := callNative(t, ormMod, "updateModel", model, data, conditions)
	query := mustRuntimeObject(t, result, "updateModel result")

	// 验证 SQL
	sql, ok := query["sql"].(string)
	if !ok {
		t.Fatal("Query should have sql field")
	}

	// 验证包含 UPDATE 和 SET
	if len(sql) < 10 || sql[:6] != "UPDATE" {
		t.Fatalf("SQL should start with UPDATE, got: %s", sql)
	}

	if !strings.Contains(sql, "SET") {
		t.Fatalf("SQL should contain SET, got: %s", sql)
	}

	// 验证参数
	params, ok := query["params"].(Array)
	if !ok {
		t.Fatal("Query should have params field")
	}

	if len(params) != 3 {
		t.Fatalf("Params length = %d, want 3", len(params))
	}
}

// TestORMDeleteModel 测试删除记录
func TestORMDeleteModel(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ormMod := mustModuleObject(t, modules, "orm")

	// 定义模型
	fields := Object{
		"id": Object{
			"primaryKey":    true,
			"autoIncrement": true,
		},
		"name": Object{
			"type": "TEXT",
		},
	}

	modelResult := callNative(t, ormMod, "defineModel", "User", fields, Object{
		"timestamps": false,
	})
	model := mustRuntimeObject(t, modelResult, "defineModel result")

	// 测试删除查询
	conditions := Object{
		"id": 1.0,
	}

	result := callNative(t, ormMod, "deleteModel", model, conditions)
	query := mustRuntimeObject(t, result, "deleteModel result")

	// 验证 SQL
	sql, ok := query["sql"].(string)
	if !ok {
		t.Fatal("Query should have sql field")
	}

	// 验证包含 DELETE FROM
	if len(sql) < 11 || sql[:11] != "DELETE FROM" {
		t.Fatalf("SQL should start with DELETE FROM, got: %s", sql)
	}

	// 验证参数
	params, ok := query["params"].(Array)
	if !ok {
		t.Fatal("Query should have params field")
	}

	if len(params) != 1 {
		t.Fatalf("Params length = %d, want 1", len(params))
	}
}

// TestORMQueryBuilder 测试查询构建器
func TestORMQueryBuilder(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ormMod := mustModuleObject(t, modules, "orm")

	// 定义模型
	fields := Object{
		"id": Object{
			"primaryKey":    true,
			"autoIncrement": true,
		},
		"name": Object{
			"type": "TEXT",
		},
		"age": Object{
			"type": "INTEGER",
		},
	}

	modelResult := callNative(t, ormMod, "defineModel", "User", fields, Object{
		"timestamps": false,
	})
	model := mustRuntimeObject(t, modelResult, "defineModel result")

	// 创建查询构建器
	qbResult := callNative(t, ormMod, "QueryBuilder", model)
	qb := mustRuntimeObject(t, qbResult, "QueryBuilder result")

	// 注意：由于 query builder 是对象，我们需要获取它的状态
	// 这里我们简化测试，只验证 qb 对象创建成功
	if qb == nil {
		t.Fatal("QueryBuilder should return non-nil result")
	}
}

// TestORMBuildQuery 测试查询构建
func TestORMBuildQuery(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ormMod := mustModuleObject(t, modules, "orm")

	// 定义模型
	fields := Object{
		"id": Object{
			"primaryKey":    true,
			"autoIncrement": true,
		},
		"name": Object{
			"type": "TEXT",
		},
		"isActive": Object{
			"type": "INTEGER",
		},
	}

	modelResult := callNative(t, ormMod, "defineModel", "User", fields, Object{
		"timestamps": false,
	})
	model := mustRuntimeObject(t, modelResult, "defineModel result")

	// 创建查询构建器并设置状态
	qbResult := callNative(t, ormMod, "QueryBuilder", model)
	qb := mustRuntimeObject(t, qbResult, "QueryBuilder result")

	// 修改状态
	state := Object{
		"conditions": Object{
			"name":     "Alice",
			"isActive": true,
		},
		"orderBy": Array{
			Object{
				"field":     "name",
				"direction": "ASC",
			},
		},
		"limit":  10.0,
		"offset": 20.0,
	}
	qb["state"] = state

	// 构建查询
	result := callNative(t, ormMod, "buildQuery", qb)
	query := mustRuntimeObject(t, result, "buildQuery result")

	// 验证 SQL
	sql, ok := query["sql"].(string)
	if !ok {
		t.Fatal("Query should have sql field")
	}

	// 验证包含 SELECT 和 WHERE
	if len(sql) < 10 || sql[:6] != "SELECT" {
		t.Fatalf("SQL should start with SELECT, got: %s", sql)
	}

	// 验证参数
	params, ok := query["params"].(Array)
	if !ok {
		t.Fatal("Query should have params field")
	}

	if len(params) != 2 {
		t.Fatalf("Params length = %d, want 2", len(params))
	}
}

// TestORMUnderscore 测试下划线转换
func TestORMUnderscore(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ormMod := mustModuleObject(t, modules, "orm")

	testCases := []struct {
		input    string
		expected string
	}{
		{"userName", "user_name"},
		{"createdAt", "created_at"},
		{"firstName", "first_name"},
		{"User", "user"},
	}

	for _, tc := range testCases {
		result := callNative(t, ormMod, "underscore", tc.input)
		if s, ok := result.(string); !ok || s != tc.expected {
			t.Fatalf("underscore(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

// TestORMCamelize 测试驼峰转换
func TestORMCamelize(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ormMod := mustModuleObject(t, modules, "orm")

	testCases := []struct {
		input    string
		expected string
	}{
		{"user_name", "UserName"},
		{"created_at", "CreatedAt"},
		{"first_name", "FirstName"},
	}

	for _, tc := range testCases {
		result := callNative(t, ormMod, "camelize", tc.input)
		if s, ok := result.(string); !ok || s != tc.expected {
			t.Fatalf("camelize(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

// TestORMDataTypes 测试数据类型定义
func TestORMDataTypes(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ormMod := mustModuleObject(t, modules, "orm")

	// 获取 DataTypes
	if dataTypesObj, ok := ormMod["DataTypes"]; ok {
		dataTypes := dataTypesObj.(Object)

		// 验证常用类型
		expectedTypes := map[string]string{
			"INTEGER":  "INTEGER",
			"STRING":   "TEXT",
			"TEXT":     "TEXT",
			"BOOLEAN":  "INTEGER",
			"FLOAT":    "FLOAT",
			"DATETIME": "TEXT",
		}

		for name, expected := range expectedTypes {
			if val, ok := dataTypes[name]; !ok || val != expected {
				t.Fatalf("DataTypes[%s] = %v, want %v", name, val, expected)
			}
		}
	} else {
		t.Fatal("ORM module should have DataTypes")
	}
}

package builtin

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func sortedObjectKeys(obj Object) []string {
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func buildORMQuery(qb Object) (Value, error) {
	tableName, _ := qb["tableName"].(string)
	state, _ := qb["state"].(Object)

	sql := "SELECT * FROM " + tableName
	params := make(Array, 0)

	if conditions, ok := state["conditions"].(Object); ok && len(conditions) > 0 {
		whereClauses := make([]string, 0)
		for _, key := range sortedObjectKeys(conditions) {
			val := conditions[key]
			whereClauses = append(whereClauses, underscore(key)+" = ?")
			params = append(params, val)
		}
		sql += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	if orderBy, ok := state["orderBy"].(Array); ok && len(orderBy) > 0 {
		orderClauses := make([]string, 0)
		for _, item := range orderBy {
			if obj, ok := item.(Object); ok {
				field, _ := obj["field"].(string)
				dir, _ := obj["direction"].(string)
				orderClauses = append(orderClauses, underscore(field)+" "+dir)
			}
		}
		sql += " ORDER BY " + strings.Join(orderClauses, ", ")
	}

	if limit, ok := state["limit"]; ok && limit != nil {
		if n, ok := limit.(float64); ok {
			sql += fmt.Sprintf(" LIMIT %d", int(n))
		}
	}

	if offset, ok := state["offset"]; ok && offset != nil {
		if n, ok := offset.(float64); ok {
			sql += fmt.Sprintf(" OFFSET %d", int(n))
		}
	}

	return Object{
		"sql":    sql,
		"params": params,
	}, nil
}

// ormModule 创建 ORM 模块
func newORMModule() Value {
	// DataTypes
	dataTypesObj := Object{
		"INTEGER":   "INTEGER",
		"INT":       "INTEGER",
		"BIGINT":    "INTEGER",
		"FLOAT":     "FLOAT",
		"DOUBLE":    "FLOAT",
		"DECIMAL":   "FLOAT",
		"STRING":    "TEXT",
		"VARCHAR":   "TEXT",
		"TEXT":      "TEXT",
		"BOOLEAN":   "INTEGER",
		"DATE":      "TEXT",
		"DATETIME":  "TEXT",
		"TIMESTAMP": "TEXT",
		"JSON":      "TEXT",
		"BLOB":      "BLOB",
	}

	// 初始化 ORM
	initORMFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("orm.init expects database object")
		}

		database, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("orm.init expects database object")
		}

		// 创建 ORM 实例
		orm := Object{
			"DataTypes":    dataTypesObj,
			"database":     database,
			"models":       Object{},
			"associations": Object{},
		}

		return orm, nil
	})

	// defineModel 函数
	defineModelFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("orm.defineModel expects at least 2 args: name, fields")
		}

		name, err := asStringValue("orm.defineModel", args[0])
		if err != nil {
			return nil, err
		}

		fieldsObj, ok := args[1].(Object)
		if !ok {
			return nil, fmt.Errorf("orm.defineModel expects fields object as second arg")
		}

		optionsObj := Object{}
		if len(args) > 2 {
			if opts, ok := args[2].(Object); ok {
				optionsObj = opts
			}
		}

		// 生成表名
		tableName := generateTableName(name)
		if tn, ok := optionsObj["tableName"]; ok {
			if s, ok := tn.(string); ok {
				tableName = s
			}
		}

		// 处理字段
		schemaFields := make(Object)
		fieldList := make(Array, 0)

		// 默认添加 id
		if _, hasID := fieldsObj["id"]; !hasID {
			fieldsObj["id"] = Object{
				"primaryKey":    true,
				"autoIncrement": true,
			}
		}

		// 默认时间戳
		timestamps := true
		if ts, ok := optionsObj["timestamps"]; ok {
			if b, ok := ts.(bool); ok {
				timestamps = b
			}
		}

		if timestamps {
			if _, has := fieldsObj["createdAt"]; !has {
				fieldsObj["createdAt"] = Object{"type": "TEXT"}
			}
			if _, has := fieldsObj["updatedAt"]; !has {
				fieldsObj["updatedAt"] = Object{"type": "TEXT"}
			}
		}

		// 处理每个字段
		for _, fieldName := range sortedObjectKeys(fieldsObj) {
			fieldDef := fieldsObj[fieldName]
			fieldType := "TEXT"
			primaryKey := false
			autoIncrement := false
			notNull := false
			defaultValue := ""
			unique := false

			if fieldObj, ok := fieldDef.(Object); ok {
				if t, ok := fieldObj["type"]; ok {
					if s, ok := t.(string); ok {
						fieldType = s
					}
				}
				if v, ok := fieldObj["primaryKey"]; ok {
					if b, ok := v.(bool); ok {
						primaryKey = b
					}
				}
				if v, ok := fieldObj["autoIncrement"]; ok {
					if b, ok := v.(bool); ok {
						autoIncrement = b
					}
				}
				if v, ok := fieldObj["notNull"]; ok {
					if b, ok := v.(bool); ok {
						notNull = b
					}
				}
				if v, ok := fieldObj["defaultValue"]; ok {
					if s, ok := v.(string); ok {
						defaultValue = s
					} else if f, ok := v.(float64); ok {
						defaultValue = fmt.Sprintf("%v", f)
					}
				}
				if v, ok := fieldObj["unique"]; ok {
					if b, ok := v.(bool); ok {
						unique = b
					}
				}
			} else if s, ok := fieldDef.(string); ok {
				fieldType = s
			}

			field := Object{
				"name":          fieldName,
				"type":          fieldType,
				"primaryKey":    primaryKey,
				"autoIncrement": autoIncrement,
				"notNull":       notNull || primaryKey,
				"defaultValue":  defaultValue,
				"unique":        unique,
			}

			schemaFields[fieldName] = field
			fieldList = append(fieldList, field)
		}

		// 创建模型对象
		model := Object{
			"name":         name,
			"tableName":    tableName,
			"fields":       schemaFields,
			"fieldList":    fieldList,
			"timestamps":   timestamps,
			"database":     nil,
			"associations": Array{},
		}
		model["hasOne"] = NativeFunction(func(args []Value) (Value, error) {
			return model, nil
		})
		model["belongsTo"] = NativeFunction(func(args []Value) (Value, error) {
			return model, nil
		})
		model["hasMany"] = NativeFunction(func(args []Value) (Value, error) {
			return model, nil
		})
		model["belongsToMany"] = NativeFunction(func(args []Value) (Value, error) {
			return model, nil
		})

		return model, nil
	})

	// createModel 方法 - 创建记录
	createModelFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("model.create expects model and data")
		}

		model, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("model.create expects model object")
		}

		data, ok := args[1].(Object)
		if !ok {
			return nil, fmt.Errorf("model.create expects data object")
		}

		tableName, _ := model["tableName"].(string)
		fields, _ := model["fields"].(Object)

		// 构建 INSERT 语句
		columns := make([]string, 0)
		values := make(Array, 0)
		placeholders := make([]string, 0)

		for _, fieldName := range sortedObjectKeys(data) {
			value := data[fieldName]
			if field, ok := fields[fieldName]; ok {
				fieldObj := field.(Object)
				colName := underscore(fieldName)

				// 跳过自增主键
				if fieldObj["autoIncrement"] == true {
					continue
				}

				columns = append(columns, colName)
				values = append(values, value)
				placeholders = append(placeholders, "?")
			}
		}

		// 添加时间戳
		if timestamps, _ := model["timestamps"].(bool); timestamps {
			now := time.Now().Format("2006-01-02 15:04:05")
			columns = append(columns, "created_at", "updated_at")
			values = append(values, now, now)
			placeholders = append(placeholders, "?", "?")
		}

		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "))

		return Object{
			"sql":    sql,
			"params": values,
		}, nil
	})

	// updateModel 方法 - 更新记录
	updateModelFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("model.update expects model, data, conditions")
		}

		model, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("model.update expects model object")
		}

		data, ok := args[1].(Object)
		if !ok {
			return nil, fmt.Errorf("model.update expects data object")
		}

		conditions, ok := args[2].(Object)
		if !ok {
			return nil, fmt.Errorf("model.update expects conditions object")
		}

		tableName, _ := model["tableName"].(string)

		// 构建 UPDATE 语句
		setClauses := make([]string, 0)
		values := make(Array, 0)

		for _, fieldName := range sortedObjectKeys(data) {
			value := data[fieldName]
			colName := underscore(fieldName)
			setClauses = append(setClauses, colName+" = ?")
			values = append(values, value)
		}

		// 添加更新时间
		if timestamps, _ := model["timestamps"].(bool); timestamps {
			now := time.Now().Format("2006-01-02 15:04:05")
			setClauses = append(setClauses, "updated_at = ?")
			values = append(values, now)
		}

		sql := fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(setClauses, ", "))

		// WHERE 子句
		whereClauses := make([]string, 0)
		for _, fieldName := range sortedObjectKeys(conditions) {
			value := conditions[fieldName]
			colName := underscore(fieldName)
			whereClauses = append(whereClauses, colName+" = ?")
			values = append(values, value)
		}

		if len(whereClauses) > 0 {
			sql += " WHERE " + strings.Join(whereClauses, " AND ")
		}

		return Object{
			"sql":    sql,
			"params": values,
		}, nil
	})

	// deleteModel 方法 - 删除记录
	deleteModelFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("model.delete expects model and conditions")
		}

		model, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("model.delete expects model object")
		}

		conditions, ok := args[1].(Object)
		if !ok {
			return nil, fmt.Errorf("model.delete expects conditions object")
		}

		tableName, _ := model["tableName"].(string)

		sql := fmt.Sprintf("DELETE FROM %s", tableName)
		values := make(Array, 0)

		// WHERE 子句
		whereClauses := make([]string, 0)
		for _, fieldName := range sortedObjectKeys(conditions) {
			value := conditions[fieldName]
			colName := underscore(fieldName)
			whereClauses = append(whereClauses, colName+" = ?")
			values = append(values, value)
		}

		if len(whereClauses) > 0 {
			sql += " WHERE " + strings.Join(whereClauses, " AND ")
		}

		return Object{
			"sql":    sql,
			"params": values,
		}, nil
	})

	// createQueryBuilder 函数
	createQueryBuilderFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("orm.QueryBuilder expects model object")
		}

		model, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("orm.QueryBuilder expects model object")
		}

		tableName, _ := model["tableName"].(string)

		qb := Object{
			"tableName": tableName,
			"model":     model,
			"state": Object{
				"conditions": Object{},
				"orderBy":    Array{},
				"limit":      nil,
				"offset":     nil,
				"include":    Array{},
			},
		}
		qb["where"] = NativeFunction(func(args []Value) (Value, error) {
			state, _ := qb["state"].(Object)
			if len(args) > 0 && args[0] != nil {
				if conditions, ok := args[0].(Object); ok {
					state["conditions"] = conditions
				}
			}
			return qb, nil
		})
		qb["order"] = NativeFunction(func(args []Value) (Value, error) {
			state, _ := qb["state"].(Object)
			orderBy, _ := state["orderBy"].(Array)
			field := ""
			direction := "ASC"
			if len(args) > 0 && args[0] != nil {
				if s, ok := args[0].(string); ok {
					field = s
				}
			}
			if len(args) > 1 && args[1] != nil {
				if s, ok := args[1].(string); ok && s != "" {
					direction = s
				}
			}
			state["orderBy"] = append(orderBy, Object{"field": field, "direction": direction})
			return qb, nil
		})
		qb["limit"] = NativeFunction(func(args []Value) (Value, error) {
			state, _ := qb["state"].(Object)
			if len(args) > 0 {
				state["limit"] = args[0]
			}
			return qb, nil
		})
		qb["offset"] = NativeFunction(func(args []Value) (Value, error) {
			state, _ := qb["state"].(Object)
			if len(args) > 0 {
				state["offset"] = args[0]
			}
			return qb, nil
		})
		qb["include"] = NativeFunction(func(args []Value) (Value, error) {
			state, _ := qb["state"].(Object)
			include, _ := state["include"].(Array)
			if len(args) > 0 {
				include = append(include, args[0])
				state["include"] = include
			}
			return qb, nil
		})
		qb["buildQuery"] = NativeFunction(func(args []Value) (Value, error) {
			return buildORMQuery(qb)
		})

		return qb, nil
	})

	// buildQuery 函数
	buildQueryFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("orm.buildQuery expects query builder")
		}

		qb, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("orm.buildQuery expects query builder")
		}
		return buildORMQuery(qb)
	})

	// Migration helper
	migrationHelperFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("orm.migration expects database object")
		}

		_ = args[0]

		return Object{
			"createTable": NativeFunction(func(args []Value) (Value, error) {
				if len(args) == 0 {
					return nil, fmt.Errorf("createTable expects model")
				}
				_, ok := args[0].(Object)
				if !ok {
					return nil, fmt.Errorf("createTable expects model")
				}
				return true, nil
			}),
			"sync": NativeFunction(func(args []Value) (Value, error) {
				return true, nil
			}),
		}, nil
	})

	// Create namespace with functions
	namespace := Object{
		"init":            initORMFn,
		"DataTypes":       dataTypesObj,
		"defineModel":     defineModelFn,
		"createModel":     createModelFn,
		"updateModel":     updateModelFn,
		"deleteModel":     deleteModelFn,
		"QueryBuilder":    createQueryBuilderFn,
		"buildQuery":      buildQueryFn,
		"migrationHelper": migrationHelperFn,
		"underscore": NativeFunction(func(args []Value) (Value, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("orm.underscore expects 1 arg")
			}
			s, err := asStringValue("orm.underscore", args[0])
			if err != nil {
				return nil, err
			}
			return underscore(s), nil
		}),
		"camelize": NativeFunction(func(args []Value) (Value, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("orm.camelize expects 1 arg")
			}
			s, err := asStringValue("orm.camelize", args[0])
			if err != nil {
				return nil, err
			}
			return camelize(s), nil
		}),
	}

	// Create module with self-reference to namespace (non-circular)
	module := cloneObject(namespace)
	module["orm"] = namespace
	return module
}

// underscore 转换为蛇形命名
func underscore(s string) string {
	var result strings.Builder
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteByte(byte(c + 32))
		} else {
			result.WriteRune(c)
		}
	}
	return result.String()
}

// camelize 转换为驼峰命名
func camelize(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// generateTableName 生成表名
func generateTableName(name string) string {
	underscored := underscore(name)
	if strings.HasSuffix(underscored, "y") {
		return underscored[:len(underscored)-1] + "ies"
	}
	return underscored + "s"
}

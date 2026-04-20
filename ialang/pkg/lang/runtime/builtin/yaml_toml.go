package builtin

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// yamlModule 创建 YAML 模块
func newYAMLModule() Value {
	// parse 方法
	parseFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("yaml.parse expects 1 arg (string), got %d", len(args))
		}

		str, err := asStringValue("yaml.parse", args[0])
		if err != nil {
			return nil, err
		}

		var result interface{}
		if err := yaml.Unmarshal([]byte(str), &result); err != nil {
			return nil, fmt.Errorf("yaml.parse failed: %w", err)
		}

		return yamlToValue(result), nil
	})

	// fromFile 方法
	fromFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("yaml.fromFile expects 1 arg (path), got %d", len(args))
		}

		path, err := asStringValue("yaml.fromFile", args[0])
		if err != nil {
			return nil, err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("yaml.fromFile failed: %w", err)
		}

		var result interface{}
		if err := yaml.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("yaml.fromFile parse error: %w", err)
		}

		return yamlToValue(result), nil
	})

	// stringify 方法
	stringifyFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("yaml.stringify expects 1 arg, got %d", len(args))
		}

		data := yamlToGoValue(args[0])
		bytes, err := yaml.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("yaml.stringify failed: %w", err)
		}

		return string(bytes), nil
	})

	// saveToFile 方法
	saveToFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("yaml.saveToFile expects 2 args: data, path")
		}

		path, err := asStringValue("yaml.saveToFile", args[1])
		if err != nil {
			return nil, err
		}

		data := yamlToGoValue(args[0])
		bytes, err := yaml.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("yaml.saveToFile marshal error: %w", err)
		}

		if err := os.WriteFile(path, bytes, 0644); err != nil {
			return nil, fmt.Errorf("yaml.saveToFile write error: %w", err)
		}

		return true, nil
	})

	return Object{
		"parse":      parseFn,
		"fromFile":   fromFileFn,
		"stringify":  stringifyFn,
		"saveToFile": saveToFileFn,
		"dump":       stringifyFn,
		"load":       parseFn,
	}
}

// tomlModule 创建 TOML 模块
func newTOMLModule() Value {
	// parse 方法
	parseFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("toml.parse expects 1 arg (string), got %d", len(args))
		}

		str, err := asStringValue("toml.parse", args[0])
		if err != nil {
			return nil, err
		}

		var result map[string]interface{}
		if _, err := toml.Decode(str, &result); err != nil {
			return nil, fmt.Errorf("toml.parse failed: %w", err)
		}

		return yamlToObject(result), nil
	})

	// fromFile 方法
	fromFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("toml.fromFile expects 1 arg (path), got %d", len(args))
		}

		path, err := asStringValue("toml.fromFile", args[0])
		if err != nil {
			return nil, err
		}

		var result map[string]interface{}
		if _, err := toml.DecodeFile(path, &result); err != nil {
			return nil, fmt.Errorf("toml.fromFile failed: %w", err)
		}

		return yamlToObject(result), nil
	})

	// stringify 方法
	stringifyFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("toml.stringify expects 1 arg, got %d", len(args))
		}

		data := yamlToGoValue(args[0])

		// 转换为 map 类型
		var m map[string]interface{}
		switch v := data.(type) {
		case map[string]interface{}:
			m = v
		default:
			return nil, fmt.Errorf("toml.stringify requires an object")
		}

		// 使用 TOML 编码器
		buf := new(strings.Builder)
		enc := toml.NewEncoder(buf)
		if err := enc.Encode(m); err != nil {
			return nil, fmt.Errorf("toml.stringify failed: %w", err)
		}

		return buf.String(), nil
	})

	// saveToFile 方法
	saveToFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("toml.saveToFile expects 2 args: data, path")
		}

		path, err := asStringValue("toml.saveToFile", args[1])
		if err != nil {
			return nil, err
		}

		data := yamlToGoValue(args[0])

		var m map[string]interface{}
		switch v := data.(type) {
		case map[string]interface{}:
			m = v
		default:
			return nil, fmt.Errorf("toml.saveToFile requires an object")
		}

		// 直接写入文件
		f, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("toml.saveToFile create error: %w", err)
		}
		defer f.Close()

		enc := toml.NewEncoder(f)
		if err := enc.Encode(m); err != nil {
			return nil, fmt.Errorf("toml.saveToFile encode error: %w", err)
		}

		return true, nil
	})

	return Object{
		"parse":      parseFn,
		"fromFile":   fromFileFn,
		"stringify":  stringifyFn,
		"saveToFile": saveToFileFn,
		"encode":     stringifyFn,
		"decode":     parseFn,
	}
}

// yamlToValue 将 Go 值转换为 ialang 值
func yamlToValue(v interface{}) Value {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case bool:
		return val
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case string:
		return val
	case []byte:
		return string(val)
	case []interface{}:
		arr := make(Array, len(val))
		for i, item := range val {
			arr[i] = yamlToValue(item)
		}
		return arr
	case map[string]interface{}:
		return yamlToObject(val)
	case map[interface{}]interface{}:
		// YAML 可能返回这种类型
		obj := make(Object)
		for k, v := range val {
			keyStr := fmt.Sprintf("%v", k)
			obj[keyStr] = yamlToValue(v)
		}
		return obj
	default:
		return fmt.Sprintf("%v", val)
	}
}

// yamlToObject 将 Go map 转换为 ialang Object
func yamlToObject(m map[string]interface{}) Object {
	obj := make(Object)
	for k, v := range m {
		obj[k] = yamlToValue(v)
	}
	return obj
}

// yamlToGoValue 将 ialang 值转换为 Go 值
func yamlToGoValue(v Value) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case bool:
		return val
	case float64:
		// 检查是否是整数
		if val == float64(int(val)) {
			return int(val)
		}
		return val
	case string:
		return val
	case Object:
		m := make(map[string]interface{})
		for k, v := range val {
			m[k] = yamlToGoValue(v)
		}
		return m
	case Array:
		arr := make([]interface{}, len(val))
		for i, item := range val {
			arr[i] = yamlToGoValue(item)
		}
		return arr
	default:
		return val
	}
}

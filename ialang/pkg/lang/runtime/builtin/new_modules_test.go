package builtin

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ialang/pkg/express"
	comp "ialang/pkg/lang/compiler"
	"ialang/pkg/lang/frontend"
	rt "ialang/pkg/lang/runtime"
	rtvm "ialang/pkg/lang/runtime/vm"
)

// TestArrayNewMethods 测试数组新增方法
func TestArrayNewMethods(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "array")

	// 测试 map
	t.Run("map", func(t *testing.T) {
		arr := Array{float64(1), float64(2), float64(3)}
		callback := NativeFunction(func(args []Value) (Value, error) {
			item := args[0].(float64)
			return item * 2, nil
		})
		result := callNative(t, mod, "map", arr, callback)
		resultArr, ok := result.(Array)
		if !ok || len(resultArr) != 3 {
			t.Fatalf("map result length = %d, want 3", len(resultArr))
		}
		if got := resultArr[0].(float64); got != 2 {
			t.Fatalf("map[0] = %v, want 2", got)
		}
	})

	// 测试 filter
	t.Run("filter", func(t *testing.T) {
		arr := Array{float64(1), float64(2), float64(3), float64(4)}
		callback := NativeFunction(func(args []Value) (Value, error) {
			item := args[0].(float64)
			return item > 2, nil
		})
		result := callNative(t, mod, "filter", arr, callback)
		resultArr, ok := result.(Array)
		if !ok || len(resultArr) != 2 {
			t.Fatalf("filter result length = %d, want 2", len(resultArr))
		}
	})

	// 测试 find
	t.Run("find", func(t *testing.T) {
		arr := Array{float64(10), float64(20), float64(30)}
		callback := NativeFunction(func(args []Value) (Value, error) {
			item := args[0].(float64)
			return item > 15, nil
		})
		result := callNative(t, mod, "find", arr, callback)
		if got := result.(float64); got != 20 {
			t.Fatalf("find = %v, want 20", got)
		}
	})

	// 测试 findIndex
	t.Run("findIndex", func(t *testing.T) {
		arr := Array{float64(10), float64(20), float64(30)}
		callback := NativeFunction(func(args []Value) (Value, error) {
			item := args[0].(float64)
			return item > 15, nil
		})
		result := callNative(t, mod, "findIndex", arr, callback)
		if got := result.(float64); got != 1 {
			t.Fatalf("findIndex = %v, want 1", got)
		}
	})

	// 测试 forEach
	t.Run("forEach", func(t *testing.T) {
		arr := Array{float64(1), float64(2), float64(3)}
		count := 0
		callback := NativeFunction(func(args []Value) (Value, error) {
			count++
			return true, nil
		})
		callNative(t, mod, "forEach", arr, callback)
		if count != 3 {
			t.Fatalf("forEach count = %d, want 3", count)
		}
	})

	// 测试 some
	t.Run("some", func(t *testing.T) {
		arr := Array{float64(1), float64(2), float64(3)}
		callback := NativeFunction(func(args []Value) (Value, error) {
			item := args[0].(float64)
			return item > 2, nil
		})
		result := callNative(t, mod, "some", arr, callback)
		if result != true {
			t.Fatalf("some = %v, want true", result)
		}
	})

	// 测试 every
	t.Run("every", func(t *testing.T) {
		arr := Array{float64(1), float64(2), float64(3)}
		callback := NativeFunction(func(args []Value) (Value, error) {
			item := args[0].(float64)
			return item > 0, nil
		})
		result := callNative(t, mod, "every", arr, callback)
		if result != true {
			t.Fatalf("every = %v, want true", result)
		}
	})

	// 测试 reduce
	t.Run("reduce", func(t *testing.T) {
		arr := Array{float64(1), float64(2), float64(3), float64(4), float64(5)}
		callback := NativeFunction(func(args []Value) (Value, error) {
			acc := args[0].(float64)
			item := args[1].(float64)
			return acc + item, nil
		})
		result := callNative(t, mod, "reduce", arr, callback, float64(0))
		if got := result.(float64); got != 15 {
			t.Fatalf("reduce = %v, want 15", got)
		}
	})

	// 测试 includes
	t.Run("includes", func(t *testing.T) {
		arr := Array{float64(1), float64(2), float64(3)}
		result := callNative(t, mod, "includes", arr, float64(2))
		if result != true {
			t.Fatalf("includes = %v, want true", result)
		}
		result = callNative(t, mod, "includes", arr, float64(5))
		if result != false {
			t.Fatalf("includes = %v, want false", result)
		}
	})

	// 测试 indexOf
	t.Run("indexOf", func(t *testing.T) {
		arr := Array{float64(10), float64(20), float64(30)}
		result := callNative(t, mod, "indexOf", arr, float64(20))
		if got := result.(float64); got != 1 {
			t.Fatalf("indexOf = %v, want 1", got)
		}
	})

	// 测试 slice
	t.Run("slice", func(t *testing.T) {
		arr := Array{float64(1), float64(2), float64(3), float64(4), float64(5)}
		result := callNative(t, mod, "slice", arr, float64(1), float64(4))
		resultArr := result.(Array)
		if len(resultArr) != 3 {
			t.Fatalf("slice length = %d, want 3", len(resultArr))
		}
	})

	// 测试 join
	t.Run("join", func(t *testing.T) {
		arr := Array{"a", "b", "c"}
		result := callNative(t, mod, "join", arr, "-")
		if got := result.(string); got != "a-b-c" {
			t.Fatalf("join = %v, want a-b-c", got)
		}
	})
}

// TestJSONFromfile 测试 JSON fromFile 方法
func TestJSONFromFile(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "json")

	// 创建临时 JSON 文件
	fsMod := mustModuleObject(t, modules, "fs")
	callNative(t, fsMod, "writeFile", "./test_temp.json", `{"name":"test","value":123}`)

	// 测试 fromFile
	result := callNative(t, mod, "fromFile", "./test_temp.json")
	obj := mustRuntimeObject(t, result, "fromFile result")
	if got := obj["name"].(string); got != "test" {
		t.Fatalf("fromFile name = %v, want test", got)
	}

	// 清理临时文件
	callNative(t, fsMod, "writeFile", "./test_temp.json", "")
}

// TestYAMLFromFile 测试 YAML fromFile 方法
func TestYAMLFromFile(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "yaml")

	// 创建临时 YAML 文件
	fsMod := mustModuleObject(t, modules, "fs")
	callNative(t, fsMod, "writeFile", "./test_temp.yaml", "name: test\nvalue: 123")

	// 测试 fromFile
	result := callNative(t, mod, "fromFile", "./test_temp.yaml")
	obj := mustRuntimeObject(t, result, "fromFile result")
	if got := obj["name"].(string); got != "test" {
		t.Fatalf("yaml fromFile name = %v, want test", got)
	}

	// 清理
	callNative(t, fsMod, "writeFile", "./test_temp.yaml", "")
}

// TestTOMLFromFile 测试 TOML fromFile 方法
func TestTOMLFromFile(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "toml")

	// 创建临时 TOML 文件
	fsMod := mustModuleObject(t, modules, "fs")
	callNative(t, fsMod, "writeFile", "./test_temp.toml", "name = \"test\"\nvalue = 123")

	// 测试 fromFile
	result := callNative(t, mod, "fromFile", "./test_temp.toml")
	obj := mustRuntimeObject(t, result, "fromFile result")
	if got := obj["name"].(string); got != "test" {
		t.Fatalf("toml fromFile name = %v, want test", got)
	}

	// 清理
	callNative(t, fsMod, "writeFile", "./test_temp.toml", "")
}

// TestXMLFromFile 测试 XML fromFile 方法
func TestXMLFromFile(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "xml")

	// 创建临时 XML 文件
	fsMod := mustModuleObject(t, modules, "fs")
	callNative(t, fsMod, "writeFile", "./test_temp.xml", `<root name="test" value="123"/>`)

	// 测试 fromFile
	result := callNative(t, mod, "fromFile", "./test_temp.xml")
	obj := mustRuntimeObject(t, result, "fromFile result")
	if got := obj["name"].(string); got != "root" {
		t.Fatalf("xml fromFile name = %v, want root", got)
	}

	// 清理
	callNative(t, fsMod, "writeFile", "./test_temp.xml", "")
}

// TestLogOutputPath 测试日志模块设置输出路径
func TestLogOutputPath(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "log")

	// 测试 getOutputPath 默认值
	result := callNative(t, mod, "getOutputPath")
	if got := result.(string); got != "stdout" {
		t.Fatalf("getOutputPath default = %v, want stdout", got)
	}

	// 测试 setOutputPath
	callNative(t, mod, "setOutputPath", "./test_log.txt")

	// 验证路径已更改
	result = callNative(t, mod, "getOutputPath")
	if got := result.(string); got != "./test_log.txt" {
		t.Fatalf("getOutputPath = %v, want ./test_log.txt", got)
	}

	// 写一条日志
	callNative(t, mod, "info", "Test log message")

	// 清理
	fsMod := mustModuleObject(t, modules, "fs")
	callNative(t, fsMod, "writeFile", "./test_log.txt", "")
	callNative(t, mod, "setOutputPath", "stdout")
}

func TestLogOutputPathCreatesMissingDirectories(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "log")

	baseDir := t.TempDir()
	logPath := filepath.Join(baseDir, "logs", "nested", "app.log")

	callNative(t, mod, "setOutputPath", logPath)
	callNative(t, mod, "info", "Test log message")

	if _, err := os.Stat(filepath.Dir(logPath)); err != nil {
		t.Fatalf("log output dir not created: %v", err)
	}
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log output file not created: %v", err)
	}

	callNative(t, mod, "setOutputPath", "stdout")
}

func TestSignalModuleRegistered(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "signal")

	if got, ok := mod["SIGINT"].(string); !ok || got != "SIGINT" {
		t.Fatalf("signal.SIGINT = %#v, want SIGINT", mod["SIGINT"])
	}
}

func TestSignalNotifyRejectsUnsupportedSignal(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "signal")

	if _, err := callNativeWithError(mod, "notify", "SIGUNKNOWN"); err == nil {
		t.Fatal("signal.notify expected error for unsupported signal, got nil")
	}
}

// TestExpressModule 测试 Express 模块基本功能
func TestExpressModule(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "express")

	// 测试创建应用
	app := callNative(t, mod, "express")
	appObj := mustRuntimeObject(t, app, "express result")

	// 测试 get 方法注册路由
	handler := NativeFunction(func(args []Value) (Value, error) {
		return true, nil
	})
	callNative(t, appObj, "get", "/", handler)
}

func TestCreateHandlerFromValueSupportsBoundMethod(t *testing.T) {
	chunk := compileBuiltinTestChunk(t, `
class UserController {
  test(ctx) {
    ctx.json({
      message: "Hello, User!"
    });
  }
}

let controller = new UserController();
export let handler = controller.test;
`)

	runtime := rt.NewGoroutineRuntime()
	vm := rtvm.NewVM(chunk, DefaultModules(runtime), nil, "express_bound_method_test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}

	handlerVal, ok := vm.Exports()["handler"]
	if !ok {
		t.Fatal("expected exported handler")
	}

	handler, ok := createHandlerFromValue(handlerVal)
	if !ok {
		t.Fatalf("createHandlerFromValue(%T) = false, want true", handlerVal)
	}

	req := httptest.NewRequest("GET", "/user/test", nil)
	rec := httptest.NewRecorder()
	ctx := express.NewContext(req, rec)

	handler(ctx)

	if rec.Code != 200 {
		t.Fatalf("response status = %d, want 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"message":"Hello, User!"}` {
		t.Fatalf("response body = %q, want %q", got, `{"message":"Hello, User!"}`)
	}
}

// TestYAMLModule 测试 YAML 模块
func TestYAMLModule(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "yaml")

	// 测试 parse
	result := callNative(t, mod, "parse", "name: test\nvalue: 123")
	obj := mustRuntimeObject(t, result, "parse result")
	if got := obj["name"].(string); got != "test" {
		t.Fatalf("yaml parse name = %v, want test", got)
	}

	// 测试 stringify
	data := Object{"name": "test", "value": float64(123)}
	result = callNative(t, mod, "stringify", data)
	if _, ok := result.(string); !ok {
		t.Fatalf("yaml stringify result type = %T, want string", result)
	}
}

// TestTOMLModule 测试 TOML 模块
func TestTOMLModule(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "toml")

	// 测试 parse
	result := callNative(t, mod, "parse", "name = \"test\"\nvalue = 123")
	obj := mustRuntimeObject(t, result, "parse result")
	if got := obj["name"].(string); got != "test" {
		t.Fatalf("toml parse name = %v, want test", got)
	}
}

// TestArrayFlat 测试数组扁平化
func TestArrayFlat(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "array")

	// 创建嵌套数组
	nested := Array{
		float64(1),
		Array{float64(2), float64(3)},
		Array{float64(4), float64(5)},
	}

	result := callNative(t, mod, "flat", nested, float64(1))
	resultArr := result.(Array)
	if len(resultArr) != 5 {
		t.Fatalf("flat length = %d, want 5", len(resultArr))
	}
}

// TestArraySliceNegativeIndex 测试负数索引
func TestArraySliceNegativeIndex(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "array")

	arr := Array{float64(1), float64(2), float64(3), float64(4), float64(5)}

	// 测试负数索引
	result := callNative(t, mod, "slice", arr, float64(-2))
	resultArr := result.(Array)
	if len(resultArr) != 2 {
		t.Fatalf("slice with -2 length = %d, want 2", len(resultArr))
	}
}

// TestArrayReduceWithoutInitialValue 测试 reduce 无初始值
func TestArrayReduceWithoutInitialValue(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "array")

	arr := Array{float64(1), float64(2), float64(3)}
	callback := NativeFunction(func(args []Value) (Value, error) {
		acc := args[0].(float64)
		item := args[1].(float64)
		return acc + item, nil
	})

	result := callNative(t, mod, "reduce", arr, callback)
	if got := result.(float64); got != 6 {
		t.Fatalf("reduce = %v, want 6", got)
	}
}

// TestJSONSaveToFile 测试 JSON 保存到文件
func TestJSONSaveToFile(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	jsonMod := mustModuleObject(t, modules, "json")
	fsMod := mustModuleObject(t, modules, "fs")

	// 测试保存数据
	data := Object{"name": "test", "value": float64(123)}
	result := callNative(t, jsonMod, "saveToFile", data, "./test_save.json", true)
	if result != true {
		t.Fatalf("saveToFile result = %v, want true", result)
	}

	// 验证文件内容
	content := callNative(t, fsMod, "readFile", "./test_save.json")
	contentStr, ok := content.(string)
	if !ok {
		t.Fatalf("readFile content type = %T, want string", content)
	}
	if len(contentStr) == 0 {
		t.Fatal("saved file is empty")
	}

	// 清理
	callNative(t, fsMod, "writeFile", "./test_save.json", "")
}

// TestYAMLSaveToFile 测试 YAML 保存到文件
func TestYAMLSaveToFile(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "yaml")
	fsMod := mustModuleObject(t, modules, "fs")

	// 测试保存数据
	data := Object{"name": "test", "value": float64(123)}
	result := callNative(t, mod, "saveToFile", data, "./test_save.yaml")
	if result != true {
		t.Fatalf("yaml saveToFile result = %v, want true", result)
	}

	// 验证文件内容
	content := callNative(t, fsMod, "readFile", "./test_save.yaml")
	contentStr, ok := content.(string)
	if !ok {
		t.Fatalf("readFile content type = %T, want string", content)
	}
	if len(contentStr) == 0 {
		t.Fatal("saved file is empty")
	}

	// 验证可以重新读取
	parsed := callNative(t, mod, "fromFile", "./test_save.yaml")
	parsedObj := mustRuntimeObject(t, parsed, "yaml fromFile")
	if got := parsedObj["name"].(string); got != "test" {
		t.Fatalf("yaml fromFile name = %v, want test", got)
	}

	// 清理
	callNative(t, fsMod, "writeFile", "./test_save.yaml", "")
}

// TestTOMLSaveToFile 测试 TOML 保存到文件
func TestTOMLSaveToFile(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "toml")
	fsMod := mustModuleObject(t, modules, "fs")

	// 测试保存数据
	data := Object{
		"title": "Test App",
		"owner": Object{
			"name": "John",
		},
	}
	result := callNative(t, mod, "saveToFile", data, "./test_save.toml")
	if result != true {
		t.Fatalf("toml saveToFile result = %v, want true", result)
	}

	// 验证文件内容
	content := callNative(t, fsMod, "readFile", "./test_save.toml")
	contentStr, ok := content.(string)
	if !ok {
		t.Fatalf("readFile content type = %T, want string", content)
	}
	if len(contentStr) == 0 {
		t.Fatal("saved file is empty")
	}

	// 验证可以重新读取
	parsed := callNative(t, mod, "fromFile", "./test_save.toml")
	parsedObj := mustRuntimeObject(t, parsed, "toml fromFile")
	if got := parsedObj["title"].(string); got != "Test App" {
		t.Fatalf("toml fromFile title = %v, want Test App", got)
	}

	// 清理
	callNative(t, fsMod, "writeFile", "./test_save.toml", "")
}

// TestXMLSaveToFile 测试 XML 保存到文件
func TestXMLSaveToFile(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "xml")
	fsMod := mustModuleObject(t, modules, "fs")

	// 创建 XML 节点
	node := Object{
		"name": "root",
		"attrs": Object{
			"version": "1.0",
		},
		"text": "",
		"children": Array{
			Object{
				"name":     "item",
				"attrs":    Object{},
				"text":     "Hello",
				"children": Array{},
			},
		},
	}

	// 测试保存
	result := callNative(t, mod, "saveToFile", node, "./test_save.xml", true)
	if result != true {
		t.Fatalf("xml saveToFile result = %v, want true", result)
	}

	// 验证文件内容
	content := callNative(t, fsMod, "readFile", "./test_save.xml")
	contentStr, ok := content.(string)
	if !ok {
		t.Fatalf("readFile content type = %T, want string", content)
	}
	if len(contentStr) == 0 {
		t.Fatal("saved file is empty")
	}

	// 验证可以重新读取
	parsed := callNative(t, mod, "fromFile", "./test_save.xml")
	parsedObj := mustRuntimeObject(t, parsed, "xml fromFile")
	if got := parsedObj["name"].(string); got != "root" {
		t.Fatalf("xml fromFile name = %v, want root", got)
	}

	// 清理
	callNative(t, fsMod, "writeFile", "./test_save.xml", "")
}

// TestJSONSaveToFileErrorCase 测试 JSON 保存错误情况
func TestJSONSaveToFileErrorCase(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "json")

	// 测试参数数量错误
	_, err := callNativeWithError(mod, "saveToFile", Object{"test": "value"})
	if err == nil {
		t.Fatal("expected error for missing path, got nil")
	}

	// 测试无效路径
	_, err = callNativeWithError(mod, "saveToFile", Object{"test": "value"}, "/invalid/path/that/does/not/exist/file.json")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

func compileBuiltinTestChunk(t *testing.T, source string) *rt.Chunk {
	t.Helper()
	l := frontend.NewLexer(source)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	c := comp.NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	return chunk
}

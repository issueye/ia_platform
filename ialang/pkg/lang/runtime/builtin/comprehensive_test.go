package builtin

import (
	"math"
	"path/filepath"
	"strings"
	"testing"

	rt "ialang/pkg/lang/runtime"
)

func getModules(t *testing.T) map[string]Value {
	t.Helper()
	return DefaultModules(rt.NewGoroutineRuntime())
}

// ─── Math Module ────────────────────────────────────────────────────────────

func TestMathModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "math")

	t.Run("constants", func(t *testing.T) {
		if v, ok := mod["PI"].(float64); !ok || v != math.Pi {
			t.Fatalf("math.PI = %v, want %v", mod["PI"], math.Pi)
		}
		if v, ok := mod["E"].(float64); !ok || v != math.E {
			t.Fatalf("math.E = %v, want %v", mod["E"], math.E)
		}
		if v, ok := mod["sqrt2"].(float64); !ok || v != math.Sqrt2 {
			t.Fatalf("math.sqrt2 = %v, want %v", mod["sqrt2"], math.Sqrt2)
		}
		if v, ok := mod["NaN"].(float64); !ok || !math.IsNaN(v) {
			t.Fatalf("math.NaN = %v, want NaN", mod["NaN"])
		}
		if v, ok := mod["Infinity"].(float64); !ok || !math.IsInf(v, 1) {
			t.Fatalf("math.Infinity = %v, want +Inf", mod["Infinity"])
		}
		if v, ok := mod["NEG_INFINITY"].(float64); !ok || !math.IsInf(v, -1) {
			t.Fatalf("math.NEG_INFINITY = %v, want -Inf", mod["NEG_INFINITY"])
		}
	})

	t.Run("basic_single_arg", func(t *testing.T) {
		tests := []struct {
			name, fn string
			in, want float64
		}{
			{"abs", "abs", -5, 5},
			{"ceil", "ceil", 1.2, 2},
			{"floor", "floor", 1.8, 1},
			{"round", "round", 1.5, 2},
			{"sqrt", "sqrt", 9, 3},
			{"log", "log", math.E, 1},
			{"log10", "log10", 100, 2},
			{"log2", "log2", 8, 3},
			{"exp", "exp", 0, 1},
			{"trunc", "trunc", 1.9, 1},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				got := callNative(t, mod, tc.fn, tc.in).(float64)
				if math.Abs(got-tc.want) > 1e-9 {
					t.Fatalf("math.%s(%v) = %v, want %v", tc.fn, tc.in, got, tc.want)
				}
			})
		}
	})

	t.Run("trig", func(t *testing.T) {
		if v := callNative(t, mod, "sin", 0.0).(float64); v != 0 {
			t.Fatalf("math.sin(0) = %v, want 0", v)
		}
		if v := callNative(t, mod, "cos", 0.0).(float64); math.Abs(v-1) > 1e-9 {
			t.Fatalf("math.cos(0) = %v, want 1", v)
		}
		if v := callNative(t, mod, "asin", 1.0).(float64); math.Abs(v-math.Pi/2) > 1e-9 {
			t.Fatalf("math.asin(1) = %v, want pi/2", v)
		}
		if v := callNative(t, mod, "acos", 1.0).(float64); v != 0 {
			t.Fatalf("math.acos(1) = %v, want 0", v)
		}
		if v := callNative(t, mod, "atan", 1.0).(float64); math.Abs(v-math.Pi/4) > 1e-9 {
			t.Fatalf("math.atan(1) = %v, want pi/4", v)
		}
	})

	t.Run("atan2", func(t *testing.T) {
		v := callNative(t, mod, "atan2", 1.0, 1.0).(float64)
		if math.Abs(v-math.Pi/4) > 1e-9 {
			t.Fatalf("math.atan2(1,1) = %v, want pi/4", v)
		}
	})

	t.Run("sign", func(t *testing.T) {
		if v := callNative(t, mod, "sign", 5.0).(float64); v != 1 {
			t.Fatalf("math.sign(5) = %v, want 1", v)
		}
		if v := callNative(t, mod, "sign", -3.0).(float64); v != -1 {
			t.Fatalf("math.sign(-3) = %v, want -1", v)
		}
		if v := callNative(t, mod, "sign", 0.0).(float64); v != 0 {
			t.Fatalf("math.sign(0) = %v, want 0", v)
		}
	})

	t.Run("pow", func(t *testing.T) {
		v := callNative(t, mod, "pow", 2.0, 10.0).(float64)
		if v != 1024 {
			t.Fatalf("math.pow(2,10) = %v, want 1024", v)
		}
	})

	t.Run("mod", func(t *testing.T) {
		v := callNative(t, mod, "mod", 10.0, 3.0).(float64)
		if v != 1 {
			t.Fatalf("math.mod(10,3) = %v, want 1", v)
		}
		_, err := callNativeWithError(mod, "mod", 10.0, 0.0)
		if err == nil || !strings.Contains(err.Error(), "division by zero") {
			t.Fatalf("math.mod(10,0) error = %v, want division by zero", err)
		}
	})

	t.Run("max_variadic", func(t *testing.T) {
		v := callNative(t, mod, "max", 3.0, 1.0, 4.0, 1.5).(float64)
		if v != 4 {
			t.Fatalf("math.max(3,1,4,1.5) = %v, want 4", v)
		}
	})

	t.Run("min_variadic", func(t *testing.T) {
		v := callNative(t, mod, "min", 3.0, 1.0, 4.0, 1.5).(float64)
		if v != 1 {
			t.Fatalf("math.min(3,1,4,1.5) = %v, want 1", v)
		}
	})

	t.Run("random", func(t *testing.T) {
		v := callNative(t, mod, "random").(float64)
		if v < 0 || v >= 1 {
			t.Fatalf("math.random() = %v, want [0,1)", v)
		}
		v2 := callNative(t, mod, "random", 10.0, 20.0).(float64)
		if v2 < 10 || v2 >= 20 {
			t.Fatalf("math.random(10,20) = %v, want [10,20)", v2)
		}
	})

	t.Run("errors", func(t *testing.T) {
		_, err := callNativeWithError(mod, "abs")
		if err == nil {
			t.Fatal("math.abs() expected error")
		}
		_, err = callNativeWithError(mod, "abs", "not-a-number")
		if err == nil {
			t.Fatal("math.abs('x') expected error")
		}
		_, err = callNativeWithError(mod, "max")
		if err == nil {
			t.Fatal("math.max() expected error")
		}
		_, err = callNativeWithError(mod, "pow", 2.0)
		if err == nil {
			t.Fatal("math.pow(2) expected error")
		}
	})
}

// ─── String Module ──────────────────────────────────────────────────────────

func TestStringModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "string")

	t.Run("split", func(t *testing.T) {
		v := callNative(t, mod, "split", "a,b,c", ",")
		arr := v.(Array)
		if len(arr) != 3 || arr[0].(string) != "a" {
			t.Fatalf("string.split = %v", v)
		}
	})

	t.Run("join", func(t *testing.T) {
		v := callNative(t, mod, "join", Array{"a", "b", "c"}, "-")
		if v.(string) != "a-b-c" {
			t.Fatalf("string.join = %v", v)
		}
	})

	t.Run("parseInt", func(t *testing.T) {
		v := callNative(t, mod, "parseInt", "42")
		if v.(float64) != 42 {
			t.Fatalf("string.parseInt = %v", v)
		}
	})

	t.Run("parseFloat", func(t *testing.T) {
		v := callNative(t, mod, "parseFloat", "3.14")
		if v.(float64) != 3.14 {
			t.Fatalf("string.parseFloat = %v", v)
		}
	})

	t.Run("fromCodePoint", func(t *testing.T) {
		v := callNative(t, mod, "fromCodePoint", 65.0)
		if v.(string) != "A" {
			t.Fatalf("string.fromCodePoint(65) = %v", v)
		}
	})

	t.Run("trim", func(t *testing.T) {
		v := callNative(t, mod, "trim", "  hi  ")
		if v.(string) != "hi" {
			t.Fatalf("string.trim = %v", v)
		}
	})

	t.Run("toLowerCase", func(t *testing.T) {
		v := callNative(t, mod, "toLowerCase", "ABC")
		if v.(string) != "abc" {
			t.Fatalf("string.toLowerCase = %v", v)
		}
	})

	t.Run("toUpperCase", func(t *testing.T) {
		v := callNative(t, mod, "toUpperCase", "abc")
		if v.(string) != "ABC" {
			t.Fatalf("string.toUpperCase = %v", v)
		}
	})

	t.Run("contains", func(t *testing.T) {
		if !callNative(t, mod, "contains", "hello world", "world").(bool) {
			t.Fatal("string.contains should be true")
		}
	})

	t.Run("indexOf", func(t *testing.T) {
		v := callNative(t, mod, "indexOf", "hello", "ll")
		if v.(float64) != 2 {
			t.Fatalf("string.indexOf = %v", v)
		}
	})

	t.Run("lastIndexOf", func(t *testing.T) {
		v := callNative(t, mod, "lastIndexOf", "abcabc", "bc")
		if v.(float64) != 4 {
			t.Fatalf("string.lastIndexOf = %v", v)
		}
	})

	t.Run("startsWith", func(t *testing.T) {
		if !callNative(t, mod, "startsWith", "hello", "hel").(bool) {
			t.Fatal("string.startsWith should be true")
		}
	})

	t.Run("endsWith", func(t *testing.T) {
		if !callNative(t, mod, "endsWith", "hello", "llo").(bool) {
			t.Fatal("string.endsWith should be true")
		}
	})

	t.Run("replace", func(t *testing.T) {
		v := callNative(t, mod, "replace", "aabaa", "aa", "x")
		if v.(string) != "xbx" {
			t.Fatalf("string.replace = %v", v)
		}
	})

	t.Run("length", func(t *testing.T) {
		v := callNative(t, mod, "length", "abc")
		if v.(float64) != 3 {
			t.Fatalf("string.length = %v", v)
		}
		v2 := callNative(t, mod, "length", "你好")
		if v2.(float64) != 2 {
			t.Fatalf("string.length(unicode) = %v, want 2", v2)
		}
	})

	t.Run("repeat", func(t *testing.T) {
		v := callNative(t, mod, "repeat", "ab", 3.0)
		if v.(string) != "ababab" {
			t.Fatalf("string.repeat = %v", v)
		}
	})

	t.Run("substring", func(t *testing.T) {
		v := callNative(t, mod, "substring", "hello", 1.0, 4.0)
		if v.(string) != "ell" {
			t.Fatalf("string.substring(1,4) = %v", v)
		}
		v2 := callNative(t, mod, "substring", "hello", 2.0)
		if v2.(string) != "llo" {
			t.Fatalf("string.substring(2) = %v", v2)
		}
		v3 := callNative(t, mod, "substring", "你好世界", 1.0, 3.0)
		if v3.(string) != "好世" {
			t.Fatalf("string.substring(unicode) = %v", v3)
		}
	})

	t.Run("padStart", func(t *testing.T) {
		v := callNative(t, mod, "padStart", "5", 3.0, "0")
		if v.(string) != "005" {
			t.Fatalf("string.padStart = %v", v)
		}
	})

	t.Run("padEnd", func(t *testing.T) {
		v := callNative(t, mod, "padEnd", "hi", 5.0, ".")
		if v.(string) != "hi..." {
			t.Fatalf("string.padEnd = %v", v)
		}
	})

	t.Run("charCodeAt", func(t *testing.T) {
		v := callNative(t, mod, "charCodeAt", "ABC", 1.0)
		if v.(float64) != 66 {
			t.Fatalf("string.charCodeAt = %v", v)
		}
		v2 := callNative(t, mod, "charCodeAt", "AB", 99.0)
		if v2.(float64) != -1 {
			t.Fatalf("string.charCodeAt(out of range) = %v, want -1", v2)
		}
	})
}

// ─── Rand Module ────────────────────────────────────────────────────────────

func TestRandModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "rand")

	t.Run("int_no_args", func(t *testing.T) {
		v := callNative(t, mod, "int").(float64)
		if v < 0 {
			t.Fatalf("rand.int() = %v, want >= 0", v)
		}
	})

	t.Run("int_one_arg", func(t *testing.T) {
		v := callNative(t, mod, "int", 100.0).(float64)
		if v < 0 || v >= 100 {
			t.Fatalf("rand.int(100) = %v, want [0,100)", v)
		}
	})

	t.Run("int_two_args", func(t *testing.T) {
		v := callNative(t, mod, "int", 50.0, 100.0).(float64)
		if v < 50 || v >= 100 {
			t.Fatalf("rand.int(50,100) = %v, want [50,100)", v)
		}
	})

	t.Run("float_no_args", func(t *testing.T) {
		v := callNative(t, mod, "float").(float64)
		if v < 0 || v >= 1 {
			t.Fatalf("rand.float() = %v", v)
		}
	})

	t.Run("float_range", func(t *testing.T) {
		v := callNative(t, mod, "float", 10.0, 20.0).(float64)
		if v < 10 || v >= 20 {
			t.Fatalf("rand.float(10,20) = %v", v)
		}
	})

	t.Run("pick", func(t *testing.T) {
		arr := Array{"x", "y", "z"}
		v := callNative(t, mod, "pick", arr)
		s := v.(string)
		if s != "x" && s != "y" && s != "z" {
			t.Fatalf("rand.pick = %v", v)
		}
	})

	t.Run("string", func(t *testing.T) {
		v := callNative(t, mod, "string", 8.0).(string)
		if len(v) != 8 {
			t.Fatalf("rand.string(8) len = %d", len(v))
		}
	})

	t.Run("shuffle", func(t *testing.T) {
		orig := Array{float64(1), float64(2), float64(3), float64(4), float64(5), float64(6), float64(7), float64(8), float64(9), float64(10)}
		result := callNative(t, mod, "shuffle", orig)
		arr := result.(Array)
		if len(arr) != len(orig) {
			t.Fatalf("rand.shuffle length changed: %d", len(arr))
		}
		if len(arr) <= 1 {
			return
		}
		sameCount := 0
		for i := range arr {
			if arr[i] == orig[i] {
				sameCount++
			}
		}
		if sameCount == len(orig) {
			t.Fatal("rand.shuffle did not shuffle at all (very unlikely)")
		}
	})

	t.Run("seed", func(t *testing.T) {
		callNative(t, mod, "seed", 42.0)
		v1 := callNative(t, mod, "float").(float64)
		callNative(t, mod, "seed", 42.0)
		v2 := callNative(t, mod, "float").(float64)
		if v1 != v2 {
			t.Fatalf("rand.seed should produce deterministic results: %v vs %v", v1, v2)
		}
	})
}

// ─── JSON Module ────────────────────────────────────────────────────────────

func TestJSONModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "json")

	t.Run("parse_and_stringify", func(t *testing.T) {
		v := callNative(t, mod, "parse", `{"x":1,"y":"hi"}`)
		obj := v.(Object)
		if obj["x"].(float64) != 1 || obj["y"].(string) != "hi" {
			t.Fatalf("json.parse = %v", v)
		}
		s := callNative(t, mod, "stringify", obj).(string)
		if !strings.Contains(s, `"x":1`) {
			t.Fatalf("json.stringify = %v", s)
		}
	})

	t.Run("stringify_pretty", func(t *testing.T) {
		v := callNative(t, mod, "stringify", Object{"a": 1.0}, true).(string)
		if !strings.Contains(v, "\n") {
			t.Fatalf("json.stringify pretty = %v, want newlines", v)
		}
	})

	t.Run("valid", func(t *testing.T) {
		if !callNative(t, mod, "valid", `{"ok":true}`).(bool) {
			t.Fatal("json.valid should be true")
		}
		if callNative(t, mod, "valid", `{bad}`).(bool) {
			t.Fatal("json.valid({bad}) should be false")
		}
	})

	t.Run("fromFile_saveToFile", func(t *testing.T) {
		tmp := t.TempDir()
		fp := filepath.Join(tmp, "test.json")
		data := Object{"hello": "world", "n": 42.0}
		callNative(t, mod, "saveToFile", data, fp, true)
		loaded := callNative(t, mod, "fromFile", fp)
		obj := loaded.(Object)
		if obj["hello"].(string) != "world" {
			t.Fatalf("json roundtrip = %v", obj)
		}
	})

	t.Run("fromFileAsync_saveToFileAsync", func(t *testing.T) {
		tmp := t.TempDir()
		fp := filepath.Join(tmp, "async_test.json")
		data := Object{"async": true}
		awaitValue(t, callNative(t, mod, "saveToFileAsync", data, fp))
		loaded := awaitValue(t, callNative(t, mod, "fromFileAsync", fp))
		obj := loaded.(Object)
		if obj["async"].(bool) != true {
			t.Fatalf("json async roundtrip = %v", obj)
		}
	})
}

// ─── Time Module ────────────────────────────────────────────────────────────

func TestTimeModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "time")

	t.Run("nowUnix", func(t *testing.T) {
		v := callNative(t, mod, "nowUnix").(float64)
		if v <= 0 {
			t.Fatalf("time.nowUnix = %v", v)
		}
	})

	t.Run("nowUnixMilli", func(t *testing.T) {
		v := callNative(t, mod, "nowUnixMilli").(float64)
		if v <= 0 {
			t.Fatalf("time.nowUnixMilli = %v", v)
		}
	})

	t.Run("nowUnixMicro", func(t *testing.T) {
		v := callNative(t, mod, "nowUnixMicro").(float64)
		if v <= 0 {
			t.Fatalf("time.nowUnixMicro = %v", v)
		}
	})

	t.Run("nowISO", func(t *testing.T) {
		v := callNative(t, mod, "nowISO").(string)
		if v == "" {
			t.Fatal("time.nowISO empty")
		}
	})

	t.Run("parseISO", func(t *testing.T) {
		v := callNative(t, mod, "parseISO", "2024-01-15T10:30:00Z")
		obj := v.(Object)
		if obj["year"].(float64) != 2024 {
			t.Fatalf("time.parseISO year = %v", obj["year"])
		}
		if obj["month"].(float64) != 1 {
			t.Fatalf("time.parseISO month = %v", obj["month"])
		}
		if obj["day"].(float64) != 15 {
			t.Fatalf("time.parseISO day = %v", obj["day"])
		}
	})

	t.Run("format", func(t *testing.T) {
		v := callNative(t, mod, "format", 1705312200000.0)
		s := v.(string)
		if s == "" {
			t.Fatal("time.format empty")
		}
	})

	t.Run("parse", func(t *testing.T) {
		v := callNative(t, mod, "parse", "2024-01-15 10:30:00")
		ms := v.(float64)
		if ms <= 0 {
			t.Fatalf("time.parse = %v", ms)
		}
	})

	t.Run("add", func(t *testing.T) {
		base := 1000.0
		v := callNative(t, mod, "add", base, 500.0).(float64)
		if v != 1500 {
			t.Fatalf("time.add(1000,500) = %v, want 1500", v)
		}
	})

	t.Run("diff", func(t *testing.T) {
		v := callNative(t, mod, "diff", 1000.0, 2000.0).(float64)
		if v != 1000 {
			t.Fatalf("time.diff(1000,2000) = %v, want 1000", v)
		}
	})
}

// ─── Path Module ────────────────────────────────────────────────────────────

func TestPathModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "path")

	t.Run("join", func(t *testing.T) {
		v := callNative(t, mod, "join", "a", "b", "c").(string)
		if v != filepath.Join("a", "b", "c") {
			t.Fatalf("path.join = %v", v)
		}
	})

	t.Run("base", func(t *testing.T) {
		v := callNative(t, mod, "base", filepath.Join("a", "b.txt")).(string)
		if v != "b.txt" {
			t.Fatalf("path.base = %v", v)
		}
	})

	t.Run("dir", func(t *testing.T) {
		v := callNative(t, mod, "dir", filepath.Join("a", "b", "c.txt")).(string)
		if v != filepath.Join("a", "b") {
			t.Fatalf("path.dir = %v", v)
		}
	})

	t.Run("ext", func(t *testing.T) {
		v := callNative(t, mod, "ext", "file.tar.gz").(string)
		if v != ".gz" {
			t.Fatalf("path.ext = %v", v)
		}
	})

	t.Run("clean", func(t *testing.T) {
		v := callNative(t, mod, "clean", "a/../b/./c").(string)
		if v != filepath.Join("b", "c") {
			t.Fatalf("path.clean = %v", v)
		}
	})

	t.Run("abs", func(t *testing.T) {
		v := callNative(t, mod, "abs", ".").(string)
		if !filepath.IsAbs(v) {
			t.Fatalf("path.abs = %v, want absolute", v)
		}
	})

	t.Run("isAbsolute", func(t *testing.T) {
		if !callNative(t, mod, "isAbsolute", "C:\\Windows").(bool) {
			t.Fatal("path.isAbsolute('C:\\Windows') should be true")
		}
		if callNative(t, mod, "isAbsolute", "foo").(bool) {
			t.Fatal("path.isAbsolute('foo') should be false")
		}
	})

	t.Run("relative", func(t *testing.T) {
		v := callNative(t, mod, "relative", "/a/b", "/a/c").(string)
		if v == "" {
			t.Fatal("path.relative empty")
		}
	})

	t.Run("sep_and_listSep", func(t *testing.T) {
		if _, ok := mod["sep"].(string); !ok {
			t.Fatal("path.sep should be string")
		}
		if _, ok := mod["listSep"].(string); !ok {
			t.Fatal("path.listSep should be string")
		}
	})
}

// ─── FS Module Extended ────────────────────────────────────────────────────

func TestFSModuleExtended(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "fs")
	tmp := t.TempDir()

	t.Run("appendFile", func(t *testing.T) {
		fp := filepath.Join(tmp, "append.txt")
		callNative(t, mod, "writeFile", fp, "hello")
		callNative(t, mod, "appendFile", fp, " world")
		got := callNative(t, mod, "readFile", fp).(string)
		if got != "hello world" {
			t.Fatalf("appendFile = %v", got)
		}
	})

	t.Run("mkdir", func(t *testing.T) {
		dp := filepath.Join(tmp, "subdir")
		callNative(t, mod, "mkdir", dp, false)
		exists := callNative(t, mod, "exists", dp).(bool)
		if !exists {
			t.Fatal("mkdir did not create dir")
		}
	})

	t.Run("mkdir_recursive", func(t *testing.T) {
		dp := filepath.Join(tmp, "a", "b", "c")
		callNative(t, mod, "mkdir", dp, true)
		exists := callNative(t, mod, "exists", dp).(bool)
		if !exists {
			t.Fatal("mkdir recursive did not create dir")
		}
	})

	t.Run("rename", func(t *testing.T) {
		fp1 := filepath.Join(tmp, "before.txt")
		fp2 := filepath.Join(tmp, "after.txt")
		callNative(t, mod, "writeFile", fp1, "data")
		callNative(t, mod, "rename", fp1, fp2)
		if callNative(t, mod, "exists", fp1).(bool) {
			t.Fatal("rename: old file still exists")
		}
		got := callNative(t, mod, "readFile", fp2).(string)
		if got != "data" {
			t.Fatalf("rename: content = %v", got)
		}
	})

	t.Run("remove", func(t *testing.T) {
		fp := filepath.Join(tmp, "to_remove.txt")
		callNative(t, mod, "writeFile", fp, "x")
		callNative(t, mod, "remove", fp)
		if callNative(t, mod, "exists", fp).(bool) {
			t.Fatal("remove did not work")
		}
	})

	t.Run("removeAll", func(t *testing.T) {
		dp := filepath.Join(tmp, "nested", "deep")
		callNative(t, mod, "mkdir", dp, true)
		callNative(t, mod, "removeAll", filepath.Join(tmp, "nested"))
		if callNative(t, mod, "exists", filepath.Join(tmp, "nested")).(bool) {
			t.Fatal("removeAll did not work")
		}
	})

	t.Run("copy", func(t *testing.T) {
		src := filepath.Join(tmp, "src.txt")
		dst := filepath.Join(tmp, "dst.txt")
		callNative(t, mod, "writeFile", src, "copy me")
		callNative(t, mod, "copy", src, dst)
		got := callNative(t, mod, "readFile", dst).(string)
		if got != "copy me" {
			t.Fatalf("copy = %v", got)
		}
	})

	t.Run("async_variants", func(t *testing.T) {
		fp := filepath.Join(tmp, "async_test.txt")
		awaitValue(t, callNative(t, mod, "writeFileAsync", fp, "async hello"))
		got := awaitValue(t, callNative(t, mod, "readFileAsync", fp)).(string)
		if got != "async hello" {
			t.Fatalf("async roundtrip = %v", got)
		}
		exists := awaitValue(t, callNative(t, mod, "existsAsync", fp)).(bool)
		if !exists {
			t.Fatal("existsAsync should be true")
		}
		dirEntries := awaitValue(t, callNative(t, mod, "readDirAsync", tmp)).(Array)
		if len(dirEntries) == 0 {
			t.Fatal("readDirAsync empty")
		}
		stat := awaitValue(t, callNative(t, mod, "statAsync", fp)).(Object)
		if stat["size"].(float64) <= 0 {
			t.Fatal("statAsync size should be > 0")
		}
	})
}

// ─── Sort Module ────────────────────────────────────────────────────────────

func TestSortModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "sort")

	t.Run("asc_numbers", func(t *testing.T) {
		v := callNative(t, mod, "asc", Array{3.0, 1.0, 2.0}).(Array)
		if v[0].(float64) != 1 || v[2].(float64) != 3 {
			t.Fatalf("sort.asc = %v", v)
		}
	})

	t.Run("asc_strings", func(t *testing.T) {
		v := callNative(t, mod, "asc", Array{"c", "a", "b"}).(Array)
		if v[0].(string) != "a" {
			t.Fatalf("sort.asc strings = %v", v)
		}
	})

	t.Run("desc", func(t *testing.T) {
		v := callNative(t, mod, "desc", Array{1.0, 3.0, 2.0}).(Array)
		if v[0].(float64) != 3 || v[2].(float64) != 1 {
			t.Fatalf("sort.desc = %v", v)
		}
	})

	t.Run("reverse", func(t *testing.T) {
		v := callNative(t, mod, "reverse", Array{1.0, 2.0, 3.0}).(Array)
		if v[0].(float64) != 3 {
			t.Fatalf("sort.reverse = %v", v)
		}
	})

	t.Run("unique", func(t *testing.T) {
		v := callNative(t, mod, "unique", Array{1.0, 2.0, 1.0, 3.0, 2.0}).(Array)
		if len(v) != 3 {
			t.Fatalf("sort.unique len = %d, want 3", len(v))
		}
	})

	t.Run("sortBy", func(t *testing.T) {
		arr := Array{
			Object{"name": "b", "age": 30.0},
			Object{"name": "a", "age": 20.0},
		}
		keyFn := NativeFunction(func(args []Value) (Value, error) {
			return args[0].(Object)["age"], nil
		})
		v := callNative(t, mod, "sortBy", arr, keyFn).(Array)
		first := v[0].(Object)
		if first["name"].(string) != "a" {
			t.Fatalf("sort.sortBy = %v", v)
		}
	})
}

// ─── Set Module ─────────────────────────────────────────────────────────────

func TestSetModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "set")

	t.Run("union", func(t *testing.T) {
		v := callNative(t, mod, "union", Array{"a", "b"}, Array{"b", "c"}).(Array)
		if len(v) != 3 {
			t.Fatalf("set.union len = %d, want 3", len(v))
		}
	})

	t.Run("intersect", func(t *testing.T) {
		v := callNative(t, mod, "intersect", Array{"a", "b"}, Array{"b", "c"}).(Array)
		if len(v) != 1 || v[0].(string) != "b" {
			t.Fatalf("set.intersect = %v", v)
		}
	})

	t.Run("diff", func(t *testing.T) {
		v := callNative(t, mod, "diff", Array{"a", "b", "c"}, Array{"b"}).(Array)
		if len(v) != 2 {
			t.Fatalf("set.diff len = %d, want 2", len(v))
		}
	})

	t.Run("has", func(t *testing.T) {
		if !callNative(t, mod, "has", Array{"a", "b"}, "a").(bool) {
			t.Fatal("set.has should be true")
		}
		if callNative(t, mod, "has", Array{"a", "b"}, "z").(bool) {
			t.Fatal("set.has should be false")
		}
	})

	t.Run("symmetricDiff", func(t *testing.T) {
		v := callNative(t, mod, "symmetricDiff", Array{"a", "b"}, Array{"b", "c"}).(Array)
		if len(v) != 2 {
			t.Fatalf("set.symmetricDiff len = %d, want 2", len(v))
		}
	})

	t.Run("fromArray", func(t *testing.T) {
		v := callNative(t, mod, "fromArray", Array{1.0, 2.0, 1.0, 3.0}).(Array)
		if len(v) != 3 {
			t.Fatalf("set.fromArray len = %d, want 3", len(v))
		}
	})

	t.Run("toArray", func(t *testing.T) {
		v := callNative(t, mod, "toArray", Array{"x", "x", "y"}).(Array)
		if len(v) != 2 {
			t.Fatalf("set.toArray len = %d, want 2", len(v))
		}
	})
}

// ─── Array Module ───────────────────────────────────────────────────────────

func TestArrayModuleExtended(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "array")

	t.Run("sort_default", func(t *testing.T) {
		v := callNative(t, mod, "sort", Array{3.0, 1.0, 2.0}).(Array)
		if v[0].(float64) != 1 || v[2].(float64) != 3 {
			t.Fatalf("array.sort = %v", v)
		}
	})

	t.Run("shuffle", func(t *testing.T) {
		bigArr := make(Array, 100)
		for i := range bigArr {
			bigArr[i] = float64(i)
		}
		v := callNative(t, mod, "shuffle", bigArr).(Array)
		if len(v) != 100 {
			t.Fatalf("array.shuffle len = %d", len(v))
		}
		sameCount := 0
		for i := range v {
			if v[i] == bigArr[i] {
				sameCount++
			}
		}
		if sameCount == 100 {
			t.Fatal("array.shuffle did not shuffle")
		}
	})

	t.Run("flat", func(t *testing.T) {
		v := callNative(t, mod, "flat", Array{1.0, Array{2.0, 3.0}, 4.0}, 1.0).(Array)
		if len(v) != 4 {
			t.Fatalf("array.flat = %v", v)
		}
	})

	t.Run("includes", func(t *testing.T) {
		if !callNative(t, mod, "includes", Array{1.0, 2.0, 3.0}, 2.0).(bool) {
			t.Fatal("array.includes should be true")
		}
		if callNative(t, mod, "includes", Array{1.0, 2.0, 3.0}, 5.0).(bool) {
			t.Fatal("array.includes should be false")
		}
	})

	t.Run("indexOf_lastIndexOf", func(t *testing.T) {
		arr := Array{10.0, 20.0, 30.0, 20.0}
		idx := callNative(t, mod, "indexOf", arr, 20.0).(float64)
		if idx != 1 {
			t.Fatalf("array.indexOf = %v, want 1", idx)
		}
		lastIdx := callNative(t, mod, "lastIndexOf", arr, 20.0).(float64)
		if lastIdx != 3 {
			t.Fatalf("array.lastIndexOf = %v, want 3", lastIdx)
		}
	})

	t.Run("reverse", func(t *testing.T) {
		v := callNative(t, mod, "reverse", Array{1.0, 2.0, 3.0}).(Array)
		if v[0].(float64) != 3 || v[2].(float64) != 1 {
			t.Fatalf("array.reverse = %v", v)
		}
	})

	t.Run("fill", func(t *testing.T) {
		v := callNative(t, mod, "fill", Array{0.0, 0.0, 0.0}, 7.0).(Array)
		for _, item := range v {
			if item.(float64) != 7 {
				t.Fatalf("array.fill = %v", v)
			}
		}
	})
}

// ─── Encoding Module ────────────────────────────────────────────────────────

func TestEncodingModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "encoding")

	t.Run("base64_roundtrip", func(t *testing.T) {
		enc := callNative(t, mod, "base64Encode", "hello world").(string)
		dec := callNative(t, mod, "base64Decode", enc).(string)
		if dec != "hello world" {
			t.Fatalf("base64 roundtrip = %v", dec)
		}
	})

	t.Run("url_roundtrip", func(t *testing.T) {
		enc := callNative(t, mod, "urlEncode", "a b&c=1").(string)
		dec := callNative(t, mod, "urlDecode", enc).(string)
		if dec != "a b&c=1" {
			t.Fatalf("url roundtrip = %v", dec)
		}
	})

	t.Run("hex_roundtrip", func(t *testing.T) {
		enc := callNative(t, mod, "hexEncode", "AB").(string)
		dec := callNative(t, mod, "hexDecode", enc).(string)
		if dec != "AB" {
			t.Fatalf("hex roundtrip = %v", dec)
		}
	})
}

// ─── Crypto Module ──────────────────────────────────────────────────────────

func TestCryptoModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "crypto")

	t.Run("sha256", func(t *testing.T) {
		v := callNative(t, mod, "sha256", "abc").(string)
		if v != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
			t.Fatalf("crypto.sha256 = %v", v)
		}
	})

	t.Run("md5", func(t *testing.T) {
		v := callNative(t, mod, "md5", "abc").(string)
		if v != "900150983cd24fb0d6963f7d28e17f72" {
			t.Fatalf("crypto.md5 = %v", v)
		}
	})
}

// ─── Hash Module ────────────────────────────────────────────────────────────

func TestHashModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "hash")

	t.Run("sha1", func(t *testing.T) {
		v := callNative(t, mod, "sha1", "abc").(string)
		if v != "a9993e364706816aba3e25717850c26c9cd0d89d" {
			t.Fatalf("hash.sha1 = %v", v)
		}
	})

	t.Run("sha256_sha512_fnv", func(t *testing.T) {
		sha256Value := callNative(t, mod, "sha256", "abc").(string)
		if sha256Value != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
			t.Fatalf("hash.sha256 = %v", sha256Value)
		}
		sha512Value := callNative(t, mod, "sha512", "abc").(string)
		if sha512Value != "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f" {
			t.Fatalf("hash.sha512 = %v", sha512Value)
		}
		if v := callNative(t, mod, "fnv32a", "abc").(float64); v <= 0 {
			t.Fatalf("hash.fnv32a = %v, want positive", v)
		}
		if v := callNative(t, mod, "fnv64a", "abc").(float64); v <= 0 {
			t.Fatalf("hash.fnv64a = %v, want positive", v)
		}
	})

	t.Run("errors", func(t *testing.T) {
		if _, err := callNativeWithError(mod, "sha1"); err == nil || !strings.Contains(err.Error(), "expects 1 arg") {
			t.Fatalf("hash.sha1 arity error = %v, want expects 1 arg", err)
		}
		if _, err := callNativeWithError(mod, "crc32", Object{}); err == nil || !strings.Contains(err.Error(), "expects string-like value") {
			t.Fatalf("hash.crc32 type error = %v, want string-like value", err)
		}
		if _, err := computeHash("unknown", "abc"); err == nil || !strings.Contains(err.Error(), "unsupported hash algorithm") {
			t.Fatalf("computeHash unsupported error = %v, want unsupported hash algorithm", err)
		}
		if v, err := computeHash("md5", "abc"); err != nil || v != "900150983cd24fb0d6963f7d28e17f72" {
			t.Fatalf("computeHash md5 = %v, %v", v, err)
		}
	})

	t.Run("crc32", func(t *testing.T) {
		v := callNative(t, mod, "crc32", "abc").(float64)
		if v <= 0 {
			t.Fatalf("hash.crc32 = %v", v)
		}
	})
}

// ─── Regexp Module ──────────────────────────────────────────────────────────

func TestRegexpModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "regexp")

	t.Run("test", func(t *testing.T) {
		if !callNative(t, mod, "test", "h.llo", "hello").(bool) {
			t.Fatal("regexp.test should be true")
		}
	})

	t.Run("find", func(t *testing.T) {
		v := callNative(t, mod, "find", `\d+`, "abc123xyz").(string)
		if v != "123" {
			t.Fatalf("regexp.find = %v", v)
		}
	})

	t.Run("findAll", func(t *testing.T) {
		v := callNative(t, mod, "findAll", `\d+`, "a1b22c333", -1.0).(Array)
		if len(v) != 3 {
			t.Fatalf("regexp.findAll len = %d", len(v))
		}
	})

	t.Run("replaceAll", func(t *testing.T) {
		v := callNative(t, mod, "replaceAll", `\s+`, "a  b   c", "-").(string)
		if v != "a-b-c" {
			t.Fatalf("regexp.replaceAll = %v", v)
		}
	})

	t.Run("split", func(t *testing.T) {
		v := callNative(t, mod, "split", `\s+`, "a b  c").(Array)
		if len(v) != 3 {
			t.Fatalf("regexp.split = %v", v)
		}
	})

	t.Run("compile", func(t *testing.T) {
		compiled := mustRuntimeObject(t, callNative(t, mod, "compile", `\d+`), "compiled")
		if !callNative(t, compiled, "test", "abc123").(bool) {
			t.Fatal("compiled regexp should match")
		}
	})
}

// ─── UUID Module ────────────────────────────────────────────────────────────

func TestUUIDModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "uuid")

	t.Run("v4", func(t *testing.T) {
		v := callNative(t, mod, "v4").(string)
		if len(v) != 36 {
			t.Fatalf("uuid.v4 len = %d, want 36", len(v))
		}
	})

	t.Run("isValid", func(t *testing.T) {
		id := callNative(t, mod, "v4").(string)
		if !callNative(t, mod, "isValid", id).(bool) {
			t.Fatal("uuid.isValid should be true for v4")
		}
		if callNative(t, mod, "isValid", "nope").(bool) {
			t.Fatal("uuid.isValid should be false for invalid")
		}
	})
}

// ─── URL Module ─────────────────────────────────────────────────────────────

func TestURLModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "url")

	t.Run("parse", func(t *testing.T) {
		v := callNative(t, mod, "parse", "https://example.com/path?q=1#frag").(Object)
		if v["host"].(string) != "example.com" {
			t.Fatalf("url.parse.host = %v", v["host"])
		}
	})

	t.Run("escape_unescape", func(t *testing.T) {
		enc := callNative(t, mod, "escape", "a b").(string)
		dec := callNative(t, mod, "unescape", enc).(string)
		if dec != "a b" {
			t.Fatalf("url roundtrip = %v", dec)
		}
	})

	t.Run("queryEncode_queryDecode", func(t *testing.T) {
		enc := callNative(t, mod, "queryEncode", Object{"k": "v 1"}).(string)
		dec := callNative(t, mod, "queryDecode", enc).(Object)
		if dec["k"].(string) != "v 1" {
			t.Fatalf("url query roundtrip = %v", dec)
		}
	})
}

// ─── Strconv Module ─────────────────────────────────────────────────────────

func TestStrconvModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "strconv")

	t.Run("atoi_itoa", func(t *testing.T) {
		v := callNative(t, mod, "atoi", "42").(float64)
		if v != 42 {
			t.Fatalf("strconv.atoi = %v", v)
		}
		s := callNative(t, mod, "itoa", 42.0).(string)
		if s != "42" {
			t.Fatalf("strconv.itoa = %v", s)
		}
	})

	t.Run("parseFloat", func(t *testing.T) {
		v := callNative(t, mod, "parseFloat", "3.5").(float64)
		if v != 3.5 {
			t.Fatalf("strconv.parseFloat = %v", v)
		}
	})

	t.Run("formatFloat", func(t *testing.T) {
		v := callNative(t, mod, "formatFloat", 3.5, 2.0).(string)
		if v != "3.50" {
			t.Fatalf("strconv.formatFloat = %v", v)
		}
	})

	t.Run("parseBool_formatBool", func(t *testing.T) {
		if !callNative(t, mod, "parseBool", "true").(bool) {
			t.Fatal("strconv.parseBool should be true")
		}
		if callNative(t, mod, "formatBool", true).(string) != "true" {
			t.Fatal("strconv.formatBool(true)")
		}
	})
}

// ─── CSV Module ─────────────────────────────────────────────────────────────

func TestCSVModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "csv")

	t.Run("stringify_parse", func(t *testing.T) {
		data := Array{
			Array{"name", "age"},
			Array{"alice", 20.0},
		}
		text := callNative(t, mod, "stringify", data).(string)
		parsed := callNative(t, mod, "parse", text).(Array)
		if len(parsed) != 2 {
			t.Fatalf("csv.parse len = %d", len(parsed))
		}
		row := parsed[1].(Array)
		if row[0].(string) != "alice" {
			t.Fatalf("csv row = %v", row)
		}
	})
}

func TestCSVModuleBranches(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "csv")

	text := callNative(t, mod, "stringify", Array{
		Array{"name", "age"},
		Array{"alice", float64(20)},
	}, Object{"delimiter": ";"}).(string)
	if !strings.Contains(text, "name;age") {
		t.Fatalf("csv.stringify custom delimiter = %q, want semicolon separated header", text)
	}
	parsed := callNative(t, mod, "parse", text, Object{"delimiter": ";"}).(Array)
	row := parsed[1].(Array)
	if row[0] != "alice" || row[1] != "20" {
		t.Fatalf("csv.parse custom delimiter row = %v", row)
	}

	if _, err := callNativeWithError(mod, "parse"); err == nil || !strings.Contains(err.Error(), "expects 1-2 args") {
		t.Fatalf("csv.parse arity error = %v, want expects 1-2 args", err)
	}
	if _, err := callNativeWithError(mod, "parse", "a,b", "bad-options"); err == nil || !strings.Contains(err.Error(), "expects object") {
		t.Fatalf("csv.parse options error = %v, want expects object", err)
	}
	if _, err := callNativeWithError(mod, "parse", "a,b", Object{"delimiter": "::"}); err == nil || !strings.Contains(err.Error(), "single rune string") {
		t.Fatalf("csv.parse delimiter error = %v, want single rune string", err)
	}
	if _, err := callNativeWithError(mod, "stringify", Object{}); err == nil || !strings.Contains(err.Error(), "arg[0] expects array") {
		t.Fatalf("csv.stringify rows error = %v, want arg[0] expects array", err)
	}
	if _, err := callNativeWithError(mod, "stringify", Array{"bad-row"}); err == nil || !strings.Contains(err.Error(), "row[0] expects array") {
		t.Fatalf("csv.stringify row error = %v, want row[0] expects array", err)
	}
	if _, err := callNativeWithError(mod, "stringify", Array{Array{Object{}}}); err == nil || !strings.Contains(err.Error(), "expects string-like value") {
		t.Fatalf("csv.stringify cell error = %v, want string-like value", err)
	}
}

// ─── XML Module ─────────────────────────────────────────────────────────────

func TestXMLModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "xml")

	t.Run("parse", func(t *testing.T) {
		v := callNative(t, mod, "parse", `<root id="1"><item>ok</item></root>`).(Object)
		if v["name"].(string) != "root" {
			t.Fatalf("xml.parse name = %v", v["name"])
		}
	})

	t.Run("stringify", func(t *testing.T) {
		node := Object{
			"name": "root", "attrs": Object{}, "text": "", "children": Array{},
		}
		v := callNative(t, mod, "stringify", node).(string)
		if !strings.Contains(v, "<root") {
			t.Fatalf("xml.stringify = %v", v)
		}
	})

	t.Run("valid", func(t *testing.T) {
		if !callNative(t, mod, "valid", `<a><b/></a>`).(bool) {
			t.Fatal("xml.valid should be true")
		}
	})
}

// ─── Hex Module ─────────────────────────────────────────────────────────────

func TestHexModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "hex")

	t.Run("encode_decode", func(t *testing.T) {
		enc := callNative(t, mod, "encode", "Hi").(string)
		if enc != "4869" {
			t.Fatalf("hex.encode = %v", enc)
		}
		dec := callNative(t, mod, "decode", enc).(string)
		if dec != "Hi" {
			t.Fatalf("hex.decode = %v", dec)
		}
	})

	t.Run("decodeBytes", func(t *testing.T) {
		v := callNative(t, mod, "decodeBytes", "4869").(Array)
		if len(v) != 2 {
			t.Fatalf("hex.decodeBytes len = %d", len(v))
		}
	})
}

// ─── Net Module ─────────────────────────────────────────────────────────────

func TestNetModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "net")

	t.Run("isIP", func(t *testing.T) {
		if !callNative(t, mod, "isIP", "127.0.0.1").(bool) {
			t.Fatal("net.isIP should be true")
		}
		if callNative(t, mod, "isIP", "not-an-ip").(bool) {
			t.Fatal("net.isIP invalid input should be false")
		}
		if !callNative(t, mod, "isIPv4", "127.0.0.1").(bool) {
			t.Fatal("net.isIPv4 should be true")
		}
		if callNative(t, mod, "isIPv4", "2001:db8::1").(bool) {
			t.Fatal("net.isIPv4 should be false for ipv6")
		}
		if !callNative(t, mod, "isIPv6", "2001:db8::1").(bool) {
			t.Fatal("net.isIPv6 should be true")
		}
		if callNative(t, mod, "isIPv6", "127.0.0.1").(bool) {
			t.Fatal("net.isIPv6 should be false for ipv4")
		}
	})

	t.Run("joinHostPort_parseHostPort", func(t *testing.T) {
		v := callNative(t, mod, "joinHostPort", "127.0.0.1", 8080.0).(string)
		parsed := callNative(t, mod, "parseHostPort", v).(Object)
		if parsed["host"].(string) != "127.0.0.1" || parsed["port"].(float64) != 8080 {
			t.Fatalf("net.parseHostPort = %v", parsed)
		}

		v6 := callNative(t, mod, "joinHostPort", "2001:db8::1", 443.0).(string)
		parsedV6 := callNative(t, mod, "parseHostPort", v6).(Object)
		if parsedV6["host"].(string) != "2001:db8::1" || parsedV6["port"].(float64) != 443 {
			t.Fatalf("net.parseHostPort ipv6 = %v", parsedV6)
		}

		if _, err := callNativeWithError(mod, "joinHostPort", "127.0.0.1", 70000.0); err == nil || !strings.Contains(err.Error(), "port out of range") {
			t.Fatalf("net.joinHostPort range error = %v, want port out of range", err)
		}
		if _, err := callNativeWithError(mod, "parseHostPort", "bad-addr"); err == nil {
			t.Fatal("net.parseHostPort invalid address should error")
		}
	})

	t.Run("lookup_parseCIDR_containsCIDR", func(t *testing.T) {
		lookup := callNative(t, mod, "lookupIP", "localhost").(Array)
		if len(lookup) == 0 {
			t.Fatal("net.lookupIP localhost should return at least one ip")
		}

		parsed := callNative(t, mod, "parseCIDR", "10.0.0.0/8").(Object)
		if parsed["network"].(string) != "10.0.0.0" || parsed["maskOnes"].(float64) != 8 {
			t.Fatalf("net.parseCIDR = %v", parsed)
		}

		if !callNative(t, mod, "containsCIDR", "10.0.0.0/8", "10.2.3.4").(bool) {
			t.Fatal("net.containsCIDR should be true")
		}
		if callNative(t, mod, "containsCIDR", "10.0.0.0/8", "192.168.1.1").(bool) {
			t.Fatal("net.containsCIDR should be false")
		}
		if _, err := callNativeWithError(mod, "containsCIDR", "10.0.0.0/8", "bad-ip"); err == nil || !strings.Contains(err.Error(), "invalid ip") {
			t.Fatalf("net.containsCIDR invalid ip error = %v, want invalid ip", err)
		}
		if _, err := callNativeWithError(mod, "parseCIDR", "bad-cidr"); err == nil {
			t.Fatal("net.parseCIDR bad cidr should error")
		}
	})
}

// ─── MIME Module ────────────────────────────────────────────────────────────

func TestMIMEModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "mime")

	t.Run("typeByExt", func(t *testing.T) {
		v := callNative(t, mod, "typeByExt", ".json").(string)
		if !strings.Contains(v, "application/json") {
			t.Fatalf("mime.typeByExt = %v", v)
		}

		vNoDot := callNative(t, mod, "typeByExt", "json").(string)
		if !strings.Contains(vNoDot, "application/json") {
			t.Fatalf("mime.typeByExt without dot = %v", vNoDot)
		}

		if v := callNative(t, mod, "typeByExt", ".definitely-unknown-ext"); v != nil {
			t.Fatalf("mime.typeByExt unknown = %#v, want nil", v)
		}
	})

	t.Run("extByType", func(t *testing.T) {
		v := callNative(t, mod, "extByType", "application/json").(Array)
		if len(v) == 0 {
			t.Fatal("mime.extByType should return extensions")
		}
		foundJSON := false
		for _, item := range v {
			if item.(string) == ".json" {
				foundJSON = true
				break
			}
		}
		if !foundJSON {
			t.Fatalf("mime.extByType(application/json) = %#v, want .json", v)
		}

		if v := callNative(t, mod, "extByType", "application/x-definitely-unknown").(Array); len(v) != 0 {
			t.Fatalf("mime.extByType unknown = %#v, want empty", v)
		}
	})

	t.Run("detectType", func(t *testing.T) {
		v := callNative(t, mod, "detectType", "plain text payload").(string)
		if !strings.Contains(v, "text/plain") {
			t.Fatalf("mime.detectType = %v", v)
		}
	})

	t.Run("detectByPath", func(t *testing.T) {
		v := callNative(t, mod, "detectByPath", "a.txt").(string)
		if !strings.Contains(v, "text/plain") {
			t.Fatalf("mime.detectByPath = %v", v)
		}

		if v := callNative(t, mod, "detectByPath", "README"); v != nil {
			t.Fatalf("mime.detectByPath without ext = %#v, want nil", v)
		}
		if v := callNative(t, mod, "detectByPath", "file.definitely-unknown-ext"); v != nil {
			t.Fatalf("mime.detectByPath unknown ext = %#v, want nil", v)
		}
	})

	t.Run("errors", func(t *testing.T) {
		if _, err := callNativeWithError(mod, "typeByExt"); err == nil {
			t.Fatal("mime.typeByExt() expected error")
		}
		if _, err := callNativeWithError(mod, "extByType", Object{}); err == nil {
			t.Fatal("mime.extByType(object) expected error")
		}
		if _, err := callNativeWithError(mod, "detectType", Object{}); err == nil {
			t.Fatal("mime.detectType(object) expected error")
		}
		if _, err := callNativeWithError(mod, "detectByPath", "a", "b"); err == nil {
			t.Fatal("mime.detectByPath with extra arg expected error")
		}
	})
}

// ─── Compress Module ────────────────────────────────────────────────────────

func TestCompressModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "compress")

	t.Run("gzip_roundtrip", func(t *testing.T) {
		compressed := callNative(t, mod, "gzipCompress", "hello", 9.0)
		decompressed := callNative(t, mod, "gzipDecompress", compressed).(string)
		if decompressed != "hello" {
			t.Fatalf("gzip roundtrip = %v", decompressed)
		}
	})

	t.Run("zlib_roundtrip", func(t *testing.T) {
		compressed := callNative(t, mod, "zlibCompress", "world", 1.0)
		decompressed := callNative(t, mod, "zlibDecompress", compressed).(string)
		if decompressed != "world" {
			t.Fatalf("zlib roundtrip = %v", decompressed)
		}
	})

	t.Run("aliases", func(t *testing.T) {
		compressed := callNative(t, mod, "gzip", "alias")
		decompressed := callNative(t, mod, "gunzip", compressed).(string)
		if decompressed != "alias" {
			t.Fatalf("gzip alias = %v", decompressed)
		}
		compressed2 := callNative(t, mod, "deflate", "deflate alias")
		decompressed2 := callNative(t, mod, "inflate", compressed2).(string)
		if decompressed2 != "deflate alias" {
			t.Fatalf("deflate alias = %v", decompressed2)
		}
	})

	t.Run("invalid_level", func(t *testing.T) {
		_, err := callNativeWithError(mod, "gzipCompress", "x", 99.0)
		if err == nil {
			t.Fatal("expected error for invalid level")
		}
	})
}

// ─── HMAC Module ────────────────────────────────────────────────────────────

func TestHMACModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "hmac")

	t.Run("sha_algorithms", func(t *testing.T) {
		sha1Sig := callNative(t, mod, "sha1", "key", "data").(string)
		if len(sha1Sig) != 40 {
			t.Fatalf("hmac.sha1 length = %d, want 40", len(sha1Sig))
		}

		sha256Sig := callNative(t, mod, "sha256", "key", "data").(string)
		if len(sha256Sig) != 64 {
			t.Fatalf("hmac.sha256 length = %d, want 64", len(sha256Sig))
		}

		sha512Sig := callNative(t, mod, "sha512", "key", "data").(string)
		if len(sha512Sig) != 128 {
			t.Fatalf("hmac.sha512 length = %d, want 128", len(sha512Sig))
		}
	})

	t.Run("sha256_sign_verify", func(t *testing.T) {
		sig := callNative(t, mod, "sha256", "key", "data").(string)
		if sig == "" {
			t.Fatal("hmac.sha256 empty")
		}
		if !callNative(t, mod, "verifySha256", "key", "data", sig).(bool) {
			t.Fatal("hmac.verifySha256 should be true")
		}
		if callNative(t, mod, "verifySha256", "key", "wrong", sig).(bool) {
			t.Fatal("hmac.verifySha256 should be false for wrong data")
		}
		if callNative(t, mod, "verifySha256", "key", "data", "not-hex").(bool) {
			t.Fatal("hmac.verifySha256 should be false for invalid hex")
		}
	})

	t.Run("namespace_alias", func(t *testing.T) {
		ns := mustObject(t, mod, "hmac")
		if sig := callNative(t, ns, "sha256", "key", "data").(string); sig == "" {
			t.Fatal("hmac.hmac.sha256 empty")
		}
	})

	t.Run("errors", func(t *testing.T) {
		if _, err := callNativeWithError(mod, "sha1", "key"); err == nil {
			t.Fatal("hmac.sha1 arity expected error")
		}
		if _, err := callNativeWithError(mod, "sha1", Object{}, "data"); err == nil {
			t.Fatal("hmac.sha1 key type expected error")
		}
		if _, err := callNativeWithError(mod, "sha256", "key", Object{}); err == nil {
			t.Fatal("hmac.sha256 data type expected error")
		}
		if _, err := callNativeWithError(mod, "sha512", "key"); err == nil {
			t.Fatal("hmac.sha512 arity expected error")
		}
		if _, err := callNativeWithError(mod, "sha512", Object{}, "data"); err == nil {
			t.Fatal("hmac.sha512 key type expected error")
		}
		if _, err := callNativeWithError(mod, "verifySha256", "key", "data"); err == nil {
			t.Fatal("hmac.verifySha256 arity expected error")
		}
		if _, err := callNativeWithError(mod, "verifySha256", Object{}, "data", "00"); err == nil {
			t.Fatal("hmac.verifySha256 key type expected error")
		}
		if _, err := callNativeWithError(mod, "verifySha256", "key", Object{}, "00"); err == nil {
			t.Fatal("hmac.verifySha256 data type expected error")
		}
		if _, err := callNativeWithError(mod, "verifySha256", "key", "data", Object{}); err == nil {
			t.Fatal("hmac.verifySha256 sig type expected error")
		}
	})
}

// ─── Bytes Module ───────────────────────────────────────────────────────────

func TestBytesModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "bytes")

	t.Run("fromString_toString", func(t *testing.T) {
		arr := callNative(t, mod, "fromString", "Hi").(Array)
		if len(arr) != 2 {
			t.Fatalf("bytes.fromString len = %d", len(arr))
		}
		s := callNative(t, mod, "toString", arr).(string)
		if s != "Hi" {
			t.Fatalf("bytes.toString = %v", s)
		}
	})

	t.Run("base64_roundtrip", func(t *testing.T) {
		arr := callNative(t, mod, "fromString", "Hi").(Array)
		b64 := callNative(t, mod, "toBase64", arr).(string)
		decoded := callNative(t, mod, "fromBase64", b64).(Array)
		s := callNative(t, mod, "toString", decoded).(string)
		if s != "Hi" {
			t.Fatalf("bytes base64 roundtrip = %v", s)
		}
	})

	t.Run("slice", func(t *testing.T) {
		arr := callNative(t, mod, "fromString", "Hello").(Array)
		sliced := callNative(t, mod, "slice", arr, 1.0, 4.0).(Array)
		s := callNative(t, mod, "toString", sliced).(string)
		if s != "ell" {
			t.Fatalf("bytes.slice = %v", s)
		}
	})
}

// ─── YAML/TOML Modules ─────────────────────────────────────────────────────

func TestYAMLModuleExtended(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "yaml")

	t.Run("parse_stringify", func(t *testing.T) {
		v := callNative(t, mod, "parse", "key: value\nnum: 42")
		obj := v.(Object)
		if obj["key"].(string) != "value" {
			t.Fatalf("yaml.parse = %v", obj)
		}
		s := callNative(t, mod, "stringify", obj).(string)
		if !strings.Contains(s, "key") {
			t.Fatalf("yaml.stringify = %v", s)
		}
	})

	t.Run("file_roundtrip", func(t *testing.T) {
		tmp := t.TempDir()
		fp := filepath.Join(tmp, "test.yaml")
		data := Object{"name": "yaml_test"}
		callNative(t, mod, "saveToFile", data, fp)
		loaded := callNative(t, mod, "fromFile", fp).(Object)
		if loaded["name"].(string) != "yaml_test" {
			t.Fatalf("yaml file roundtrip = %v", loaded)
		}
	})
}

func TestTOMLModuleExtended(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "toml")

	t.Run("parse_stringify", func(t *testing.T) {
		v := callNative(t, mod, "parse", "title = \"test\"")
		obj := v.(Object)
		if obj["title"].(string) != "test" {
			t.Fatalf("toml.parse = %v", obj)
		}
		s := callNative(t, mod, "stringify", obj).(string)
		if !strings.Contains(s, "title") {
			t.Fatalf("toml.stringify = %v", s)
		}
	})

	t.Run("file_roundtrip", func(t *testing.T) {
		tmp := t.TempDir()
		fp := filepath.Join(tmp, "test.toml")
		data := Object{"title": "toml_test"}
		callNative(t, mod, "saveToFile", data, fp)
		loaded := callNative(t, mod, "fromFile", fp).(Object)
		if loaded["title"].(string) != "toml_test" {
			t.Fatalf("toml file roundtrip = %v", loaded)
		}
	})
}

// ─── Promise Module ─────────────────────────────────────────────────────────

func TestPromiseModule(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "Promise")

	t.Run("all_with_plain_values", func(t *testing.T) {
		result := awaitValue(t, callNative(t, mod, "all", Array{1.0, 2.0}))
		arr := result.(Array)
		if len(arr) != 2 || arr[0].(float64) != 1 || arr[1].(float64) != 2 {
			t.Fatalf("Promise.all = %v", result)
		}
	})

	t.Run("race", func(t *testing.T) {
		result := awaitValue(t, callNative(t, mod, "race", Array{42.0}))
		if result.(float64) != 42 {
			t.Fatalf("Promise.race = %v", result)
		}
	})

	t.Run("allSettled", func(t *testing.T) {
		result := awaitValue(t, callNative(t, mod, "allSettled", Array{1.0, 2.0}))
		arr := result.(Array)
		if len(arr) != 2 {
			t.Fatalf("Promise.allSettled len = %d", len(arr))
		}
	})
}

// ─── Log Module ─────────────────────────────────────────────────────────────

func TestLogModuleExtended(t *testing.T) {
	mod := mustModuleObject(t, getModules(t), "log")

	t.Run("level_cycle", func(t *testing.T) {
		callNative(t, mod, "setLevel", "debug")
		got := callNative(t, mod, "getLevel").(string)
		if got != "debug" {
			t.Fatalf("log.getLevel = %v", got)
		}
	})

	t.Run("json_mode", func(t *testing.T) {
		callNative(t, mod, "setJSON", true)
		if !callNative(t, mod, "isJSON").(bool) {
			t.Fatal("log.isJSON should be true")
		}
	})

	t.Run("with_scope", func(t *testing.T) {
		scoped := callNative(t, mod, "with", Object{"scope": "test"}).(Object)
		if scoped == nil {
			t.Fatal("log.with returned nil")
		}
	})
}

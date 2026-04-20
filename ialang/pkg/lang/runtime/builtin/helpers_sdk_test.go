package builtin

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	rt "ialang/pkg/lang/runtime"
)

func TestBuiltinHelpersArgumentParsing(t *testing.T) {
	if got, err := asStringArg("fn", []Value{"hello"}, 0); err != nil || got != "hello" {
		t.Fatalf("asStringArg(string) = %q, %v; want hello, nil", got, err)
	}
	if got, err := asStringArg("fn", []Value{float64(12.5)}, 0); err != nil || got != "12.5" {
		t.Fatalf("asStringArg(number) = %q, %v; want 12.5, nil", got, err)
	}
	if _, err := asStringArg("fn", nil, 0); err == nil || !strings.Contains(err.Error(), "arg[0] is missing") {
		t.Fatalf("asStringArg(missing) error = %v, want missing arg error", err)
	}
	if _, err := asStringValue("label", Object{}); err == nil || !strings.Contains(err.Error(), "expects string-like value") {
		t.Fatalf("asStringValue(object) error = %v, want type error", err)
	}

	if got, err := asBoolArg("fn", []Value{true}, 0); err != nil || got != true {
		t.Fatalf("asBoolArg(true) = %v, %v; want true, nil", got, err)
	}
	if _, err := asBoolArg("fn", []Value{"bad"}, 0); err == nil || !strings.Contains(err.Error(), "expects bool") {
		t.Fatalf("asBoolArg(string) error = %v, want type error", err)
	}

	if got, err := asIntArg("fn", []Value{float64(7)}, 0); err != nil || got != 7 {
		t.Fatalf("asIntArg(number) = %d, %v; want 7, nil", got, err)
	}
	if got, err := asIntValue("label", "23"); err != nil || got != 23 {
		t.Fatalf("asIntValue(string) = %d, %v; want 23, nil", got, err)
	}
	if _, err := asIntValue("label", "xx"); err == nil || !strings.Contains(err.Error(), "invalid integer string") {
		t.Fatalf("asIntValue(bad string) error = %v, want parse error", err)
	}
	if _, err := asIntValue("label", true); err == nil || !strings.Contains(err.Error(), "expects number or integer string") {
		t.Fatalf("asIntValue(bool) error = %v, want type error", err)
	}
}

func TestBuiltinHelpersConversionAndEnvironment(t *testing.T) {
	key := "IALANG_HELPERS_TEST_" + strconv.Itoa(os.Getpid())
	defer os.Unsetenv(key)
	if err := os.Setenv(key, "value"); err != nil {
		t.Fatalf("setenv fixture: %v", err)
	}

	headers := headersToObject(http.Header{
		"X-Test":  []string{"a", "b"},
		"X-Other": []string{"solo"},
	})
	if headers["X-Test"] != "a,b" {
		t.Fatalf("headersToObject join = %#v, want a,b", headers["X-Test"])
	}
	if headers["X-Other"] != "solo" {
		t.Fatalf("headersToObject single = %#v, want solo", headers["X-Other"])
	}

	env := envObject()
	if env[key] != "value" {
		t.Fatalf("envObject[%q] = %#v, want value", key, env[key])
	}

	args := argsArray()
	if len(args) == 0 {
		t.Fatal("argsArray() returned empty array")
	}
	if _, ok := args[0].(string); !ok {
		t.Fatalf("argsArray()[0] type = %T, want string", args[0])
	}

	converted := toRuntimeJSONValue(map[string]any{
		"n": 3,
		"a": []any{"x", 2},
	}).(Object)
	if converted["n"] != float64(3) {
		t.Fatalf("toRuntimeJSONValue number = %#v, want 3", converted["n"])
	}
	arr, ok := converted["a"].(Array)
	if !ok || len(arr) != 2 || arr[1] != float64(2) {
		t.Fatalf("toRuntimeJSONValue array = %#v, want converted array", converted["a"])
	}

	original := Object{"name": "demo"}
	cloned := cloneObject(original)
	cloned["name"] = "changed"
	if original["name"] != "demo" {
		t.Fatalf("cloneObject mutated source = %#v, want demo", original["name"])
	}

	if got := toString(nil); got != "null" {
		t.Fatalf("toString(nil) = %q, want null", got)
	}
	if got := toString("x"); got != "x" {
		t.Fatalf("toString(string) = %q, want x", got)
	}
	if got := toString(float64(12.5)); got != "12.5" {
		t.Fatalf("toString(number) = %q, want 12.5", got)
	}
	if got := toString(true); got != "true" {
		t.Fatalf("toString(true) = %q, want true", got)
	}
}

func TestBuiltinHelpersGeneratedNativeFunctions(t *testing.T) {
	key := "IALANG_MAKE_ENV_TEST_" + strconv.Itoa(os.Getpid())
	defer os.Unsetenv(key)

	cwdFn := makeCwdFn("os")
	cwd, err := cwdFn(nil)
	if err != nil {
		t.Fatalf("makeCwdFn() error: %v", err)
	}
	actualCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error: %v", err)
	}
	if cwd != actualCwd {
		t.Fatalf("makeCwdFn() = %#v, want %q", cwd, actualCwd)
	}
	if _, err := cwdFn([]Value{"extra"}); err == nil || !strings.Contains(err.Error(), "expects 0 args") {
		t.Fatalf("makeCwdFn(extra) error = %v, want arity error", err)
	}

	getEnvFn := makeGetEnvFn("os")
	if got, err := getEnvFn([]Value{key}); err != nil || got != nil {
		t.Fatalf("makeGetEnvFn(missing) = %#v, %v; want nil, nil", got, err)
	}
	if _, err := getEnvFn([]Value{}); err == nil || !strings.Contains(err.Error(), "expects 1 arg") {
		t.Fatalf("makeGetEnvFn() error = %v, want arity error", err)
	}

	setEnvFn := makeSetEnvFn("os")
	if got, err := setEnvFn([]Value{key, float64(99)}); err != nil || got != true {
		t.Fatalf("makeSetEnvFn() = %#v, %v; want true, nil", got, err)
	}
	if os.Getenv(key) != "99" {
		t.Fatalf("makeSetEnvFn set value = %q, want 99", os.Getenv(key))
	}
	if _, err := setEnvFn([]Value{key}); err == nil || !strings.Contains(err.Error(), "expects 2 args") {
		t.Fatalf("makeSetEnvFn(short) error = %v, want arity error", err)
	}

	if got, err := getEnvFn([]Value{key}); err != nil || got != "99" {
		t.Fatalf("makeGetEnvFn(existing) = %#v, %v; want 99, nil", got, err)
	}

	envFn := makeEnvFn("os")
	env, err := envFn(nil)
	if err != nil {
		t.Fatalf("makeEnvFn() error: %v", err)
	}
	envObj := mustRuntimeObject(t, env, "makeEnvFn return")
	if envObj[key] != "99" {
		t.Fatalf("makeEnvFn env[%q] = %#v, want 99", key, envObj[key])
	}
	if _, err := envFn([]Value{true}); err == nil || !strings.Contains(err.Error(), "expects 0 args") {
		t.Fatalf("makeEnvFn(extra) error = %v, want arity error", err)
	}
}

func TestBuiltinHelpersLogOSAndArrayBranches(t *testing.T) {
	fields, err := parseLogFieldsArg("log.fields", Object{
		"nested": Object{"x": float64(1)},
		"list":   Array{"a", true, Object{"ok": false}},
		"other":  NativeFunction(func(args []Value) (Value, error) { return nil, nil }),
	})
	if err != nil {
		t.Fatalf("parseLogFieldsArg(object) error: %v", err)
	}
	if len(fields) != 6 {
		t.Fatalf("parseLogFieldsArg len = %d, want 6", len(fields))
	}
	if _, err := parseLogFieldsArg("log.fields", "bad"); err == nil || !strings.Contains(err.Error(), "expects object") {
		t.Fatalf("parseLogFieldsArg(string) error = %v, want type error", err)
	}
	if got, err := parseLogFieldsArg("log.fields", nil); err != nil || got != nil {
		t.Fatalf("parseLogFieldsArg(nil) = %#v, %v; want nil, nil", got, err)
	}
	if got := toLogValue(Array{"x", Object{"n": float64(2)}}); got == nil {
		t.Fatal("toLogValue(array) returned nil")
	}
	if got := toLogValue(NativeFunction(func(args []Value) (Value, error) { return nil, nil })); !strings.Contains(got.(string), "0x") && got.(string) == "" {
		t.Fatalf("toLogValue(function) = %#v, want fallback string", got)
	}

	levelTests := []struct {
		name  string
		level slog.Level
		want  string
	}{
		{"debug", slog.LevelDebug, "debug"},
		{"debug-lower", slog.LevelDebug - 1, "debug"},
		{"info", slog.LevelInfo, "info"},
		{"warn", slog.LevelWarn, "warn"},
		{"error", slog.LevelError, "error"},
		{"error-higher", slog.LevelError + 1, "error"},
	}
	for _, tc := range levelTests {
		if got := levelName(tc.level); got != tc.want {
			t.Fatalf("levelName(%s) = %q, want %q", tc.name, got, tc.want)
		}
	}

	if level, err := parseLogLevelValue("warning"); err != nil || level != slog.LevelWarn {
		t.Fatalf("parseLogLevelValue(warning) = %v, %v; want warn, nil", level, err)
	}
	if level, err := parseLogLevelValue(float64(3)); err != nil || level != slog.Level(3) {
		t.Fatalf("parseLogLevelValue(number) = %v, %v; want 3, nil", level, err)
	}
	if _, err := parseLogLevelValue("verbose"); err == nil || !strings.Contains(err.Error(), "unsupported log level") {
		t.Fatalf("parseLogLevelValue(verbose) error = %v, want unsupported log level", err)
	}
	if _, err := parseLogLevelValue(true); err == nil || !strings.Contains(err.Error(), "expects string or number") {
		t.Fatalf("parseLogLevelValue(bool) error = %v, want type error", err)
	}

	if got, err := userDataDir(); err != nil || got == "" || !filepath.IsAbs(got) {
		t.Fatalf("userDataDir() = %q, %v; want absolute path", got, err)
	}

	if got := compareValues(nil, float64(1)); got != -1 {
		t.Fatalf("compareValues(nil,1) = %d, want -1", got)
	}
	if got := compareValues(float64(2), float64(1)); got != 1 {
		t.Fatalf("compareValues(2,1) = %d, want 1", got)
	}
	if got := compareValues("a", "b"); got != -1 {
		t.Fatalf("compareValues(a,b) = %d, want -1", got)
	}
	if got := compareValues(Object{}, Array{}); got == 0 {
		t.Fatalf("compareValues(Object{}, Array{}) = %d, want non-zero", got)
	}

	valuesEqualTests := []struct {
		name string
		a    Value
		b    Value
		want bool
	}{
		{"nil-nil", nil, nil, true},
		{"nil-value", nil, float64(1), false},
		{"bool-eq", true, true, true},
		{"bool-type-mismatch", true, "true", false},
		{"number-eq", float64(2), float64(2), true},
		{"string-ne", "a", "b", false},
		{"object", Object{}, Object{}, false},
	}
	for _, tc := range valuesEqualTests {
		if got := valuesEqual(tc.a, tc.b); got != tc.want {
			t.Fatalf("valuesEqual(%s) = %v, want %v", tc.name, got, tc.want)
		}
	}

	truthyTests := []struct {
		name string
		in   Value
		want bool
	}{
		{"nil", nil, false},
		{"false", false, false},
		{"zero", float64(0), false},
		{"empty-string", "", false},
		{"object", Object{}, true},
	}
	for _, tc := range truthyTests {
		if got := isTruthyValue(tc.in); got != tc.want {
			t.Fatalf("isTruthyValue(%s) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestBuiltinHelpersTimerBranches(t *testing.T) {
	if err := expectTimerCallback(NativeFunction(func(args []Value) (Value, error) { return true, nil })); err != nil {
		t.Fatalf("expectTimerCallback(native) error: %v", err)
	}
	if err := expectTimerCallback(&UserFunction{Name: "cb"}); err != nil {
		t.Fatalf("expectTimerCallback(userfn) error: %v", err)
	}
	if err := expectTimerCallback("bad"); err == nil || !strings.Contains(err.Error(), "expects function") {
		t.Fatalf("expectTimerCallback(string) error = %v, want type error", err)
	}

	done := make(chan struct{}, 1)
	callTimerCallback(NativeFunction(func(args []Value) (Value, error) {
		select {
		case done <- struct{}{}:
		default:
		}
		return true, nil
	}))

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("callTimerCallback(native) did not execute callback")
	}

	callTimerCallback("ignored")
	callTimerCallback(&UserFunction{Name: "bad-user-fn"})
}

func TestBuiltinHelpersYAMLToValueBranches(t *testing.T) {
	if got := yamlToValue(nil); got != nil {
		t.Fatalf("yamlToValue(nil) = %#v, want nil", got)
	}
	if got := yamlToValue(int32(7)); got != float64(7) {
		t.Fatalf("yamlToValue(int32) = %#v, want 7", got)
	}
	if got := yamlToValue(uint16(9)); got != float64(9) {
		t.Fatalf("yamlToValue(uint16) = %#v, want 9", got)
	}
	if got := yamlToValue(float32(1.5)); got != float64(1.5) {
		t.Fatalf("yamlToValue(float32) = %#v, want 1.5", got)
	}
	if got := yamlToValue([]byte("abc")); got != "abc" {
		t.Fatalf("yamlToValue([]byte) = %#v, want abc", got)
	}

	arr := yamlToValue([]interface{}{int(1), "x", map[interface{}]interface{}{"ok": true}}).(Array)
	if len(arr) != 3 || arr[0] != float64(1) || arr[1] != "x" {
		t.Fatalf("yamlToValue([]interface{}) = %#v, want converted array", arr)
	}
	objFromIface, ok := arr[2].(Object)
	if !ok || objFromIface["ok"] != true {
		t.Fatalf("yamlToValue(map[interface{}]interface{}) in array = %#v, want object with ok=true", arr[2])
	}

	obj := yamlToValue(map[string]interface{}{"name": "demo", "count": int64(2)}).(Object)
	if obj["name"] != "demo" || obj["count"] != float64(2) {
		t.Fatalf("yamlToValue(map[string]interface{}) = %#v, want converted object", obj)
	}

	mixed := yamlToValue(map[interface{}]interface{}{1: "one", "nested": []interface{}{uint8(2)}}).(Object)
	if mixed["1"] != "one" {
		t.Fatalf("yamlToValue(map[interface{}]interface{}) key conversion = %#v, want stringified key", mixed)
	}
	nested, ok := mixed["nested"].(Array)
	if !ok || len(nested) != 1 || nested[0] != float64(2) {
		t.Fatalf("yamlToValue nested array = %#v, want [2]", mixed["nested"])
	}

	if got := yamlToValue(struct{ Name string }{Name: "x"}); got != "{x}" {
		t.Fatalf("yamlToValue(struct) = %#v, want fallback string", got)
	}
}

func TestAgentSDKModule(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "@agent/sdk")

	llm := mustObject(t, mod, "llm")
	tool := mustObject(t, mod, "tool")
	memory := mustObject(t, mod, "memory")

	if got := callNative(t, llm, "chat", "hello"); got != "[mock-llm] hello" {
		t.Fatalf("llm.chat = %#v, want [mock-llm] hello", got)
	}
	async := callNative(t, llm, "chatAsync", "world")
	if got := awaitValue(t, async); got != "[mock-llm-async] world" {
		t.Fatalf("llm.chatAsync = %#v, want [mock-llm-async] world", got)
	}
	if got := callNative(t, tool, "call", "search", Object{"q": "x"}); got != "[mock-tool] called search" {
		t.Fatalf("tool.call = %#v, want [mock-tool] called search", got)
	}
	if got := callNative(t, memory, "get", "profile"); got != "[mock-memory] profile" {
		t.Fatalf("memory.get = %#v, want [mock-memory] profile", got)
	}

	if _, err := callNativeWithError(llm, "chat"); err == nil || !strings.Contains(err.Error(), "expects 1 arg") {
		t.Fatalf("llm.chat() error = %v, want arity error", err)
	}
	if _, err := callNativeWithError(llm, "chatAsync"); err == nil || !strings.Contains(err.Error(), "expects 1 arg") {
		t.Fatalf("llm.chatAsync() error = %v, want arity error", err)
	}
	if _, err := callNativeWithError(tool, "call"); err == nil || !strings.Contains(err.Error(), "expects at least 1 arg") {
		t.Fatalf("tool.call() error = %v, want arity error", err)
	}
	if _, err := callNativeWithError(memory, "get"); err == nil || !strings.Contains(err.Error(), "expects 1 arg") {
		t.Fatalf("memory.get() error = %v, want arity error", err)
	}
}

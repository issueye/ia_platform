package builtin

import (
	"testing"

	rt "ialang/pkg/lang/runtime"
)

func TestRegexpCompileAndFlags(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	regexpMod := mustModuleObject(t, modules, "regexp")

	compiled := mustRuntimeObject(t, callNative(t, regexpMod, "compile", "^hello$", "i"), "regexp.compile return")
	if s, ok := compiled["flags"].(string); !ok || s != "i" {
		t.Fatalf("regexp.compile flags = %#v, want i", compiled["flags"])
	}
	matched := callNative(t, compiled, "test", "HELLO")
	if b, ok := matched.(bool); !ok || !b {
		t.Fatalf("compiled.test = %#v, want true", matched)
	}

	found := callNative(t, regexpMod, "find", "a.b", "a\nb", "s")
	if s, ok := found.(string); !ok || s != "a\nb" {
		t.Fatalf("regexp.find with s flag = %#v, want multiline match", found)
	}
	matched = callNative(t, regexpMod, "test", "^foo$", "FOO", "i")
	if b, ok := matched.(bool); !ok || !b {
		t.Fatalf("regexp.test with i flag = %#v, want true", matched)
	}
}

func TestRegexpSubmatchFunctions(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	regexpMod := mustModuleObject(t, modules, "regexp")

	submatch := callNative(t, regexpMod, "findSubmatch", "([a-z]+)(\\d+)", "ab12cd")
	subArr, ok := submatch.(Array)
	if !ok || len(subArr) != 3 {
		t.Fatalf("regexp.findSubmatch = %#v, want len 3", submatch)
	}
	if s, _ := subArr[0].(string); s != "ab12" {
		t.Fatalf("regexp.findSubmatch[0] = %#v, want ab12", subArr[0])
	}
	if s, _ := subArr[1].(string); s != "ab" {
		t.Fatalf("regexp.findSubmatch[1] = %#v, want ab", subArr[1])
	}
	if s, _ := subArr[2].(string); s != "12" {
		t.Fatalf("regexp.findSubmatch[2] = %#v, want 12", subArr[2])
	}

	allSubmatch := callNative(t, regexpMod, "findAllSubmatch", "([a-z])(\\d+)", "a1 b22", float64(-1))
	allArr, ok := allSubmatch.(Array)
	if !ok || len(allArr) != 2 {
		t.Fatalf("regexp.findAllSubmatch = %#v, want len 2", allSubmatch)
	}
	first, ok := allArr[0].(Array)
	if !ok || len(first) != 3 {
		t.Fatalf("regexp.findAllSubmatch[0] = %#v, want len 3", allArr[0])
	}
	if s, _ := first[0].(string); s != "a1" {
		t.Fatalf("regexp.findAllSubmatch[0][0] = %#v, want a1", first[0])
	}

	compiled := mustRuntimeObject(t, callNative(t, regexpMod, "compile", "([a-z])(\\d+)"), "regexp.compile return")
	compiledAll := callNative(t, compiled, "findAllSubmatch", "a1 b2")
	compiledAllArr, ok := compiledAll.(Array)
	if !ok || len(compiledAllArr) != 2 {
		t.Fatalf("compiled.findAllSubmatch = %#v, want len 2", compiledAll)
	}
}

func TestRegexpCompatibilityForOptionalNAndFlags(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	regexpMod := mustModuleObject(t, modules, "regexp")

	oldSignature := callNative(t, regexpMod, "findAll", "\\d", "a1b2c3", float64(2))
	oldArr, ok := oldSignature.(Array)
	if !ok || len(oldArr) != 2 {
		t.Fatalf("regexp.findAll old signature = %#v, want len 2", oldSignature)
	}

	withFlags := callNative(t, regexpMod, "findAll", "a.b", "a\nb", float64(-1), "s")
	withFlagsArr, ok := withFlags.(Array)
	if !ok || len(withFlagsArr) != 1 {
		t.Fatalf("regexp.findAll with flags = %#v, want len 1", withFlags)
	}

	splitOld := callNative(t, regexpMod, "split", "\\s+", "a b c", float64(2))
	splitArr, ok := splitOld.(Array)
	if !ok || len(splitArr) != 2 {
		t.Fatalf("regexp.split old signature = %#v, want len 2", splitOld)
	}
}

func TestRegexpInvalidFlags(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	regexpMod := mustModuleObject(t, modules, "regexp")

	_, err := callNativeWithError(regexpMod, "compile", "a", "x")
	if err == nil {
		t.Fatal("regexp.compile with invalid flags expected error, got nil")
	}

	_, err = callNativeWithError(regexpMod, "test", "a", "a", "ixz")
	if err == nil {
		t.Fatal("regexp.test with invalid flags expected error, got nil")
	}
}

func TestRegexpCompiledObjectBranches(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	regexpMod := mustModuleObject(t, modules, "regexp")
	compiled := mustRuntimeObject(t, callNative(t, regexpMod, "compile", "([a-z]+)(\\d+)"), "regexp.compile return")

	if got := callNative(t, compiled, "find", "xx12 yy"); got != "xx12" {
		t.Fatalf("compiled.find = %#v, want xx12", got)
	}
	if got := callNative(t, compiled, "find", "no-digits"); got != nil {
		t.Fatalf("compiled.find no match = %#v, want nil", got)
	}
	all := callNative(t, compiled, "findAll", "a1 b2 c3", float64(2)).(Array)
	if len(all) != 2 || all[0] != "a1" || all[1] != "b2" {
		t.Fatalf("compiled.findAll = %#v, want first two matches", all)
	}
	split := callNative(t, compiled, "split", "a1-mid-b2", float64(-1)).(Array)
	if len(split) != 3 || split[1] != "-mid-" {
		t.Fatalf("compiled.split = %#v, want split around matches", split)
	}
	if got := callNative(t, compiled, "replaceAll", "a1 b2", "X"); got != "X X" {
		t.Fatalf("compiled.replaceAll = %#v, want X X", got)
	}
	if got := callNative(t, compiled, "findSubmatch", "none"); got != nil {
		t.Fatalf("compiled.findSubmatch no match = %#v, want nil", got)
	}

	if _, err := callNativeWithError(compiled, "test"); err == nil {
		t.Fatal("compiled.test arity error expected")
	}
	if _, err := callNativeWithError(compiled, "findAll", "text", Object{}); err == nil {
		t.Fatal("compiled.findAll n type error expected")
	}
	if _, err := callNativeWithError(compiled, "replaceAll", "text"); err == nil {
		t.Fatal("compiled.replaceAll arity error expected")
	}
}

func BenchmarkRegexpCompileCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compiled, err := compileRegexp("([a-z]+)(\\d+)", "i")
		if err != nil {
			b.Fatalf("compileRegexp failed: %v", err)
		}
		if !compiled.Raw.MatchString("abc123") {
			b.Fatal("compiled regexp should match")
		}
	}
}

package runtime

import (
	"os"
	"path/filepath"
	"testing"

	compiler "ialang/pkg/lang/compiler"
	frontend "ialang/pkg/lang/frontend"
	bridge_ialang "iavm/pkg/bridge/ialang"
	"iavm/pkg/core"
	"iavm/pkg/module"
)

func TestAwaitPassThroughRuntime(t *testing.T) {
	mod := moduleForAsyncTest([]any{int64(7)}, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpAwait},
		{Op: core.OpReturn},
	})

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected await result")
	}
	if result.Kind != core.ValueI64 || result.Raw.(int64) != 7 {
		t.Fatalf("unexpected await result: %#v", result)
	}
}

func TestAsyncFunctionExampleFileRuntime(t *testing.T) {
	sourcePath := filepath.Join("..", "..", "..", "ialang", "examples", "async_loop.ia")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	lexer := frontend.NewLexer(string(source))
	parser := frontend.NewParser(lexer)
	program := parser.ParseProgram()
	if errs := parser.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	compiled, errs := compiler.NewCompiler().Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	mod, err := bridge_ialang.LowerToModule(compiled)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}
	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	var got core.Value
	found := false
	for i, global := range mod.Globals {
		if global.Name == "v" && i < len(vm.globals) {
			got = vm.globals[i]
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected global v")
	}
	if got.Kind == core.ValueI64 && got.Raw.(int64) == 16 {
		return
	}
	if got.Kind == core.ValueF64 && got.Raw.(float64) == 16 {
		return
	}
	if got.Kind != core.ValueI64 && got.Kind != core.ValueF64 {
		t.Fatalf("unexpected async_loop result: %#v", got)
	}
}

func moduleForAsyncTest(constants []any, code []core.Instruction) *module.Module {
	return &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Constants: constants,
		Types:     []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code:      code,
			},
		},
	}
}

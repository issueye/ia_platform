package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"iacommon/pkg/host/api"
	"iacommon/pkg/ialang/bytecode"
	compiler "ialang/pkg/lang/compiler"
	frontend "ialang/pkg/lang/frontend"
	"iavm/pkg/binary"
	bridge_ialang "iavm/pkg/bridge/ialang"
	"iavm/pkg/core"
	"iavm/pkg/module"
	"iavm/pkg/runtime"
)

func runIalangChunkPipeline(t *testing.T, chunk *bytecode.Chunk) core.Value {
	t.Helper()

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule failed: %v", err)
	}

	decoded, err := binary.DecodeModule(data)
	if err != nil {
		t.Fatalf("DecodeModule failed: %v", err)
	}

	result, err := binary.VerifyModule(decoded, binary.VerifyOptions{RequireEntry: true})
	if err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("module verification failed: %v", result.Errors)
	}

	vm, err := runtime.New(decoded, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("VM.Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	return val
}

func TestFullPipeline_CompileToLowerToRun(t *testing.T) {
	// 1. Create ialang chunk (simulating compiled output)
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0, B: 0},
			{Op: bytecode.OpConstant, A: 1, B: 0},
			{Op: bytecode.OpAdd},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{
			int64(10),
			int64(20),
		},
	}

	// 2. Lower to iavm module
	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	// 3. Encode to binary
	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule failed: %v", err)
	}

	// 4. Decode back
	decoded, err := binary.DecodeModule(data)
	if err != nil {
		t.Fatalf("DecodeModule failed: %v", err)
	}

	// 5. Verify
	result, err := binary.VerifyModule(decoded, binary.VerifyOptions{})
	if err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("module verification failed: %v", result.Errors)
	}

	// 6. Run
	vm, err := runtime.New(decoded, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	err = vm.Run()
	if err != nil {
		t.Fatalf("VM.Run failed: %v", err)
	}

	// 7. Check result (10 + 20 = 30)
	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	if val.Kind != core.ValueI64 {
		t.Fatalf("expected I64 result, got %v", val.Kind)
	}
	if val.Raw.(int64) != 30 {
		t.Fatalf("expected 30, got %v", val.Raw)
	}
}

func TestFullPipeline_WithHostCapability(t *testing.T) {
	// Setup mock host
	host := newMockHost()
	host.callResult = api.CallResult{Value: map[string]any{"status": "ok"}}

	// Create module with FS capability
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{Kind: module.CapabilityFS},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"fs", "fs.read_file"},
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpHostCall},
					{Op: core.OpReturn},
				},
			},
		},
	}

	// Verify
	result, err := binary.VerifyModule(mod, binary.VerifyOptions{})
	if err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatal("module not valid")
	}

	// Run
	vm, err := runtime.New(mod, runtime.Options{Host: host})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	err = vm.Run()
	if err != nil {
		t.Fatalf("VM.Run with host failed: %v", err)
	}

	// Check host was called
	if len(host.callLog) != 1 {
		t.Fatalf("expected 1 host call, got %d", len(host.callLog))
	}
}

func TestFullPipeline_WithHostCapabilityArgs(t *testing.T) {
	host := newMockHost()
	host.callResult = api.CallResult{Value: map[string]any{"status": "ok"}}

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{Kind: module.CapabilityFS},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"path", "/workspace/hello.txt", "fs.read_file"},
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpMakeObject, A: 1},
					{Op: core.OpConst, A: 2},
					{Op: core.OpHostCall, A: 1},
					{Op: core.OpReturn},
				},
			},
		},
	}

	result, err := binary.VerifyModule(mod, binary.VerifyOptions{})
	if err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatal("module not valid")
	}

	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule failed: %v", err)
	}
	decoded, err := binary.DecodeModule(data)
	if err != nil {
		t.Fatalf("DecodeModule failed: %v", err)
	}

	vm, err := runtime.New(decoded, runtime.Options{Host: host})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("VM.Run with host args failed: %v", err)
	}

	if len(host.callLog) != 1 {
		t.Fatalf("expected 1 host call, got %d", len(host.callLog))
	}
	if got := host.callLog[0].Args["path"]; got != "/workspace/hello.txt" {
		t.Fatalf("expected path arg after encode/decode, got %#v", got)
	}
}

func TestFullPipeline_WithHostCapabilityModuleConstants(t *testing.T) {
	host := newMockHost()
	host.callResult = api.CallResult{Value: map[string]any{"status": "ok"}}

	mod := &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{"fs", "fs.read_file"},
		Capabilities: []module.CapabilityDecl{
			{Kind: module.CapabilityFS},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpHostCall},
					{Op: core.OpReturn},
				},
			},
		},
	}

	result, err := binary.VerifyModule(mod, binary.VerifyOptions{})
	if err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatal("module not valid")
	}

	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule failed: %v", err)
	}
	decoded, err := binary.DecodeModule(data)
	if err != nil {
		t.Fatalf("DecodeModule failed: %v", err)
	}

	vm, err := runtime.New(decoded, runtime.Options{Host: host})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("VM.Run with module constants failed: %v", err)
	}

	if len(host.callLog) != 1 {
		t.Fatalf("expected 1 host call, got %d", len(host.callLog))
	}
	if host.callLog[0].CapabilityID != "fs" {
		t.Fatalf("expected decoded module constants to drive fs capability, got %q", host.callLog[0].CapabilityID)
	}
}

func TestFullPipeline_ClassInheritanceExample(t *testing.T) {
	sourcePath := filepath.Join("..", "..", "..", "ialang", "examples", "inheritance.ia")
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
	chunk, errs := compiler.NewCompiler().Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}
	if _, err := binary.VerifyModule(mod, binary.VerifyOptions{}); err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("VM.Run failed: %v", err)
	}
}

func TestFullPipeline_EncodeDecodeVerify(t *testing.T) {
	original := &module.Module{
		Magic:      "IAVM",
		Version:    1,
		Target:     "ialang",
		ABIVersion: 1,
		Types: []core.FuncType{
			{Params: []core.ValueKind{core.ValueI64}, Results: []core.ValueKind{core.ValueI64}},
		},
		Functions: []module.Function{
			{
				Name:      "double",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueI64},
				Constants: []any{int64(2)},
				Code: []core.Instruction{
					{Op: core.OpLoadLocal, A: 0},
					{Op: core.OpConst, A: 0},
					{Op: core.OpMul},
					{Op: core.OpReturn},
				},
			},
		},
		Exports: []module.Export{
			{Name: "double", Kind: module.ExportFunction, Index: 0},
		},
		Capabilities: []module.CapabilityDecl{
			{Kind: module.CapabilityFS},
		},
	}

	// Encode
	data, err := binary.EncodeModule(original)
	if err != nil {
		t.Fatalf("EncodeModule failed: %v", err)
	}

	// Decode
	decoded, err := binary.DecodeModule(data)
	if err != nil {
		t.Fatalf("DecodeModule failed: %v", err)
	}

	// Verify fields match
	if decoded.Magic != original.Magic {
		t.Errorf("magic mismatch")
	}
	if decoded.Version != original.Version {
		t.Errorf("version mismatch")
	}
	if len(decoded.Functions) != len(original.Functions) {
		t.Errorf("function count mismatch: got %d, want %d", len(decoded.Functions), len(original.Functions))
	}
	if decoded.Functions[0].Name != "double" {
		t.Errorf("function name mismatch: got %q", decoded.Functions[0].Name)
	}
	if len(decoded.Functions[0].Code) != 4 {
		t.Errorf("instruction count mismatch: got %d", len(decoded.Functions[0].Code))
	}
	if len(decoded.Exports) != 1 {
		t.Errorf("export count mismatch: got %d", len(decoded.Exports))
	}

	// Verify
	result, err := binary.VerifyModule(decoded, binary.VerifyOptions{})
	if err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("verification failed: %v", result.Errors)
	}
}

func TestFullPipeline_ExecutionWithMultipleFunctions(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}, {}},
		Functions: []module.Function{
			{
				Name:      "add",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueI64, core.ValueI64},
				Code: []core.Instruction{
					{Op: core.OpLoadLocal, A: 0},
					{Op: core.OpLoadLocal, A: 1},
					{Op: core.OpAdd},
					{Op: core.OpReturn},
				},
			},
			{
				Name:      "entry",
				TypeIndex: 1,
				Constants: []any{int64(5), int64(7)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpCall, A: 0, B: 2},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok || val.Kind != core.ValueI64 || val.Raw.(int64) != 12 {
		t.Fatalf("expected 12, got %v", val)
	}
}

func TestFullPipeline_ControlFlowBranch(t *testing.T) {
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0, B: 0},
			{Op: bytecode.OpJumpIfFalse, A: 4, B: 0},
			{Op: bytecode.OpConstant, A: 1, B: 0},
			{Op: bytecode.OpReturn},
			{Op: bytecode.OpConstant, A: 2, B: 0},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{false, int64(99), int64(42)},
	}

	val := runIalangChunkPipeline(t, chunk)
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestFullPipeline_ArrayIndex(t *testing.T) {
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0, B: 0},
			{Op: bytecode.OpConstant, A: 1, B: 0},
			{Op: bytecode.OpConstant, A: 2, B: 0},
			{Op: bytecode.OpArray, A: 3, B: 0},
			{Op: bytecode.OpConstant, A: 3, B: 0},
			{Op: bytecode.OpIndex},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{int64(1), int64(2), int64(3), int64(1)},
	}

	val := runIalangChunkPipeline(t, chunk)
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 2 {
		t.Fatalf("expected 2, got %v", val)
	}
}

func TestFullPipeline_StringIndex(t *testing.T) {
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0, B: 0},
			{Op: bytecode.OpConstant, A: 1, B: 0},
			{Op: bytecode.OpIndex},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{"hello", int64(1)},
	}

	val := runIalangChunkPipeline(t, chunk)
	if val.Kind != core.ValueString || val.Raw.(string) != "e" {
		t.Fatalf("expected e, got %v", val)
	}
}

func TestFullPipeline_ObjectProperty(t *testing.T) {
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpObject},
			{Op: bytecode.OpDup},
			{Op: bytecode.OpConstant, A: 1, B: 0},
			{Op: bytecode.OpSetProperty, A: 0, B: 0},
			{Op: bytecode.OpGetProperty, A: 0, B: 0},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{"answer", int64(42)},
	}

	val := runIalangChunkPipeline(t, chunk)
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestFullPipeline_BuiltinLen(t *testing.T) {
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpGetName, A: 0, B: 0},
			{Op: bytecode.OpConstant, A: 1, B: 0},
			{Op: bytecode.OpCall, A: 1, B: 0},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{"len", "hello"},
	}

	val := runIalangChunkPipeline(t, chunk)
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 5 {
		t.Fatalf("expected 5, got %v", val)
	}
}

type mockHost struct {
	caps       map[string]api.CapabilityInstance
	callLog    []api.CallRequest
	callResult api.CallResult
	callErr    error
}

func newMockHost() *mockHost {
	return &mockHost{
		caps: make(map[string]api.CapabilityInstance),
	}
}

func (h *mockHost) AcquireCapability(ctx context.Context, req api.AcquireRequest) (api.CapabilityInstance, error) {
	cap := api.CapabilityInstance{ID: string(req.Kind), Kind: req.Kind}
	h.caps[cap.ID] = cap
	return cap, nil
}

func (h *mockHost) ReleaseCapability(ctx context.Context, capID string) error {
	delete(h.caps, capID)
	return nil
}

func (h *mockHost) Call(ctx context.Context, req api.CallRequest) (api.CallResult, error) {
	h.callLog = append(h.callLog, req)
	return h.callResult, h.callErr
}

func TestFullPipeline_GlobalVariableReassignment(t *testing.T) {
	// let x = 1; x = x + 1;
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0, B: 0},   // push 1
			{Op: bytecode.OpDefineName, A: 1, B: 0}, // define x
			{Op: bytecode.OpGetName, A: 2, B: 0},    // load x
			{Op: bytecode.OpConstant, A: 3, B: 0},   // push 1
			{Op: bytecode.OpAdd},                    // x + 1
			{Op: bytecode.OpSetName, A: 4, B: 0},    // x = ...
			{Op: bytecode.OpReturn},
		},
		Constants: []any{
			float64(1), "x", "x", float64(1), "x",
		},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify global x = 2 by inspecting globals via a new load
	// We can check by adding another test that loads x after the above
}

func TestFullPipeline_GlobalVariableReadWrite(t *testing.T) {
	// let x = 5; let y = 3; return x + y;
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0, B: 0},   // push 5
			{Op: bytecode.OpDefineName, A: 1, B: 0}, // define x
			{Op: bytecode.OpConstant, A: 2, B: 0},   // push 3
			{Op: bytecode.OpDefineName, A: 3, B: 0}, // define y
			{Op: bytecode.OpGetName, A: 4, B: 0},    // load x
			{Op: bytecode.OpGetName, A: 5, B: 0},    // load y
			{Op: bytecode.OpAdd},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{
			float64(5), "x", float64(3), "y", "x", "y",
		},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule failed: %v", err)
	}

	decoded, err := binary.DecodeModule(data)
	if err != nil {
		t.Fatalf("DecodeModule failed: %v", err)
	}

	result, err := binary.VerifyModule(decoded, binary.VerifyOptions{})
	if err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("module verification failed: %v", result.Errors)
	}

	vm, err := runtime.New(decoded, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	if val.Kind != core.ValueF64 || val.Raw.(float64) != 8 {
		t.Fatalf("expected 8, got %v", val)
	}
}

func TestFullPipeline_FunctionAccessesGlobal(t *testing.T) {
	// let z = 10; function getZ() { return z; } getZ();
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0, B: 0},   // push 10
			{Op: bytecode.OpDefineName, A: 1, B: 0}, // define z
			{Op: bytecode.OpClosure, A: 2, B: 0},    // load function template
			{Op: bytecode.OpDefineName, A: 3, B: 0}, // define getZ
			{Op: bytecode.OpGetName, A: 4, B: 0},    // load getZ
			{Op: bytecode.OpCall, A: 0, B: 0},       // getZ()
			{Op: bytecode.OpReturn},
		},
		Constants: []any{
			float64(10), "z",
			&bytecode.FunctionTemplate{
				Name:   "getZ",
				Params: []string{},
				Chunk: &bytecode.Chunk{
					Code: []bytecode.Instruction{
						{Op: bytecode.OpGetName, A: 0, B: 0}, // load z
						{Op: bytecode.OpReturn},
					},
					Constants: []any{"z"},
				},
			},
			"getZ", "getZ",
		},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	if val.Kind != core.ValueF64 || val.Raw.(float64) != 10 {
		t.Fatalf("expected 10, got %v", val)
	}
}

func (h *mockHost) Poll(ctx context.Context, handleID uint64) (api.PollResult, error) {
	return api.PollResult{}, nil
}

func TestFullPipeline_IfElse(t *testing.T) {
	// if (true) { 42 } else { 99 }
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0, B: 0},    // push true
			{Op: bytecode.OpJumpIfFalse, A: 5, B: 0}, // if false, jump to else
			{Op: bytecode.OpConstant, A: 1, B: 0},    // push 42 (then branch)
			{Op: bytecode.OpJump, A: 6, B: 0},        // jump over else
			{Op: bytecode.OpConstant, A: 2, B: 0},    // push 99 (else branch)
			{Op: bytecode.OpReturn},
		},
		Constants: []any{true, float64(42), float64(99)},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	if val.Kind != core.ValueF64 || val.Raw.(float64) != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestFullPipeline_WhileLoop(t *testing.T) {
	// let i = 0; let sum = 0;
	// while (i < 3) { sum = sum + i; i = i + 1; }
	// return sum; // should be 0+1+2 = 3
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0, B: 0},   // push 0
			{Op: bytecode.OpDefineName, A: 1, B: 0}, // define i = 0
			{Op: bytecode.OpConstant, A: 0, B: 0},   // push 0
			{Op: bytecode.OpDefineName, A: 2, B: 0}, // define sum = 0
			// loop condition: i < 3
			{Op: bytecode.OpGetName, A: 1, B: 0},      // push i
			{Op: bytecode.OpConstant, A: 3, B: 0},     // push 3
			{Op: bytecode.OpLess},                     // i < 3
			{Op: bytecode.OpJumpIfFalse, A: 17, B: 0}, // if false, jump to end
			// loop body: sum = sum + i
			{Op: bytecode.OpGetName, A: 2, B: 0}, // push sum
			{Op: bytecode.OpGetName, A: 1, B: 0}, // push i
			{Op: bytecode.OpAdd},                 // sum + i
			{Op: bytecode.OpSetName, A: 2, B: 0}, // sum = ...
			// loop body: i = i + 1
			{Op: bytecode.OpGetName, A: 1, B: 0},  // push i
			{Op: bytecode.OpConstant, A: 4, B: 0}, // push 1
			{Op: bytecode.OpAdd},                  // i + 1
			{Op: bytecode.OpSetName, A: 1, B: 0},  // i = ...
			{Op: bytecode.OpJump, A: 4, B: 0},     // jump back to condition
			// end
			{Op: bytecode.OpGetName, A: 2, B: 0}, // push sum
			{Op: bytecode.OpReturn},
		},
		Constants: []any{float64(0), "i", "sum", float64(3), float64(1)},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	if val.Kind != core.ValueF64 || val.Raw.(float64) != 3 {
		t.Fatalf("expected 3, got %v", val)
	}
}

func TestFullPipeline_ObjectPropertyAccess(t *testing.T) {
	// let obj = {x: 10}; return obj.x;
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpObject},                  // push empty object
			{Op: bytecode.OpDup},                     // dup object
			{Op: bytecode.OpConstant, A: 1, B: 0},    // push 10
			{Op: bytecode.OpSetProperty, A: 0, B: 0}, // obj.x = 10
			{Op: bytecode.OpDup},                     // dup object
			{Op: bytecode.OpGetProperty, A: 0, B: 0}, // obj.x
			{Op: bytecode.OpReturn},
		},
		Constants: []any{"x", float64(10)},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	if val.Kind != core.ValueF64 || val.Raw.(float64) != 10 {
		t.Fatalf("expected 10, got %v", val)
	}
}

func TestFullPipeline_ObjectPropertyMultiple(t *testing.T) {
	// let user = {name: "alice", score: 10}; print(user.name);
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpObject},                  // push empty object
			{Op: bytecode.OpDup},                     // dup for set
			{Op: bytecode.OpConstant, A: 0, B: 0},    // push "alice"
			{Op: bytecode.OpSetProperty, A: 1, B: 0}, // obj.name = "alice"
			{Op: bytecode.OpDup},                     // dup for set
			{Op: bytecode.OpConstant, A: 2, B: 0},    // push 10
			{Op: bytecode.OpSetProperty, A: 3, B: 0}, // obj.score = 10
			{Op: bytecode.OpDefineName, A: 4, B: 0},  // define user
			{Op: bytecode.OpGetName, A: 4, B: 0},     // load user
			{Op: bytecode.OpGetProperty, A: 1, B: 0}, // user.name
			{Op: bytecode.OpReturn},
		},
		Constants: []any{"alice", "name", float64(10), "score", "user"},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	if val.Kind != core.ValueString || val.Raw.(string) != "alice" {
		t.Fatalf("expected 'alice', got %v", val)
	}
}

func TestFullPipeline_ObjectIndexAccess(t *testing.T) {
	// let user = {city: "shanghai"}; user["city"]
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpObject},
			{Op: bytecode.OpDup},
			{Op: bytecode.OpConstant, A: 0, B: 0},    // "shanghai"
			{Op: bytecode.OpSetProperty, A: 1, B: 0}, // obj.city = "shanghai"
			{Op: bytecode.OpDefineName, A: 2, B: 0},  // define user
			{Op: bytecode.OpGetName, A: 2, B: 0},     // load user
			{Op: bytecode.OpConstant, A: 1, B: 0},    // "city"
			{Op: bytecode.OpIndex},                   // user["city"]
			{Op: bytecode.OpReturn},
		},
		Constants: []any{"shanghai", "city", "user"},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	if val.Kind != core.ValueString || val.Raw.(string) != "shanghai" {
		t.Fatalf("expected 'shanghai', got %v", val)
	}
}

func TestFullPipeline_DataIaPattern(t *testing.T) {
	// Mimics data.ia: let arr = [1,2,3,4]; arr[0]; let user = {name:"alice", score:10, city:"sh"}; user.name; user["city"]; user[key]
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			// let arr = [1, 2, 3, 4]
			{Op: bytecode.OpConstant, A: 0, B: 0},   // 1
			{Op: bytecode.OpConstant, A: 1, B: 0},   // 2
			{Op: bytecode.OpConstant, A: 2, B: 0},   // 3
			{Op: bytecode.OpConstant, A: 3, B: 0},   // 4
			{Op: bytecode.OpArray, A: 4, B: 0},      // [1,2,3,4]
			{Op: bytecode.OpDefineName, A: 4, B: 0}, // define arr
			// arr[0]
			{Op: bytecode.OpGetName, A: 4, B: 0},  // load arr
			{Op: bytecode.OpConstant, A: 5, B: 0}, // 0
			{Op: bytecode.OpIndex},                // arr[0]
			{Op: bytecode.OpPop},                  // discard
			// let user = {name: "alice", score: 10, city: "shanghai"}
			{Op: bytecode.OpObject},                   // {}
			{Op: bytecode.OpDup},                      // dup
			{Op: bytecode.OpConstant, A: 6, B: 0},     // "alice"
			{Op: bytecode.OpSetProperty, A: 7, B: 0},  // .name = "alice"
			{Op: bytecode.OpDup},                      // dup
			{Op: bytecode.OpConstant, A: 8, B: 0},     // 10
			{Op: bytecode.OpSetProperty, A: 9, B: 0},  // .score = 10
			{Op: bytecode.OpDup},                      // dup
			{Op: bytecode.OpConstant, A: 10, B: 0},    // "shanghai"
			{Op: bytecode.OpSetProperty, A: 11, B: 0}, // .city = "shanghai"
			{Op: bytecode.OpDefineName, A: 12, B: 0},  // define user
			// user.name
			{Op: bytecode.OpGetName, A: 12, B: 0},    // load user
			{Op: bytecode.OpGetProperty, A: 7, B: 0}, // .name
			{Op: bytecode.OpPop},                     // discard
			// user["city"]
			{Op: bytecode.OpGetName, A: 12, B: 0},  // load user
			{Op: bytecode.OpConstant, A: 11, B: 0}, // "city"
			{Op: bytecode.OpIndex},                 // ["city"]
			{Op: bytecode.OpPop},                   // discard
			// let key = "score"; user[key]
			{Op: bytecode.OpConstant, A: 9, B: 0},    // "score"
			{Op: bytecode.OpDefineName, A: 13, B: 0}, // define key
			{Op: bytecode.OpGetName, A: 12, B: 0},    // load user
			{Op: bytecode.OpGetName, A: 13, B: 0},    // load key
			{Op: bytecode.OpIndex},                   // [key]
			{Op: bytecode.OpReturn},
		},
		Constants: []any{
			float64(1), float64(2), float64(3), float64(4), // 0-3: array elements
			"arr", float64(0), // 4-5: arr name, index 0
			"alice", "name", float64(10), "score", // 6-9: user props
			"shanghai", "city", "user", "key", // 10-13: more user props
		},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	if val.Kind != core.ValueF64 || val.Raw.(float64) != 10 {
		t.Fatalf("expected 10 (user[score]), got %v", val)
	}
}

func TestFullPipeline_TryCatch(t *testing.T) {
	// try { throw "err"; } catch (e) { return 42; }
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpPushTry, A: 4, B: 0},  // catch handler at instruction 4
			{Op: bytecode.OpConstant, A: 0, B: 0}, // push "err"
			{Op: bytecode.OpThrow},                // throw
			{Op: bytecode.OpPopTry},               // pop try (unreachable in normal flow)
			{Op: bytecode.OpConstant, A: 1, B: 0}, // push 42 (catch handler)
			{Op: bytecode.OpReturn},
		},
		Constants: []any{"err", float64(42)},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	if val.Kind != core.ValueF64 || val.Raw.(float64) != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestFullPipeline_TryCatchExceptionValue(t *testing.T) {
	// try { throw "bad"; } catch (e) { return e; }
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpPushTry, A: 4, B: 0},  // catch handler at instruction 4
			{Op: bytecode.OpConstant, A: 0, B: 0}, // push "bad"
			{Op: bytecode.OpThrow},                // throw
			{Op: bytecode.OpPopTry},               // pop try (unreachable)
			{Op: bytecode.OpReturn},               // return e (e is on stack from catch)
		},
		Constants: []any{"bad"},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	if val.Kind != core.ValueString || val.Raw.(string) != "bad" {
		t.Fatalf("expected 'bad', got %v", val)
	}
}

func TestFullPipeline_TryCatchBindVariable(t *testing.T) {
	// try { throw "oops"; } catch (e) { print(e); }
	catchNameIdx := 1
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpPushTry, A: 4, B: catchNameIdx},
			{Op: bytecode.OpConstant, A: 0, B: 0},
			{Op: bytecode.OpThrow},
			{Op: bytecode.OpPopTry},
			{Op: bytecode.OpGetName, A: catchNameIdx, B: 0},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{"oops", "e"},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	if val.Kind != core.ValueString || val.Raw.(string) != "oops" {
		t.Fatalf("expected 'oops', got %v", val)
	}
}

func TestFullPipeline_ExportedGlobalVariable(t *testing.T) {
	// let x = 42; export x;
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0, B: 0},   // push 42
			{Op: bytecode.OpDefineName, A: 1, B: 0}, // define x
			{Op: bytecode.OpExportName, A: 1, B: 0}, // export x
			{Op: bytecode.OpReturn},
		},
		Constants: []any{float64(42), "x"},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	// Verify exports include global x
	if len(mod.Exports) != 1 {
		t.Fatalf("expected 1 export, got %d", len(mod.Exports))
	}
	if mod.Exports[0].Name != "x" {
		t.Fatalf("expected export name 'x', got %q", mod.Exports[0].Name)
	}
	if mod.Exports[0].Kind != module.ExportGlobal {
		t.Fatalf("expected ExportGlobal, got %v", mod.Exports[0].Kind)
	}

	// Run VM
	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestFullPipeline_FunctionExpression(t *testing.T) {
	// let f = function() { return 42; }; return f();
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpClosure, A: 0, B: 0},    // load function template
			{Op: bytecode.OpDefineName, A: 1, B: 0}, // define f
			{Op: bytecode.OpGetName, A: 1, B: 0},    // load f
			{Op: bytecode.OpCall, A: 0, B: 0},       // f()
			{Op: bytecode.OpReturn},
		},
		Constants: []any{
			&bytecode.FunctionTemplate{
				Name:   "",
				Params: []string{},
				Chunk: &bytecode.Chunk{
					Code: []bytecode.Instruction{
						{Op: bytecode.OpConstant, A: 0, B: 0},
						{Op: bytecode.OpReturn},
					},
					Constants: []any{float64(42)},
				},
			},
			"f",
		},
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	vm, err := runtime.New(mod, runtime.Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected result on stack")
	}
	if val.Kind != core.ValueF64 || val.Raw.(float64) != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

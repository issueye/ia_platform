package integration

import (
	"context"
	"testing"

	"iacommon/pkg/host/api"
	"iacommon/pkg/ialang/bytecode"
	bridge_ialang "iavm/pkg/bridge/ialang"
	"iavm/pkg/binary"
	"iavm/pkg/core"
	"iavm/pkg/module"
	"iavm/pkg/runtime"
)

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
			{Op: bytecode.OpConstant, A: 0, B: 0}, // push 1
			{Op: bytecode.OpDefineName, A: 1, B: 0}, // define x
			{Op: bytecode.OpGetName, A: 2, B: 0}, // load x
			{Op: bytecode.OpConstant, A: 3, B: 0}, // push 1
			{Op: bytecode.OpAdd}, // x + 1
			{Op: bytecode.OpSetName, A: 4, B: 0}, // x = ...
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
			{Op: bytecode.OpConstant, A: 0, B: 0}, // push 5
			{Op: bytecode.OpDefineName, A: 1, B: 0}, // define x
			{Op: bytecode.OpConstant, A: 2, B: 0}, // push 3
			{Op: bytecode.OpDefineName, A: 3, B: 0}, // define y
			{Op: bytecode.OpGetName, A: 4, B: 0}, // load x
			{Op: bytecode.OpGetName, A: 5, B: 0}, // load y
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
			{Op: bytecode.OpConstant, A: 0, B: 0}, // push 10
			{Op: bytecode.OpDefineName, A: 1, B: 0}, // define z
			{Op: bytecode.OpClosure, A: 2, B: 0}, // load function template
			{Op: bytecode.OpDefineName, A: 3, B: 0}, // define getZ
			{Op: bytecode.OpGetName, A: 4, B: 0}, // load getZ
			{Op: bytecode.OpCall, A: 0, B: 0}, // getZ()
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

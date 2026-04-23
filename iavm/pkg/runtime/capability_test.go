package runtime

import (
	"context"
	"testing"

	"iacommon/pkg/host/api"
	"iavm/pkg/core"
	"iavm/pkg/module"
)

type mockHost struct {
	acquireLog []api.AcquireRequest
	caps       map[string]api.CapabilityInstance
	callLog    []api.CallRequest
	callResult api.CallResult
	callErr    error
	pollLog    []uint64
	pollResult api.PollResult
	pollErr    error
}

func newMockHost() *mockHost {
	return &mockHost{
		caps: make(map[string]api.CapabilityInstance),
	}
}

func (h *mockHost) AcquireCapability(ctx context.Context, req api.AcquireRequest) (api.CapabilityInstance, error) {
	h.acquireLog = append(h.acquireLog, req)
	cap := api.CapabilityInstance{
		ID:   string(req.Kind),
		Kind: req.Kind,
	}
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

func (h *mockHost) Poll(ctx context.Context, handleID uint64) (api.PollResult, error) {
	h.pollLog = append(h.pollLog, handleID)
	return h.pollResult, h.pollErr
}

func TestCapability_AcquireAndCall(t *testing.T) {
	host := newMockHost()
	host.callResult = api.CallResult{Value: map[string]any{"data": []byte("hello world")}}

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
					{Op: core.OpImportCap, A: 0}, // acquire fs capability
					{Op: core.OpConst, A: 1},     // push operation name
					{Op: core.OpHostCall},        // call host
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Check that host.Call was invoked
	if len(host.callLog) != 1 {
		t.Fatalf("expected 1 host call, got %d", len(host.callLog))
	}
	if host.callLog[0].Operation != "fs.read_file" {
		t.Errorf("expected operation 'fs.read_file', got %q", host.callLog[0].Operation)
	}

	// Result should be on stack (handle from ImportCap + result from HostCall)
	if vm.stack.Size() < 1 {
		t.Fatalf("expected at least 1 item on stack, got %d", vm.stack.Size())
	}
	// Pop the host call result (last pushed)
	val := vm.stack.Pop()
	if val.Kind != core.ValueObjectRef {
		t.Fatalf("expected object result, got %v", val.Kind)
	}
}

func TestCapability_HostCallPassesObjectArgs(t *testing.T) {
	host := newMockHost()
	host.callResult = api.CallResult{Value: map[string]any{"ok": true}}

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
				Constants: []any{"path", "/workspace/demo.txt", "fs.read_file"},
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

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(host.callLog) != 1 {
		t.Fatalf("expected 1 host call, got %d", len(host.callLog))
	}
	if got := host.callLog[0].Args["path"]; got != "/workspace/demo.txt" {
		t.Fatalf("expected path arg to be propagated, got %#v", got)
	}
}

func TestCapability_NoHost(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"fs"},
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err == nil {
		t.Fatal("expected error when no host configured")
	}
}

func TestCapability_HostCallError(t *testing.T) {
	host := newMockHost()
	host.callErr = context.Canceled

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
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

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err == nil {
		t.Fatal("expected error from host call")
	}
}

func TestCapability_ImportCapUsesModuleConstants(t *testing.T) {
	host := newMockHost()
	host.callResult = api.CallResult{Value: map[string]any{"ok": true}}

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

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(host.callLog) != 1 {
		t.Fatalf("expected 1 host call, got %d", len(host.callLog))
	}
	if host.callLog[0].CapabilityID != "fs" {
		t.Fatalf("expected fs capability from module constants, got %q", host.callLog[0].CapabilityID)
	}
}

func TestCapability_HostCallUsesLastImportedCapability(t *testing.T) {
	host := newMockHost()
	host.callResult = api.CallResult{Value: map[string]any{"ok": true}}

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{Kind: module.CapabilityFS},
			{Kind: module.CapabilityNetwork},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"fs", "network", "network.http_fetch"},
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpImportCap, A: 1},
					{Op: core.OpConst, A: 2},
					{Op: core.OpHostCall},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(host.callLog) != 1 {
		t.Fatalf("expected 1 host call, got %d", len(host.callLog))
	}
	if host.callLog[0].CapabilityID != "network" {
		t.Fatalf("expected last imported capability to be used, got %q", host.callLog[0].CapabilityID)
	}
}

func TestCapability_HostPollUsesHandleFromStack(t *testing.T) {
	host := newMockHost()
	host.pollResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true, "data": []byte("ok")},
	}

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Constants: []any{
			int64(7),
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpHostPoll},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(host.pollLog) != 1 || host.pollLog[0] != 7 {
		t.Fatalf("unexpected poll log: %#v", host.pollLog)
	}
	if vm.stack.Size() == 0 {
		t.Fatal("expected poll result on stack")
	}
	result := vm.stack.Pop()
	if result.Kind != core.ValuePromise {
		t.Fatalf("expected promise result, got %v", result.Kind)
	}
}

func TestCapability_ImportCapPassesModuleConfig(t *testing.T) {
	host := newMockHost()

	mod := &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{"fs"},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityFS,
				Config: map[string]any{
					"rights": []string{"read"},
					"preopens": []any{
						map[string]any{
							"virtual_path": "/workspace",
							"real_path":    "C:/tmp/workspace",
							"read_only":    true,
						},
					},
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(host.acquireLog) != 1 {
		t.Fatalf("expected 1 capability acquire, got %d", len(host.acquireLog))
	}
	if got := host.acquireLog[0].Config["rights"]; got == nil {
		t.Fatal("expected rights config to be forwarded")
	}
	preopens, ok := host.acquireLog[0].Config["preopens"].([]any)
	if !ok || len(preopens) != 1 {
		t.Fatalf("expected preopens to be forwarded, got %#v", host.acquireLog[0].Config["preopens"])
	}
}

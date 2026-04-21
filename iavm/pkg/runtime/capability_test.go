package runtime

import (
	"context"
	"testing"

	"iacommon/pkg/host/api"
	"iavm/pkg/core"
	"iavm/pkg/module"
)

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
	return api.PollResult{}, nil
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
					{Op: core.OpImportCap, A: 0},  // acquire fs capability
					{Op: core.OpConst, A: 1},      // push operation name
					{Op: core.OpHostCall},         // call host
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

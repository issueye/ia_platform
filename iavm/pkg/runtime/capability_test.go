package runtime

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"iacommon/pkg/host/api"
	hostnet "iacommon/pkg/host/network"
	"iavm/pkg/core"
	"iavm/pkg/module"
)

type mockHost struct {
	acquireLog            []api.AcquireRequest
	caps                  map[string]api.CapabilityInstance
	callLog               []api.CallRequest
	callResult            api.CallResult
	callErr               error
	callDeadlineFailures  int
	callRetryableFailures int
	pollLog               []uint64
	pollResult            api.PollResult
	pollErr               error
	pollRetryableFailures int
	waitLog               []uint64
	waitResult            api.PollResult
	waitErr               error
	waitRetryableFailures int
	blockAcquire          bool
	blockCall             bool
	blockPoll             bool
	blockWait             bool
	pollDeadlineFailures  int
	waitDeadlineFailures  int
}

type pollOnlyHost struct {
	pollLog     []uint64
	pollResults []api.PollResult
}

func newMockHost() *mockHost {
	return &mockHost{
		caps: make(map[string]api.CapabilityInstance),
	}
}

func (h *mockHost) AcquireCapability(ctx context.Context, req api.AcquireRequest) (api.CapabilityInstance, error) {
	if h.blockAcquire {
		<-ctx.Done()
		return api.CapabilityInstance{}, ctx.Err()
	}
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
	if h.callDeadlineFailures > 0 {
		h.callDeadlineFailures--
		<-ctx.Done()
		return api.CallResult{}, ctx.Err()
	}
	if h.callRetryableFailures > 0 {
		h.callRetryableFailures--
		return api.CallResult{}, api.MarkRetryable(errors.New("temporary host.call failure"))
	}
	if h.blockCall {
		<-ctx.Done()
		return api.CallResult{}, ctx.Err()
	}
	return h.callResult, h.callErr
}

func (h *mockHost) Poll(ctx context.Context, handleID uint64) (api.PollResult, error) {
	h.pollLog = append(h.pollLog, handleID)
	if h.pollDeadlineFailures > 0 {
		h.pollDeadlineFailures--
		<-ctx.Done()
		return api.PollResult{}, ctx.Err()
	}
	if h.pollRetryableFailures > 0 {
		h.pollRetryableFailures--
		return api.PollResult{}, api.MarkRetryable(errors.New("temporary host.poll failure"))
	}
	if h.blockPoll {
		<-ctx.Done()
		return api.PollResult{}, ctx.Err()
	}
	return h.pollResult, h.pollErr
}

func (h *mockHost) Wait(ctx context.Context, handleID uint64) (api.PollResult, error) {
	h.waitLog = append(h.waitLog, handleID)
	if h.waitDeadlineFailures > 0 {
		h.waitDeadlineFailures--
		<-ctx.Done()
		return api.PollResult{}, ctx.Err()
	}
	if h.waitRetryableFailures > 0 {
		h.waitRetryableFailures--
		return api.PollResult{}, api.MarkRetryable(errors.New("temporary host.wait failure"))
	}
	if h.blockWait {
		<-ctx.Done()
		return api.PollResult{}, ctx.Err()
	}
	return h.waitResult, h.waitErr
}

func (h *pollOnlyHost) AcquireCapability(ctx context.Context, req api.AcquireRequest) (api.CapabilityInstance, error) {
	return api.CapabilityInstance{}, nil
}

func (h *pollOnlyHost) ReleaseCapability(ctx context.Context, capID string) error {
	return nil
}

func (h *pollOnlyHost) Call(ctx context.Context, req api.CallRequest) (api.CallResult, error) {
	return api.CallResult{}, nil
}

func (h *pollOnlyHost) Poll(ctx context.Context, handleID uint64) (api.PollResult, error) {
	h.pollLog = append(h.pollLog, handleID)
	if len(h.pollResults) == 0 {
		return api.PollResult{Done: true, Value: map[string]any{"ready": true}}, nil
	}
	result := h.pollResults[0]
	if len(h.pollResults) > 1 {
		h.pollResults = h.pollResults[1:]
	}
	return result, nil
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

func TestCapability_HostCallHonorsHostTimeout(t *testing.T) {
	host := newMockHost()
	host.blockCall = true

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

	vm, err := New(mod, Options{
		Host:        host,
		HostTimeout: 5 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run error = %v, want deadline exceeded", err)
	}
	if len(host.callLog) != 1 {
		t.Fatalf("expected 1 host call attempt, got %d", len(host.callLog))
	}
}

func TestCapability_HostCallUsesCapabilityTimeoutProfile(t *testing.T) {
	host := newMockHost()
	host.blockCall = true

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityFS,
				Config: map[string]any{
					"host_timeout_ms": int64(5),
				},
			},
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

	vm, err := New(mod, Options{
		Host:        host,
		HostTimeout: time.Hour,
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run error = %v, want deadline exceeded", err)
	}
	if len(host.callLog) != 1 {
		t.Fatalf("expected 1 host call attempt, got %d", len(host.callLog))
	}
}

func TestCapability_HostCallRetriesConfiguredSafeOperation(t *testing.T) {
	host := newMockHost()
	host.callDeadlineFailures = 1
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

	vm, err := New(mod, Options{
		Host:         host,
		HostTimeout:  5 * time.Millisecond,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
		RetryCallOps: []string{"fs.read_file"},
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if len(host.callLog) != 2 {
		t.Fatalf("expected 2 host call attempts, got %d", len(host.callLog))
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected host call result on stack")
	}
	if result.Kind != core.ValueObjectRef {
		t.Fatalf("expected object result, got %v", result.Kind)
	}
}

func TestCapability_HostCallDoesNotRetryUnlistedOperation(t *testing.T) {
	host := newMockHost()
	host.callDeadlineFailures = 1
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

	vm, err := New(mod, Options{
		Host:         host,
		HostTimeout:  5 * time.Millisecond,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
		RetryCallOps: []string{"fs.stat"},
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run error = %v, want deadline exceeded", err)
	}
	if len(host.callLog) != 1 {
		t.Fatalf("expected 1 host call attempt, got %d", len(host.callLog))
	}
}

func TestCapability_HostCallUsesCapabilityRetryAllowlist(t *testing.T) {
	host := newMockHost()
	host.callDeadlineFailures = 1
	host.callResult = api.CallResult{Value: map[string]any{"ok": true}}

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityFS,
				Config: map[string]any{
					"host_timeout_ms":  int64(5),
					"retry_count":      int64(1),
					"retry_backoff_ms": int64(1),
					"retry_call_ops":   []any{"fs.read_file"},
				},
			},
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

	vm, err := New(mod, Options{
		Host:        host,
		HostTimeout: time.Hour,
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if len(host.callLog) != 2 {
		t.Fatalf("expected 2 host call attempts, got %d", len(host.callLog))
	}
}

func TestCapability_HostCallRetriesExplicitRetryableError(t *testing.T) {
	host := newMockHost()
	host.callRetryableFailures = 1
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

	vm, err := New(mod, Options{
		Host:         host,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
		RetryCallOps: []string{"fs.read_file"},
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if len(host.callLog) != 2 {
		t.Fatalf("expected 2 host call attempts, got %d", len(host.callLog))
	}
}

func TestCapability_HostCallRetriesConfiguredHTTPStatusWithDefaultHost(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("retry later"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	host := &api.DefaultHost{
		Network: &hostnet.HTTPProvider{
			Policy: hostnet.Policy{AllowSchemes: []string{"http", "https"}},
		},
	}

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"retry_http_statuses": []any{http.StatusServiceUnavailable},
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"network", "url", server.URL, "network.http_fetch", "status"},
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpConst, A: 2},
					{Op: core.OpMakeObject, A: 1},
					{Op: core.OpConst, A: 3},
					{Op: core.OpHostCall, A: 1},
					{Op: core.OpGetProp, A: 4},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{
		Host:         host,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
		RetryCallOps: []string{"network.http_fetch"},
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 http attempts, got %d", attempts)
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected http status result on stack")
	}
	if result.Kind != core.ValueI64 || result.Raw.(int64) != http.StatusOK {
		t.Fatalf("unexpected retry result: %#v", result)
	}
}

func TestCapability_HostCallSkipsConfiguredHTTPStatusRetryForDisallowedMethod(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("retry later"))
	}))
	defer server.Close()

	host := &api.DefaultHost{
		Network: &hostnet.HTTPProvider{
			Policy: hostnet.Policy{AllowSchemes: []string{"http", "https"}},
		},
	}

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"retry_http_statuses": []any{http.StatusServiceUnavailable},
					"retry_http_methods":  []any{http.MethodGet},
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"network", "url", server.URL, "method", http.MethodPost, "network.http_fetch", "status"},
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpConst, A: 2},
					{Op: core.OpConst, A: 3},
					{Op: core.OpConst, A: 4},
					{Op: core.OpMakeObject, A: 2},
					{Op: core.OpConst, A: 5},
					{Op: core.OpHostCall, A: 1},
					{Op: core.OpGetProp, A: 6},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{
		Host:         host,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
		RetryCallOps: []string{"network.http_fetch"},
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 http attempt, got %d", attempts)
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected http status result on stack")
	}
	if result.Kind != core.ValueI64 || result.Raw.(int64) != http.StatusServiceUnavailable {
		t.Fatalf("unexpected post status result: %#v", result)
	}
}

func TestCapability_HostCallRetriesConfiguredHTTPStatusClassWithDefaultHost(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("retry later"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	host := &api.DefaultHost{
		Network: &hostnet.HTTPProvider{
			Policy: hostnet.Policy{AllowSchemes: []string{"http", "https"}},
		},
	}

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"retry_http_status_classes": []any{5},
					"retry_http_methods":        []any{http.MethodGet},
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"network", "url", server.URL, "network.http_fetch", "status"},
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpConst, A: 2},
					{Op: core.OpMakeObject, A: 1},
					{Op: core.OpConst, A: 3},
					{Op: core.OpHostCall, A: 1},
					{Op: core.OpGetProp, A: 4},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{
		Host:         host,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
		RetryCallOps: []string{"network.http_fetch"},
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 http attempts, got %d", attempts)
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected http status result on stack")
	}
	if result.Kind != core.ValueI64 || result.Raw.(int64) != http.StatusOK {
		t.Fatalf("unexpected retry result: %#v", result)
	}
}

func TestCapability_HostCallUsesDefaultSafeHTTPMethodsWhenEnabled(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("retry later"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	host := &api.DefaultHost{
		Network: &hostnet.HTTPProvider{
			Policy: hostnet.Policy{AllowSchemes: []string{"http", "https"}},
		},
	}

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"retry_http_status_classes":       []any{5},
					"retry_http_default_safe_methods": true,
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"network", "url", server.URL, "method", http.MethodDelete, "network.http_fetch", "status"},
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpConst, A: 2},
					{Op: core.OpConst, A: 3},
					{Op: core.OpConst, A: 4},
					{Op: core.OpMakeObject, A: 2},
					{Op: core.OpConst, A: 5},
					{Op: core.OpHostCall, A: 1},
					{Op: core.OpGetProp, A: 6},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{
		Host:         host,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
		RetryCallOps: []string{"network.http_fetch"},
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 http attempts, got %d", attempts)
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected http status result on stack")
	}
	if result.Kind != core.ValueI64 || result.Raw.(int64) != http.StatusOK {
		t.Fatalf("unexpected retry result: %#v", result)
	}
}

func TestCapability_HostCallExcludesConfiguredHTTPStatusFromRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("retry later"))
	}))
	defer server.Close()

	host := &api.DefaultHost{
		Network: &hostnet.HTTPProvider{
			Policy: hostnet.Policy{AllowSchemes: []string{"http", "https"}},
		},
	}

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"retry_http_status_classes":       []any{5},
					"retry_http_excluded_statuses":    []any{http.StatusServiceUnavailable},
					"retry_http_default_safe_methods": true,
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"network", "url", server.URL, "network.http_fetch", "status"},
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpConst, A: 2},
					{Op: core.OpMakeObject, A: 1},
					{Op: core.OpConst, A: 3},
					{Op: core.OpHostCall, A: 1},
					{Op: core.OpGetProp, A: 4},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{
		Host:         host,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
		RetryCallOps: []string{"network.http_fetch"},
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 http attempt, got %d", attempts)
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected http status result on stack")
	}
	if result.Kind != core.ValueI64 || result.Raw.(int64) != http.StatusServiceUnavailable {
		t.Fatalf("unexpected excluded status result: %#v", result)
	}
}

func TestCapability_HostCallExcludesConfiguredHTTPStatusClassFromRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("retry later"))
	}))
	defer server.Close()

	host := &api.DefaultHost{
		Network: &hostnet.HTTPProvider{
			Policy: hostnet.Policy{AllowSchemes: []string{"http", "https"}},
		},
	}

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"retry_http_status_classes":          []any{5},
					"retry_http_excluded_status_classes": []any{5},
					"retry_http_default_safe_methods":    true,
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"network", "url", server.URL, "network.http_fetch", "status"},
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpConst, A: 2},
					{Op: core.OpMakeObject, A: 1},
					{Op: core.OpConst, A: 3},
					{Op: core.OpHostCall, A: 1},
					{Op: core.OpGetProp, A: 4},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{
		Host:         host,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
		RetryCallOps: []string{"network.http_fetch"},
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 http attempt, got %d", attempts)
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected http status result on stack")
	}
	if result.Kind != core.ValueI64 || result.Raw.(int64) != http.StatusBadGateway {
		t.Fatalf("unexpected excluded class result: %#v", result)
	}
}

func TestCapability_HostCallExcludesConfiguredHTTPMethodFromRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("retry later"))
	}))
	defer server.Close()

	host := &api.DefaultHost{
		Network: &hostnet.HTTPProvider{
			Policy: hostnet.Policy{AllowSchemes: []string{"http", "https"}},
		},
	}

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"retry_http_statuses":         []any{http.StatusServiceUnavailable},
					"retry_http_methods":          []any{http.MethodGet},
					"retry_http_excluded_methods": []any{http.MethodGet},
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"network", "url", server.URL, "method", http.MethodGet, "network.http_fetch", "status"},
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpConst, A: 2},
					{Op: core.OpConst, A: 3},
					{Op: core.OpConst, A: 4},
					{Op: core.OpMakeObject, A: 2},
					{Op: core.OpConst, A: 5},
					{Op: core.OpHostCall, A: 1},
					{Op: core.OpGetProp, A: 6},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{
		Host:         host,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
		RetryCallOps: []string{"network.http_fetch"},
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 http attempt, got %d", attempts)
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected http status result on stack")
	}
	if result.Kind != core.ValueI64 || result.Raw.(int64) != http.StatusServiceUnavailable {
		t.Fatalf("unexpected excluded method result: %#v", result)
	}
}

func TestCapability_HostCallHonorsRetryMaxElapsedBudget(t *testing.T) {
	host := newMockHost()
	host.callRetryableFailures = 2

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"retry_count":          int64(2),
					"retry_backoff_ms":     int64(20),
					"retry_max_elapsed_ms": int64(5),
					"retry_call_ops":       []any{"network.http_fetch"},
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"network", "url", "https://example.com", "network.http_fetch"},
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpConst, A: 2},
					{Op: core.OpMakeObject, A: 1},
					{Op: core.OpConst, A: 3},
					{Op: core.OpHostCall, A: 1},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{
		Host:         host,
		RetryCount:   5,
		RetryBackoff: time.Millisecond,
		RetryCallOps: []string{"network.http_fetch"},
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err == nil || !api.IsRetryableError(err) {
		t.Fatalf("Run error = %v, want retryable host.call failure", err)
	}
	if len(host.callLog) != 1 {
		t.Fatalf("expected single call attempt once retry budget is exhausted, got %d", len(host.callLog))
	}
}

func TestCapability_HostCallDoesNotRetryPlainErrorEvenIfAllowlisted(t *testing.T) {
	host := newMockHost()
	host.callErr = errors.New("permanent host.call failure")

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

	vm, err := New(mod, Options{
		Host:         host,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
		RetryCallOps: []string{"fs.read_file"},
	})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err == nil {
		t.Fatal("expected permanent host.call error")
	}
	if len(host.callLog) != 1 {
		t.Fatalf("expected 1 host call attempt, got %d", len(host.callLog))
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

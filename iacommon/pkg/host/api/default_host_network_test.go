package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	hostnet "iacommon/pkg/host/network"
)

type mockSocketHandle struct {
	sendLog [][]byte
	recvBuf []byte
	closed  bool
}

func (h *mockSocketHandle) Send(ctx context.Context, data []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	copied := append([]byte(nil), data...)
	h.sendLog = append(h.sendLog, copied)
	return len(data), nil
}

func (h *mockSocketHandle) Recv(ctx context.Context, size int) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if size <= 0 || size > len(h.recvBuf) {
		size = len(h.recvBuf)
	}
	return append([]byte(nil), h.recvBuf[:size]...), nil
}

func (h *mockSocketHandle) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	h.closed = true
	return nil
}

type mockListenerHandle struct {
	socket *mockSocketHandle
	closed bool
}

func (h *mockListenerHandle) Accept(ctx context.Context) (hostnet.SocketHandle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return h.socket, nil
}

func (h *mockListenerHandle) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	h.closed = true
	return nil
}

type mockNetworkProvider struct {
	dialSocket    *mockSocketHandle
	listener      *mockListenerHandle
	lastEndpoint  hostnet.Endpoint
	lastDialOpts  hostnet.DialOptions
	lastListenOpt hostnet.ListenOptions
}

func (p *mockNetworkProvider) HTTPFetch(ctx context.Context, req hostnet.HTTPRequest) (*hostnet.HTTPResponse, error) {
	_, _ = ctx, req
	return nil, hostnet.ErrNetworkOperationNotSupported
}

func (p *mockNetworkProvider) Dial(ctx context.Context, endpoint hostnet.Endpoint, opts hostnet.DialOptions) (hostnet.SocketHandle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p.lastEndpoint = endpoint
	p.lastDialOpts = opts
	if p.dialSocket == nil {
		p.dialSocket = &mockSocketHandle{}
	}
	return p.dialSocket, nil
}

func (p *mockNetworkProvider) Listen(ctx context.Context, endpoint hostnet.Endpoint, opts hostnet.ListenOptions) (hostnet.ListenerHandle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p.lastEndpoint = endpoint
	p.lastListenOpt = opts
	if p.listener == nil {
		p.listener = &mockListenerHandle{socket: &mockSocketHandle{}}
	}
	return p.listener, nil
}

func TestDefaultHostCallNetworkHTTPFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("X-Test") != "ok" {
			t.Fatalf("unexpected header: %q", r.Header.Get("X-Test"))
		}
		w.Header().Set("Content-Type", "text/plain")
		if got := r.Header.Get("X-Timeout-MS"); got != "25" {
			t.Fatalf("unexpected timeout header: %q", got)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	host := &DefaultHost{Network: &hostnet.HTTPProvider{Policy: hostnet.Policy{AllowSchemes: []string{"http", "https"}}}}
	capability, err := host.AcquireCapability(context.Background(), AcquireRequest{Kind: CapabilityNetwork})
	if err != nil {
		t.Fatalf("acquire network capability: %v", err)
	}

	result, err := host.Call(context.Background(), CallRequest{
		CapabilityID: capability.ID,
		Operation:    "network.http_fetch",
		Args: map[string]any{
			"url":       server.URL,
			"method":    http.MethodPost,
			"headers":   map[string]any{"X-Test": "ok", "X-Timeout-MS": "25"},
			"body":      "ping",
			"timeoutMS": 25,
		},
	})
	if err != nil {
		t.Fatalf("call network http_fetch: %v", err)
	}

	status, ok := result.Value["status"].(int)
	if !ok || status != http.StatusOK {
		t.Fatalf("unexpected status result: %#v", result.Value["status"])
	}
	body, ok := result.Value["body"].([]byte)
	if !ok || string(body) != "ok" {
		t.Fatalf("unexpected body result: %#v", result.Value["body"])
	}
}

func TestDefaultHostRejectsUnknownNetworkOperation(t *testing.T) {
	host := &DefaultHost{Network: &hostnet.HTTPProvider{}}
	capability, err := host.AcquireCapability(context.Background(), AcquireRequest{Kind: CapabilityNetwork})
	if err != nil {
		t.Fatalf("acquire network capability: %v", err)
	}

	_, err = host.Call(context.Background(), CallRequest{
		CapabilityID: capability.ID,
		Operation:    "network.unknown",
	})
	if !errors.Is(err, ErrCapabilityUnsupported) {
		t.Fatalf("expected ErrCapabilityUnsupported, got %v", err)
	}
}

func TestDefaultHostCallNetworkDialSendRecvClose(t *testing.T) {
	provider := &mockNetworkProvider{
		dialSocket: &mockSocketHandle{recvBuf: []byte("pong")},
	}
	host := &DefaultHost{Network: provider}
	capability, err := host.AcquireCapability(context.Background(), AcquireRequest{Kind: CapabilityNetwork})
	if err != nil {
		t.Fatalf("acquire network capability: %v", err)
	}

	dialResult, err := host.Call(context.Background(), CallRequest{
		CapabilityID: capability.ID,
		Operation:    "network.dial",
		Args: map[string]any{
			"network":    "tcp",
			"host":       "example.com",
			"port":       443,
			"timeout_ms": 25,
		},
	})
	if err != nil {
		t.Fatalf("network.dial failed: %v", err)
	}
	handle, ok := dialResult.Value["handle"].(uint64)
	if !ok || handle == 0 {
		t.Fatalf("unexpected dial handle: %#v", dialResult.Value["handle"])
	}

	sendResult, err := host.Call(context.Background(), CallRequest{
		CapabilityID: capability.ID,
		Operation:    "network.send",
		Args: map[string]any{
			"handle": handle,
			"data":   []byte("ping"),
		},
	})
	if err != nil {
		t.Fatalf("network.send failed: %v", err)
	}
	if sent, ok := sendResult.Value["n"].(int64); !ok || sent != 4 {
		t.Fatalf("unexpected send result: %#v", sendResult.Value["n"])
	}

	recvResult, err := host.Call(context.Background(), CallRequest{
		CapabilityID: capability.ID,
		Operation:    "network.recv",
		Args: map[string]any{
			"handle": handle,
			"size":   4,
		},
	})
	if err != nil {
		t.Fatalf("network.recv failed: %v", err)
	}
	body, ok := recvResult.Value["data"].([]byte)
	if !ok || string(body) != "pong" {
		t.Fatalf("unexpected recv data: %#v", recvResult.Value["data"])
	}

	pollResult, err := host.Poll(context.Background(), handle)
	if err != nil {
		t.Fatalf("poll socket handle failed: %v", err)
	}
	if !pollResult.Done {
		t.Fatal("expected socket poll to be ready")
	}

	if _, err := host.Call(context.Background(), CallRequest{
		CapabilityID: capability.ID,
		Operation:    "network.close",
		Args:         map[string]any{"handle": handle},
	}); err != nil {
		t.Fatalf("network.close failed: %v", err)
	}
	if !provider.dialSocket.closed {
		t.Fatal("expected socket to be closed")
	}
}

func TestDefaultHostCallNetworkListenAcceptClose(t *testing.T) {
	listener := &mockListenerHandle{socket: &mockSocketHandle{recvBuf: []byte("ok")}}
	provider := &mockNetworkProvider{listener: listener}
	host := &DefaultHost{Network: provider}
	capability, err := host.AcquireCapability(context.Background(), AcquireRequest{Kind: CapabilityNetwork})
	if err != nil {
		t.Fatalf("acquire network capability: %v", err)
	}

	listenResult, err := host.Call(context.Background(), CallRequest{
		CapabilityID: capability.ID,
		Operation:    "network.listen",
		Args: map[string]any{
			"network": "tcp",
			"host":    "127.0.0.1",
			"port":    9000,
			"backlog": 8,
		},
	})
	if err != nil {
		t.Fatalf("network.listen failed: %v", err)
	}
	listenerHandle, ok := listenResult.Value["handle"].(uint64)
	if !ok || listenerHandle == 0 {
		t.Fatalf("unexpected listener handle: %#v", listenResult.Value["handle"])
	}

	acceptResult, err := host.Call(context.Background(), CallRequest{
		CapabilityID: capability.ID,
		Operation:    "network.accept",
		Args:         map[string]any{"handle": listenerHandle},
	})
	if err != nil {
		t.Fatalf("network.accept failed: %v", err)
	}
	socketHandle, ok := acceptResult.Value["handle"].(uint64)
	if !ok || socketHandle == 0 {
		t.Fatalf("unexpected accepted socket handle: %#v", acceptResult.Value["handle"])
	}

	if _, err := host.Call(context.Background(), CallRequest{
		CapabilityID: capability.ID,
		Operation:    "network.close",
		Args:         map[string]any{"handle": listenerHandle},
	}); err != nil {
		t.Fatalf("listener close failed: %v", err)
	}
	if !listener.closed {
		t.Fatal("expected listener to be closed")
	}

	if _, err := host.Call(context.Background(), CallRequest{
		CapabilityID: capability.ID,
		Operation:    "network.close",
		Args:         map[string]any{"handle": socketHandle},
	}); err != nil {
		t.Fatalf("socket close failed: %v", err)
	}
	if !listener.socket.closed {
		t.Fatal("expected accepted socket to be closed")
	}
}

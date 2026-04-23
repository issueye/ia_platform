package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	hostnet "iacommon/pkg/host/network"
)

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
		Operation:    "network.dial",
	})
	if !errors.Is(err, ErrCapabilityUnsupported) {
		t.Fatalf("expected ErrCapabilityUnsupported, got %v", err)
	}
}

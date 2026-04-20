package network

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPProviderHTTPFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("X-Test") != "ok" {
			t.Fatalf("unexpected header: %q", r.Header.Get("X-Test"))
		}
		_, _ = w.Write([]byte("pong"))
	}))
	defer server.Close()

	provider := &HTTPProvider{Policy: Policy{AllowSchemes: []string{"http", "https"}}}
	resp, err := provider.HTTPFetch(context.Background(), HTTPRequest{
		Method:  http.MethodPost,
		URL:     server.URL,
		Headers: map[string]string{"X-Test": "ok"},
		Body:    []byte("ping"),
	})
	if err != nil {
		t.Fatalf("http fetch failed: %v", err)
	}
	if resp.Status != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.Status)
	}
	if string(resp.Body) != "pong" {
		t.Fatalf("unexpected body: %q", string(resp.Body))
	}
}

func TestHTTPProviderRejectsDisallowedScheme(t *testing.T) {
	provider := &HTTPProvider{Policy: Policy{AllowSchemes: []string{"https"}}}

	_, err := provider.HTTPFetch(context.Background(), HTTPRequest{URL: "http://example.com"})
	if !errors.Is(err, ErrNetworkSchemeNotAllowed) {
		t.Fatalf("expected ErrNetworkSchemeNotAllowed, got %v", err)
	}
}

func TestHTTPProviderRejectsOversizedBody(t *testing.T) {
	provider := &HTTPProvider{Policy: Policy{AllowSchemes: []string{"http"}, MaxBytesPerRequest: 3}}

	_, err := provider.HTTPFetch(context.Background(), HTTPRequest{URL: "http://example.com", Body: []byte("abcd")})
	if !errors.Is(err, ErrNetworkRequestTooLarge) {
		t.Fatalf("expected ErrNetworkRequestTooLarge, got %v", err)
	}
}

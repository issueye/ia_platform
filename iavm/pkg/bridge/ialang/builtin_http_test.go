package ialang

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	hostapi "iavm/pkg/host/api"
	hostnet "iavm/pkg/host/network"
)

func TestBuildPlatformHTTPModuleWithHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte("get-response"))
		case http.MethodPost:
			if r.Header.Get("X-Test") != "ok" {
				t.Fatalf("unexpected header: %q", r.Header.Get("X-Test"))
			}
			_, _ = w.Write([]byte("response-body"))
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	}))
	defer server.Close()

	host := &hostapi.DefaultHost{Network: &hostnet.HTTPProvider{Policy: hostnet.Policy{AllowSchemes: []string{"http", "https"}}}}
	module, err := BuildPlatformHTTPModuleWithHost(host)
	if err != nil {
		t.Fatalf("build http module: %v", err)
	}

	client := module["http"].(map[string]any)["client"].(map[string]any)
	request := client["request"].(HTTPRequestFunc)
	get := client["get"].(HTTPGetFunc)
	post := client["post"].(HTTPPostFunc)

	result, err := request(server.URL, map[string]any{
		"method":    "POST",
		"headers":   map[string]string{"X-Test": "ok"},
		"body":      "request-body",
		"timeoutMS": 2000,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !result["ok"].(bool) {
		t.Fatalf("expected ok result: %#v", result)
	}
	if result["body"].(string) != "response-body" {
		t.Fatalf("unexpected body: %#v", result)
	}

	getResult, err := get(server.URL)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if getResult["body"].(string) != "get-response" {
		t.Fatalf("unexpected get body: %#v", getResult)
	}

	postResult, err := post(server.URL, map[string]any{"headers": map[string]string{"X-Test": "ok"}})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	if postResult["body"].(string) != "response-body" {
		t.Fatalf("unexpected post body: %#v", postResult)
	}
}

func TestBuildPlatformHTTPModuleWithHostPropagatesPolicyError(t *testing.T) {
	host := &hostapi.DefaultHost{Network: &hostnet.HTTPProvider{Policy: hostnet.Policy{AllowSchemes: []string{"https"}}}}
	module, err := BuildPlatformHTTPModuleWithHost(host)
	if err != nil {
		t.Fatalf("build http module: %v", err)
	}

	client := module["http"].(map[string]any)["client"].(map[string]any)
	get := client["get"].(HTTPGetFunc)
	_, err = get("http://example.com")
	if !errors.Is(err, hostnet.ErrNetworkSchemeNotAllowed) {
		t.Fatalf("expected ErrNetworkSchemeNotAllowed, got %v", err)
	}
}

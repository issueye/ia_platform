package builtin

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	rt "ialang/pkg/lang/runtime"
)

func TestHTTPServerProxyMutations(t *testing.T) {
	type upstreamObserved struct {
		Method  string
		Path    string
		Query   string
		HeaderA string
		HeaderB string
		Body    string
	}

	observedCh := make(chan upstreamObserved, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		observedCh <- upstreamObserved{
			Method:  r.Method,
			Path:    r.URL.Path,
			Query:   r.URL.RawQuery,
			HeaderA: r.Header.Get("X-Added"),
			HeaderB: r.Header.Get("X-Remove"),
			Body:    string(raw),
		}
		w.Header().Set("Server", "upstream")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("upstream-body"))
	}))
	defer upstream.Close()

	modules := DefaultModules(rt.NewGoroutineRuntime())
	httpMod := mustModuleObject(t, modules, "http")
	serverNS := mustObject(t, httpMod, "server")

	proxyValue := callNative(t, serverNS, "proxy", Object{
		"addr":   "127.0.0.1:0",
		"target": upstream.URL,
		"requestMutations": Object{
			"method":        "POST",
			"path":          "/rewritten",
			"appendPath":    "/tail",
			"setQuery":      Object{"k": "v"},
			"removeQuery":   Array{"drop"},
			"setHeaders":    Object{"X-Added": "yes"},
			"removeHeaders": Array{"x-remove"},
			"body":          "mutated-request-body",
		},
		"responseMutations": Object{
			"statusCode":    float64(http.StatusAccepted),
			"setHeaders":    Object{"X-Gateway": "ok"},
			"removeHeaders": Array{"server"},
			"body":          "mutated-response-body",
		},
	})
	proxyServer := mustRuntimeObject(t, proxyValue, "http.server.proxy with mutations return")
	addr, _ := proxyServer["addr"].(string)
	if strings.TrimSpace(addr) == "" {
		t.Fatalf("http.server.proxy(mutations) addr = %#v, want non-empty string", proxyServer["addr"])
	}
	defer func() {
		_ = callNative(t, proxyServer, "close")
	}()

	req, err := http.NewRequest(http.MethodGet, "http://"+addr+"/api/original?keep=1&drop=1", strings.NewReader("original"))
	if err != nil {
		t.Fatalf("build proxy request failed: %v", err)
	}
	req.Header.Set("X-Remove", "to-be-removed")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request via proxy failed: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read proxy response body failed: %v", err)
	}

	observed := <-observedCh
	if observed.Method != http.MethodPost {
		t.Fatalf("upstream method = %q, want POST", observed.Method)
	}
	if observed.Path != "/rewritten/tail" {
		t.Fatalf("upstream path = %q, want /rewritten/tail", observed.Path)
	}
	if observed.Query != "k=v&keep=1" && observed.Query != "keep=1&k=v" {
		t.Fatalf("upstream query = %q, want keep=1&k=v", observed.Query)
	}
	if observed.HeaderA != "yes" {
		t.Fatalf("upstream X-Added = %q, want yes", observed.HeaderA)
	}
	if observed.HeaderB != "" {
		t.Fatalf("upstream X-Remove = %q, want empty", observed.HeaderB)
	}
	if observed.Body != "mutated-request-body" {
		t.Fatalf("upstream body = %q, want mutated-request-body", observed.Body)
	}

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("proxy response status = %d, want 202", resp.StatusCode)
	}
	if resp.Header.Get("X-Gateway") != "ok" {
		t.Fatalf("proxy response X-Gateway = %q, want ok", resp.Header.Get("X-Gateway"))
	}
	if resp.Header.Get("Server") != "" {
		t.Fatalf("proxy response Server = %q, want empty", resp.Header.Get("Server"))
	}
	if string(respBody) != "mutated-response-body" {
		t.Fatalf("proxy response body = %q, want mutated-response-body", string(respBody))
	}
}

func TestHTTPServerProxyMutationsValidationErrors(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	httpMod := mustModuleObject(t, modules, "http")
	serverNS := mustObject(t, httpMod, "server")

	_, err := callNativeWithError(serverNS, "proxy", Object{
		"target": "http://example.com",
		"requestMutations": Object{
			"body":       "a",
			"bodyBase64": "Yg==",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("proxy requestMutations body/bodyBase64 err = %v, want mutually exclusive error", err)
	}

	_, err = callNativeWithError(serverNS, "proxy", Object{
		"target": "http://example.com",
		"responseMutations": Object{
			"statusCode": float64(99),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "expects 100-599") {
		t.Fatalf("proxy responseMutations statusCode err = %v, want range error", err)
	}
}

func TestHTTPServerProxyDynamicMutations(t *testing.T) {
	type upstreamObserved struct {
		Method  string
		Path    string
		Query   string
		HeaderA string
		HeaderB string
		Body    string
	}

	observedCh := make(chan upstreamObserved, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		observedCh <- upstreamObserved{
			Method:  r.Method,
			Path:    r.URL.Path,
			Query:   r.URL.RawQuery,
			HeaderA: r.Header.Get("X-Dynamic"),
			HeaderB: r.Header.Get("X-Remove"),
			Body:    string(raw),
		}
		w.Header().Set("Server", "upstream")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("upstream-body"))
	}))
	defer upstream.Close()

	modules := DefaultModules(rt.NewGoroutineRuntime())
	httpMod := mustModuleObject(t, modules, "http")
	serverNS := mustObject(t, httpMod, "server")

	proxyValue := callNative(t, serverNS, "proxy", Object{
		"addr":        "127.0.0.1:0",
		"target":      upstream.URL,
		"stripPrefix": "/api",
		"requestMutations": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("requestMutations callback expects 1 arg")
			}
			return Object{
				"method":        "POST",
				"path":          "/dyn",
				"appendPath":    "/tail",
				"setQuery":      Object{"dyn": "1"},
				"removeQuery":   Array{"drop"},
				"setHeaders":    Object{"X-Dynamic": "yes"},
				"removeHeaders": Array{"x-remove"},
				"body":          "dynamic-request-body",
			}, nil
		}),
		"responseMutations": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("responseMutations callback expects 2 args")
			}
			reqObj, ok := args[1].(Object)
			if !ok {
				return nil, fmt.Errorf("responseMutations second arg expects object")
			}
			path, _ := reqObj["path"].(string)
			return Object{
				"statusCode":    float64(http.StatusNonAuthoritativeInfo),
				"setHeaders":    Object{"X-Resp-Dynamic": "ok"},
				"removeHeaders": Array{"server"},
				"body":          "resp-for-" + path,
			}, nil
		}),
	})
	proxyServer := mustRuntimeObject(t, proxyValue, "http.server.proxy with dynamic mutations return")
	addr, _ := proxyServer["addr"].(string)
	if strings.TrimSpace(addr) == "" {
		t.Fatalf("http.server.proxy(dynamic) addr = %#v, want non-empty string", proxyServer["addr"])
	}
	defer func() {
		_ = callNative(t, proxyServer, "close")
	}()

	req, err := http.NewRequest(http.MethodGet, "http://"+addr+"/api/original?keep=1&drop=1", strings.NewReader("origin"))
	if err != nil {
		t.Fatalf("build proxy request failed: %v", err)
	}
	req.Header.Set("X-Remove", "to-be-removed")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request via proxy failed: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read proxy response body failed: %v", err)
	}

	observed := <-observedCh
	if observed.Method != http.MethodPost {
		t.Fatalf("upstream method = %q, want POST", observed.Method)
	}
	if observed.Path != "/dyn/tail" {
		t.Fatalf("upstream path = %q, want /dyn/tail", observed.Path)
	}
	if observed.Query != "dyn=1&keep=1" && observed.Query != "keep=1&dyn=1" {
		t.Fatalf("upstream query = %q, want keep=1&dyn=1", observed.Query)
	}
	if observed.HeaderA != "yes" {
		t.Fatalf("upstream X-Dynamic = %q, want yes", observed.HeaderA)
	}
	if observed.HeaderB != "" {
		t.Fatalf("upstream X-Remove = %q, want empty", observed.HeaderB)
	}
	if observed.Body != "dynamic-request-body" {
		t.Fatalf("upstream body = %q, want dynamic-request-body", observed.Body)
	}

	if resp.StatusCode != http.StatusNonAuthoritativeInfo {
		t.Fatalf("proxy response status = %d, want 203", resp.StatusCode)
	}
	if resp.Header.Get("X-Resp-Dynamic") != "ok" {
		t.Fatalf("proxy response X-Resp-Dynamic = %q, want ok", resp.Header.Get("X-Resp-Dynamic"))
	}
	if resp.Header.Get("Server") != "" {
		t.Fatalf("proxy response Server = %q, want empty", resp.Header.Get("Server"))
	}
	if string(respBody) != "resp-for-/dyn/tail" {
		t.Fatalf("proxy response body = %q, want resp-for-/dyn/tail", string(respBody))
	}
}

func TestHTTPServerProxyDynamicMutationInvalidReturn(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("upstream-ok"))
	}))
	defer upstream.Close()

	modules := DefaultModules(rt.NewGoroutineRuntime())
	httpMod := mustModuleObject(t, modules, "http")
	serverNS := mustObject(t, httpMod, "server")

	proxyValue := callNative(t, serverNS, "proxy", Object{
		"addr":   "127.0.0.1:0",
		"target": upstream.URL,
		"requestMutations": NativeFunction(func(args []Value) (Value, error) {
			return "invalid", nil
		}),
	})
	proxyServer := mustRuntimeObject(t, proxyValue, "http.server.proxy with invalid callback return")
	addr, _ := proxyServer["addr"].(string)
	defer func() {
		_ = callNative(t, proxyServer, "close")
	}()

	resp, err := http.Get("http://" + addr + "/x")
	if err != nil {
		t.Fatalf("request via proxy failed: %v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("proxy response status = %d, want 502", resp.StatusCode)
	}
	if !strings.Contains(string(raw), "must return object") {
		t.Fatalf("proxy error body = %q, want callback return type error", string(raw))
	}
}

package builtin

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	rt "ialang/pkg/lang/runtime"
)

func TestHTTPServerForwardMutations(t *testing.T) {
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
			HeaderA: r.Header.Get("X-Forward-Added"),
			HeaderB: r.Header.Get("X-Remove"),
			Body:    string(raw),
		}
		w.Header().Set("Server", "upstream")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("upstream-forward-body"))
	}))
	defer upstream.Close()

	modules := DefaultModules(rt.NewGoroutineRuntime())
	httpMod := mustModuleObject(t, modules, "http")
	serverNS := mustObject(t, httpMod, "server")

	forwardValue := callNative(t, serverNS, "forward", Object{
		"addr":   "127.0.0.1:0",
		"target": upstream.URL,
		"requestMutations": Object{
			"method":        "PUT",
			"path":          "/mutated",
			"appendPath":    "/tail",
			"setQuery":      Object{"x": "1"},
			"removeQuery":   Array{"drop"},
			"setHeaders":    Object{"X-Forward-Added": "ok"},
			"removeHeaders": Array{"x-remove"},
			"bodyBase64":    base64.StdEncoding.EncodeToString([]byte("mutated-request-forward")),
		},
		"responseMutations": Object{
			"statusCode":    float64(http.StatusNonAuthoritativeInfo),
			"setHeaders":    Object{"X-Forward-Gateway": "yes"},
			"removeHeaders": Array{"server"},
			"bodyBase64":    base64.StdEncoding.EncodeToString([]byte("mutated-response-forward")),
		},
	})
	forwardServer := mustRuntimeObject(t, forwardValue, "http.server.forward with mutations return")
	addr, _ := forwardServer["addr"].(string)
	if strings.TrimSpace(addr) == "" {
		t.Fatalf("http.server.forward(mutations) addr = %#v, want non-empty string", forwardServer["addr"])
	}
	defer func() {
		_ = callNative(t, forwardServer, "close")
	}()

	req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/origin?keep=2&drop=1", strings.NewReader("original-forward"))
	if err != nil {
		t.Fatalf("build forward request failed: %v", err)
	}
	req.Header.Set("X-Remove", "to-be-removed")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request via forward failed: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read forward response body failed: %v", err)
	}

	observed := <-observedCh
	if observed.Method != http.MethodPut {
		t.Fatalf("upstream method = %q, want PUT", observed.Method)
	}
	if observed.Path != "/mutated/tail" {
		t.Fatalf("upstream path = %q, want /mutated/tail", observed.Path)
	}
	if observed.Query != "keep=2&x=1" && observed.Query != "x=1&keep=2" {
		t.Fatalf("upstream query = %q, want keep=2&x=1", observed.Query)
	}
	if observed.HeaderA != "ok" {
		t.Fatalf("upstream X-Forward-Added = %q, want ok", observed.HeaderA)
	}
	if observed.HeaderB != "" {
		t.Fatalf("upstream X-Remove = %q, want empty", observed.HeaderB)
	}
	if observed.Body != "mutated-request-forward" {
		t.Fatalf("upstream body = %q, want mutated-request-forward", observed.Body)
	}

	if resp.StatusCode != http.StatusNonAuthoritativeInfo {
		t.Fatalf("forward response status = %d, want 203", resp.StatusCode)
	}
	if resp.Header.Get("X-Forward-Gateway") != "yes" {
		t.Fatalf("forward response X-Forward-Gateway = %q, want yes", resp.Header.Get("X-Forward-Gateway"))
	}
	if resp.Header.Get("Server") != "" {
		t.Fatalf("forward response Server = %q, want empty", resp.Header.Get("Server"))
	}
	if string(respBody) != "mutated-response-forward" {
		t.Fatalf("forward response body = %q, want mutated-response-forward", string(respBody))
	}
}

func TestHTTPServerForwardMutationsValidationErrors(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	httpMod := mustModuleObject(t, modules, "http")
	serverNS := mustObject(t, httpMod, "server")

	_, err := callNativeWithError(serverNS, "forward", Object{
		"target": "http://example.com",
		"requestMutations": Object{
			"body":       "a",
			"bodyBase64": "Yg==",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("forward requestMutations body/bodyBase64 err = %v, want mutually exclusive error", err)
	}

	_, err = callNativeWithError(serverNS, "forward", Object{
		"target": "http://example.com",
		"responseMutations": Object{
			"statusCode": float64(600),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "expects 100-599") {
		t.Fatalf("forward responseMutations statusCode err = %v, want range error", err)
	}
}

func TestHTTPServerForwardDynamicMutations(t *testing.T) {
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
			HeaderA: r.Header.Get("X-Forward-Dynamic"),
			HeaderB: r.Header.Get("X-Remove"),
			Body:    string(raw),
		}
		w.Header().Set("Server", "upstream")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("upstream-forward-body"))
	}))
	defer upstream.Close()

	modules := DefaultModules(rt.NewGoroutineRuntime())
	httpMod := mustModuleObject(t, modules, "http")
	serverNS := mustObject(t, httpMod, "server")

	forwardValue := callNative(t, serverNS, "forward", Object{
		"addr":   "127.0.0.1:0",
		"target": upstream.URL,
		"requestMutations": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("requestMutations callback expects 1 arg")
			}
			return Object{
				"method":        "PUT",
				"path":          "/fdyn",
				"appendPath":    "/tail",
				"setQuery":      Object{"dyn": "1"},
				"removeQuery":   Array{"drop"},
				"setHeaders":    Object{"X-Forward-Dynamic": "ok"},
				"removeHeaders": Array{"x-remove"},
				"body":          "dynamic-forward-body",
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
				"statusCode":    float64(http.StatusAccepted),
				"setHeaders":    Object{"X-Forward-Resp": "yes"},
				"removeHeaders": Array{"server"},
				"body":          "dynamic-forward-resp-for-" + path,
			}, nil
		}),
	})
	forwardServer := mustRuntimeObject(t, forwardValue, "http.server.forward with dynamic mutations return")
	addr, _ := forwardServer["addr"].(string)
	if strings.TrimSpace(addr) == "" {
		t.Fatalf("http.server.forward(dynamic) addr = %#v, want non-empty string", forwardServer["addr"])
	}
	defer func() {
		_ = callNative(t, forwardServer, "close")
	}()

	req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/origin?keep=2&drop=1", strings.NewReader("original-forward"))
	if err != nil {
		t.Fatalf("build forward request failed: %v", err)
	}
	req.Header.Set("X-Remove", "to-be-removed")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request via forward failed: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read forward response body failed: %v", err)
	}

	observed := <-observedCh
	if observed.Method != http.MethodPut {
		t.Fatalf("upstream method = %q, want PUT", observed.Method)
	}
	if observed.Path != "/fdyn/tail" {
		t.Fatalf("upstream path = %q, want /fdyn/tail", observed.Path)
	}
	if observed.Query != "dyn=1&keep=2" && observed.Query != "keep=2&dyn=1" {
		t.Fatalf("upstream query = %q, want keep=2&dyn=1", observed.Query)
	}
	if observed.HeaderA != "ok" {
		t.Fatalf("upstream X-Forward-Dynamic = %q, want ok", observed.HeaderA)
	}
	if observed.HeaderB != "" {
		t.Fatalf("upstream X-Remove = %q, want empty", observed.HeaderB)
	}
	if observed.Body != "dynamic-forward-body" {
		t.Fatalf("upstream body = %q, want dynamic-forward-body", observed.Body)
	}

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("forward response status = %d, want 202", resp.StatusCode)
	}
	if resp.Header.Get("X-Forward-Resp") != "yes" {
		t.Fatalf("forward response X-Forward-Resp = %q, want yes", resp.Header.Get("X-Forward-Resp"))
	}
	if resp.Header.Get("Server") != "" {
		t.Fatalf("forward response Server = %q, want empty", resp.Header.Get("Server"))
	}
	if string(respBody) != "dynamic-forward-resp-for-/fdyn/tail" {
		t.Fatalf("forward response body = %q, want dynamic-forward-resp-for-/fdyn/tail", string(respBody))
	}
}

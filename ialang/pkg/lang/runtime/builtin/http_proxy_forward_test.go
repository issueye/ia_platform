package builtin

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	rt "ialang/pkg/lang/runtime"
)

func TestHTTPClientProxyOption(t *testing.T) {
	var proxyHits int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proxy-client" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("proxy-client-ok"))
	}))
	defer upstream.Close()

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&proxyHits, 1)
		outReq, err := http.NewRequestWithContext(r.Context(), r.Method, r.URL.String(), r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		outReq.Header = r.Header.Clone()
		resp, err := http.DefaultTransport.RoundTrip(outReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		for k, vv := range resp.Header {
			for _, hv := range vv {
				w.Header().Add(k, hv)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}))
	defer proxy.Close()

	modules := DefaultModules(rt.NewGoroutineRuntime())
	httpMod := mustModuleObject(t, modules, "http")
	clientNS := mustObject(t, httpMod, "client")

	resp := mustRuntimeObject(t, callNative(t, clientNS, "request", upstream.URL+"/proxy-client", Object{
		"proxy": proxy.URL,
	}), "http.client.request with proxy")

	if code, ok := resp["statusCode"].(float64); !ok || int(code) != http.StatusOK {
		t.Fatalf("http.client.request(proxy) statusCode = %#v, want 200", resp["statusCode"])
	}
	if body, ok := resp["body"].(string); !ok || body != "proxy-client-ok" {
		t.Fatalf("http.client.request(proxy) body = %#v, want proxy-client-ok", resp["body"])
	}
	if atomic.LoadInt32(&proxyHits) == 0 {
		t.Fatal("proxy server was not hit")
	}
}

func TestHTTPServerProxy(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := r.URL.Path + "?" + r.URL.RawQuery + "|" + r.Header.Get("X-From-Proxy")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload))
	}))
	defer upstream.Close()

	modules := DefaultModules(rt.NewGoroutineRuntime())
	httpMod := mustModuleObject(t, modules, "http")
	serverNS := mustObject(t, httpMod, "server")

	proxyValue := callNative(t, serverNS, "proxy", Object{
		"addr":        "127.0.0.1:0",
		"target":      upstream.URL,
		"stripPrefix": "/api",
		"headers": Object{
			"X-From-Proxy": "yes",
		},
	})
	proxyServer := mustRuntimeObject(t, proxyValue, "http.server.proxy return")
	addr, _ := proxyServer["addr"].(string)
	if strings.TrimSpace(addr) == "" {
		t.Fatalf("http.server.proxy addr = %#v, want non-empty string", proxyServer["addr"])
	}
	defer func() {
		_ = callNative(t, proxyServer, "close")
	}()

	resp, err := http.Get("http://" + addr + "/api/orders?id=7")
	if err != nil {
		t.Fatalf("GET via proxy failed: %v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read proxy response failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("proxy status = %d, want 200", resp.StatusCode)
	}
	if string(raw) != "/orders?id=7|yes" {
		t.Fatalf("proxy body = %q, want /orders?id=7|yes", string(raw))
	}
}

func TestHTTPServerForward(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		payload := r.Method + "|" + r.URL.Path + "|" + r.URL.RawQuery + "|" + string(raw) + "|" + r.Header.Get("X-Forwarded-By")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(payload))
	}))
	defer upstream.Close()

	modules := DefaultModules(rt.NewGoroutineRuntime())
	httpMod := mustModuleObject(t, modules, "http")
	serverNS := mustObject(t, httpMod, "server")

	forwardValue := awaitValue(t, callNative(t, serverNS, "forwardAsync", Object{
		"addr":     "127.0.0.1:0",
		"target":   upstream.URL,
		"keepPath": false,
		"path":     "/fixed",
		"headers": Object{
			"X-Forwarded-By": "ialang",
		},
	}))
	forwardServer := mustRuntimeObject(t, forwardValue, "http.server.forwardAsync return")
	addr, _ := forwardServer["addr"].(string)
	if strings.TrimSpace(addr) == "" {
		t.Fatalf("http.server.forward addr = %#v, want non-empty string", forwardServer["addr"])
	}
	defer func() {
		_ = callNative(t, forwardServer, "close")
	}()

	req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/ignore?trace=1", strings.NewReader("abc"))
	if err != nil {
		t.Fatalf("build forward request failed: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST via forward failed: %v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read forward response failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("forward status = %d, want 201", resp.StatusCode)
	}
	if string(raw) != "POST|/fixed|trace=1|abc|ialang" {
		t.Fatalf("forward body = %q, want POST|/fixed|trace=1|abc|ialang", string(raw))
	}
}

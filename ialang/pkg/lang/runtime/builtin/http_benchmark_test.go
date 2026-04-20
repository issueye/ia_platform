package builtin

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	rt "ialang/pkg/lang/runtime"
)

func BenchmarkHTTPServerProxyPipeline(b *testing.B) {
	benchmarkHTTPServerPipeline(
		b,
		"proxy",
		"proxy",
		"/api/bench?keep=1&drop=1",
		func(target string) Object {
			return Object{
				"addr":        "127.0.0.1:0",
				"target":      target,
				"stripPrefix": "/api",
			}
		},
		http.StatusOK,
	)

	benchmarkHTTPServerPipeline(
		b,
		"proxy",
		"proxy_static_mutations",
		"/api/bench?keep=1&drop=1",
		func(target string) Object {
			return Object{
				"addr":        "127.0.0.1:0",
				"target":      target,
				"stripPrefix": "/api",
				"requestMutations": Object{
					"method":      "POST",
					"path":        "/bench/static",
					"setQuery":    Object{"mode": "static"},
					"removeQuery": Array{"drop"},
					"setHeaders":  Object{"X-Bench-Mode": "static"},
				},
				"responseMutations": Object{
					"statusCode": float64(http.StatusCreated),
					"setHeaders": Object{"X-Bench-Resp": "static"},
				},
			}
		},
		http.StatusCreated,
	)

	benchmarkHTTPServerPipeline(
		b,
		"proxy",
		"proxy_dynamic_mutations",
		"/api/bench?keep=1&drop=1",
		func(target string) Object {
			return Object{
				"addr":        "127.0.0.1:0",
				"target":      target,
				"stripPrefix": "/api",
				"requestMutations": NativeFunction(func(args []Value) (Value, error) {
					return Object{
						"method":      "POST",
						"path":        "/bench/dynamic",
						"setQuery":    Object{"mode": "dynamic"},
						"removeQuery": Array{"drop"},
						"setHeaders":  Object{"X-Bench-Mode": "dynamic"},
					}, nil
				}),
				"responseMutations": NativeFunction(func(args []Value) (Value, error) {
					return Object{
						"statusCode": float64(http.StatusAccepted),
						"setHeaders": Object{"X-Bench-Resp": "dynamic"},
					}, nil
				}),
			}
		},
		http.StatusAccepted,
	)
}

func BenchmarkHTTPServerForwardPipeline(b *testing.B) {
	benchmarkHTTPServerPipeline(
		b,
		"forward",
		"forward",
		"/bench?keep=1&drop=1",
		func(target string) Object {
			return Object{
				"addr":   "127.0.0.1:0",
				"target": target,
			}
		},
		http.StatusOK,
	)

	benchmarkHTTPServerPipeline(
		b,
		"forward",
		"forward_static_mutations",
		"/bench?keep=1&drop=1",
		func(target string) Object {
			return Object{
				"addr":   "127.0.0.1:0",
				"target": target,
				"requestMutations": Object{
					"method":      "POST",
					"path":        "/bench/static",
					"setQuery":    Object{"mode": "static"},
					"removeQuery": Array{"drop"},
					"setHeaders":  Object{"X-Bench-Mode": "static"},
				},
				"responseMutations": Object{
					"statusCode": float64(http.StatusCreated),
					"setHeaders": Object{"X-Bench-Resp": "static"},
				},
			}
		},
		http.StatusCreated,
	)

	benchmarkHTTPServerPipeline(
		b,
		"forward",
		"forward_dynamic_mutations",
		"/bench?keep=1&drop=1",
		func(target string) Object {
			return Object{
				"addr":   "127.0.0.1:0",
				"target": target,
				"requestMutations": NativeFunction(func(args []Value) (Value, error) {
					return Object{
						"method":      "POST",
						"path":        "/bench/dynamic",
						"setQuery":    Object{"mode": "dynamic"},
						"removeQuery": Array{"drop"},
						"setHeaders":  Object{"X-Bench-Mode": "dynamic"},
					}, nil
				}),
				"responseMutations": NativeFunction(func(args []Value) (Value, error) {
					return Object{
						"statusCode": float64(http.StatusAccepted),
						"setHeaders": Object{"X-Bench-Resp": "dynamic"},
					}, nil
				}),
			}
		},
		http.StatusAccepted,
	)
}

func benchmarkHTTPServerPipeline(
	b *testing.B,
	startFn string,
	name string,
	requestPath string,
	buildOptions func(target string) Object,
	wantStatus int,
) {
	b.Run(name, func(b *testing.B) {
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.Copy(io.Discard, r.Body)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))
		b.Cleanup(upstream.Close)

		modules := DefaultModules(rt.NewGoroutineRuntime())
		httpRaw, ok := modules["http"]
		if !ok {
			b.Fatal("module http not found")
		}
		httpMod, ok := httpRaw.(Object)
		if !ok {
			b.Fatalf("module http type = %T, want Object", httpRaw)
		}
		serverRaw, ok := httpMod["server"]
		if !ok {
			b.Fatal("module http.server not found")
		}
		serverNS, ok := serverRaw.(Object)
		if !ok {
			b.Fatalf("module http.server type = %T, want Object", serverRaw)
		}

		serverValue, err := callNativeWithError(serverNS, startFn, buildOptions(upstream.URL))
		if err != nil {
			b.Fatalf("http.server.%s setup error: %v", startFn, err)
		}
		serverObj, ok := serverValue.(Object)
		if !ok {
			b.Fatalf("http.server.%s return type = %T, want Object", startFn, serverValue)
		}
		b.Cleanup(func() {
			_, _ = callNativeWithError(serverObj, "close")
		})

		addr, ok := serverObj["addr"].(string)
		if !ok || addr == "" {
			b.Fatalf("http.server.%s addr = %#v, want non-empty string", startFn, serverObj["addr"])
		}

		client := &http.Client{Timeout: 5 * time.Second}
		url := "http://" + addr + requestPath
		payload := []byte("benchmark-payload")

		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				req, reqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
				if reqErr != nil {
					b.Errorf("build request error: %v", reqErr)
					return
				}
				req.Header.Set("X-Bench", "1")

				resp, doErr := client.Do(req)
				if doErr != nil {
					b.Errorf("request error: %v", doErr)
					return
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if resp.StatusCode != wantStatus {
					b.Errorf("status = %d, want %d", resp.StatusCode, wantStatus)
					return
				}
			}
		})
	})
}

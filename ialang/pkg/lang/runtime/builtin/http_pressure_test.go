package builtin

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	rt "ialang/pkg/lang/runtime"
)

func TestHTTPServerPipelinePressure(t *testing.T) {
	t.Parallel()

	runHTTPPipelinePressureCase(t, "proxy", httpPressureCase{
		startFn:     "proxy",
		requestPath: "/api/pressure?keep=1&drop=1",
		wantStatus:  http.StatusOK,
		buildOptions: func(target string) Object {
			return Object{
				"addr":        "127.0.0.1:0",
				"target":      target,
				"stripPrefix": "/api",
			}
		},
	})

	runHTTPPipelinePressureCase(t, "proxy_static_mutations", httpPressureCase{
		startFn:     "proxy",
		requestPath: "/api/pressure?keep=1&drop=1",
		wantStatus:  http.StatusCreated,
		buildOptions: func(target string) Object {
			return Object{
				"addr":        "127.0.0.1:0",
				"target":      target,
				"stripPrefix": "/api",
				"requestMutations": Object{
					"method":      "POST",
					"path":        "/pressure/static",
					"setQuery":    Object{"mode": "static"},
					"removeQuery": Array{"drop"},
				},
				"responseMutations": Object{
					"statusCode": float64(http.StatusCreated),
				},
			}
		},
	})

	runHTTPPipelinePressureCase(t, "proxy_dynamic_mutations", httpPressureCase{
		startFn:     "proxy",
		requestPath: "/api/pressure?keep=1&drop=1",
		wantStatus:  http.StatusAccepted,
		buildOptions: func(target string) Object {
			return Object{
				"addr":        "127.0.0.1:0",
				"target":      target,
				"stripPrefix": "/api",
				"requestMutations": NativeFunction(func(args []Value) (Value, error) {
					return Object{
						"method":      "POST",
						"path":        "/pressure/dynamic",
						"setQuery":    Object{"mode": "dynamic"},
						"removeQuery": Array{"drop"},
					}, nil
				}),
				"responseMutations": NativeFunction(func(args []Value) (Value, error) {
					return Object{
						"statusCode": float64(http.StatusAccepted),
					}, nil
				}),
			}
		},
	})

	runHTTPPipelinePressureCase(t, "forward", httpPressureCase{
		startFn:     "forward",
		requestPath: "/pressure?keep=1&drop=1",
		wantStatus:  http.StatusOK,
		buildOptions: func(target string) Object {
			return Object{
				"addr":   "127.0.0.1:0",
				"target": target,
			}
		},
	})

	runHTTPPipelinePressureCase(t, "forward_static_mutations", httpPressureCase{
		startFn:     "forward",
		requestPath: "/pressure?keep=1&drop=1",
		wantStatus:  http.StatusCreated,
		buildOptions: func(target string) Object {
			return Object{
				"addr":   "127.0.0.1:0",
				"target": target,
				"requestMutations": Object{
					"method":      "POST",
					"path":        "/pressure/static",
					"setQuery":    Object{"mode": "static"},
					"removeQuery": Array{"drop"},
				},
				"responseMutations": Object{
					"statusCode": float64(http.StatusCreated),
				},
			}
		},
	})

	runHTTPPipelinePressureCase(t, "forward_dynamic_mutations", httpPressureCase{
		startFn:     "forward",
		requestPath: "/pressure?keep=1&drop=1",
		wantStatus:  http.StatusAccepted,
		buildOptions: func(target string) Object {
			return Object{
				"addr":   "127.0.0.1:0",
				"target": target,
				"requestMutations": NativeFunction(func(args []Value) (Value, error) {
					return Object{
						"method":      "POST",
						"path":        "/pressure/dynamic",
						"setQuery":    Object{"mode": "dynamic"},
						"removeQuery": Array{"drop"},
					}, nil
				}),
				"responseMutations": NativeFunction(func(args []Value) (Value, error) {
					return Object{
						"statusCode": float64(http.StatusAccepted),
					}, nil
				}),
			}
		},
	})
}

type httpPressureCase struct {
	startFn      string
	requestPath  string
	wantStatus   int
	buildOptions func(target string) Object
}

func runHTTPPipelinePressureCase(t *testing.T, name string, cfg httpPressureCase) {
	t.Helper()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	modules := DefaultModules(rt.NewGoroutineRuntime())
	httpMod := mustModuleObject(t, modules, "http")
	serverNS := mustObject(t, httpMod, "server")
	serverValue := callNative(t, serverNS, cfg.startFn, cfg.buildOptions(upstream.URL))
	serverObj := mustRuntimeObject(t, serverValue, "http pressure server return")
	defer func() {
		_ = callNative(t, serverObj, "close")
	}()

	addr, _ := serverObj["addr"].(string)
	if addr == "" {
		t.Fatalf("%s addr is empty", name)
	}

	client := &http.Client{}
	url := "http://" + addr + cfg.requestPath
	payload := []byte("pressure-payload")

	const workers = 12
	const requestsPerWorker = 120

	var requestErrors int64
	var badStatus int64
	var firstErr atomic.Value

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < requestsPerWorker; i++ {
				req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
				if err != nil {
					atomic.AddInt64(&requestErrors, 1)
					firstErr.CompareAndSwap(nil, err)
					continue
				}
				resp, err := client.Do(req)
				if err != nil {
					atomic.AddInt64(&requestErrors, 1)
					firstErr.CompareAndSwap(nil, err)
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if resp.StatusCode != cfg.wantStatus {
					atomic.AddInt64(&badStatus, 1)
					firstErr.CompareAndSwap(nil, fmt.Errorf("status=%d want=%d", resp.StatusCode, cfg.wantStatus))
				}
			}
		}()
	}
	wg.Wait()

	total := int64(workers * requestsPerWorker)
	errN := atomic.LoadInt64(&requestErrors)
	badN := atomic.LoadInt64(&badStatus)
	if errN > 0 || badN > 0 {
		first, _ := firstErr.Load().(error)
		t.Fatalf("%s pressure failed: total=%d requestErrors=%d badStatus=%d firstErr=%v", name, total, errN, badN, first)
	}
}

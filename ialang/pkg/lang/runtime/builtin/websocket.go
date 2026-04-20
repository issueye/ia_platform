package builtin

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type websocketClientConfig struct {
	URL       string
	Headers   Object
	TimeoutMs int
}

type websocketServerConfig struct {
	Addr    string
	Path    string
	Echo    bool
	Welcome string
}

func newWebSocketModule(asyncRuntime AsyncRuntime) Object {
	connectFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseWebSocketClientArgs("websocket.client.connect", args)
		if err != nil {
			return nil, err
		}
		return connectWebSocket(cfg, asyncRuntime)
	})
	connectAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return connectFn(args)
		}), nil
	})

	serveFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseWebSocketServerArgs("websocket.server.serve", args)
		if err != nil {
			return nil, err
		}
		return startWebSocketServer(cfg)
	})
	serveAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return serveFn(args)
		}), nil
	})

	clientNamespace := Object{
		"connect":      connectFn,
		"connectAsync": connectAsyncFn,
	}
	serverNamespace := Object{
		"serve":      serveFn,
		"serveAsync": serveAsyncFn,
	}

	namespace := Object{
		"client": clientNamespace,
		"server": serverNamespace,
	}
	module := cloneObject(namespace)
	module["websocket"] = namespace
	return module
}

func parseWebSocketClientArgs(fn string, args []Value) (websocketClientConfig, error) {
	if len(args) < 1 || len(args) > 2 {
		return websocketClientConfig{}, fmt.Errorf("%s expects 1-2 args: url, [options]", fn)
	}
	url, err := asStringArg(fn, args, 0)
	if err != nil {
		return websocketClientConfig{}, err
	}
	cfg := websocketClientConfig{
		URL:       url,
		Headers:   Object{},
		TimeoutMs: 15000,
	}
	if len(args) == 1 || args[1] == nil {
		return cfg, nil
	}
	options, ok := args[1].(Object)
	if !ok {
		return websocketClientConfig{}, fmt.Errorf("%s arg[1] expects object options, got %T", fn, args[1])
	}
	if v, ok := options["headers"]; ok && v != nil {
		headers, ok := v.(Object)
		if !ok {
			return websocketClientConfig{}, fmt.Errorf("websocket.client options.headers expects object, got %T", v)
		}
		cfg.Headers = cloneObject(headers)
	}
	if v, ok := options["timeoutMs"]; ok && v != nil {
		timeoutMs, err := asIntValue("websocket.client options.timeoutMs", v)
		if err != nil {
			return websocketClientConfig{}, err
		}
		if timeoutMs <= 0 {
			return websocketClientConfig{}, fmt.Errorf("websocket.client options.timeoutMs expects positive integer, got %d", timeoutMs)
		}
		cfg.TimeoutMs = timeoutMs
	}
	return cfg, nil
}

func parseWebSocketServerArgs(fn string, args []Value) (websocketServerConfig, error) {
	if len(args) > 1 {
		return websocketServerConfig{}, fmt.Errorf("%s expects 0-1 args: [options]", fn)
	}
	cfg := websocketServerConfig{
		Addr:    "127.0.0.1:0",
		Path:    "/",
		Echo:    true,
		Welcome: "",
	}
	if len(args) == 0 || args[0] == nil {
		return cfg, nil
	}
	options, ok := args[0].(Object)
	if !ok {
		return websocketServerConfig{}, fmt.Errorf("%s arg[0] expects object options, got %T", fn, args[0])
	}
	if v, ok := options["addr"]; ok && v != nil {
		addr, err := asStringValue("websocket.server options.addr", v)
		if err != nil {
			return websocketServerConfig{}, err
		}
		cfg.Addr = addr
	}
	if v, ok := options["path"]; ok && v != nil {
		path, err := asStringValue("websocket.server options.path", v)
		if err != nil {
			return websocketServerConfig{}, err
		}
		if path == "" || path[0] != '/' {
			return websocketServerConfig{}, fmt.Errorf("websocket.server options.path must start with '/', got %q", path)
		}
		cfg.Path = path
	}
	if v, ok := options["echo"]; ok && v != nil {
		echo, ok := v.(bool)
		if !ok {
			return websocketServerConfig{}, fmt.Errorf("websocket.server options.echo expects bool, got %T", v)
		}
		cfg.Echo = echo
	}
	if v, ok := options["welcome"]; ok && v != nil {
		welcome, err := asStringValue("websocket.server options.welcome", v)
		if err != nil {
			return websocketServerConfig{}, err
		}
		cfg.Welcome = welcome
	}
	return cfg, nil
}

func connectWebSocket(cfg websocketClientConfig, asyncRuntime AsyncRuntime) (Value, error) {
	headers := http.Header{}
	for k, v := range cfg.Headers {
		s, err := asStringValue("websocket.client options.headers["+k+"]", v)
		if err != nil {
			return nil, err
		}
		headers.Set(k, s)
	}
	dialer := websocket.Dialer{
		HandshakeTimeout: time.Duration(cfg.TimeoutMs) * time.Millisecond,
	}
	conn, _, err := dialer.Dial(cfg.URL, headers)
	if err != nil {
		return nil, err
	}

	var ioMu sync.Mutex
	var closeOnce sync.Once
	closeConn := func() error {
		var closeErr error
		closeOnce.Do(func() {
			closeErr = conn.Close()
		})
		return closeErr
	}

	sendFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("websocket.client.send expects 1 arg: message")
		}
		msg, err := asStringValue("websocket.client.send arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		ioMu.Lock()
		defer ioMu.Unlock()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			return nil, err
		}
		return true, nil
	})
	recvFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("websocket.client.recv expects 0 args, got %d", len(args))
		}
		ioMu.Lock()
		_, payload, err := conn.ReadMessage()
		ioMu.Unlock()
		if err != nil {
			return nil, err
		}
		return string(payload), nil
	})
	closeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("websocket.client.close expects 0 args, got %d", len(args))
		}
		if err := closeConn(); err != nil {
			return nil, err
		}
		return true, nil
	})
	sendAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return sendFn(args)
		}), nil
	})
	recvAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return recvFn(args)
		}), nil
	})

	return Object{
		"send":      sendFn,
		"recv":      recvFn,
		"close":     closeFn,
		"sendAsync": sendAsyncFn,
		"recvAsync": recvAsyncFn,
	}, nil
}

func startWebSocketServer(cfg websocketServerConfig) (Value, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc(cfg.Path, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		if cfg.Welcome != "" {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(cfg.Welcome)); err != nil {
				return
			}
		}

		for {
			msgType, payload, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if cfg.Echo {
				if msgType != websocket.TextMessage && msgType != websocket.BinaryMessage {
					continue
				}
				if err := conn.WriteMessage(msgType, payload); err != nil {
					return
				}
			}
		}
	})

	server := &http.Server{Handler: mux}
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return nil, err
	}
	go func() {
		_ = server.Serve(ln)
	}()
	addr := ln.Addr().String()
	url := "ws://" + addr + cfg.Path
	return Object{
		"addr": addr,
		"path": cfg.Path,
		"url":  url,
		"close": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("websocket.server.close expects 0 args, got %d", len(args))
			}
			err := server.Close()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return nil, err
			}
			return true, nil
		}),
	}, nil
}

package builtin

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type sseClientConfig struct {
	URL       string
	Headers   Object
	TimeoutMs int
}

type sseServerConfig struct {
	Addr    string
	Path    string
	Headers Object
}

type sseEvent struct {
	Event string
	Data  string
	ID    string
	Retry int
}

type sseServerHub struct {
	mu      sync.Mutex
	nextID  int
	closed  bool
	clients map[int]chan sseEvent
}

func newSSEModule(asyncRuntime AsyncRuntime) Object {
	connectFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseSSEClientArgs("sse.client.connect", args)
		if err != nil {
			return nil, err
		}
		return connectSSE(cfg, asyncRuntime)
	})
	connectAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return connectFn(args)
		}), nil
	})

	serveFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseSSEServerArgs("sse.server.serve", args)
		if err != nil {
			return nil, err
		}
		return startSSEServer(cfg, asyncRuntime)
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
	module["sse"] = namespace
	return module
}

func parseSSEClientArgs(fn string, args []Value) (sseClientConfig, error) {
	if len(args) < 1 || len(args) > 2 {
		return sseClientConfig{}, fmt.Errorf("%s expects 1-2 args: url, [options]", fn)
	}
	url, err := asStringArg(fn, args, 0)
	if err != nil {
		return sseClientConfig{}, err
	}
	cfg := sseClientConfig{
		URL:       url,
		Headers:   Object{},
		TimeoutMs: 15000,
	}
	if len(args) == 1 || args[1] == nil {
		return cfg, nil
	}
	options, ok := args[1].(Object)
	if !ok {
		return sseClientConfig{}, fmt.Errorf("%s arg[1] expects object options, got %T", fn, args[1])
	}
	if v, ok := options["headers"]; ok && v != nil {
		headers, ok := v.(Object)
		if !ok {
			return sseClientConfig{}, fmt.Errorf("sse.client options.headers expects object, got %T", v)
		}
		cfg.Headers = cloneObject(headers)
	}
	if v, ok := options["timeoutMs"]; ok && v != nil {
		timeoutMs, err := asIntValue("sse.client options.timeoutMs", v)
		if err != nil {
			return sseClientConfig{}, err
		}
		if timeoutMs <= 0 {
			return sseClientConfig{}, fmt.Errorf("sse.client options.timeoutMs expects positive integer, got %d", timeoutMs)
		}
		cfg.TimeoutMs = timeoutMs
	}
	return cfg, nil
}

func parseSSEServerArgs(fn string, args []Value) (sseServerConfig, error) {
	if len(args) > 1 {
		return sseServerConfig{}, fmt.Errorf("%s expects 0-1 args: [options]", fn)
	}
	cfg := sseServerConfig{
		Addr:    "127.0.0.1:0",
		Path:    "/events",
		Headers: Object{},
	}
	if len(args) == 0 || args[0] == nil {
		return cfg, nil
	}
	options, ok := args[0].(Object)
	if !ok {
		return sseServerConfig{}, fmt.Errorf("%s arg[0] expects object options, got %T", fn, args[0])
	}
	if v, ok := options["addr"]; ok && v != nil {
		addr, err := asStringValue("sse.server options.addr", v)
		if err != nil {
			return sseServerConfig{}, err
		}
		cfg.Addr = addr
	}
	if v, ok := options["path"]; ok && v != nil {
		path, err := asStringValue("sse.server options.path", v)
		if err != nil {
			return sseServerConfig{}, err
		}
		if path == "" || path[0] != '/' {
			return sseServerConfig{}, fmt.Errorf("sse.server options.path must start with '/', got %q", path)
		}
		cfg.Path = path
	}
	if v, ok := options["headers"]; ok && v != nil {
		headers, ok := v.(Object)
		if !ok {
			return sseServerConfig{}, fmt.Errorf("sse.server options.headers expects object, got %T", v)
		}
		cfg.Headers = cloneObject(headers)
	}
	return cfg, nil
}

func connectSSE(cfg sseClientConfig, asyncRuntime AsyncRuntime) (Value, error) {
	req, err := http.NewRequest(http.MethodGet, cfg.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	for k, v := range cfg.Headers {
		s, err := asStringValue("sse.client options.headers["+k+"]", v)
		if err != nil {
			return nil, err
		}
		req.Header.Set(k, s)
	}

	client := &http.Client{
		Timeout: time.Duration(cfg.TimeoutMs) * time.Millisecond,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("sse.client.connect unexpected status %s", resp.Status)
	}

	reader := bufio.NewReader(resp.Body)
	var readMu sync.Mutex
	var closeOnce sync.Once
	closeStream := func() error {
		var closeErr error
		closeOnce.Do(func() {
			closeErr = resp.Body.Close()
		})
		return closeErr
	}

	recvFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("sse.client.recv expects 0 args, got %d", len(args))
		}
		readMu.Lock()
		defer readMu.Unlock()
		ev, err := readSSEEvent(reader)
		if err != nil {
			return nil, err
		}
		return Object{
			"event": ev.Event,
			"data":  ev.Data,
			"id":    ev.ID,
			"retry": float64(ev.Retry),
		}, nil
	})
	closeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("sse.client.close expects 0 args, got %d", len(args))
		}
		if err := closeStream(); err != nil {
			return nil, err
		}
		return true, nil
	})
	recvAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return recvFn(args)
		}), nil
	})

	return Object{
		"recv":      recvFn,
		"close":     closeFn,
		"recvAsync": recvAsyncFn,
	}, nil
}

func readSSEEvent(reader *bufio.Reader) (sseEvent, error) {
	ev := sseEvent{Retry: -1}
	var dataLines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return sseEvent{}, io.EOF
			}
			return sseEvent{}, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if len(dataLines) == 0 && ev.Event == "" && ev.ID == "" && ev.Retry < 0 {
				continue
			}
			ev.Data = strings.Join(dataLines, "\n")
			if ev.Retry < 0 {
				ev.Retry = 0
			}
			return ev, nil
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		field := line
		value := ""
		if idx := strings.IndexByte(line, ':'); idx >= 0 {
			field = line[:idx]
			value = line[idx+1:]
			if strings.HasPrefix(value, " ") {
				value = value[1:]
			}
		}
		switch field {
		case "event":
			ev.Event = value
		case "data":
			dataLines = append(dataLines, value)
		case "id":
			ev.ID = value
		case "retry":
			if ms, err := strconv.Atoi(value); err == nil {
				ev.Retry = ms
			}
		}
	}
}

func startSSEServer(cfg sseServerConfig, asyncRuntime AsyncRuntime) (Value, error) {
	hub := &sseServerHub{
		clients: map[int]chan sseEvent{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc(cfg.Path, func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		for k, v := range cfg.Headers {
			s, err := asStringValue("sse.server options.headers["+k+"]", v)
			if err != nil {
				continue
			}
			w.Header().Set(k, s)
		}

		clientID, ch, err := hub.register()
		if err != nil {
			http.Error(w, "server closed", http.StatusServiceUnavailable)
			return
		}
		defer hub.unregister(clientID)

		_, _ = w.Write([]byte(": connected\n\n"))
		flusher.Flush()

		for {
			select {
			case <-r.Context().Done():
				return
			case ev, ok := <-ch:
				if !ok {
					return
				}
				if err := writeSSEEvent(w, ev); err != nil {
					return
				}
				flusher.Flush()
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

	sendFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("sse.server.send expects 1-2 args: data, [event]")
		}
		data, err := asStringValue("sse.server.send arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		ev := sseEvent{Data: data}
		if len(args) == 2 && args[1] != nil {
			eventName, err := asStringValue("sse.server.send arg[1]", args[1])
			if err != nil {
				return nil, err
			}
			ev.Event = eventName
		}
		return float64(hub.broadcast(ev)), nil
	})
	sendAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return sendFn(args)
		}), nil
	})
	closeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("sse.server.close expects 0 args, got %d", len(args))
		}
		hub.closeAll()
		err := server.Close()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return nil, err
		}
		return true, nil
	})

	addr := ln.Addr().String()
	return Object{
		"addr":      addr,
		"path":      cfg.Path,
		"url":       "http://" + addr + cfg.Path,
		"send":      sendFn,
		"sendAsync": sendAsyncFn,
		"close":     closeFn,
	}, nil
}

func writeSSEEvent(w io.Writer, ev sseEvent) error {
	var b strings.Builder
	if ev.Event != "" {
		b.WriteString("event: ")
		b.WriteString(ev.Event)
		b.WriteString("\n")
	}
	if ev.ID != "" {
		b.WriteString("id: ")
		b.WriteString(ev.ID)
		b.WriteString("\n")
	}
	if ev.Retry > 0 {
		b.WriteString("retry: ")
		b.WriteString(strconv.Itoa(ev.Retry))
		b.WriteString("\n")
	}
	if ev.Data == "" {
		b.WriteString("data:\n\n")
	} else {
		for _, line := range strings.Split(ev.Data, "\n") {
			b.WriteString("data: ")
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func (h *sseServerHub) register() (int, chan sseEvent, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return 0, nil, errors.New("sse server closed")
	}
	id := h.nextID
	h.nextID++
	ch := make(chan sseEvent, 16)
	h.clients[id] = ch
	return id, ch, nil
}

func (h *sseServerHub) unregister(id int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch, ok := h.clients[id]
	if !ok {
		return
	}
	delete(h.clients, id)
	close(ch)
}

func (h *sseServerHub) broadcast(ev sseEvent) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return 0
	}
	delivered := 0
	for _, ch := range h.clients {
		select {
		case ch <- ev:
			delivered++
		default:
		}
	}
	return delivered
}

func (h *sseServerHub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.closed = true
	for id, ch := range h.clients {
		delete(h.clients, id)
		close(ch)
	}
}

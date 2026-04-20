package builtin

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	goJSON "encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	ipcDefaultListenAddr = "127.0.0.1:0"
	ipcDefaultDialMs     = 5000
	ipcMessageKindReq    = "request"
	ipcMessageKindResp   = "response"
)

type ipcServerConfig struct {
	Addr string
}

type ipcClientConfig struct {
	TimeoutMs int
}

type ipcConnection struct {
	conn         net.Conn
	reader       *bufio.Reader
	asyncRuntime AsyncRuntime
	readMu       sync.Mutex
	writeMu      sync.Mutex
	closeOnce    sync.Once
}

func newIPCModule(asyncRuntime AsyncRuntime) Object {
	listenFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseIPCServerArgs("ipc.server.listen", args)
		if err != nil {
			return nil, err
		}
		return startIPCServer(cfg, asyncRuntime)
	})
	listenAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return listenFn(args)
		}), nil
	})
	connectFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseIPCClientArgs("ipc.client.connect", args)
		if err != nil {
			return nil, err
		}
		addr, err := asStringArg("ipc.client.connect", args, 0)
		if err != nil {
			return nil, err
		}
		return connectIPC(addr, cfg, asyncRuntime)
	})
	connectAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return connectFn(args)
		}), nil
	})

	buildRequestFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 || len(args) > 3 {
			return nil, fmt.Errorf("ipc.buildRequest expects 2-3 args: method, payload, [options]")
		}
		method, err := asStringArg("ipc.buildRequest", args, 0)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(method) == "" {
			return nil, fmt.Errorf("ipc.buildRequest method cannot be empty")
		}

		reqID := ""
		if len(args) == 3 && args[2] != nil {
			options, ok := args[2].(Object)
			if !ok {
				return nil, fmt.Errorf("ipc.buildRequest arg[2] expects object options, got %T", args[2])
			}
			if v, ok := options["id"]; ok && v != nil {
				parsed, err := asStringValue("ipc.buildRequest options.id", v)
				if err != nil {
					return nil, err
				}
				reqID = strings.TrimSpace(parsed)
			}
		}
		if reqID == "" {
			reqID, err = ipcNewID()
			if err != nil {
				return nil, fmt.Errorf("ipc.buildRequest generate id error: %w", err)
			}
		}

		return Object{
			"kind":    ipcMessageKindReq,
			"id":      reqID,
			"method":  method,
			"payload": args[1],
		}, nil
	})

	buildResponseFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 3 || len(args) > 4 {
			return nil, fmt.Errorf("ipc.buildResponse expects 3-4 args: requestId, ok, payload, [error]")
		}
		requestID, err := asStringArg("ipc.buildResponse", args, 0)
		if err != nil {
			return nil, err
		}
		okFlag, ok := args[1].(bool)
		if !ok {
			return nil, fmt.Errorf("ipc.buildResponse arg[1] expects bool, got %T", args[1])
		}

		out := Object{
			"kind":      ipcMessageKindResp,
			"requestId": requestID,
			"ok":        okFlag,
			"payload":   args[2],
		}
		if !okFlag {
			errText := "request failed"
			if len(args) == 4 && args[3] != nil {
				parsed, err := asStringValue("ipc.buildResponse arg[3]", args[3])
				if err != nil {
					return nil, err
				}
				if strings.TrimSpace(parsed) != "" {
					errText = parsed
				}
			}
			out["error"] = errText
		}
		return out, nil
	})

	serverNS := Object{
		"listen":      listenFn,
		"listenAsync": listenAsyncFn,
	}
	clientNS := Object{
		"connect":      connectFn,
		"connectAsync": connectAsyncFn,
	}

	namespace := Object{
		"server":        serverNS,
		"client":        clientNS,
		"buildRequest":  buildRequestFn,
		"buildResponse": buildResponseFn,
	}
	module := cloneObject(namespace)
	module["ipc"] = namespace
	return module
}

func parseIPCServerArgs(fn string, args []Value) (ipcServerConfig, error) {
	if len(args) > 1 {
		return ipcServerConfig{}, fmt.Errorf("%s expects 0-1 args: [options]", fn)
	}
	cfg := ipcServerConfig{
		Addr: ipcDefaultListenAddr,
	}
	if len(args) == 0 || args[0] == nil {
		return cfg, nil
	}
	options, ok := args[0].(Object)
	if !ok {
		return ipcServerConfig{}, fmt.Errorf("%s arg[0] expects object options, got %T", fn, args[0])
	}
	if v, ok := options["addr"]; ok && v != nil {
		addr, err := asStringValue("ipc.server options.addr", v)
		if err != nil {
			return ipcServerConfig{}, err
		}
		cfg.Addr = addr
	}
	if err := ipcValidateLocalAddr(cfg.Addr); err != nil {
		return ipcServerConfig{}, err
	}
	return cfg, nil
}

func parseIPCClientArgs(fn string, args []Value) (ipcClientConfig, error) {
	if len(args) < 1 || len(args) > 2 {
		return ipcClientConfig{}, fmt.Errorf("%s expects 1-2 args: addr, [options]", fn)
	}
	cfg := ipcClientConfig{
		TimeoutMs: ipcDefaultDialMs,
	}
	if len(args) == 1 || args[1] == nil {
		return cfg, nil
	}
	options, ok := args[1].(Object)
	if !ok {
		return ipcClientConfig{}, fmt.Errorf("%s arg[1] expects object options, got %T", fn, args[1])
	}
	if v, ok := options["timeoutMs"]; ok && v != nil {
		timeout, err := asIntValue("ipc.client options.timeoutMs", v)
		if err != nil {
			return ipcClientConfig{}, err
		}
		if timeout <= 0 {
			return ipcClientConfig{}, fmt.Errorf("ipc.client options.timeoutMs expects positive integer, got %d", timeout)
		}
		cfg.TimeoutMs = timeout
	}
	return cfg, nil
}

func startIPCServer(cfg ipcServerConfig, asyncRuntime AsyncRuntime) (Value, error) {
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return nil, err
	}

	acceptFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("ipc.server.accept expects 0 args, got %d", len(args))
		}
		conn, err := ln.Accept()
		if err != nil {
			return nil, err
		}
		if !ipcRemoteLoopback(conn.RemoteAddr()) {
			_ = conn.Close()
			return nil, fmt.Errorf("ipc.server.accept only allows loopback connections, got %s", conn.RemoteAddr().String())
		}
		return newIPCConnectionObject(conn, asyncRuntime), nil
	})
	acceptAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return acceptFn(args)
		}), nil
	})
	closeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("ipc.server.close expects 0 args, got %d", len(args))
		}
		if err := ln.Close(); err != nil {
			return nil, err
		}
		return true, nil
	})

	return Object{
		"network":     "tcp",
		"addr":        ln.Addr().String(),
		"accept":      acceptFn,
		"acceptAsync": acceptAsyncFn,
		"close":       closeFn,
	}, nil
}

func connectIPC(addr string, cfg ipcClientConfig, asyncRuntime AsyncRuntime) (Value, error) {
	if err := ipcValidateLocalAddr(addr); err != nil {
		return nil, err
	}
	dialer := net.Dialer{
		Timeout: time.Duration(cfg.TimeoutMs) * time.Millisecond,
	}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	if !ipcRemoteLoopback(conn.RemoteAddr()) {
		_ = conn.Close()
		return nil, fmt.Errorf("ipc.client.connect only allows loopback endpoint, got %s", conn.RemoteAddr().String())
	}
	return newIPCConnectionObject(conn, asyncRuntime), nil
}

func newIPCConnectionObject(conn net.Conn, asyncRuntime AsyncRuntime) Object {
	state := &ipcConnection{
		conn:         conn,
		reader:       bufio.NewReader(conn),
		asyncRuntime: asyncRuntime,
	}

	sendFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("ipc.conn.send expects 1 arg: value")
		}
		if err := state.send(args[0]); err != nil {
			return nil, err
		}
		return true, nil
	})
	sendAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return sendFn(args)
		}), nil
	})
	recvFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("ipc.conn.recv expects 0 args, got %d", len(args))
		}
		return state.recv()
	})
	recvAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return recvFn(args)
		}), nil
	})
	callFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 || len(args) > 3 {
			return nil, fmt.Errorf("ipc.conn.call expects 2-3 args: method, payload, [options]")
		}
		method, err := asStringArg("ipc.conn.call", args, 0)
		if err != nil {
			return nil, err
		}
		reqID := ""
		if len(args) == 3 && args[2] != nil {
			options, ok := args[2].(Object)
			if !ok {
				return nil, fmt.Errorf("ipc.conn.call arg[2] expects object options, got %T", args[2])
			}
			if v, ok := options["id"]; ok && v != nil {
				parsed, err := asStringValue("ipc.conn.call options.id", v)
				if err != nil {
					return nil, err
				}
				reqID = strings.TrimSpace(parsed)
			}
		}
		if reqID == "" {
			reqID, err = ipcNewID()
			if err != nil {
				return nil, fmt.Errorf("ipc.conn.call generate id error: %w", err)
			}
		}

		req := Object{
			"kind":    ipcMessageKindReq,
			"id":      reqID,
			"method":  method,
			"payload": args[1],
		}
		if err := state.send(req); err != nil {
			return nil, err
		}
		respVal, err := state.recv()
		if err != nil {
			return nil, err
		}
		respObj, ok := respVal.(Object)
		if !ok {
			return nil, fmt.Errorf("ipc.conn.call expects response object, got %T", respVal)
		}
		kind, _ := respObj["kind"].(string)
		if kind != ipcMessageKindResp {
			return nil, fmt.Errorf("ipc.conn.call expects response kind %q, got %q", ipcMessageKindResp, kind)
		}
		responseReqID, _ := respObj["requestId"].(string)
		if responseReqID != reqID {
			return nil, fmt.Errorf("ipc.conn.call response requestId mismatch: got %q, want %q", responseReqID, reqID)
		}
		return respObj, nil
	})
	callAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return callFn(args)
		}), nil
	})
	replyFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 3 || len(args) > 4 {
			return nil, fmt.Errorf("ipc.conn.reply expects 3-4 args: request, ok, payload, [error]")
		}
		request, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("ipc.conn.reply arg[0] expects request object, got %T", args[0])
		}
		reqID, _ := request["id"].(string)
		if strings.TrimSpace(reqID) == "" {
			return nil, fmt.Errorf("ipc.conn.reply request.id is required")
		}
		okFlag, ok := args[1].(bool)
		if !ok {
			return nil, fmt.Errorf("ipc.conn.reply arg[1] expects bool, got %T", args[1])
		}
		response := Object{
			"kind":      ipcMessageKindResp,
			"requestId": reqID,
			"ok":        okFlag,
			"payload":   args[2],
		}
		if !okFlag {
			errText := "request failed"
			if len(args) == 4 && args[3] != nil {
				parsed, err := asStringValue("ipc.conn.reply arg[3]", args[3])
				if err != nil {
					return nil, err
				}
				if strings.TrimSpace(parsed) != "" {
					errText = parsed
				}
			}
			response["error"] = errText
		}
		if err := state.send(response); err != nil {
			return nil, err
		}
		return true, nil
	})
	replyAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return replyFn(args)
		}), nil
	})
	closeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("ipc.conn.close expects 0 args, got %d", len(args))
		}
		if err := state.close(); err != nil {
			return nil, err
		}
		return true, nil
	})

	return Object{
		"send":       sendFn,
		"sendAsync":  sendAsyncFn,
		"recv":       recvFn,
		"recvAsync":  recvAsyncFn,
		"call":       callFn,
		"callAsync":  callAsyncFn,
		"reply":      replyFn,
		"replyAsync": replyAsyncFn,
		"close":      closeFn,
	}
}

func (c *ipcConnection) send(v Value) error {
	raw, err := goJSON.Marshal(v)
	if err != nil {
		return fmt.Errorf("ipc.conn.send json encode error: %w", err)
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if _, err := c.conn.Write(raw); err != nil {
		return err
	}
	if _, err := c.conn.Write([]byte{'\n'}); err != nil {
		return err
	}
	return nil
}

func (c *ipcConnection) recv() (Value, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	for {
		line, err := c.reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) && len(line) > 0 {
				// continue parse trailing bytes without newline
			} else {
				return nil, err
			}
		}
		line = bytesTrimLine(line)
		if len(line) == 0 {
			if err != nil {
				return nil, io.EOF
			}
			continue
		}
		var parsed any
		if unmarshalErr := goJSON.Unmarshal(line, &parsed); unmarshalErr != nil {
			return nil, fmt.Errorf("ipc.conn.recv json decode error: %w", unmarshalErr)
		}
		return toRuntimeJSONValue(parsed), nil
	}
}

func (c *ipcConnection) close() error {
	var closeErr error
	c.closeOnce.Do(func() {
		closeErr = c.conn.Close()
	})
	return closeErr
}

func bytesTrimLine(in []byte) []byte {
	for len(in) > 0 {
		last := in[len(in)-1]
		if last == '\n' || last == '\r' {
			in = in[:len(in)-1]
			continue
		}
		break
	}
	return in
}

func ipcValidateLocalAddr(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("ipc address must be host:port, got %q", addr)
	}
	if !ipcIsLoopbackHost(host) {
		return fmt.Errorf("ipc only allows loopback host, got %q", host)
	}
	return nil
}

func ipcRemoteLoopback(addr net.Addr) bool {
	if addr == nil {
		return false
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return false
	}
	return ipcIsLoopbackHost(host)
}

func ipcIsLoopbackHost(host string) bool {
	if host == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func ipcNewID() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

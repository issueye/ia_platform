package builtin

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const (
	socketDefaultListenAddr = "127.0.0.1:0"
	socketDefaultNetwork    = "tcp"
	socketDefaultUDPNetwork = "udp"
	socketDefaultDialMs     = 5000
	socketDefaultReadSize   = 4096
	socketMaxReadSize       = 1 << 20
)

type socketServerConfig struct {
	Network string
	Addr    string
}

type socketClientConfig struct {
	Network   string
	TimeoutMs int
}

type socketConnection struct {
	conn         net.Conn
	asyncRuntime AsyncRuntime
	readMu       sync.Mutex
	writeMu      sync.Mutex
	closeOnce    sync.Once
}

type socketPacketEndpoint struct {
	conn         net.PacketConn
	network      string
	asyncRuntime AsyncRuntime
	readMu       sync.Mutex
	writeMu      sync.Mutex
	closeOnce    sync.Once
}

func newSocketModule(asyncRuntime AsyncRuntime) Object {
	listenFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseSocketServerArgs("socket.server.listen", args)
		if err != nil {
			return nil, err
		}
		return startSocketServer(cfg, asyncRuntime)
	})
	listenAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return listenFn(args)
		}), nil
	})
	connectFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseSocketClientArgs("socket.client.connect", args)
		if err != nil {
			return nil, err
		}
		addr, err := asStringArg("socket.client.connect", args, 0)
		if err != nil {
			return nil, err
		}
		return connectSocket(addr, cfg, asyncRuntime)
	})
	connectAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return connectFn(args)
		}), nil
	})
	udpBindFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseSocketUDPServerArgs("socket.udp.bind", args)
		if err != nil {
			return nil, err
		}
		return bindUDPSocket(cfg, asyncRuntime)
	})
	udpBindAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return udpBindFn(args)
		}), nil
	})
	udpConnectFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseSocketUDPClientArgs("socket.udp.connect", args)
		if err != nil {
			return nil, err
		}
		addr, err := asStringArg("socket.udp.connect", args, 0)
		if err != nil {
			return nil, err
		}
		return connectSocket(addr, cfg, asyncRuntime)
	})
	udpConnectAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return udpConnectFn(args)
		}), nil
	})

	serverNS := Object{
		"listen":      listenFn,
		"listenAsync": listenAsyncFn,
	}
	clientNS := Object{
		"connect":      connectFn,
		"connectAsync": connectAsyncFn,
	}
	udpNS := Object{
		"bind":         udpBindFn,
		"bindAsync":    udpBindAsyncFn,
		"connect":      udpConnectFn,
		"connectAsync": udpConnectAsyncFn,
	}

	namespace := Object{
		"server": serverNS,
		"client": clientNS,
		"udp":    udpNS,
	}
	module := cloneObject(namespace)
	module["socket"] = namespace
	return module
}

func parseSocketServerArgs(fn string, args []Value) (socketServerConfig, error) {
	if len(args) > 1 {
		return socketServerConfig{}, fmt.Errorf("%s expects 0-1 args: [options]", fn)
	}
	cfg := socketServerConfig{
		Network: socketDefaultNetwork,
		Addr:    socketDefaultListenAddr,
	}
	if len(args) == 0 || args[0] == nil {
		return cfg, nil
	}
	options, ok := args[0].(Object)
	if !ok {
		return socketServerConfig{}, fmt.Errorf("%s arg[0] expects object options, got %T", fn, args[0])
	}
	if v, ok := options["network"]; ok && v != nil {
		network, err := asStringValue("socket.server options.network", v)
		if err != nil {
			return socketServerConfig{}, err
		}
		if !socketTCPNetworkSupported(network) {
			return socketServerConfig{}, fmt.Errorf("socket.server options.network only supports tcp/tcp4/tcp6, got %q", network)
		}
		cfg.Network = network
	}
	if v, ok := options["addr"]; ok && v != nil {
		addr, err := asStringValue("socket.server options.addr", v)
		if err != nil {
			return socketServerConfig{}, err
		}
		cfg.Addr = addr
	}
	return cfg, nil
}

func parseSocketClientArgs(fn string, args []Value) (socketClientConfig, error) {
	if len(args) < 1 || len(args) > 2 {
		return socketClientConfig{}, fmt.Errorf("%s expects 1-2 args: addr, [options]", fn)
	}
	cfg := socketClientConfig{
		Network:   socketDefaultNetwork,
		TimeoutMs: socketDefaultDialMs,
	}
	if len(args) == 1 || args[1] == nil {
		return cfg, nil
	}
	options, ok := args[1].(Object)
	if !ok {
		return socketClientConfig{}, fmt.Errorf("%s arg[1] expects object options, got %T", fn, args[1])
	}
	if v, ok := options["network"]; ok && v != nil {
		network, err := asStringValue("socket.client options.network", v)
		if err != nil {
			return socketClientConfig{}, err
		}
		if !socketTCPNetworkSupported(network) {
			return socketClientConfig{}, fmt.Errorf("socket.client options.network only supports tcp/tcp4/tcp6, got %q", network)
		}
		cfg.Network = network
	}
	if v, ok := options["timeoutMs"]; ok && v != nil {
		timeout, err := asIntValue("socket.client options.timeoutMs", v)
		if err != nil {
			return socketClientConfig{}, err
		}
		if timeout <= 0 {
			return socketClientConfig{}, fmt.Errorf("socket.client options.timeoutMs expects positive integer, got %d", timeout)
		}
		cfg.TimeoutMs = timeout
	}
	return cfg, nil
}

func parseSocketUDPServerArgs(fn string, args []Value) (socketServerConfig, error) {
	if len(args) > 1 {
		return socketServerConfig{}, fmt.Errorf("%s expects 0-1 args: [options]", fn)
	}
	cfg := socketServerConfig{
		Network: socketDefaultUDPNetwork,
		Addr:    socketDefaultListenAddr,
	}
	if len(args) == 0 || args[0] == nil {
		return cfg, nil
	}
	options, ok := args[0].(Object)
	if !ok {
		return socketServerConfig{}, fmt.Errorf("%s arg[0] expects object options, got %T", fn, args[0])
	}
	if v, ok := options["network"]; ok && v != nil {
		network, err := asStringValue("socket.udp options.network", v)
		if err != nil {
			return socketServerConfig{}, err
		}
		if !socketUDPNetworkSupported(network) {
			return socketServerConfig{}, fmt.Errorf("socket.udp options.network only supports udp/udp4/udp6, got %q", network)
		}
		cfg.Network = network
	}
	if v, ok := options["addr"]; ok && v != nil {
		addr, err := asStringValue("socket.udp options.addr", v)
		if err != nil {
			return socketServerConfig{}, err
		}
		cfg.Addr = addr
	}
	return cfg, nil
}

func parseSocketUDPClientArgs(fn string, args []Value) (socketClientConfig, error) {
	if len(args) < 1 || len(args) > 2 {
		return socketClientConfig{}, fmt.Errorf("%s expects 1-2 args: addr, [options]", fn)
	}
	cfg := socketClientConfig{
		Network:   socketDefaultUDPNetwork,
		TimeoutMs: socketDefaultDialMs,
	}
	if len(args) == 1 || args[1] == nil {
		return cfg, nil
	}
	options, ok := args[1].(Object)
	if !ok {
		return socketClientConfig{}, fmt.Errorf("%s arg[1] expects object options, got %T", fn, args[1])
	}
	if v, ok := options["network"]; ok && v != nil {
		network, err := asStringValue("socket.udp options.network", v)
		if err != nil {
			return socketClientConfig{}, err
		}
		if !socketUDPNetworkSupported(network) {
			return socketClientConfig{}, fmt.Errorf("socket.udp options.network only supports udp/udp4/udp6, got %q", network)
		}
		cfg.Network = network
	}
	if v, ok := options["timeoutMs"]; ok && v != nil {
		timeout, err := asIntValue("socket.udp options.timeoutMs", v)
		if err != nil {
			return socketClientConfig{}, err
		}
		if timeout <= 0 {
			return socketClientConfig{}, fmt.Errorf("socket.udp options.timeoutMs expects positive integer, got %d", timeout)
		}
		cfg.TimeoutMs = timeout
	}
	return cfg, nil
}

func socketTCPNetworkSupported(network string) bool {
	return network == "tcp" || network == "tcp4" || network == "tcp6"
}

func socketUDPNetworkSupported(network string) bool {
	return network == "udp" || network == "udp4" || network == "udp6"
}

func startSocketServer(cfg socketServerConfig, asyncRuntime AsyncRuntime) (Value, error) {
	ln, err := net.Listen(cfg.Network, cfg.Addr)
	if err != nil {
		return nil, err
	}

	acceptFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("socket.server.accept expects 0 args, got %d", len(args))
		}
		conn, err := ln.Accept()
		if err != nil {
			return nil, err
		}
		return newSocketConnectionObject(conn, asyncRuntime), nil
	})
	acceptAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return acceptFn(args)
		}), nil
	})
	closeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("socket.server.close expects 0 args, got %d", len(args))
		}
		if err := ln.Close(); err != nil {
			return nil, err
		}
		return true, nil
	})

	return Object{
		"network":     cfg.Network,
		"addr":        ln.Addr().String(),
		"accept":      acceptFn,
		"acceptAsync": acceptAsyncFn,
		"close":       closeFn,
	}, nil
}

func bindUDPSocket(cfg socketServerConfig, asyncRuntime AsyncRuntime) (Value, error) {
	packetConn, err := net.ListenPacket(cfg.Network, cfg.Addr)
	if err != nil {
		return nil, err
	}
	return newSocketPacketEndpointObject(packetConn, cfg.Network, asyncRuntime), nil
}

func connectSocket(addr string, cfg socketClientConfig, asyncRuntime AsyncRuntime) (Value, error) {
	dialer := net.Dialer{
		Timeout: time.Duration(cfg.TimeoutMs) * time.Millisecond,
	}
	conn, err := dialer.Dial(cfg.Network, addr)
	if err != nil {
		return nil, err
	}
	return newSocketConnectionObject(conn, asyncRuntime), nil
}

func newSocketConnectionObject(conn net.Conn, asyncRuntime AsyncRuntime) Object {
	state := &socketConnection{
		conn:         conn,
		asyncRuntime: asyncRuntime,
	}

	writeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("socket.conn.write expects 1 arg: data")
		}
		data, err := asStringValue("socket.conn.write arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		written, err := state.write(data)
		if err != nil {
			return nil, err
		}
		return float64(written), nil
	})
	writeAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return writeFn(args)
		}), nil
	})
	readFn := NativeFunction(func(args []Value) (Value, error) {
		size, err := parseSocketReadSize("socket.conn.read", args)
		if err != nil {
			return nil, err
		}
		return state.read(size)
	})
	readAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return readFn(args)
		}), nil
	})
	sendFn := NativeFunction(func(args []Value) (Value, error) {
		if _, err := writeFn(args); err != nil {
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
		return readFn(args)
	})
	recvAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return recvFn(args)
		}), nil
	})
	localAddrFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("socket.conn.localAddr expects 0 args, got %d", len(args))
		}
		return state.conn.LocalAddr().String(), nil
	})
	remoteAddrFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("socket.conn.remoteAddr expects 0 args, got %d", len(args))
		}
		return state.conn.RemoteAddr().String(), nil
	})
	closeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("socket.conn.close expects 0 args, got %d", len(args))
		}
		if err := state.close(); err != nil {
			return nil, err
		}
		return true, nil
	})

	return Object{
		"write":      writeFn,
		"writeAsync": writeAsyncFn,
		"read":       readFn,
		"readAsync":  readAsyncFn,
		"send":       sendFn,
		"sendAsync":  sendAsyncFn,
		"recv":       recvFn,
		"recvAsync":  recvAsyncFn,
		"localAddr":  localAddrFn,
		"remoteAddr": remoteAddrFn,
		"close":      closeFn,
	}
}

func newSocketPacketEndpointObject(conn net.PacketConn, network string, asyncRuntime AsyncRuntime) Object {
	state := &socketPacketEndpoint{
		conn:         conn,
		network:      network,
		asyncRuntime: asyncRuntime,
	}

	sendToFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("socket.udp.sendTo expects 2 args: data, addr")
		}
		data, err := asStringValue("socket.udp.sendTo arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		addr, err := asStringValue("socket.udp.sendTo arg[1]", args[1])
		if err != nil {
			return nil, err
		}
		n, err := state.sendTo(data, addr)
		if err != nil {
			return nil, err
		}
		return float64(n), nil
	})
	sendToAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return sendToFn(args)
		}), nil
	})
	recvFromFn := NativeFunction(func(args []Value) (Value, error) {
		size, err := parseSocketReadSize("socket.udp.recvFrom", args)
		if err != nil {
			return nil, err
		}
		return state.recvFrom(size)
	})
	recvFromAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return recvFromFn(args)
		}), nil
	})
	localAddrFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("socket.udp.localAddr expects 0 args, got %d", len(args))
		}
		return state.conn.LocalAddr().String(), nil
	})
	closeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("socket.udp.close expects 0 args, got %d", len(args))
		}
		if err := state.close(); err != nil {
			return nil, err
		}
		return true, nil
	})

	return Object{
		"network":       network,
		"addr":          conn.LocalAddr().String(),
		"sendTo":        sendToFn,
		"sendToAsync":   sendToAsyncFn,
		"recvFrom":      recvFromFn,
		"recvFromAsync": recvFromAsyncFn,
		"localAddr":     localAddrFn,
		"close":         closeFn,
	}
}

func parseSocketReadSize(fn string, args []Value) (int, error) {
	if len(args) > 1 {
		return 0, fmt.Errorf("%s expects 0-1 args: [size]", fn)
	}
	size := socketDefaultReadSize
	if len(args) == 1 && args[0] != nil {
		parsed, err := asIntValue(fn+" arg[0]", args[0])
		if err != nil {
			return 0, err
		}
		size = parsed
	}
	if size <= 0 {
		return 0, fmt.Errorf("%s size expects positive integer, got %d", fn, size)
	}
	if size > socketMaxReadSize {
		return 0, fmt.Errorf("%s size exceeds limit %d, got %d", fn, socketMaxReadSize, size)
	}
	return size, nil
}

func (c *socketConnection) write(data string) (int, error) {
	bytes := []byte(data)
	total := 0

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	for total < len(bytes) {
		n, err := c.conn.Write(bytes[total:])
		if n > 0 {
			total += n
		}
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func (c *socketConnection) read(size int) (string, error) {
	buf := make([]byte, size)

	c.readMu.Lock()
	n, err := c.conn.Read(buf)
	c.readMu.Unlock()

	if err != nil {
		if errors.Is(err, io.EOF) && n > 0 {
			return string(buf[:n]), nil
		}
		return "", err
	}
	return string(buf[:n]), nil
}

func (c *socketConnection) close() error {
	var closeErr error
	c.closeOnce.Do(func() {
		closeErr = c.conn.Close()
	})
	return closeErr
}

func (c *socketPacketEndpoint) sendTo(data string, addr string) (int, error) {
	remote, err := net.ResolveUDPAddr(c.network, addr)
	if err != nil {
		return 0, err
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteTo([]byte(data), remote)
}

func (c *socketPacketEndpoint) recvFrom(size int) (Value, error) {
	buf := make([]byte, size)

	c.readMu.Lock()
	n, addr, err := c.conn.ReadFrom(buf)
	c.readMu.Unlock()
	if err != nil {
		return nil, err
	}

	return Object{
		"data": string(buf[:n]),
		"addr": addr.String(),
	}, nil
}

func (c *socketPacketEndpoint) close() error {
	var closeErr error
	c.closeOnce.Do(func() {
		closeErr = c.conn.Close()
	})
	return closeErr
}

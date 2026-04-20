package builtin

import (
	"testing"

	rt "ialang/pkg/lang/runtime"
)

func TestSocketModuleSendRecv(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	socketMod := mustModuleObject(t, modules, "socket")
	serverNS := mustObject(t, socketMod, "server")
	clientNS := mustObject(t, socketMod, "client")

	serverValue := callNative(t, serverNS, "listen", Object{
		"addr": "127.0.0.1:0",
	})
	server := mustRuntimeObject(t, serverValue, "socket.server.listen return")
	addr, ok := server["addr"].(string)
	if !ok || addr == "" {
		t.Fatalf("socket.server.listen addr = %#v, want non-empty string", server["addr"])
	}

	acceptAwait := callNative(t, server, "acceptAsync")
	clientValue := callNative(t, clientNS, "connect", addr)
	clientConn := mustRuntimeObject(t, clientValue, "socket.client.connect return")
	serverConnValue := awaitValue(t, acceptAwait)
	serverConn := mustRuntimeObject(t, serverConnValue, "socket.server.accept return")

	_ = callNative(t, clientConn, "send", "ping")
	serverRecv := callNative(t, serverConn, "recv", float64(4))
	if s, ok := serverRecv.(string); !ok || s != "ping" {
		t.Fatalf("socket server recv = %#v, want ping", serverRecv)
	}

	_ = callNative(t, serverConn, "send", "pong")
	clientRecv := callNative(t, clientConn, "recv", float64(4))
	if s, ok := clientRecv.(string); !ok || s != "pong" {
		t.Fatalf("socket client recv = %#v, want pong", clientRecv)
	}

	written := callNative(t, clientConn, "write", "abc")
	if n, ok := written.(float64); !ok || int(n) != 3 {
		t.Fatalf("socket write bytes = %#v, want 3", written)
	}
	readBack := callNative(t, serverConn, "read", float64(3))
	if s, ok := readBack.(string); !ok || s != "abc" {
		t.Fatalf("socket read = %#v, want abc", readBack)
	}

	localAddr := callNative(t, clientConn, "localAddr")
	if s, ok := localAddr.(string); !ok || s == "" {
		t.Fatalf("socket localAddr = %#v, want non-empty string", localAddr)
	}
	remoteAddr := callNative(t, clientConn, "remoteAddr")
	if s, ok := remoteAddr.(string); !ok || s == "" {
		t.Fatalf("socket remoteAddr = %#v, want non-empty string", remoteAddr)
	}

	_ = callNative(t, clientConn, "close")
	_ = callNative(t, serverConn, "close")
	_ = callNative(t, server, "close")
}

func TestSocketModuleAsync(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	socketMod := mustModuleObject(t, modules, "socket")
	serverNS := mustObject(t, socketMod, "server")
	clientNS := mustObject(t, socketMod, "client")

	serverValue := awaitValue(t, callNative(t, serverNS, "listenAsync", Object{
		"addr": "127.0.0.1:0",
	}))
	server := mustRuntimeObject(t, serverValue, "socket.server.listenAsync return")
	addr, ok := server["addr"].(string)
	if !ok || addr == "" {
		t.Fatalf("socket.server.listenAsync addr = %#v, want non-empty string", server["addr"])
	}

	acceptAwait := callNative(t, server, "acceptAsync")
	clientValue := awaitValue(t, callNative(t, clientNS, "connectAsync", addr, Object{
		"timeoutMs": float64(3000),
	}))
	clientConn := mustRuntimeObject(t, clientValue, "socket.client.connectAsync return")
	serverConn := mustRuntimeObject(t, awaitValue(t, acceptAwait), "socket.server.accept return")

	_ = awaitValue(t, callNative(t, clientConn, "sendAsync", "hello"))
	gotServer := awaitValue(t, callNative(t, serverConn, "recvAsync", float64(5)))
	if s, ok := gotServer.(string); !ok || s != "hello" {
		t.Fatalf("socket async recv on server = %#v, want hello", gotServer)
	}

	_ = awaitValue(t, callNative(t, serverConn, "sendAsync", "world"))
	gotClient := awaitValue(t, callNative(t, clientConn, "recvAsync", float64(5)))
	if s, ok := gotClient.(string); !ok || s != "world" {
		t.Fatalf("socket async recv on client = %#v, want world", gotClient)
	}

	_ = callNative(t, clientConn, "close")
	_ = callNative(t, serverConn, "close")
	_ = callNative(t, server, "close")
}

func TestSocketUDPBindSendRecv(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	socketMod := mustModuleObject(t, modules, "socket")
	udpNS := mustObject(t, socketMod, "udp")

	serverValue := callNative(t, udpNS, "bind", Object{
		"addr": "127.0.0.1:0",
	})
	server := mustRuntimeObject(t, serverValue, "socket.udp.bind(server) return")
	serverAddrValue := callNative(t, server, "localAddr")
	serverAddr, ok := serverAddrValue.(string)
	if !ok || serverAddr == "" {
		t.Fatalf("socket.udp.localAddr(server) = %#v, want non-empty string", serverAddrValue)
	}

	clientValue := callNative(t, udpNS, "bind", Object{
		"addr": "127.0.0.1:0",
	})
	client := mustRuntimeObject(t, clientValue, "socket.udp.bind(client) return")

	sent := callNative(t, client, "sendTo", "ping", serverAddr)
	if n, ok := sent.(float64); !ok || int(n) != 4 {
		t.Fatalf("socket.udp.sendTo bytes = %#v, want 4", sent)
	}
	serverRecv := mustRuntimeObject(t, callNative(t, server, "recvFrom", float64(4)), "socket.udp.recvFrom(server)")
	if s, ok := serverRecv["data"].(string); !ok || s != "ping" {
		t.Fatalf("socket.udp.recvFrom(server).data = %#v, want ping", serverRecv["data"])
	}
	replyAddr, ok := serverRecv["addr"].(string)
	if !ok || replyAddr == "" {
		t.Fatalf("socket.udp.recvFrom(server).addr = %#v, want non-empty string", serverRecv["addr"])
	}

	replySent := callNative(t, server, "sendTo", "pong", replyAddr)
	if n, ok := replySent.(float64); !ok || int(n) != 4 {
		t.Fatalf("socket.udp.sendTo reply bytes = %#v, want 4", replySent)
	}
	clientRecv := mustRuntimeObject(t, callNative(t, client, "recvFrom", float64(4)), "socket.udp.recvFrom(client)")
	if s, ok := clientRecv["data"].(string); !ok || s != "pong" {
		t.Fatalf("socket.udp.recvFrom(client).data = %#v, want pong", clientRecv["data"])
	}

	_ = callNative(t, client, "close")
	_ = callNative(t, server, "close")
}

func TestSocketUDPAsyncAndConnect(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	socketMod := mustModuleObject(t, modules, "socket")
	udpNS := mustObject(t, socketMod, "udp")

	serverValue := awaitValue(t, callNative(t, udpNS, "bindAsync", Object{
		"addr": "127.0.0.1:0",
	}))
	server := mustRuntimeObject(t, serverValue, "socket.udp.bindAsync return")
	serverAddrValue := callNative(t, server, "localAddr")
	serverAddr, ok := serverAddrValue.(string)
	if !ok || serverAddr == "" {
		t.Fatalf("socket.udp.localAddr = %#v, want non-empty string", serverAddrValue)
	}

	clientConnValue := awaitValue(t, callNative(t, udpNS, "connectAsync", serverAddr, Object{
		"timeoutMs": float64(3000),
	}))
	clientConn := mustRuntimeObject(t, clientConnValue, "socket.udp.connectAsync return")

	_ = awaitValue(t, callNative(t, clientConn, "sendAsync", "hello"))
	serverRecv := mustRuntimeObject(t, awaitValue(t, callNative(t, server, "recvFromAsync", float64(5))), "socket.udp.recvFromAsync(server)")
	if s, ok := serverRecv["data"].(string); !ok || s != "hello" {
		t.Fatalf("socket.udp.recvFromAsync(server).data = %#v, want hello", serverRecv["data"])
	}
	replyAddr, _ := serverRecv["addr"].(string)
	_ = awaitValue(t, callNative(t, server, "sendToAsync", "world", replyAddr))
	gotClient := awaitValue(t, callNative(t, clientConn, "recvAsync", float64(5)))
	if s, ok := gotClient.(string); !ok || s != "world" {
		t.Fatalf("socket.udp connect recvAsync = %#v, want world", gotClient)
	}

	_ = callNative(t, clientConn, "close")
	_ = callNative(t, server, "close")
}

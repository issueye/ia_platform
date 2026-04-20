package builtin

import (
	"testing"

	rt "ialang/pkg/lang/runtime"
)

func TestIPCModuleSendRecv(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ipcMod := mustModuleObject(t, modules, "ipc")
	serverNS := mustObject(t, ipcMod, "server")
	clientNS := mustObject(t, ipcMod, "client")

	serverValue := callNative(t, serverNS, "listen", Object{
		"addr": "127.0.0.1:0",
	})
	server := mustRuntimeObject(t, serverValue, "ipc.server.listen return")
	addr, ok := server["addr"].(string)
	if !ok || addr == "" {
		t.Fatalf("ipc.server.listen addr = %#v, want non-empty string", server["addr"])
	}

	acceptAwait := callNative(t, server, "acceptAsync")
	clientValue := callNative(t, clientNS, "connect", addr)
	clientConn := mustRuntimeObject(t, clientValue, "ipc.client.connect return")
	serverConnValue := awaitValue(t, acceptAwait)
	serverConn := mustRuntimeObject(t, serverConnValue, "ipc.server.accept return")

	_ = callNative(t, clientConn, "send", Object{
		"type": "ping",
		"n":    float64(7),
	})
	recvValue := callNative(t, serverConn, "recv")
	recvObj := mustRuntimeObject(t, recvValue, "ipc.conn.recv on server")
	if got, _ := recvObj["type"].(string); got != "ping" {
		t.Fatalf("ipc recv type = %#v, want ping", recvObj["type"])
	}
	if got, _ := recvObj["n"].(float64); got != 7 {
		t.Fatalf("ipc recv n = %#v, want 7", recvObj["n"])
	}

	_ = callNative(t, serverConn, "send", Object{
		"type": "pong",
		"ok":   true,
	})
	clientRecvValue := callNative(t, clientConn, "recv")
	clientRecvObj := mustRuntimeObject(t, clientRecvValue, "ipc.conn.recv on client")
	if got, _ := clientRecvObj["type"].(string); got != "pong" {
		t.Fatalf("ipc client recv type = %#v, want pong", clientRecvObj["type"])
	}

	_ = callNative(t, clientConn, "close")
	_ = callNative(t, serverConn, "close")
	_ = callNative(t, server, "close")
}

func TestIPCModuleCallReply(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ipcMod := mustModuleObject(t, modules, "ipc")
	serverNS := mustObject(t, ipcMod, "server")
	clientNS := mustObject(t, ipcMod, "client")

	serverValue := callNative(t, serverNS, "listen", Object{
		"addr": "127.0.0.1:0",
	})
	server := mustRuntimeObject(t, serverValue, "ipc.server.listen return")
	addr := server["addr"].(string)

	acceptAwait := callNative(t, server, "acceptAsync")
	clientValue := callNative(t, clientNS, "connect", addr)
	clientConn := mustRuntimeObject(t, clientValue, "ipc.client.connect return")
	serverConn := mustRuntimeObject(t, awaitValue(t, acceptAwait), "ipc.server.accept return")

	callAwait := callNative(t, clientConn, "callAsync", "sum", Object{
		"a": float64(2),
		"b": float64(5),
	})
	reqValue := callNative(t, serverConn, "recv")
	reqObj := mustRuntimeObject(t, reqValue, "ipc server recv request")
	if kind, _ := reqObj["kind"].(string); kind != ipcMessageKindReq {
		t.Fatalf("ipc request kind = %#v, want request", reqObj["kind"])
	}
	if method, _ := reqObj["method"].(string); method != "sum" {
		t.Fatalf("ipc request method = %#v, want sum", reqObj["method"])
	}
	_ = callNative(t, serverConn, "reply", reqObj, true, Object{
		"sum": float64(7),
	})

	respValue := awaitValue(t, callAwait)
	respObj := mustRuntimeObject(t, respValue, "ipc client call response")
	if kind, _ := respObj["kind"].(string); kind != ipcMessageKindResp {
		t.Fatalf("ipc response kind = %#v, want response", respObj["kind"])
	}
	if okFlag, _ := respObj["ok"].(bool); !okFlag {
		t.Fatalf("ipc response ok = %#v, want true", respObj["ok"])
	}
	payload := mustRuntimeObject(t, respObj["payload"], "ipc response payload")
	if sum, _ := payload["sum"].(float64); sum != 7 {
		t.Fatalf("ipc response payload.sum = %#v, want 7", payload["sum"])
	}

	_ = callNative(t, clientConn, "close")
	_ = callNative(t, serverConn, "close")
	_ = callNative(t, server, "close")
}

func TestIPCBuildRequestAndResponse(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	ipcMod := mustModuleObject(t, modules, "ipc")

	reqValue := callNative(t, ipcMod, "buildRequest", "hello", Object{
		"name": "ialang",
	}, Object{
		"id": "req-1",
	})
	reqObj := mustRuntimeObject(t, reqValue, "ipc.buildRequest return")
	if kind, _ := reqObj["kind"].(string); kind != ipcMessageKindReq {
		t.Fatalf("ipc.buildRequest kind = %#v, want request", reqObj["kind"])
	}
	if id, _ := reqObj["id"].(string); id != "req-1" {
		t.Fatalf("ipc.buildRequest id = %#v, want req-1", reqObj["id"])
	}

	respValue := callNative(t, ipcMod, "buildResponse", "req-1", false, nil, "bad input")
	respObj := mustRuntimeObject(t, respValue, "ipc.buildResponse return")
	if kind, _ := respObj["kind"].(string); kind != ipcMessageKindResp {
		t.Fatalf("ipc.buildResponse kind = %#v, want response", respObj["kind"])
	}
	if okFlag, _ := respObj["ok"].(bool); okFlag {
		t.Fatalf("ipc.buildResponse ok = %#v, want false", respObj["ok"])
	}
	if errText, _ := respObj["error"].(string); errText != "bad input" {
		t.Fatalf("ipc.buildResponse error = %#v, want bad input", respObj["error"])
	}
}

package builtin

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	rt "ialang/pkg/lang/runtime"
)

func TestIAXBuildRequestRouteModes(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	iaxMod := mustModuleObject(t, modules, "iax")

	dotReq := mustRuntimeObject(t, callNative(t, iaxMod, "buildRequest",
		"order", "create", Object{"ok": true},
	), "iax.buildRequest dot return")
	if route, _ := dotReq["route"].(string); route != "order.create" {
		t.Fatalf("dot route = %#v, want order.create", dotReq["route"])
	}

	slashReq := mustRuntimeObject(t, callNative(t, iaxMod, "buildRequest",
		"order", "create", Object{"ok": true}, Object{"routeMode": "slash"},
	), "iax.buildRequest slash return")
	if route, _ := slashReq["route"].(string); route != "/order/create" {
		t.Fatalf("slash route = %#v, want /order/create", slashReq["route"])
	}

	colonReq := mustRuntimeObject(t, callNative(t, iaxMod, "buildRequest",
		"order", "create", Object{"ok": true}, Object{"routeMode": "colon"},
	), "iax.buildRequest colon return")
	if route, _ := colonReq["route"].(string); route != "order:create" {
		t.Fatalf("colon route = %#v, want order:create", colonReq["route"])
	}

	expressReq := mustRuntimeObject(t, callNative(t, iaxMod, "buildRequest",
		"order", "create", Object{"ok": true}, Object{"routeMode": "express"},
	), "iax.buildRequest express return")
	if route, _ := expressReq["route"].(string); route != "POST /order/create" {
		t.Fatalf("express route = %#v, want POST /order/create", expressReq["route"])
	}

	expressTemplateReq := mustRuntimeObject(t, callNative(t, iaxMod, "buildRequest",
		"order", "create", Object{"ok": true}, Object{
			"routeMode":     "express",
			"routeMethod":   "patch",
			"routeTemplate": "/v1/:service/:action",
			"routePrefix":   "/api",
		},
	), "iax.buildRequest express template return")
	if route, _ := expressTemplateReq["route"].(string); route != "PATCH /api/v1/order/create" {
		t.Fatalf("express template route = %#v, want PATCH /api/v1/order/create", expressTemplateReq["route"])
	}
}

func TestIAXCallReceiveReplyFlow(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	iaxMod := mustModuleObject(t, modules, "iax")
	ipcMod := mustModuleObject(t, modules, "ipc")
	serverNS := mustObject(t, ipcMod, "server")
	clientNS := mustObject(t, ipcMod, "client")

	serverValue := callNative(t, serverNS, "listen", Object{
		"addr": "127.0.0.1:0",
	})
	server := mustRuntimeObject(t, serverValue, "ipc.server.listen return")
	addr := server["addr"].(string)

	acceptAwait := callNative(t, server, "acceptAsync")
	clientConnValue := callNative(t, clientNS, "connect", addr)
	clientConn := mustRuntimeObject(t, clientConnValue, "ipc.client.connect return")
	serverConn := mustRuntimeObject(t, awaitValue(t, acceptAwait), "ipc.server.accept return")

	callAwait := callNative(t, iaxMod, "callAsync",
		clientConn,
		"inventory",
		"reserve",
		Object{
			"sku": "SKU-42",
			"qty": float64(3),
		},
		Object{
			"requestOptions": Object{
				"from":        "producer-app",
				"routeMode":   "express",
				"routeMethod": "post",
				"routePrefix": "/api",
			},
		},
	)

	recvValue := callNative(t, iaxMod, "receive", serverConn)
	recvObj := mustRuntimeObject(t, recvValue, "iax.receive return")
	if ok, _ := recvObj["ok"].(bool); !ok {
		t.Fatalf("iax.receive ok = %#v, want true, code=%v", recvObj["ok"], recvObj["code"])
	}
	if service, _ := recvObj["service"].(string); service != "inventory" {
		t.Fatalf("iax.receive service = %#v, want inventory", recvObj["service"])
	}
	if action, _ := recvObj["action"].(string); action != "reserve" {
		t.Fatalf("iax.receive action = %#v, want reserve", recvObj["action"])
	}
	if route, _ := recvObj["route"].(string); route != "POST /api/inventory/reserve" {
		t.Fatalf("iax.receive route = %#v, want POST /api/inventory/reserve", recvObj["route"])
	}
	if routePath, _ := recvObj["routePath"].(string); routePath != "/api/inventory/reserve" {
		t.Fatalf("iax.receive routePath = %#v, want /api/inventory/reserve", recvObj["routePath"])
	}
	payload := mustRuntimeObject(t, recvObj["payload"], "iax.receive payload")
	if sku, _ := payload["sku"].(string); sku != "SKU-42" {
		t.Fatalf("iax.receive payload.sku = %#v, want SKU-42", payload["sku"])
	}

	_ = callNative(t, iaxMod, "reply",
		serverConn,
		recvObj,
		true,
		Object{
			"reserved": true,
			"left":     float64(97),
		},
	)

	callResult := mustRuntimeObject(t, awaitValue(t, callAwait), "iax.call return")
	if ok, _ := callResult["ok"].(bool); !ok {
		t.Fatalf("iax.call ok = %#v, want true, code=%v", callResult["ok"], callResult["code"])
	}
	dataObj := mustRuntimeObject(t, callResult["data"], "iax.call data")
	if reserved, _ := dataObj["reserved"].(bool); !reserved {
		t.Fatalf("iax.call data.reserved = %#v, want true", dataObj["reserved"])
	}

	_ = callNative(t, clientConn, "close")
	_ = callNative(t, serverConn, "close")
	_ = callNative(t, server, "close")
}

func TestIAXCallReceiveReplyFlowRemoteError(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	iaxMod := mustModuleObject(t, modules, "iax")
	ipcMod := mustModuleObject(t, modules, "ipc")
	serverNS := mustObject(t, ipcMod, "server")
	clientNS := mustObject(t, ipcMod, "client")

	serverValue := callNative(t, serverNS, "listen", Object{
		"addr": "127.0.0.1:0",
	})
	server := mustRuntimeObject(t, serverValue, "ipc.server.listen return")
	addr := server["addr"].(string)

	acceptAwait := callNative(t, server, "acceptAsync")
	clientConnValue := callNative(t, clientNS, "connect", addr)
	clientConn := mustRuntimeObject(t, clientConnValue, "ipc.client.connect return")
	serverConn := mustRuntimeObject(t, awaitValue(t, acceptAwait), "ipc.server.accept return")

	callAwait := callNative(t, iaxMod, "callAsync",
		clientConn,
		"payment",
		"charge",
		Object{"amount": float64(10)},
	)

	recvValue := callNative(t, iaxMod, "receive", serverConn)
	recvObj := mustRuntimeObject(t, recvValue, "iax.receive return")
	_ = callNative(t, iaxMod, "reply",
		serverConn,
		recvObj,
		false,
		nil,
		"insufficient balance",
	)

	callResult := mustRuntimeObject(t, awaitValue(t, callAwait), "iax.call return")
	if ok, _ := callResult["ok"].(bool); ok {
		t.Fatalf("iax.call ok = %#v, want false", callResult["ok"])
	}
	if code, _ := callResult["code"].(string); code != "REMOTE_ERROR" {
		t.Fatalf("iax.call code = %#v, want REMOTE_ERROR", callResult["code"])
	}
	if message, _ := callResult["message"].(string); message != "insufficient balance" {
		t.Fatalf("iax.call message = %#v, want insufficient balance", callResult["message"])
	}

	_ = callNative(t, clientConn, "close")
	_ = callNative(t, serverConn, "close")
	_ = callNative(t, server, "close")
}

func TestIAXBuildEvent(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	iaxMod := mustModuleObject(t, modules, "iax")

	event := mustRuntimeObject(t, callNative(t, iaxMod, "buildEvent",
		"orders.created",
		Object{"orderId": "A-1"},
		Object{"from": "order-service"},
	), "iax.buildEvent return")

	if protocol, _ := event["protocol"].(string); protocol != "iax/1" {
		t.Fatalf("iax.buildEvent protocol = %#v, want iax/1", event["protocol"])
	}
	if topic, _ := event["topic"].(string); topic != "orders.created" {
		t.Fatalf("iax.buildEvent topic = %#v, want orders.created", event["topic"])
	}
	if from, _ := event["from"].(string); from != "order-service" {
		t.Fatalf("iax.buildEvent from = %#v, want order-service", event["from"])
	}
}

func TestIAXPublishSubscribeWithTopic(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	iaxMod := mustModuleObject(t, modules, "iax")
	ipcMod := mustModuleObject(t, modules, "ipc")
	serverNS := mustObject(t, ipcMod, "server")
	clientNS := mustObject(t, ipcMod, "client")

	serverValue := callNative(t, serverNS, "listen", Object{
		"addr": "127.0.0.1:0",
	})
	server := mustRuntimeObject(t, serverValue, "ipc.server.listen return")
	addr := server["addr"].(string)

	acceptAwait := callNative(t, server, "acceptAsync")
	clientConnValue := callNative(t, clientNS, "connect", addr)
	clientConn := mustRuntimeObject(t, clientConnValue, "ipc.client.connect return")
	serverConn := mustRuntimeObject(t, awaitValue(t, acceptAwait), "ipc.server.accept return")

	sub := mustRuntimeObject(t, callNative(t, iaxMod, "subscribe",
		serverConn,
		"orders.*",
	), "iax.subscribe return")

	_ = callNative(t, iaxMod, "publish",
		clientConn,
		"orders.created",
		Object{"orderId": "A-200"},
		Object{"from": "order-service"},
	)
	next := callNative(t, sub, "next")
	eventObj := mustRuntimeObject(t, next, "iax.subscription.next return")
	if topic, _ := eventObj["topic"].(string); topic != "orders.created" {
		t.Fatalf("iax.subscribe next topic = %#v, want orders.created", eventObj["topic"])
	}
	payload := mustRuntimeObject(t, eventObj["payload"], "iax.subscribe next payload")
	if orderID, _ := payload["orderId"].(string); orderID != "A-200" {
		t.Fatalf("iax.subscribe payload.orderId = %#v, want A-200", payload["orderId"])
	}

	_ = callNative(t, sub, "close")
	_ = callNative(t, clientConn, "close")
	_ = callNative(t, serverConn, "close")
	_ = callNative(t, server, "close")
}

func TestIAXWorksWithSendRecvStringSocket(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	iaxMod := mustModuleObject(t, modules, "iax")

	clientConn, serverConn := mockStringSocketPair()

	callAwait := callNative(t, iaxMod, "callAsync",
		clientConn,
		"billing",
		"charge",
		Object{"amount": float64(23)},
		Object{
			"requestOptions": Object{
				"routeMode":   "express",
				"routeMethod": "post",
				"routePrefix": "/api",
			},
		},
	)

	recv := mustRuntimeObject(t, callNative(t, iaxMod, "receive", serverConn), "iax.receive over socket")
	if route, _ := recv["route"].(string); route != "POST /api/billing/charge" {
		t.Fatalf("iax.receive route over socket = %#v, want POST /api/billing/charge", recv["route"])
	}
	_ = callNative(t, iaxMod, "reply", serverConn, recv, true, Object{"ok": true})

	callResult := mustRuntimeObject(t, awaitValue(t, callAwait), "iax.call over socket")
	if ok, _ := callResult["ok"].(bool); !ok {
		t.Fatalf("iax.call over socket ok = %#v, want true", callResult["ok"])
	}

	sub := mustRuntimeObject(t, callNative(t, iaxMod, "subscribe", serverConn, "orders.*"), "iax.subscribe over socket")
	_ = callNative(t, iaxMod, "publish", clientConn, "orders.created", Object{"id": "A-1"})
	event := mustRuntimeObject(t, callNative(t, sub, "next"), "iax.subscribe next over socket")
	if topic, _ := event["topic"].(string); topic != "orders.created" {
		t.Fatalf("iax.subscribe topic over socket = %#v, want orders.created", event["topic"])
	}
}

func TestIAXPersistencePublishLoadReplay(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	iaxMod := mustModuleObject(t, modules, "iax")

	logPath := filepath.Join(t.TempDir(), "iax-events.jsonl")
	t.Cleanup(func() {
		_ = callNative(t, iaxMod, "configurePersistence", Object{
			"enabled": false,
			"path":    "",
		})
	})
	_ = callNative(t, iaxMod, "configurePersistence", Object{
		"enabled": true,
		"path":    logPath,
	})

	clientConn, serverConn := mockStringSocketPair()
	_ = callNative(t, iaxMod, "publish", clientConn, "orders.created", Object{"id": "A-1"})
	_ = callNative(t, iaxMod, "publish", clientConn, "orders.cancelled", Object{"id": "A-2"})

	eventsVal := callNative(t, iaxMod, "loadEvents", Object{
		"path":   logPath,
		"topics": "orders.*",
	})
	events, ok := eventsVal.(Array)
	if !ok || len(events) < 2 {
		t.Fatalf("iax.loadEvents = %#v, want >=2 events", eventsVal)
	}

	sub := mustRuntimeObject(t, callNative(t, iaxMod, "subscribe", serverConn, "orders.cancelled"), "iax.subscribe replay")
	replayResult := mustRuntimeObject(t, callNative(t, iaxMod, "replay", clientConn, Object{
		"path":   logPath,
		"topics": "orders.cancelled",
	}), "iax.replay result")
	if count, _ := replayResult["count"].(float64); count != 1 {
		t.Fatalf("iax.replay count = %#v, want 1", replayResult["count"])
	}

	ev := mustRuntimeObject(t, callNative(t, sub, "next"), "iax.subscription.next replay")
	if topic, _ := ev["topic"].(string); topic != "orders.cancelled" {
		t.Fatalf("replayed topic = %#v, want orders.cancelled", ev["topic"])
	}

	stat, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("persistence leveldb path should exist, err=%v", err)
	}
	if !stat.IsDir() {
		t.Fatalf("persistence path should be leveldb directory, got file mode=%v", stat.Mode())
	}
}

func TestIAXConcurrentPublishWithPersistence(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	iaxMod := mustModuleObject(t, modules, "iax")

	logPath := filepath.Join(t.TempDir(), "iax-events-concurrent.jsonl")
	t.Cleanup(func() {
		_ = callNative(t, iaxMod, "configurePersistence", Object{
			"enabled": false,
			"path":    "",
		})
	})
	_ = callNative(t, iaxMod, "configurePersistence", Object{
		"enabled": true,
		"path":    logPath,
	})

	clientConn, serverConn := mockStringSocketPair()
	sub := mustRuntimeObject(t, callNative(t, iaxMod, "subscribe", serverConn, "orders.*"), "iax.subscribe concurrent")

	const workers = 6
	const perWorker = 20
	total := workers * perWorker

	consumeDone := make(chan error, 1)
	go func() {
		seen := 0
		for seen < total {
			if _, err := callNativeWithError(sub, "next"); err != nil {
				consumeDone <- err
				return
			}
			seen++
		}
		consumeDone <- nil
	}()

	var wg sync.WaitGroup
	pubErr := make(chan error, workers)
	for w := 0; w < workers; w++ {
		wid := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				if _, err := callNativeWithError(iaxMod, "publish",
					clientConn,
					"orders.created",
					Object{"worker": float64(wid), "index": float64(i)},
				); err != nil {
					pubErr <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(pubErr)
	for err := range pubErr {
		if err != nil {
			t.Fatalf("concurrent publish error: %v", err)
		}
	}
	if err := <-consumeDone; err != nil {
		t.Fatalf("concurrent consume error: %v", err)
	}

	eventsVal, err := callNativeWithError(iaxMod, "loadEvents", Object{
		"path":  logPath,
		"topic": "orders.created",
	})
	if err != nil {
		t.Fatalf("iax.loadEvents concurrent error: %v", err)
	}
	events, ok := eventsVal.(Array)
	if !ok {
		t.Fatalf("iax.loadEvents concurrent type = %T, want Array", eventsVal)
	}
	if len(events) != total {
		t.Fatalf("iax.loadEvents concurrent count = %d, want %d", len(events), total)
	}
}

func TestIAXConcurrentReceiveAndSubscribeOnSameConnection(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	iaxMod := mustModuleObject(t, modules, "iax")

	clientConn, serverConn := mockStringSocketPair()
	sub := mustRuntimeObject(t, callNative(t, iaxMod, "subscribe", serverConn, "orders.*"), "iax.subscribe same conn")

	eventCh := make(chan Object, 1)
	errCh := make(chan error, 1)
	go func() {
		v, err := callNativeWithError(sub, "next")
		if err != nil {
			errCh <- err
			return
		}
		ev, ok := v.(Object)
		if !ok {
			errCh <- fmt.Errorf("subscription next type = %T, want Object", v)
			return
		}
		eventCh <- ev
	}()

	callAwait := callNative(t, iaxMod, "callAsync",
		clientConn,
		"inventory",
		"reserve",
		Object{"sku": "SKU-100", "qty": float64(1)},
	)

	recv, err := callNativeWithError(iaxMod, "receive", serverConn)
	if err != nil {
		t.Fatalf("iax.receive concurrent error: %v", err)
	}
	recvObj, ok := recv.(Object)
	if !ok {
		t.Fatalf("iax.receive concurrent type = %T, want Object", recv)
	}
	_, _ = callNativeWithError(iaxMod, "reply", serverConn, recvObj, true, Object{"ok": true})

	_, _ = callNativeWithError(iaxMod, "publish", clientConn, "orders.created", Object{"id": "A-300"})
	_ = mustRuntimeObject(t, awaitValue(t, callAwait), "iax.call concurrent result")

	select {
	case err := <-errCh:
		t.Fatalf("iax.subscribe concurrent next error: %v", err)
	case ev := <-eventCh:
		if topic, _ := ev["topic"].(string); topic != "orders.created" {
			t.Fatalf("iax.subscribe concurrent topic = %#v, want orders.created", ev["topic"])
		}
	}
}

func TestIAXReaderWriterSplitHighConcurrency(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	iaxMod := mustModuleObject(t, modules, "iax")

	clientConn, serverConn := mockStringSocketPair()
	sub := mustRuntimeObject(t, callNative(t, iaxMod, "subscribe", serverConn, "orders.*"), "iax.subscribe high concurrency")

	const reqCount = 40
	const evtCount = 80

	serverReqErr := make(chan error, 1)
	go func() {
		for i := 0; i < reqCount; i++ {
			recv, err := callNativeWithError(iaxMod, "receive", serverConn)
			if err != nil {
				serverReqErr <- err
				return
			}
			recvObj, ok := recv.(Object)
			if !ok {
				serverReqErr <- fmt.Errorf("iax.receive type = %T, want Object", recv)
				return
			}
			if _, err := callNativeWithError(iaxMod, "reply", serverConn, recvObj, true, Object{"ok": true}); err != nil {
				serverReqErr <- err
				return
			}
		}
		serverReqErr <- nil
	}()

	serverEvtErr := make(chan error, 1)
	go func() {
		for i := 0; i < evtCount; i++ {
			if _, err := callNativeWithError(sub, "next"); err != nil {
				serverEvtErr <- err
				return
			}
		}
		serverEvtErr <- nil
	}()

	clientErr := make(chan error, reqCount+evtCount)
	var wg sync.WaitGroup
	for i := 0; i < reqCount; i++ {
		idx := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := callNativeWithError(iaxMod, "call",
				clientConn,
				"inventory",
				"reserve",
				Object{"sku": "SKU", "index": float64(idx)},
			)
			if err != nil {
				clientErr <- err
			}
		}()
	}
	for i := 0; i < evtCount; i++ {
		idx := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := callNativeWithError(iaxMod, "publish",
				clientConn,
				"orders.created",
				Object{"id": float64(idx)},
			)
			if err != nil {
				clientErr <- err
			}
		}()
	}
	wg.Wait()
	close(clientErr)
	for err := range clientErr {
		if err != nil {
			t.Fatalf("client concurrent error: %v", err)
		}
	}

	if err := <-serverReqErr; err != nil {
		t.Fatalf("server request loop error: %v", err)
	}
	if err := <-serverEvtErr; err != nil {
		t.Fatalf("server event loop error: %v", err)
	}
}

func mockStringSocketPair() (Object, Object) {
	chAB := make(chan string, 16)
	chBA := make(chan string, 16)

	build := func(out chan<- string, in <-chan string) Object {
		sendFn := NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("mock.socket.send expects 1 arg")
			}
			msg, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("mock.socket.send expects string, got %T", args[0])
			}
			out <- msg
			return true, nil
		})
		recvFn := NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("mock.socket.recv expects 0 args")
			}
			msg := <-in
			return msg, nil
		})
		closeFn := NativeFunction(func(args []Value) (Value, error) {
			return true, nil
		})
		return Object{
			"send":  sendFn,
			"recv":  recvFn,
			"close": closeFn,
		}
	}

	return build(chAB, chBA), build(chBA, chAB)
}

package builtin

import (
	"path/filepath"
	"testing"
	"time"

	rt "ialang/pkg/lang/runtime"
)

func BenchmarkIAXPublishParallel(b *testing.B) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	iaxMod, ok := modules["iax"].(Object)
	if !ok {
		b.Fatalf("module iax type = %T, want Object", modules["iax"])
	}

	clientConn, serverConn := mockStringSocketPair()
	subVal, err := callNativeWithError(iaxMod, "subscribe", serverConn, "orders.*")
	if err != nil {
		b.Fatalf("iax.subscribe setup error: %v", err)
	}
	sub, ok := subVal.(Object)
	if !ok {
		b.Fatalf("iax.subscribe return type = %T, want Object", subVal)
	}

	consumerDone := make(chan error, 1)
	go func() {
		for i := 0; i < b.N; i++ {
			if _, err := callNativeWithError(sub, "next"); err != nil {
				consumerDone <- err
				return
			}
		}
		consumerDone <- nil
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := callNativeWithError(iaxMod, "publish",
				clientConn,
				"orders.created",
				Object{"id": float64(1)},
			); err != nil {
				b.Errorf("iax.publish benchmark error: %v", err)
				return
			}
		}
	})
	b.StopTimer()

	select {
	case err := <-consumerDone:
		if err != nil {
			b.Fatalf("consume benchmark error: %v", err)
		}
	case <-time.After(10 * time.Second):
		b.Fatal("consume benchmark timeout")
	}
}

func BenchmarkIAXPublishParallelWithPersistence(b *testing.B) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	iaxMod, ok := modules["iax"].(Object)
	if !ok {
		b.Fatalf("module iax type = %T, want Object", modules["iax"])
	}

	logPath := filepath.Join(b.TempDir(), "iax-bench-leveldb")
	if _, err := callNativeWithError(iaxMod, "configurePersistence", Object{
		"enabled": true,
		"path":    logPath,
	}); err != nil {
		b.Fatalf("iax.configurePersistence setup error: %v", err)
	}
	b.Cleanup(func() {
		_, _ = callNativeWithError(iaxMod, "configurePersistence", Object{
			"enabled": false,
			"path":    "",
		})
	})

	clientConn, serverConn := mockStringSocketPair()
	subVal, err := callNativeWithError(iaxMod, "subscribe", serverConn, "orders.*")
	if err != nil {
		b.Fatalf("iax.subscribe setup error: %v", err)
	}
	sub, ok := subVal.(Object)
	if !ok {
		b.Fatalf("iax.subscribe return type = %T, want Object", subVal)
	}

	consumerDone := make(chan error, 1)
	go func() {
		for i := 0; i < b.N; i++ {
			if _, err := callNativeWithError(sub, "next"); err != nil {
				consumerDone <- err
				return
			}
		}
		consumerDone <- nil
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := callNativeWithError(iaxMod, "publish",
				clientConn,
				"orders.created",
				Object{"id": float64(1)},
			); err != nil {
				b.Errorf("iax.publish persistence benchmark error: %v", err)
				return
			}
		}
	})
	b.StopTimer()

	select {
	case err := <-consumerDone:
		if err != nil {
			b.Fatalf("consume benchmark error: %v", err)
		}
	case <-time.After(10 * time.Second):
		b.Fatal("consume benchmark timeout")
	}
}

func BenchmarkIAXCallParallel(b *testing.B) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	iaxMod, ok := modules["iax"].(Object)
	if !ok {
		b.Fatalf("module iax type = %T, want Object", modules["iax"])
	}

	clientConn, serverConn := mockStringSocketPair()
	serverDone := make(chan error, 1)
	go func() {
		for i := 0; i < b.N; i++ {
			recvVal, err := callNativeWithError(iaxMod, "receive", serverConn)
			if err != nil {
				serverDone <- err
				return
			}
			recvObj, ok := recvVal.(Object)
			if !ok {
				serverDone <- &testError{msg: "iax.receive return is not Object"}
				return
			}
			if _, err := callNativeWithError(iaxMod, "reply", serverConn, recvObj, true, Object{"ok": true}); err != nil {
				serverDone <- err
				return
			}
		}
		serverDone <- nil
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := callNativeWithError(iaxMod, "call",
				clientConn,
				"inventory",
				"reserve",
				Object{"sku": "SKU-1", "qty": float64(1)},
			); err != nil {
				b.Errorf("iax.call benchmark error: %v", err)
				return
			}
		}
	})
	b.StopTimer()

	select {
	case err := <-serverDone:
		if err != nil {
			b.Fatalf("server benchmark error: %v", err)
		}
	case <-time.After(10 * time.Second):
		b.Fatal("server benchmark timeout")
	}
}

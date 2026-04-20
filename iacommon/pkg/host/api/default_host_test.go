package api

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	hostfs "iacommon/pkg/host/fs"
)

func TestDefaultHostAcquireAndCallFS(t *testing.T) {
	host := &DefaultHost{FS: newTestFSProvider(t, false)}
	ctx := context.Background()

	capability, err := host.AcquireCapability(ctx, AcquireRequest{
		Kind: CapabilityFS,
		Config: map[string]any{
			"rights": []string{"read", "write"},
		},
	})
	if err != nil {
		t.Fatalf("acquire fs capability: %v", err)
	}
	if capability.ID == "" {
		t.Fatal("expected capability id to be assigned")
	}
	if capability.Kind != CapabilityFS {
		t.Fatalf("unexpected capability kind: %s", capability.Kind)
	}

	_, err = host.Call(ctx, CallRequest{
		CapabilityID: capability.ID,
		Operation:    "fs.write_file",
		Args: map[string]any{
			"path":   "/workspace/hello.txt",
			"data":   "hello",
			"create": true,
			"trunc":  true,
		},
	})
	if err != nil {
		t.Fatalf("write file through host: %v", err)
	}

	result, err := host.Call(ctx, CallRequest{
		CapabilityID: capability.ID,
		Operation:    "fs.read_file",
		Args:         map[string]any{"path": "/workspace/hello.txt"},
	})
	if err != nil {
		t.Fatalf("read file through host: %v", err)
	}

	data, ok := result.Value["data"].([]byte)
	if !ok {
		t.Fatalf("expected []byte result, got %#v", result.Value["data"])
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected file contents: %q", string(data))
	}

	if err := host.ReleaseCapability(ctx, capability.ID); err != nil {
		t.Fatalf("release capability: %v", err)
	}

	_, err = host.Call(ctx, CallRequest{
		CapabilityID: capability.ID,
		Operation:    "fs.read_file",
		Args:         map[string]any{"path": "/workspace/hello.txt"},
	})
	if !errors.Is(err, ErrCapabilityNotFound) {
		t.Fatalf("expected ErrCapabilityNotFound after release, got %v", err)
	}
}

func TestDefaultHostRejectsAcquireWithoutProvider(t *testing.T) {
	host := &DefaultHost{}

	_, err := host.AcquireCapability(context.Background(), AcquireRequest{Kind: CapabilityFS})
	if !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("expected ErrProviderUnavailable, got %v", err)
	}
}

func TestDefaultHostPropagatesReadOnlyFSRestrictions(t *testing.T) {
	host := &DefaultHost{FS: newTestFSProvider(t, true)}
	capability, err := host.AcquireCapability(context.Background(), AcquireRequest{Kind: CapabilityFS})
	if err != nil {
		t.Fatalf("acquire fs capability: %v", err)
	}

	_, err = host.Call(context.Background(), CallRequest{
		CapabilityID: capability.ID,
		Operation:    "fs.write_file",
		Args: map[string]any{
			"path":   "/workspace/blocked.txt",
			"data":   []byte("blocked"),
			"create": true,
			"trunc":  true,
		},
	})
	if !errors.Is(err, hostfs.ErrReadOnlyPreopen) {
		t.Fatalf("expected ErrReadOnlyPreopen, got %v", err)
	}
}

func newTestFSProvider(t *testing.T, readOnly bool) hostfs.Provider {
	t.Helper()

	root := filepath.Join(t.TempDir(), "workspace")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}

	mapper, err := hostfs.NewPreopenPathMapper([]hostfs.Preopen{{
		VirtualPath: "/workspace",
		RealPath:    root,
		ReadOnly:    readOnly,
	}})
	if err != nil {
		t.Fatalf("create preopen mapper: %v", err)
	}

	return &hostfs.LocalFSProvider{Mapper: mapper}
}

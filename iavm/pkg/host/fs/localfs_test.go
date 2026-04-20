package fs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFSProviderReadWriteRoundTrip(t *testing.T) {
	provider := newTestLocalFSProvider(t, false)
	ctx := context.Background()

	if err := provider.WriteFile(ctx, "/workspace/hello.txt", []byte("hello"), WriteOptions{Create: true, Trunc: true}); err != nil {
		t.Fatalf("write file: %v", err)
	}

	data, err := provider.ReadFile(ctx, "/workspace/hello.txt")
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("read data mismatch: got %q", string(data))
	}

	entries, err := provider.ReadDir(ctx, "/workspace")
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "hello.txt" || entries[0].IsDir {
		t.Fatalf("unexpected dir entries: %+v", entries)
	}

	info, err := provider.Stat(ctx, "/workspace/hello.txt")
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	if info.Name != "hello.txt" || info.IsDir || info.Size != int64(len("hello")) {
		t.Fatalf("unexpected file info: %+v", info)
	}
}

func TestLocalFSProviderRejectsWriteOnReadOnlyPreopen(t *testing.T) {
	provider := newTestLocalFSProvider(t, true)

	err := provider.WriteFile(context.Background(), "/workspace/blocked.txt", []byte("blocked"), WriteOptions{Create: true, Trunc: true})
	if !errors.Is(err, ErrReadOnlyPreopen) {
		t.Fatalf("expected ErrReadOnlyPreopen, got %v", err)
	}
}

func TestLocalFSProviderRejectsPathOutsidePreopen(t *testing.T) {
	provider := newTestLocalFSProvider(t, false)

	_, err := provider.ReadFile(context.Background(), "/outside/file.txt")
	if !errors.Is(err, ErrPathNotMapped) {
		t.Fatalf("expected ErrPathNotMapped, got %v", err)
	}
}

func newTestLocalFSProvider(t *testing.T, readOnly bool) *LocalFSProvider {
	t.Helper()

	root := filepath.Join(t.TempDir(), "workspace")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}

	mapper, err := NewPreopenPathMapper([]Preopen{{
		VirtualPath: "/workspace",
		RealPath:    root,
		ReadOnly:    readOnly,
	}})
	if err != nil {
		t.Fatalf("create mapper: %v", err)
	}

	return &LocalFSProvider{Mapper: mapper}
}

package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPreopenPathMapperResolveLongestMatch(t *testing.T) {
	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspace")
	cacheDir := filepath.Join(workspaceDir, "cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("mkdir cache dir: %v", err)
	}

	mapper, err := NewPreopenPathMapper([]Preopen{
		{VirtualPath: "/workspace", RealPath: workspaceDir},
		{VirtualPath: "/workspace/cache", RealPath: cacheDir, ReadOnly: true},
	})
	if err != nil {
		t.Fatalf("create mapper: %v", err)
	}

	realPath, preopen, err := mapper.Resolve("/workspace/cache/data.txt")
	if err != nil {
		t.Fatalf("resolve path: %v", err)
	}

	expected := filepath.Join(cacheDir, "data.txt")
	if realPath != expected {
		t.Fatalf("resolved path mismatch: got %q want %q", realPath, expected)
	}
	if preopen.VirtualPath != "/workspace/cache" {
		t.Fatalf("matched preopen mismatch: got %q", preopen.VirtualPath)
	}
	if !preopen.ReadOnly {
		t.Fatal("expected matched preopen to be read-only")
	}
}

func TestPreopenPathMapperRejectsUnmappedPath(t *testing.T) {
	mapper, err := NewPreopenPathMapper([]Preopen{{VirtualPath: "/workspace", RealPath: t.TempDir()}})
	if err != nil {
		t.Fatalf("create mapper: %v", err)
	}

	_, _, err = mapper.Resolve("/outside/file.txt")
	if !errors.Is(err, ErrPathNotMapped) {
		t.Fatalf("expected ErrPathNotMapped, got %v", err)
	}
}

func TestPreopenPathMapperNormalizesRelativeWindowsStylePath(t *testing.T) {
	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("mkdir workspace dir: %v", err)
	}

	mapper, err := NewPreopenPathMapper([]Preopen{{VirtualPath: "/workspace", RealPath: workspaceDir}})
	if err != nil {
		t.Fatalf("create mapper: %v", err)
	}

	realPath, _, err := mapper.Resolve(`workspace\\nested\\file.txt`)
	if err != nil {
		t.Fatalf("resolve path: %v", err)
	}

	expected := filepath.Join(workspaceDir, "nested", "file.txt")
	if realPath != expected {
		t.Fatalf("resolved path mismatch: got %q want %q", realPath, expected)
	}
}

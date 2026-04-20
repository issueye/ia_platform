package ialang

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	hostapi "iacommon/pkg/host/api"
	hostfs "iacommon/pkg/host/fs"
)

func TestBuildPlatformFSModuleWithHost(t *testing.T) {
	host := &hostapi.DefaultHost{FS: newBridgeTestFSProvider(t, false)}
	module, err := BuildPlatformFSModuleWithHost(host)
	if err != nil {
		t.Fatalf("build fs module: %v", err)
	}

	writeFile := module["writeFile"].(FSWriteFileFunc)
	readFile := module["readFile"].(FSReadFileFunc)
	appendFile := module["appendFile"].(FSAppendFileFunc)
	exists := module["exists"].(FSExistsFunc)
	mkdir := module["mkdir"].(FSMkdirFunc)
	readDir := module["readDir"].(FSReadDirFunc)
	stat := module["stat"].(FSStatFunc)
	rename := module["rename"].(FSRenameFunc)
	remove := module["remove"].(FSRemoveFunc)
	copyFile := module["copy"].(FSCopyFunc)

	ok, err := mkdir("/workspace/nested", true)
	if err != nil || !ok {
		t.Fatalf("mkdir failed: ok=%v err=%v", ok, err)
	}

	ok, err = writeFile("/workspace/nested/hello.txt", "hello")
	if err != nil || !ok {
		t.Fatalf("writeFile failed: ok=%v err=%v", ok, err)
	}

	ok, err = appendFile("/workspace/nested/hello.txt", " world")
	if err != nil || !ok {
		t.Fatalf("appendFile failed: ok=%v err=%v", ok, err)
	}

	content, err := readFile("/workspace/nested/hello.txt")
	if err != nil {
		t.Fatalf("readFile failed: %v", err)
	}
	if content != "hello world" {
		t.Fatalf("unexpected content: %q", content)
	}

	present, err := exists("/workspace/nested/hello.txt")
	if err != nil {
		t.Fatalf("exists failed: %v", err)
	}
	if !present {
		t.Fatal("expected file to exist")
	}

	entries, err := readDir("/workspace/nested")
	if err != nil {
		t.Fatalf("readDir failed: %v", err)
	}
	if len(entries) != 1 || entries[0] != "hello.txt" {
		t.Fatalf("unexpected directory entries: %#v", entries)
	}

	info, err := stat("/workspace/nested/hello.txt")
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info["name"] != "hello.txt" {
		t.Fatalf("unexpected stat name: %#v", info)
	}

	ok, err = copyFile("/workspace/nested/hello.txt", "/workspace/nested/copy.txt")
	if err != nil || !ok {
		t.Fatalf("copy failed: ok=%v err=%v", ok, err)
	}

	ok, err = rename("/workspace/nested/copy.txt", "/workspace/nested/renamed.txt")
	if err != nil || !ok {
		t.Fatalf("rename failed: ok=%v err=%v", ok, err)
	}

	ok, err = remove("/workspace/nested/renamed.txt")
	if err != nil || !ok {
		t.Fatalf("remove failed: ok=%v err=%v", ok, err)
	}
}

func TestBuildPlatformFSModuleWithHostPropagatesReadOnlyRestriction(t *testing.T) {
	host := &hostapi.DefaultHost{FS: newBridgeTestFSProvider(t, true)}
	module, err := BuildPlatformFSModuleWithHost(host)
	if err != nil {
		t.Fatalf("build fs module: %v", err)
	}

	writeFile := module["writeFile"].(FSWriteFileFunc)
	_, err = writeFile("/workspace/blocked.txt", "blocked")
	if !errors.Is(err, hostfs.ErrReadOnlyPreopen) {
		t.Fatalf("expected ErrReadOnlyPreopen, got %v", err)
	}
}

func TestResolveModuleReturnsPlatformFSModule(t *testing.T) {
	resolved, err := ResolveModule(PlatformFSModuleName)
	if err != nil {
		t.Fatalf("resolve module: %v", err)
	}

	module, ok := resolved.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", resolved)
	}
	if _, ok := module["readFile"]; !ok {
		t.Fatal("expected readFile export in resolved module")
	}
	if _, ok := module["fs"]; !ok {
		t.Fatal("expected fs namespace in resolved module")
	}
}

func newBridgeTestFSProvider(t *testing.T, readOnly bool) hostfs.Provider {
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

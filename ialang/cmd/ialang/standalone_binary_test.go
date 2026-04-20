package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"ialang/pkg/lang/packagefile"
)

func TestAppendAndExtractEmbeddedPackagePayload(t *testing.T) {
	executableBytes := []byte{1, 2, 3, 4}
	pkgBytes := []byte("demo package bytes")

	combined, err := appendEmbeddedPackagePayload(executableBytes, pkgBytes)
	if err != nil {
		t.Fatalf("appendEmbeddedPackagePayload unexpected error: %v", err)
	}
	got, found, err := extractEmbeddedPackagePayload(combined)
	if err != nil {
		t.Fatalf("extractEmbeddedPackagePayload unexpected error: %v", err)
	}
	if !found {
		t.Fatal("extractEmbeddedPackagePayload should find embedded payload")
	}
	if !bytes.Equal(got, pkgBytes) {
		t.Fatalf("embedded payload mismatch\n got: %q\nwant: %q", string(got), string(pkgBytes))
	}
}

func TestExtractEmbeddedPackagePayloadNotFound(t *testing.T) {
	got, found, err := extractEmbeddedPackagePayload([]byte{1, 2, 3})
	if err != nil {
		t.Fatalf("extractEmbeddedPackagePayload unexpected error: %v", err)
	}
	if found {
		t.Fatal("extractEmbeddedPackagePayload should not find payload")
	}
	if got != nil {
		t.Fatalf("extractEmbeddedPackagePayload got payload %v, want nil", got)
	}
}

func TestExecuteBuildBinCommand(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "main.ia")
	helperPath := filepath.Join(dir, "helper.ia")
	outputPath := filepath.Join(dir, "standalone")
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}

	if err := os.WriteFile(helperPath, []byte("export let answer = 42;"), 0o644); err != nil {
		t.Fatalf("write helper file error: %v", err)
	}
	if err := os.WriteFile(entryPath, []byte("import { answer } from \"./helper\";\nprint(answer);\n"), 0o644); err != nil {
		t.Fatalf("write entry file error: %v", err)
	}

	var stderr bytes.Buffer
	if err := executeBuildBinCommand(entryPath, outputPath, &stderr); err != nil {
		t.Fatalf("executeBuildBinCommand unexpected error: %v", err)
	}

	outputBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read standalone output error: %v", err)
	}
	pkgBytes, found, err := extractEmbeddedPackagePayload(outputBytes)
	if err != nil {
		t.Fatalf("extractEmbeddedPackagePayload from output unexpected error: %v", err)
	}
	if !found {
		t.Fatal("standalone output missing embedded package payload")
	}
	pkg, err := packagefile.Decode(pkgBytes)
	if err != nil {
		t.Fatalf("decode embedded package unexpected error: %v", err)
	}
	if pkg.Entry == "" {
		t.Fatal("embedded package entry should not be empty")
	}
	if len(pkg.Modules) == 0 {
		t.Fatal("embedded package modules should not be empty")
	}
}

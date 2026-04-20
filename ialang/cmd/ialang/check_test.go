package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCLICheckCommandSuccess(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "main.ia")
	helperPath := filepath.Join(dir, "helper.ia")

	if err := os.WriteFile(helperPath, []byte("export let value = 1;"), 0o644); err != nil {
		t.Fatalf("write helper file error: %v", err)
	}
	if err := os.WriteFile(entryPath, []byte("import { value } from \"./helper\";\nlet x = value;\n"), 0o644); err != nil {
		t.Fatalf("write entry file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "check", entryPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI check code = %d, want 0, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "syntax check passed:") {
		t.Fatalf("stdout = %q, want syntax check passed message", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunCLICheckCommandParseFailure(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "main.ia")
	// Use a real syntax error: missing identifier after 'let'
	if err := os.WriteFile(entryPath, []byte("let = 1;"), 0o644); err != nil {
		t.Fatalf("write entry file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "check", entryPath}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI check code = %d, want 1", code)
	}
	stderrText := stderr.String()
	if !strings.Contains(stderrText, "parse errors in "+entryPath+":") {
		t.Fatalf("stderr = %q, want module unit output", stderrText)
	}
	if !strings.Contains(stderrText, "line 1, col") {
		t.Fatalf("stderr = %q, want line/column output", stderrText)
	}
}

func TestRunCLICheckCommandCompileFailureShowsUnit(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "main.ia")
	if err := os.WriteFile(entryPath, []byte("break;"), 0o644); err != nil {
		t.Fatalf("write entry file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "check", entryPath}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI check code = %d, want 1", code)
	}
	stderrText := stderr.String()
	if !strings.Contains(stderrText, "compile errors in "+entryPath+":") {
		t.Fatalf("stderr = %q, want compile unit output", stderrText)
	}
	if !strings.Contains(stderrText, "line 1, col 1") {
		t.Fatalf("stderr = %q, want compile error position output", stderrText)
	}
}

func TestRunCLICheckProjectDirUsesPkgTomlEntry(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "app")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir app dir error: %v", err)
	}

	pkgToml := "name = \"demo\"\nversion = \"0.1.0\"\nentry = \"app/main.ia\"\n"
	if err := os.WriteFile(filepath.Join(dir, "pkg.toml"), []byte(pkgToml), 0o644); err != nil {
		t.Fatalf("write pkg.toml error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "main.ia"), []byte("let ok = 1;\n"), 0o644); err != nil {
		t.Fatalf("write app main file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "check", dir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI check project dir code = %d, want 0, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "syntax check passed:") {
		t.Fatalf("stdout = %q, want syntax check passed message", stdout.String())
	}
}

func TestRunCLICheckProjectDirSupportsRootAliasImport(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "app")
	sharedDir := filepath.Join(dir, "shared")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("mkdir app dir error: %v", err)
	}
	if err := os.MkdirAll(sharedDir, 0o755); err != nil {
		t.Fatalf("mkdir shared dir error: %v", err)
	}

	pkgToml := "name = \"demo\"\nversion = \"0.1.0\"\nentry = \"app/main.ia\"\n\n[imports]\nroot_alias = \"@\"\n"
	if err := os.WriteFile(filepath.Join(dir, "pkg.toml"), []byte(pkgToml), 0o644); err != nil {
		t.Fatalf("write pkg.toml error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sharedDir, "helper.ia"), []byte("export let value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write helper file error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "main.ia"), []byte("import { value } from \"@/shared/helper\";\nlet x = value;\n"), 0o644); err != nil {
		t.Fatalf("write app main file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "check", dir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI check project dir code = %d, want 0, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "syntax check passed:") {
		t.Fatalf("stdout = %q, want syntax check passed message", stdout.String())
	}
}

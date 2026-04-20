package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCLIFmtCommandSuccess(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "main.ia")
	input := "let  x=1+2;function add(a,b){return a+b;}"
	if err := os.WriteFile(entryPath, []byte(input), 0o644); err != nil {
		t.Fatalf("write entry file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "fmt", entryPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI fmt code = %d, want 0, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "formatted: "+entryPath) {
		t.Fatalf("stdout = %q, want formatted message", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	got, err := os.ReadFile(entryPath)
	if err != nil {
		t.Fatalf("read formatted file error: %v", err)
	}
	want := "let x = 1 + 2;\n\nfunction add(a, b) {\n  return a + b;\n}\n"
	if string(got) != want {
		t.Fatalf("formatted content = %q, want %q", string(got), want)
	}
}

func TestRunCLIFmtCommandAlreadyFormatted(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "main.ia")
	input := "let x = 1;\n"
	if err := os.WriteFile(entryPath, []byte(input), 0o644); err != nil {
		t.Fatalf("write entry file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "fmt", entryPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI fmt code = %d, want 0, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "already formatted: "+entryPath) {
		t.Fatalf("stdout = %q, want already formatted message", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunCLIFmtCommandParseFailure(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "main.ia")
	if err := os.WriteFile(entryPath, []byte("let = 1;"), 0o644); err != nil {
		t.Fatalf("write entry file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "fmt", entryPath}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI fmt code = %d, want 1", code)
	}
	stderrText := stderr.String()
	if !strings.Contains(stderrText, "parse errors in "+entryPath+":") {
		t.Fatalf("stderr = %q, want parse unit output", stderrText)
	}
	if !strings.Contains(stderrText, "line 1, col") {
		t.Fatalf("stderr = %q, want line/column output", stderrText)
	}
}

package main

import (
	"bytes"
	"errors"
	bc "iacommon/pkg/ialang/bytecode"
	moduleapi "iacommon/pkg/ialang/module"
	"iacommon/pkg/ialang/packagefile"
	"ialang/pkg/lang"
	"iavm/pkg/binary"
	"iavm/pkg/core"
	"iavm/pkg/module"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseCLIArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "valid run", args: []string{"ialang", "run", "examples/hello.ia"}, wantErr: false},
		{name: "valid build", args: []string{"ialang", "build", "examples/hello.ia"}, wantErr: false},
		{name: "valid build with out flag", args: []string{"ialang", "build", "examples/hello.ia", "-o", "app.iapkg"}, wantErr: false},
		{name: "valid build-bin", args: []string{"ialang", "build-bin", "examples/hello.ia"}, wantErr: false},
		{name: "valid build-bin with out flag", args: []string{"ialang", "build-bin", "examples/hello.ia", "-o", "app.exe"}, wantErr: false},
		{name: "valid run-pkg", args: []string{"ialang", "run-pkg", "app.iapkg"}, wantErr: false},
		{name: "valid init default dir", args: []string{"ialang", "init"}, wantErr: false},
		{name: "valid init target dir", args: []string{"ialang", "init", "demo-app"}, wantErr: false},
		{name: "valid check default target", args: []string{"ialang", "check"}, wantErr: false},
		{name: "valid check target", args: []string{"ialang", "check", "examples/hello.ia"}, wantErr: false},
		{name: "valid fmt", args: []string{"ialang", "fmt", "examples/hello.ia"}, wantErr: false},
		{name: "valid verify-iavm", args: []string{"ialang", "verify-iavm", "app.iavm"}, wantErr: false},
		{name: "valid verify-iavm strict", args: []string{"ialang", "verify-iavm", "app.iavm", "--strict"}, wantErr: false},
		{name: "valid verify-iavm with limits", args: []string{"ialang", "verify-iavm", "app.iavm", "--max-functions", "2", "--max-constants", "3", "--max-code-size", "4", "--max-locals", "5", "--max-stack", "6", "--allow-capability", "fs"}, wantErr: false},
		{name: "valid inspect-iavm", args: []string{"ialang", "inspect-iavm", "app.iavm"}, wantErr: false},
		{name: "valid inspect-iavm verbose", args: []string{"ialang", "inspect-iavm", "app.iavm", "--verbose"}, wantErr: false},
		{name: "valid run-iavm", args: []string{"ialang", "run-iavm", "app.iavm"}, wantErr: false},
		{name: "valid run-iavm strict capability", args: []string{"ialang", "run-iavm", "app.iavm", "--strict", "--allow-capability", "network"}, wantErr: false},
		{name: "missing command", args: []string{"ialang"}, wantErr: true},
		{name: "unsupported command", args: []string{"ialang", "test"}, wantErr: true},
		{name: "run missing file", args: []string{"ialang", "run"}, wantErr: true},
		{name: "run with script args", args: []string{"ialang", "run", "a.ia", "b.ia"}, wantErr: false},
		{name: "run-pkg missing file", args: []string{"ialang", "run-pkg"}, wantErr: true},
		{name: "run-pkg with script args", args: []string{"ialang", "run-pkg", "a.iapkg", "b"}, wantErr: false},
		{name: "build missing entry", args: []string{"ialang", "build"}, wantErr: true},
		{name: "build out missing value", args: []string{"ialang", "build", "entry.ia", "-o"}, wantErr: true},
		{name: "build unknown option", args: []string{"ialang", "build", "entry.ia", "--bad"}, wantErr: true},
		{name: "build-bin missing entry", args: []string{"ialang", "build-bin"}, wantErr: true},
		{name: "build-bin out missing value", args: []string{"ialang", "build-bin", "entry.ia", "-o"}, wantErr: true},
		{name: "build-bin unknown option", args: []string{"ialang", "build-bin", "entry.ia", "--bad"}, wantErr: true},
		{name: "init too many args", args: []string{"ialang", "init", "a", "b"}, wantErr: true},
		{name: "check too many args", args: []string{"ialang", "check", "a", "b"}, wantErr: true},
		{name: "fmt missing file", args: []string{"ialang", "fmt"}, wantErr: false},
		{name: "fmt too many args", args: []string{"ialang", "fmt", "a.ia", "b.ia"}, wantErr: true},
		{name: "verify-iavm missing file", args: []string{"ialang", "verify-iavm"}, wantErr: true},
		{name: "verify-iavm unknown option", args: []string{"ialang", "verify-iavm", "app.iavm", "--bad"}, wantErr: true},
		{name: "verify-iavm duplicate strict", args: []string{"ialang", "verify-iavm", "app.iavm", "--strict", "--strict"}, wantErr: true},
		{name: "verify-iavm missing max-functions value", args: []string{"ialang", "verify-iavm", "app.iavm", "--max-functions"}, wantErr: true},
		{name: "verify-iavm invalid max-functions value", args: []string{"ialang", "verify-iavm", "app.iavm", "--max-functions", "0"}, wantErr: true},
		{name: "verify-iavm invalid capability", args: []string{"ialang", "verify-iavm", "app.iavm", "--allow-capability", "gpu"}, wantErr: true},
		{name: "inspect-iavm missing file", args: []string{"ialang", "inspect-iavm"}, wantErr: true},
		{name: "inspect-iavm unknown option", args: []string{"ialang", "inspect-iavm", "app.iavm", "--bad"}, wantErr: true},
		{name: "inspect-iavm duplicate verbose", args: []string{"ialang", "inspect-iavm", "app.iavm", "--verbose", "--verbose"}, wantErr: true},
		{name: "run-iavm missing file", args: []string{"ialang", "run-iavm"}, wantErr: true},
		{name: "run-iavm missing max-stack value", args: []string{"ialang", "run-iavm", "app.iavm", "--max-stack"}, wantErr: true},
		{name: "run-iavm invalid max-stack value", args: []string{"ialang", "run-iavm", "app.iavm", "--max-stack", "-1"}, wantErr: true},
		{name: "run-iavm invalid capability", args: []string{"ialang", "run-iavm", "app.iavm", "--allow-capability", "gpu"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCLIArgs(tt.args)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseCLIArgsRunAndRunPkgScriptArgs(t *testing.T) {
	runCmd, err := parseCLIArgs([]string{"ialang", "run", "examples/main.ia", "install", "pkg@1.0.0"})
	if err != nil {
		t.Fatalf("parse run args unexpected error: %v", err)
	}
	if runCmd.file != "examples/main.ia" {
		t.Fatalf("run file = %q, want examples/main.ia", runCmd.file)
	}
	if len(runCmd.args) != 2 || runCmd.args[0] != "install" || runCmd.args[1] != "pkg@1.0.0" {
		t.Fatalf("run args = %#v, want [install pkg@1.0.0]", runCmd.args)
	}

	runPkgCmd, err := parseCLIArgs([]string{"ialang", "run-pkg", "app.iapkg", "list"})
	if err != nil {
		t.Fatalf("parse run-pkg args unexpected error: %v", err)
	}
	if runPkgCmd.file != "app.iapkg" {
		t.Fatalf("run-pkg file = %q, want app.iapkg", runPkgCmd.file)
	}
	if len(runPkgCmd.args) != 1 || runPkgCmd.args[0] != "list" {
		t.Fatalf("run-pkg args = %#v, want [list]", runPkgCmd.args)
	}
}

func TestParseCLIArgsIavmVerifierOptions(t *testing.T) {
	cmd, err := parseCLIArgs([]string{
		"ialang", "verify-iavm", "app.iavm",
		"--strict",
		"--max-functions", "2",
		"--max-constants", "3",
		"--max-code-size", "4",
		"--max-locals", "5",
		"--max-stack", "6",
		"--allow-capability", "fs",
		"--allow-capability", "network",
	})
	if err != nil {
		t.Fatalf("parse verify-iavm options unexpected error: %v", err)
	}
	if !cmd.strict {
		t.Fatal("strict = false, want true")
	}
	if cmd.maxFunctions != 2 || cmd.maxConstants != 3 || cmd.maxCodeSize != 4 || cmd.maxLocals != 5 || cmd.maxStack != 6 {
		t.Fatalf("parsed limits = %+v, want maxFunctions=2 maxConstants=3 maxCodeSize=4 maxLocals=5 maxStack=6", cmd)
	}
	if len(cmd.allowedCapabilities) != 2 || cmd.allowedCapabilities[0] != module.CapabilityFS || cmd.allowedCapabilities[1] != module.CapabilityNetwork {
		t.Fatalf("allowedCapabilities = %#v, want [fs network]", cmd.allowedCapabilities)
	}
}

func TestRunCLIInvalidArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "usage:") {
		t.Fatalf("stderr = %q, want usage message", stderr.String())
	}
}

func TestRunCLIRunFileNotFound(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "run", "__not_found__.ia"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "read file error:") {
		t.Fatalf("stderr = %q, want read file error", stderr.String())
	}
}

func TestRunCLIRunParseErrorIncludesUnitAndPosition(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "main.ia")
	// Use a real syntax error: missing identifier after 'let'
	if err := os.WriteFile(entryPath, []byte("let = 1;"), 0o644); err != nil {
		t.Fatalf("write entry file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "run", entryPath}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI run code = %d, want 1", code)
	}
	stderrText := stderr.String()
	if !strings.Contains(stderrText, "parse errors in "+entryPath+":") {
		t.Fatalf("stderr = %q, want parse unit output", stderrText)
	}
	if !strings.Contains(stderrText, "line 1, col") {
		t.Fatalf("stderr = %q, want line/column output", stderrText)
	}
}

func TestRunCLIVerifyIavmSuccess(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "app.iavm")

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code:      []core.Instruction{{Op: core.OpReturn}},
			},
		},
	}
	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule unexpected error: %v", err)
	}
	if err := os.WriteFile(modulePath, data, 0o644); err != nil {
		t.Fatalf("write module file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "verify-iavm", modulePath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI verify-iavm code = %d, want 0, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "module verification passed:") {
		t.Fatalf("stdout = %q, want success output", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunCLIVerifyIavmStrictRequiresEntry(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "app.iavm")

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "helper",
				TypeIndex: 0,
				Code:      []core.Instruction{{Op: core.OpReturn}},
			},
		},
	}
	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule unexpected error: %v", err)
	}
	if err := os.WriteFile(modulePath, data, 0o644); err != nil {
		t.Fatalf("write module file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "verify-iavm", modulePath, "--strict"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI verify-iavm strict code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "verify module error: no entry point function found") {
		t.Fatalf("stderr = %q, want strict verification error", stderr.String())
	}
}

func TestRunCLIVerifyIavmFileNotFound(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "verify-iavm", "__not_found__.iavm"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI verify-iavm code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "read file error:") {
		t.Fatalf("stderr = %q, want read file error", stderr.String())
	}
}

func TestRunCLIVerifyIavmDecodeError(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "bad.iavm")
	if err := os.WriteFile(modulePath, []byte("not a valid iavm module"), 0o644); err != nil {
		t.Fatalf("write module file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "verify-iavm", modulePath}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI verify-iavm code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "decode module error:") {
		t.Fatalf("stderr = %q, want decode error", stderr.String())
	}
}

func TestRunCLIVerifyIavmLimitAndCapabilityOptions(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "limits.iavm")

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{Kind: module.CapabilityFS},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code:      []core.Instruction{{Op: core.OpReturn}},
			},
		},
	}
	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule unexpected error: %v", err)
	}
	if err := os.WriteFile(modulePath, data, 0o644); err != nil {
		t.Fatalf("write module file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "verify-iavm", modulePath, "--max-functions", "1", "--max-code-size", "1", "--allow-capability", "fs"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI verify-iavm with limits code = %d, want 0, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "module verification passed:") {
		t.Fatalf("stdout = %q, want success output", stdout.String())
	}
}

func TestRunCLIVerifyIavmCapabilityDenied(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "cap-denied.iavm")

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{Kind: module.CapabilityFS},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code:      []core.Instruction{{Op: core.OpReturn}},
			},
		},
	}
	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule unexpected error: %v", err)
	}
	if err := os.WriteFile(modulePath, data, 0o644); err != nil {
		t.Fatalf("write module file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "verify-iavm", modulePath, "--allow-capability", "network"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI verify-iavm capability deny code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "verify module error: capability[0]: kind \"fs\" is not allowed") {
		t.Fatalf("stderr = %q, want capability deny error", stderr.String())
	}
}

func TestRunCLIVerifyIavmFunctionLimitExceeded(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "too-many-functions.iavm")

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{Name: "entry", TypeIndex: 0, Code: []core.Instruction{{Op: core.OpReturn}}},
			{Name: "helper", TypeIndex: 0, Code: []core.Instruction{{Op: core.OpReturn}}},
		},
	}
	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule unexpected error: %v", err)
	}
	if err := os.WriteFile(modulePath, data, 0o644); err != nil {
		t.Fatalf("write module file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "verify-iavm", modulePath, "--max-functions", "1"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI verify-iavm function limit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "verify module error: function count 2 exceeds limit 1") {
		t.Fatalf("stderr = %q, want function limit error", stderr.String())
	}
}

func TestRunCLIRunIavmSuccess(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "run-success.iavm")

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:         "entry",
				TypeIndex:    0,
				Code:         []core.Instruction{{Op: core.OpReturn}},
				IsEntryPoint: true,
			},
		},
	}
	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule unexpected error: %v", err)
	}
	if err := os.WriteFile(modulePath, data, 0o644); err != nil {
		t.Fatalf("write module file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "run-iavm", modulePath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI run-iavm code = %d, want 0, stderr=%q", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunCLIRunIavmCapabilityDenied(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "run-cap-denied.iavm")

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Capabilities: []module.CapabilityDecl{
			{Kind: module.CapabilityFS},
		},
		Functions: []module.Function{
			{
				Name:         "entry",
				TypeIndex:    0,
				Code:         []core.Instruction{{Op: core.OpReturn}},
				IsEntryPoint: true,
			},
		},
	}
	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule unexpected error: %v", err)
	}
	if err := os.WriteFile(modulePath, data, 0o644); err != nil {
		t.Fatalf("write module file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "run-iavm", modulePath, "--allow-capability", "network"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI run-iavm capability deny code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "verify module error: capability[0]: kind \"fs\" is not allowed") {
		t.Fatalf("stderr = %q, want capability deny error", stderr.String())
	}
}

func TestRunCLIInspectIavmSuccess(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "app.iavm")

	mod := &module.Module{
		Magic:        "IAVM",
		Version:      1,
		Target:       "ialang",
		ABIVersion:   1,
		FeatureFlags: 3,
		Types:        []core.FuncType{{}},
		Globals: []module.Global{
			{Name: "g0", Type: core.ValueI64},
		},
		Exports: []module.Export{
			{Name: "entry", Kind: module.ExportFunction, Index: 0},
		},
		Capabilities: []module.CapabilityDecl{
			{Kind: module.CapabilityFS},
			{Kind: module.CapabilityNetwork},
		},
		Constants: []any{int64(1)},
		Custom: map[string][]byte{
			"meta": []byte("v1"),
		},
		Functions: []module.Function{
			{
				Name:         "entry",
				TypeIndex:    0,
				Locals:       []core.ValueKind{core.ValueI64},
				Code:         []core.Instruction{{Op: core.OpReturn}},
				MaxStack:     1,
				IsEntryPoint: true,
			},
		},
	}
	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule unexpected error: %v", err)
	}
	if err := os.WriteFile(modulePath, data, 0o644); err != nil {
		t.Fatalf("write module file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "inspect-iavm", modulePath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI inspect-iavm code = %d, want 0, stderr=%q", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "IAVM module summary") {
		t.Fatalf("stdout = %q, want summary header", out)
	}
	if !strings.Contains(out, "  target: ialang") {
		t.Fatalf("stdout = %q, want target summary", out)
	}
	if !strings.Contains(out, "  entry: entry") {
		t.Fatalf("stdout = %q, want entry summary", out)
	}
	if !strings.Contains(out, "  capability_kinds: fs,network") {
		t.Fatalf("stdout = %q, want capability summary", out)
	}
	if strings.Contains(out, "function[0]:") {
		t.Fatalf("stdout = %q, want non-verbose output without function details", out)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunCLIInspectIavmVerboseShowsFunctions(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "app.iavm")

	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:         "",
				TypeIndex:    0,
				Locals:       []core.ValueKind{core.ValueI64, core.ValueI64},
				Code:         []core.Instruction{{Op: core.OpReturn}},
				MaxStack:     2,
				IsEntryPoint: true,
			},
		},
	}
	data, err := binary.EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule unexpected error: %v", err)
	}
	if err := os.WriteFile(modulePath, data, 0o644); err != nil {
		t.Fatalf("write module file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "inspect-iavm", modulePath, "--verbose"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI inspect-iavm verbose code = %d, want 0, stderr=%q", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "function[0]: name=<anonymous> type=0 locals=2 code=1 max_stack=2 entry=true") {
		t.Fatalf("stdout = %q, want verbose function details", out)
	}
}

func TestRunCLIInspectIavmFileNotFound(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "inspect-iavm", "__not_found__.iavm"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI inspect-iavm code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "read file error:") {
		t.Fatalf("stderr = %q, want read file error", stderr.String())
	}
}

func TestRunCLIInspectIavmDecodeError(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "bad-inspect.iavm")
	if err := os.WriteFile(modulePath, []byte("not a valid iavm module"), 0o644); err != nil {
		t.Fatalf("write module file error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "inspect-iavm", modulePath}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI inspect-iavm code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "decode module error:") {
		t.Fatalf("stderr = %q, want decode error", stderr.String())
	}
}

func TestReadRunSource(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ok.ia")
	if err := os.WriteFile(path, []byte("let x = 1;"), 0o644); err != nil {
		t.Fatalf("write file error: %v", err)
	}

	got, err := readRunSource(path)
	if err != nil {
		t.Fatalf("readRunSource unexpected error: %v", err)
	}
	if got != "let x = 1;" {
		t.Fatalf("readRunSource = %q, want %q", got, "let x = 1;")
	}

	_, err = readRunSource(filepath.Join(dir, "missing.ia"))
	if err == nil {
		t.Fatal("readRunSource missing file expected error, got nil")
	}
}

func TestCompileRunSourceParseError(t *testing.T) {
	var stderr bytes.Buffer
	_, err := compileRunSource("let = 1;", &stderr)
	if err == nil {
		t.Fatal("compileRunSource parse error expected, got nil")
	}
	if !strings.Contains(stderr.String(), "parse errors:") {
		t.Fatalf("stderr = %q, want parse errors", stderr.String())
	}
}

func TestCompileRunSourceCompileError(t *testing.T) {
	var stderr bytes.Buffer
	_, err := compileRunSource("break;", &stderr)
	if err == nil {
		t.Fatal("compileRunSource compile error expected, got nil")
	}
	if !strings.Contains(stderr.String(), "compile errors:") {
		t.Fatalf("stderr = %q, want compile errors", stderr.String())
	}
	if !strings.Contains(stderr.String(), "line 1, col 1") {
		t.Fatalf("stderr = %q, want compile error position", stderr.String())
	}
}

func TestExecuteRunCommandSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ok.ia")
	if err := os.WriteFile(path, []byte("let x = 1 + 2;"), 0o644); err != nil {
		t.Fatalf("write file error: %v", err)
	}

	var stderr bytes.Buffer
	if err := executeRunCommand(path, nil, &stderr); err != nil {
		t.Fatalf("executeRunCommand unexpected error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestExecuteRunCommandPassesScriptArgs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "args.ia")
	src := "import { args } from \"process\";\n" +
		"function main() {\n" +
		"  let argv = args();\n" +
		"  if (argv.length < 4) { missing_len(); }\n" +
		"  if (argv[2] != \"install\") { missing_cmd(); }\n" +
		"  if (argv[3] != \"demo@1.2.3\") { missing_target(); }\n" +
		"}\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write file error: %v", err)
	}

	var stderr bytes.Buffer
	if err := executeRunCommand(path, []string{"install", "demo@1.2.3"}, &stderr); err != nil {
		t.Fatalf("executeRunCommand with script args unexpected error: %v", err)
	}
}

func TestExecuteRunCommandAutoCallsMain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auto_main.ia")
	src := "function main() {\n  missing_symbol();\n}\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write file error: %v", err)
	}

	var stderr bytes.Buffer
	err := executeRunCommand(path, nil, &stderr)
	if err == nil {
		t.Fatal("executeRunCommand expected runtime error from auto-called main, got nil")
	}
	if !strings.Contains(err.Error(), "entry main() call error") {
		t.Fatalf("executeRunCommand error = %v, want entry main() call error", err)
	}
}

func TestExecuteBuildAndRunPkgCommand(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "main.ia")
	helperPath := filepath.Join(dir, "helper.ia")
	pkgPath := filepath.Join(dir, "app.iapkg")

	if err := os.WriteFile(helperPath, []byte("export let answer = 40 + 2;"), 0o644); err != nil {
		t.Fatalf("write helper file error: %v", err)
	}
	if err := os.WriteFile(entryPath, []byte("import { answer } from \"./helper\";\nprint(answer);\n"), 0o644); err != nil {
		t.Fatalf("write entry file error: %v", err)
	}

	var stderr bytes.Buffer
	if err := executeBuildCommand(entryPath, pkgPath, &stderr); err != nil {
		t.Fatalf("executeBuildCommand unexpected error: %v", err)
	}
	if _, err := os.Stat(pkgPath); err != nil {
		t.Fatalf("package output not found: %v", err)
	}

	if err := executeRunPkgCommand(pkgPath, nil, &stderr); err != nil {
		t.Fatalf("executeRunPkgCommand unexpected error: %v", err)
	}
}

func TestExecuteBuildCommandParseErrorIncludesUnit(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "main.ia")
	// Use a real syntax error: missing identifier after 'let'
	if err := os.WriteFile(entryPath, []byte("let = 1;"), 0o644); err != nil {
		t.Fatalf("write entry file error: %v", err)
	}

	var stderr bytes.Buffer
	err := executeBuildCommand(entryPath, filepath.Join(dir, "app.iapkg"), &stderr)
	if err == nil {
		t.Fatal("executeBuildCommand expected parse error, got nil")
	}
	stderrText := stderr.String()
	if !strings.Contains(stderrText, "parse errors in "+entryPath+":") {
		t.Fatalf("stderr = %q, want parse unit output", stderrText)
	}
}

func TestExecuteRunPkgCommandAutoCallsMain(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "main.ia")
	pkgPath := filepath.Join(dir, "app.iapkg")
	src := "function main() {\n  missing_symbol();\n}\n"
	if err := os.WriteFile(entryPath, []byte(src), 0o644); err != nil {
		t.Fatalf("write entry file error: %v", err)
	}

	var stderr bytes.Buffer
	if err := executeBuildCommand(entryPath, pkgPath, &stderr); err != nil {
		t.Fatalf("executeBuildCommand unexpected error: %v", err)
	}

	err := executeRunPkgCommand(pkgPath, nil, &stderr)
	if err == nil {
		t.Fatal("executeRunPkgCommand expected runtime error from auto-called main, got nil")
	}
	if !strings.Contains(err.Error(), "entry main() call error") {
		t.Fatalf("executeRunPkgCommand error = %v, want entry main() call error", err)
	}
}

func TestPackageModuleLoaderResolvesBuiltinFromSharedRegistry(t *testing.T) {
	loader := newPackageModuleLoader(&packagefile.Package{}, map[string]lang.Value{"builtin": float64(9)}, nil, lang.VMOptions{})

	loaded, err := loader.Resolve("/entry.ia", "builtin")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if loaded != float64(9) {
		t.Fatalf("Resolve() = %#v, want 9", loaded)
	}
}

func TestPackageModuleLoaderUsesSharedModuleNotFoundError(t *testing.T) {
	loader := newPackageModuleLoader(&packagefile.Package{
		Entry:   "/entry.ia",
		Modules: map[string]*bc.Chunk{},
		Imports: map[string]map[string]string{},
	}, nil, nil, lang.VMOptions{})

	_, err := loader.Resolve("/entry.ia", "./missing")
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
	if !errors.Is(err, moduleapi.ErrModuleNotFound) {
		t.Fatalf("errors.Is(err, ErrModuleNotFound) = false, err = %v", err)
	}
}

func TestPackageModuleLoaderUsesSharedCyclicImportError(t *testing.T) {
	targetPath := "/dep.ia"
	loader := newPackageModuleLoader(&packagefile.Package{
		Entry:   "/entry.ia",
		Modules: map[string]*bc.Chunk{targetPath: {}},
		Imports: map[string]map[string]string{"/entry.ia": {"./dep": targetPath}},
	}, nil, nil, lang.VMOptions{})
	loader.loading[targetPath] = true

	_, err := loader.Resolve("/entry.ia", "./dep")
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
	if !errors.Is(err, moduleapi.ErrCyclicImport) {
		t.Fatalf("errors.Is(err, ErrCyclicImport) = false, err = %v", err)
	}
}

func TestPackageModuleLoaderUsesSharedRuntimeModuleError(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "main.ia")
	depPath := filepath.Join(dir, "dep.ia")
	if err := os.WriteFile(depPath, []byte("missing_symbol();\nexport let value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write dep file error: %v", err)
	}
	if err := os.WriteFile(entryPath, []byte("import { value } from \"./dep\";\nprint(value);\n"), 0o644); err != nil {
		t.Fatalf("write entry file error: %v", err)
	}

	var stderr bytes.Buffer
	pkg, err := buildPackage(entryPath, &stderr)
	if err != nil {
		t.Fatalf("buildPackage unexpected error: %v", err)
	}

	loader := newPackageModuleLoader(pkg, nil, nil, lang.VMOptions{})
	_, err = loader.Resolve(pkg.Entry, "./dep")
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
	if !errors.Is(err, moduleapi.ErrRuntimeModule) {
		t.Fatalf("errors.Is(err, ErrRuntimeModule) = false, err = %v", err)
	}
}

func TestReadTimeoutMsEnv(t *testing.T) {
	const name = "IALANG_TEST_TIMEOUT_MS"
	t.Setenv(name, "")
	got, err := readTimeoutMsEnv(name)
	if err != nil {
		t.Fatalf("empty env unexpected error: %v", err)
	}
	if got != 0 {
		t.Fatalf("empty env = %s, want 0", got)
	}

	t.Setenv(name, "15")
	got, err = readTimeoutMsEnv(name)
	if err != nil {
		t.Fatalf("valid env unexpected error: %v", err)
	}
	if got != 15*time.Millisecond {
		t.Fatalf("valid env = %s, want 15ms", got)
	}

	t.Setenv(name, "-1")
	_, err = readTimeoutMsEnv(name)
	if err == nil {
		t.Fatal("negative env expected error, got nil")
	}

	t.Setenv(name, "abc")
	_, err = readTimeoutMsEnv(name)
	if err == nil {
		t.Fatal("non-int env expected error, got nil")
	}
}

func TestBuildAsyncRuntimeFromEnv(t *testing.T) {
	// Use real env names to exercise full config path.
	t.Setenv("IALANG_ASYNC_TASK_TIMEOUT_MS", "20")
	t.Setenv("IALANG_ASYNC_AWAIT_TIMEOUT_MS", "30")

	rt, err := buildAsyncRuntimeFromEnv()
	if err != nil {
		t.Fatalf("buildAsyncRuntimeFromEnv unexpected error: %v", err)
	}
	if rt == nil {
		t.Fatal("buildAsyncRuntimeFromEnv returned nil runtime")
	}
}

func TestReadBoolEnv(t *testing.T) {
	const name = "IALANG_TEST_BOOL"

	t.Setenv(name, "")
	v, err := readBoolEnv(name)
	if err != nil {
		t.Fatalf("empty env unexpected error: %v", err)
	}
	if v {
		t.Fatalf("empty env = true, want false")
	}

	t.Setenv(name, "true")
	v, err = readBoolEnv(name)
	if err != nil {
		t.Fatalf("true env unexpected error: %v", err)
	}
	if !v {
		t.Fatalf("true env = false, want true")
	}

	t.Setenv(name, "0")
	v, err = readBoolEnv(name)
	if err != nil {
		t.Fatalf("0 env unexpected error: %v", err)
	}
	if v {
		t.Fatalf("0 env = true, want false")
	}

	t.Setenv(name, "maybe")
	_, err = readBoolEnv(name)
	if err == nil {
		t.Fatal("invalid bool env expected error, got nil")
	}
}

func TestBuildVMOptionsFromEnv(t *testing.T) {
	t.Setenv("IALANG_STRUCTURED_RUNTIME_ERRORS", "1")
	options, err := buildVMOptionsFromEnv()
	if err != nil {
		t.Fatalf("buildVMOptionsFromEnv unexpected error: %v", err)
	}
	if !options.StructuredRuntimeErrors {
		t.Fatal("StructuredRuntimeErrors = false, want true")
	}

	t.Setenv("IALANG_STRUCTURED_RUNTIME_ERRORS", "invalid")
	_, err = buildVMOptionsFromEnv()
	if err == nil {
		t.Fatal("buildVMOptionsFromEnv invalid env expected error, got nil")
	}
}

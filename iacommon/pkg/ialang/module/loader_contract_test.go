package module

import (
	"errors"
	"strings"
	"testing"
)

func TestPlatformModuleNames(t *testing.T) {
	if PlatformFSModuleName != "@platform/fs" {
		t.Fatalf("PlatformFSModuleName = %q, want %q", PlatformFSModuleName, "@platform/fs")
	}
	if PlatformHTTPModuleName != "@platform/http" {
		t.Fatalf("PlatformHTTPModuleName = %q, want %q", PlatformHTTPModuleName, "@platform/http")
	}
}

func TestLoaderContractErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantIs  error
		wantMsg string
	}{
		{name: "unknown", err: UnknownModuleError("@platform/missing"), wantIs: ErrUnknownModule, wantMsg: "unknown module: @platform/missing"},
		{name: "not found", err: ModuleNotFoundError("./missing"), wantIs: ErrModuleNotFound, wantMsg: "module not found: ./missing"},
		{name: "cyclic", err: CyclicImportError("/tmp/mod.ia"), wantIs: ErrCyclicImport, wantMsg: "cyclic import detected: /tmp/mod.ia"},
		{name: "read", err: ReadModuleError("/tmp/mod.ia", errors.New("permission denied")), wantIs: ErrReadModule, wantMsg: "read module error (/tmp/mod.ia): permission denied"},
		{name: "parse", err: ParseModuleError("/tmp/mod.ia", "unexpected token"), wantIs: ErrParseModule, wantMsg: "parse module error (/tmp/mod.ia): unexpected token"},
		{name: "compile", err: CompileModuleError("/tmp/mod.ia", "unknown symbol"), wantIs: ErrCompileModule, wantMsg: "compile module error (/tmp/mod.ia): unknown symbol"},
		{name: "runtime", err: RuntimeModuleError("/tmp/mod.ia", errors.New("boom")), wantIs: ErrRuntimeModule, wantMsg: "runtime module error (/tmp/mod.ia): boom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, tt.wantIs) {
				t.Fatalf("errors.Is(%v, %v) = false", tt.err, tt.wantIs)
			}
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Fatalf("error message = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestLoaderContractWrappedCause(t *testing.T) {
	cause := errors.New("disk failure")
	err := RuntimeModuleError("/tmp/main.ia", cause)
	if !errors.Is(err, cause) {
		t.Fatal("errors.Is(runtimeErr, cause) = false")
	}
	if !strings.Contains(err.Error(), "disk failure") {
		t.Fatalf("error message = %q, want contains %q", err.Error(), "disk failure")
	}
}

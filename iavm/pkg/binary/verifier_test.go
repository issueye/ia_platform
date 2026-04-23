package binary

import (
	"iavm/pkg/core"
	"iavm/pkg/module"
	"testing"
)

func TestVerifyModule_ValidMinimal(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
	}
	result, err := VerifyModule(mod, VerifyOptions{})
	if err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatal("expected valid module")
	}
}

func TestVerifyModule_InvalidMagic(t *testing.T) {
	mod := &module.Module{
		Magic:   "BAD!",
		Version: 1,
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid magic")
	}
}

func TestVerifyModule_InvalidVersion(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 99,
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestVerifyModule_InvalidTypeRef(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 5,
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid type reference")
	}
}

func TestVerifyModule_InvalidExportRef(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Exports: []module.Export{
			{Name: "main", Kind: module.ExportFunction, Index: 10},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid export reference")
	}
}

func TestVerifyModule_InvalidOpcode(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types: []core.FuncType{
			{},
		},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpCode(255)},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid opcode")
	}
}

func TestVerifyModule_EmptyImportModule(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Imports: []module.Import{
			{Module: "", Name: "test", Kind: module.ImportFunction},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for empty import module name")
	}
}

func TestVerifyModule_InvalidCapability(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Capabilities: []module.CapabilityDecl{
			{Kind: "invalid"},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid capability kind")
	}
}

func TestVerifyModule_RequireEntry_NoEntry(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types: []core.FuncType{
			{},
		},
		Functions: []module.Function{
			{
				Name:      "helper",
				TypeIndex: 0,
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{RequireEntry: true})
	if err == nil {
		t.Fatal("expected error for missing entry point")
	}
}

func TestVerifyModule_RequireEntry_WithMain(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types: []core.FuncType{
			{},
		},
		Functions: []module.Function{
			{
				Name:      "main",
				TypeIndex: 0,
			},
		},
	}
	result, err := VerifyModule(mod, VerifyOptions{RequireEntry: true})
	if err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatal("expected valid module with main function")
	}
}

func TestVerifyModule_InvalidJumpTarget(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpJump, A: 100},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid jump target")
	}
}

func TestVerifyModule_InvalidLocalIndex(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueNull},
				Code: []core.Instruction{
					{Op: core.OpLoadLocal, A: 5},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid local index")
	}
}

func TestVerifyModule_InvalidConstantIndex(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Constants: []any{int64(1)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 10},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid constant index")
	}
}

func TestVerifyModule_InvalidTryHandlerTarget(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpPushTry, A: 99},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid try handler target")
	}
}

func TestVerifyModule_ValidControlFlow(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueI64},
				Constants: []any{int64(42)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpStoreLocal, A: 0},
					{Op: core.OpLoadLocal, A: 0},
					{Op: core.OpJump, A: 3},
					{Op: core.OpReturn},
				},
			},
		},
	}
	result, err := VerifyModule(mod, VerifyOptions{})
	if err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid module, got errors: %v", result.Errors)
	}
}
func TestVerifyModule_StackUnderflow(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpAdd},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for stack underflow")
	}
}

func TestVerifyModule_StackHeightMismatchAtJoin(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Constants: []any{false, int64(1)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpJumpIfFalse, A: 4},
					{Op: core.OpConst, A: 1},
					{Op: core.OpJump, A: 4},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for stack height mismatch")
	}
}

func TestVerifyModule_DirectCallArgumentMismatch(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types: []core.FuncType{
			{Params: []core.ValueKind{core.ValueI64, core.ValueI64}, Results: []core.ValueKind{core.ValueI64}},
			{},
		},
		Functions: []module.Function{
			{
				Name:      "add",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueI64, core.ValueI64},
				Code: []core.Instruction{
					{Op: core.OpLoadLocal, A: 0},
					{Op: core.OpLoadLocal, A: 1},
					{Op: core.OpAdd},
					{Op: core.OpReturn},
				},
			},
			{
				Name:      "entry",
				TypeIndex: 1,
				Constants: []any{int64(5)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpCall, A: 0, B: 1},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for direct call argument mismatch")
	}
}

func TestVerifyModule_StackExceedsMaxStack(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Constants: []any{int64(1), int64(2)},
				MaxStack:  1,
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for max stack overflow")
	}
}

func TestVerifyModule_ResourceLimitsAllowValidModule(t *testing.T) {
	mod := resourceLimitTestModule()

	opts := VerifyOptions{
		MaxFunctions:           2,
		MaxConstants:           2,
		MaxCodeSizePerFunction: 4,
		MaxLocalsPerFunction:   2,
		MaxStackPerFunction:    3,
	}
	if _, err := VerifyModule(mod, opts); err != nil {
		t.Fatalf("expected resource limits to allow module, got %v", err)
	}
}

func TestVerifyModule_ResourceLimitFailures(t *testing.T) {
	cases := []struct {
		name string
		opts VerifyOptions
	}{
		{name: "max functions", opts: VerifyOptions{MaxFunctions: 1}},
		{name: "max constants", opts: VerifyOptions{MaxConstants: 1}},
		{name: "max code size", opts: VerifyOptions{MaxCodeSizePerFunction: 3}},
		{name: "max locals", opts: VerifyOptions{MaxLocalsPerFunction: 1}},
		{name: "max stack declaration", opts: VerifyOptions{MaxStackPerFunction: 2}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mod := resourceLimitTestModule()
			if tc.name == "max functions" {
				tc.opts.MaxFunctions = 0
				if _, err := VerifyModule(mod, tc.opts); err != nil {
					t.Fatalf("zero max functions should mean unlimited, got %v", err)
				}
				tc.opts.MaxFunctions = len(mod.Functions) - 1
			}
			if _, err := VerifyModule(mod, tc.opts); err == nil {
				t.Fatalf("expected %s limit error", tc.name)
			}
		})
	}
}

func TestVerifyModule_PerFunctionConstantLimit(t *testing.T) {
	mod := resourceLimitTestModule()
	mod.Constants = nil
	mod.Functions[0].Constants = []any{int64(1), int64(2)}

	if _, err := VerifyModule(mod, VerifyOptions{MaxConstants: 1}); err == nil {
		t.Fatal("expected per-function constant limit error")
	}
}

func TestVerifyModule_CapabilityAllowlist(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Capabilities: []module.CapabilityDecl{
			{Kind: module.CapabilityFS},
		},
	}

	if _, err := VerifyModule(mod, VerifyOptions{CapabilityAllowlistSet: true, AllowedCapabilities: []module.CapabilityKind{module.CapabilityFS}}); err != nil {
		t.Fatalf("expected fs capability to be allowed, got %v", err)
	}
	if _, err := VerifyModule(mod, VerifyOptions{CapabilityAllowlistSet: true, AllowedCapabilities: []module.CapabilityKind{module.CapabilityNetwork}}); err == nil {
		t.Fatal("expected fs capability to be denied by allowlist")
	}
}

func TestBuildVerifyOptions_DefaultProfile(t *testing.T) {
	opts, err := BuildVerifyOptions(VerifyProfileDefault, VerifyPolicyOverrides{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.RequireEntry {
		t.Fatal("default profile should not require entry")
	}
	if opts.CapabilityAllowlistSet {
		t.Fatal("default profile should not set capability allowlist")
	}
}

func TestBuildVerifyOptions_StrictProfile(t *testing.T) {
	opts, err := BuildVerifyOptions(VerifyProfileStrict, VerifyPolicyOverrides{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.RequireEntry {
		t.Fatal("strict profile should require entry")
	}
	if opts.CapabilityAllowlistSet {
		t.Fatal("strict profile should not set capability allowlist")
	}
}

func TestBuildVerifyOptions_SandboxProfile(t *testing.T) {
	opts, err := BuildVerifyOptions(VerifyProfileSandbox, VerifyPolicyOverrides{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.RequireEntry {
		t.Fatal("sandbox profile should require entry")
	}
	if !opts.CapabilityAllowlistSet {
		t.Fatal("sandbox profile should set capability allowlist")
	}
	if len(opts.AllowedCapabilities) != 0 {
		t.Fatal("sandbox profile should have empty allowlist (deny-all)")
	}
	if opts.MaxFunctions != 128 {
		t.Fatalf("sandbox profile should set max functions to 128, got %d", opts.MaxFunctions)
	}
}

func TestBuildVerifyOptions_OverridesApplyToSandbox(t *testing.T) {
	opts, err := BuildVerifyOptions(VerifyProfileSandbox, VerifyPolicyOverrides{
		MaxFunctions: 256,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.MaxFunctions != 256 {
		t.Fatalf("override should set max functions to 256, got %d", opts.MaxFunctions)
	}
}

func TestBuildVerifyOptions_CapabilityAllowlistOverride(t *testing.T) {
	opts, err := BuildVerifyOptions(VerifyProfileDefault, VerifyPolicyOverrides{
		CapabilityAllowlistSet: true,
		AllowedCapabilities:    []module.CapabilityKind{module.CapabilityFS},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.CapabilityAllowlistSet {
		t.Fatal("override should set capability allowlist")
	}
	if len(opts.AllowedCapabilities) != 1 || opts.AllowedCapabilities[0] != module.CapabilityFS {
		t.Fatalf("override should set allowed capabilities to [fs], got %v", opts.AllowedCapabilities)
	}
}

func TestBuildVerifyOptions_ExplicitDenyAll(t *testing.T) {
	opts, err := BuildVerifyOptions(VerifyProfileDefault, VerifyPolicyOverrides{
		CapabilityAllowlistSet: true,
		AllowedCapabilities:    []module.CapabilityKind{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.CapabilityAllowlistSet {
		t.Fatal("explicit deny-all should set CapabilityAllowlistSet")
	}
	if len(opts.AllowedCapabilities) != 0 {
		t.Fatal("explicit deny-all should have empty AllowedCapabilities")
	}
}

func TestBuildVerifyOptions_UnsupportedProfile(t *testing.T) {
	_, err := BuildVerifyOptions(VerifyProfile("custom"), VerifyPolicyOverrides{})
	if err == nil {
		t.Fatal("expected error for unsupported profile")
	}
}

func TestCapabilityPolicy_AllowAll(t *testing.T) {
	opts := VerifyOptions{CapabilityAllowlistSet: false}
	if opts.CapabilityPolicy() != "allow-all" {
		t.Fatalf("expected 'allow-all', got %q", opts.CapabilityPolicy())
	}
}

func TestCapabilityPolicy_DenyAll(t *testing.T) {
	opts := VerifyOptions{CapabilityAllowlistSet: true, AllowedCapabilities: nil}
	if opts.CapabilityPolicy() != "deny-all" {
		t.Fatalf("expected 'deny-all', got %q", opts.CapabilityPolicy())
	}
}

func TestCapabilityPolicy_Allowlist(t *testing.T) {
	opts := VerifyOptions{CapabilityAllowlistSet: true, AllowedCapabilities: []module.CapabilityKind{module.CapabilityFS, module.CapabilityNetwork}}
	if opts.CapabilityPolicy() != "allowlist:fs,network" {
		t.Fatalf("expected 'allowlist:fs,network', got %q", opts.CapabilityPolicy())
	}
}

func TestPolicySummary_Default(t *testing.T) {
	opts := VerifyOptions{}
	s := opts.PolicySummary()
	if s != "policy: capabilities=allow-all" {
		t.Fatalf("unexpected summary: %q", s)
	}
}

func TestPolicySummary_Sandbox(t *testing.T) {
	opts, _ := BuildVerifyOptions(VerifyProfileSandbox, VerifyPolicyOverrides{})
	s := opts.PolicySummary()
	if s == "" {
		t.Fatal("expected non-empty summary")
	}
}

func TestVerifyModule_CapabilityDenyAll(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Capabilities: []module.CapabilityDecl{
			{Kind: module.CapabilityFS},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{CapabilityAllowlistSet: true, AllowedCapabilities: nil})
	if err == nil {
		t.Fatal("expected capability to be denied with deny-all policy")
	}
}

func TestVerifyModule_CapabilityAllowAllByDefault(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Capabilities: []module.CapabilityDecl{
			{Kind: module.CapabilityFS},
			{Kind: module.CapabilityNetwork},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err != nil {
		t.Fatalf("expected all capabilities allowed when allowlist not set, got %v", err)
	}
}

func resourceLimitTestModule() *module.Module {
	return &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{int64(1), int64(2)},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueI64, core.ValueI64},
				MaxStack:  3,
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpAdd},
					{Op: core.OpReturn},
				},
			},
			{
				Name:      "helper",
				TypeIndex: 0,
				Code:      []core.Instruction{{Op: core.OpReturn}},
			},
		},
	}
}

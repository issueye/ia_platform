package binary

import (
	"fmt"
	"iavm/pkg/core"
	"iavm/pkg/module"
)

type VerifyOptions struct {
	RequireEntry bool
	AllowCustom  bool
}

func VerifyModule(m *module.Module, opts VerifyOptions) (*VerifyResult, error) {
	result := &VerifyResult{Valid: true}

	if err := verifyHeader(m); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	if err := verifyTypes(m); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	if err := verifyFunctions(m); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	if err := verifyExports(m); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	if err := verifyImports(m); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	if err := verifyGlobals(m); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	if err := verifyCapabilities(m); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	if opts.RequireEntry {
		if err := verifyEntry(m); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
			return result, err
		}
	}

	return result, nil
}

func verifyHeader(m *module.Module) error {
	if m.Magic != "IAVM" {
		return fmt.Errorf("invalid magic: %q, expected 'IAVM'", m.Magic)
	}
	if m.Version < 1 || m.Version > 2 {
		return fmt.Errorf("unsupported version: %d", m.Version)
	}
	if m.Target == "" {
		return fmt.Errorf("empty target")
	}
	return nil
}

func verifyTypes(m *module.Module) error {
	for i, ft := range m.Types {
		for j, p := range ft.Params {
			if !isValidValueKind(p) {
				return fmt.Errorf("type[%d]: invalid param kind %v at index %d", i, p, j)
			}
		}
		for j, r := range ft.Results {
			if !isValidValueKind(r) {
				return fmt.Errorf("type[%d]: invalid result kind %v at index %d", i, r, j)
			}
		}
	}
	return nil
}

func verifyFunctions(m *module.Module) error {
	for i, fn := range m.Functions {
		if int(fn.TypeIndex) >= len(m.Types) {
			return fmt.Errorf("function[%d]: type index %d out of range (types: %d)",
				i, fn.TypeIndex, len(m.Types))
		}

		for j, local := range fn.Locals {
			if !isValidValueKind(local) {
				return fmt.Errorf("function[%d]: invalid local kind %v at index %d", i, local, j)
			}
		}

		for j, inst := range fn.Code {
			if !isValidOpcode(inst.Op) {
				return fmt.Errorf("function[%d]: invalid opcode %v at instruction %d", i, inst.Op, j)
			}
		}
	}
	return nil
}

func verifyExports(m *module.Module) error {
	for i, exp := range m.Exports {
		switch exp.Kind {
		case module.ExportFunction:
			if exp.Index >= uint32(len(m.Functions)) {
				return fmt.Errorf("export[%d]: function index %d out of range (functions: %d)", i, exp.Index, len(m.Functions))
			}
		case module.ExportGlobal:
			if exp.Index >= uint32(len(m.Globals)) {
				return fmt.Errorf("export[%d]: global index %d out of range (globals: %d)", i, exp.Index, len(m.Globals))
			}
		}
	}
	return nil
}

func verifyImports(m *module.Module) error {
	for i, imp := range m.Imports {
		if imp.Module == "" {
			return fmt.Errorf("import[%d]: empty module name", i)
		}
		if imp.Name == "" {
			return fmt.Errorf("import[%d]: empty name", i)
		}
	}
	return nil
}

func verifyGlobals(m *module.Module) error {
	for i, g := range m.Globals {
		if g.Name == "" {
			return fmt.Errorf("global[%d]: empty name", i)
		}
		if !isValidValueKind(g.Type) {
			return fmt.Errorf("global[%d]: invalid type kind %v", i, g.Type)
		}
	}
	return nil
}

func verifyCapabilities(m *module.Module) error {
	for i, cap := range m.Capabilities {
		if cap.Kind != module.CapabilityFS && cap.Kind != module.CapabilityNetwork {
			return fmt.Errorf("capability[%d]: invalid kind %q", i, cap.Kind)
		}
	}
	return nil
}

func verifyEntry(m *module.Module) error {
	hasEntry := false
	for _, fn := range m.Functions {
		if fn.IsEntryPoint || fn.Name == "main" || fn.Name == "entry" {
			hasEntry = true
			break
		}
	}
	if !hasEntry {
		return fmt.Errorf("no entry point function found")
	}
	return nil
}

func isValidValueKind(kind core.ValueKind) bool {
	switch kind {
	case core.ValueNull, core.ValueBool, core.ValueI64, core.ValueF64,
		core.ValueString, core.ValueBytes, core.ValueArrayRef,
		core.ValueObjectRef, core.ValueFuncRef, core.ValueHostHandle:
		return true
	default:
		return false
	}
}

func isValidOpcode(op core.OpCode) bool {
	return op >= core.OpNop && op <= core.OpHostPoll
}

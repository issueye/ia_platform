package binary

import (
	"fmt"
	"iavm/pkg/core"
	"iavm/pkg/module"
)

type VerifyProfile string

const (
	VerifyProfileDefault VerifyProfile = "default"
	VerifyProfileStrict  VerifyProfile = "strict"
	VerifyProfileSandbox VerifyProfile = "sandbox"
)

func (p VerifyProfile) String() string {
	if p == "" {
		return string(VerifyProfileDefault)
	}
	return string(p)
}

type VerifyPolicyOverrides struct {
	RequireEntry *bool

	MaxFunctions           int
	MaxConstants           int
	MaxCodeSizePerFunction int
	MaxLocalsPerFunction   int
	MaxStackPerFunction    int
	AllowedCapabilities    []module.CapabilityKind
	CapabilityAllowlistSet bool
}

type VerifyOptions struct {
	RequireEntry bool
	AllowCustom  bool

	MaxFunctions           int
	MaxConstants           int
	MaxCodeSizePerFunction int
	MaxLocalsPerFunction   int
	MaxStackPerFunction    int
	AllowedCapabilities    []module.CapabilityKind
	CapabilityAllowlistSet bool
}

func (o VerifyOptions) CapabilityPolicy() string {
	if !o.CapabilityAllowlistSet {
		return "allow-all"
	}
	if len(o.AllowedCapabilities) == 0 {
		return "deny-all"
	}
	return "allowlist:" + formatCapabilityKinds(o.AllowedCapabilities)
}

func (o VerifyOptions) PolicySummary() string {
	s := "policy:"
	if o.RequireEntry {
		s += " require-entry"
	}
	if o.MaxFunctions > 0 {
		s += fmt.Sprintf(" max-functions=%d", o.MaxFunctions)
	}
	if o.MaxConstants > 0 {
		s += fmt.Sprintf(" max-constants=%d", o.MaxConstants)
	}
	if o.MaxCodeSizePerFunction > 0 {
		s += fmt.Sprintf(" max-code-size=%d", o.MaxCodeSizePerFunction)
	}
	if o.MaxLocalsPerFunction > 0 {
		s += fmt.Sprintf(" max-locals=%d", o.MaxLocalsPerFunction)
	}
	if o.MaxStackPerFunction > 0 {
		s += fmt.Sprintf(" max-stack=%d", o.MaxStackPerFunction)
	}
	s += " capabilities=" + o.CapabilityPolicy()
	return s
}

func formatCapabilityKinds(kinds []module.CapabilityKind) string {
	if len(kinds) == 0 {
		return ""
	}
	names := make([]string, len(kinds))
	for i, k := range kinds {
		names[i] = string(k)
	}
	result := names[0]
	for i := 1; i < len(names); i++ {
		result += "," + names[i]
	}
	return result
}

func BuildVerifyOptions(profile VerifyProfile, overrides VerifyPolicyOverrides) (VerifyOptions, error) {
	opts := VerifyOptions{}

	switch profile {
	case "", VerifyProfileDefault:
	case VerifyProfileStrict:
		opts.RequireEntry = true
	case VerifyProfileSandbox:
		opts.RequireEntry = true
		opts.MaxFunctions = 128
		opts.MaxConstants = 512
		opts.MaxCodeSizePerFunction = 4096
		opts.MaxLocalsPerFunction = 64
		opts.MaxStackPerFunction = 128
		opts.CapabilityAllowlistSet = true
		opts.AllowedCapabilities = []module.CapabilityKind{}
	default:
		return VerifyOptions{}, fmt.Errorf("unsupported verify profile %q", profile)
	}

	if overrides.RequireEntry != nil {
		opts.RequireEntry = *overrides.RequireEntry
	}
	if overrides.MaxFunctions > 0 {
		opts.MaxFunctions = overrides.MaxFunctions
	}
	if overrides.MaxConstants > 0 {
		opts.MaxConstants = overrides.MaxConstants
	}
	if overrides.MaxCodeSizePerFunction > 0 {
		opts.MaxCodeSizePerFunction = overrides.MaxCodeSizePerFunction
	}
	if overrides.MaxLocalsPerFunction > 0 {
		opts.MaxLocalsPerFunction = overrides.MaxLocalsPerFunction
	}
	if overrides.MaxStackPerFunction > 0 {
		opts.MaxStackPerFunction = overrides.MaxStackPerFunction
	}
	if overrides.CapabilityAllowlistSet {
		opts.CapabilityAllowlistSet = true
		opts.AllowedCapabilities = append([]module.CapabilityKind(nil), overrides.AllowedCapabilities...)
	}

	return opts, nil
}

func VerifyOptionsProfileName(profile VerifyProfile, strict bool) string {
	if strict && (profile == "" || profile == VerifyProfileDefault) {
		return string(VerifyProfileStrict)
	}
	if profile != "" {
		return string(profile)
	}
	return string(VerifyProfileDefault)
}

func VerifyModule(m *module.Module, opts VerifyOptions) (*VerifyResult, error) {
	result := &VerifyResult{Valid: true}

	if err := verifyHeader(m); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	if err := verifyResourceLimits(m, opts); err != nil {
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

	if err := verifyCapabilities(m, opts); err != nil {
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

func verifyResourceLimits(m *module.Module, opts VerifyOptions) error {
	if opts.MaxFunctions > 0 && len(m.Functions) > opts.MaxFunctions {
		return fmt.Errorf("function count %d exceeds limit %d", len(m.Functions), opts.MaxFunctions)
	}
	if opts.MaxConstants > 0 && len(m.Constants) > opts.MaxConstants {
		return fmt.Errorf("module constant count %d exceeds limit %d", len(m.Constants), opts.MaxConstants)
	}
	for i, fn := range m.Functions {
		if opts.MaxCodeSizePerFunction > 0 && len(fn.Code) > opts.MaxCodeSizePerFunction {
			return fmt.Errorf("function[%d]: code size %d exceeds limit %d", i, len(fn.Code), opts.MaxCodeSizePerFunction)
		}
		if opts.MaxLocalsPerFunction > 0 && len(fn.Locals) > opts.MaxLocalsPerFunction {
			return fmt.Errorf("function[%d]: local count %d exceeds limit %d", i, len(fn.Locals), opts.MaxLocalsPerFunction)
		}
		if opts.MaxConstants > 0 && len(m.Constants) == 0 && len(fn.Constants) > opts.MaxConstants {
			return fmt.Errorf("function[%d]: constant count %d exceeds limit %d", i, len(fn.Constants), opts.MaxConstants)
		}
		if opts.MaxStackPerFunction > 0 && fn.MaxStack > uint32(opts.MaxStackPerFunction) {
			return fmt.Errorf("function[%d]: declared max stack %d exceeds limit %d", i, fn.MaxStack, opts.MaxStackPerFunction)
		}
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

		if err := verifyControlFlow(&fn, m); err != nil {
			return fmt.Errorf("function[%d]: %w", i, err)
		}

		if err := verifyConstantRefs(&fn); err != nil {
			return fmt.Errorf("function[%d]: %w", i, err)
		}

		if err := verifyStackEffects(i, &fn, m); err != nil {
			return fmt.Errorf("function[%d]: %w", i, err)
		}
	}
	return nil
}

func verifyControlFlow(fn *module.Function, m *module.Module) error {
	codeLen := len(fn.Code)
	if codeLen == 0 {
		return nil
	}

	for i, inst := range fn.Code {
		switch inst.Op {
		case core.OpJump, core.OpJumpIfFalse, core.OpJumpIfTrue,
			core.OpJumpIfNullish, core.OpJumpIfNotNullish:
			if int(inst.A) < 0 || int(inst.A) >= codeLen {
				return fmt.Errorf("instruction[%d]: jump target %d out of range [0, %d)", i, inst.A, codeLen)
			}

		case core.OpLoadLocal, core.OpStoreLocal:
			if int(inst.A) >= len(fn.Locals) {
				return fmt.Errorf("instruction[%d]: local index %d out of range (locals: %d)", i, inst.A, len(fn.Locals))
			}

		case core.OpLoadGlobal, core.OpStoreGlobal:
			// Global index validation deferred to runtime (globals are dynamic)

		case core.OpConst:
			if len(m.Constants) > 0 {
				if int(inst.A) >= len(m.Constants) {
					return fmt.Errorf("instruction[%d]: module constant index %d out of range (constants: %d)", i, inst.A, len(m.Constants))
				}
			} else {
				if int(inst.A) >= len(fn.Constants) {
					return fmt.Errorf("instruction[%d]: constant index %d out of range (constants: %d)", i, inst.A, len(fn.Constants))
				}
			}

		case core.OpCall:
			if inst.B > 0 {
				if int(inst.A) >= len(fn.Code) {
					// A is function index, validated at runtime
				}
			}

		case core.OpGetProp, core.OpSetProp:
			if len(m.Constants) > 0 {
				if int(inst.A) >= len(m.Constants) {
					return fmt.Errorf("instruction[%d]: property name constant index %d out of range (module constants: %d)", i, inst.A, len(m.Constants))
				}
			} else {
				if int(inst.A) >= len(fn.Constants) {
					return fmt.Errorf("instruction[%d]: property name constant index %d out of range (constants: %d)", i, inst.A, len(fn.Constants))
				}
			}

		case core.OpPushTry:
			if int(inst.A) < 0 || int(inst.A) >= codeLen {
				return fmt.Errorf("instruction[%d]: try handler target %d out of range [0, %d)", i, inst.A, codeLen)
			}
		}
	}
	return nil
}

func verifyStackEffects(fnIndex int, fn *module.Function, m *module.Module) error {
	codeLen := len(fn.Code)
	if codeLen == 0 {
		return nil
	}

	entryHeight := 0
	stackHeights := make([]int, codeLen)
	for i := range stackHeights {
		stackHeights[i] = -1
	}

	worklist := []int{0}
	stackHeights[0] = entryHeight

	for len(worklist) > 0 {
		pc := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]

		inst := fn.Code[pc]
		height := stackHeights[pc]

		nextHeight, err := applyStackEffect(inst, height, pc, fn, m)
		if err != nil {
			return err
		}
		if fn.MaxStack > 0 && nextHeight > int(fn.MaxStack) {
			return fmt.Errorf("instruction[%d]: stack height %d exceeds max stack %d", pc, nextHeight, fn.MaxStack)
		}

		for _, nextPC := range stackSuccessors(inst, pc, codeLen) {
			if nextPC < 0 || nextPC >= codeLen {
				continue
			}
			if stackHeights[nextPC] == -1 {
				stackHeights[nextPC] = nextHeight
				worklist = append(worklist, nextPC)
				continue
			}
			if stackHeights[nextPC] != nextHeight {
				return fmt.Errorf("instruction[%d]: stack height mismatch at target %d: existing %d, incoming %d", pc, nextPC, stackHeights[nextPC], nextHeight)
			}
		}
	}

	return nil
}

func applyStackEffect(inst core.Instruction, height int, pc int, fn *module.Function, m *module.Module) (int, error) {
	pop, push, err := stackEffect(inst, m)
	if err != nil {
		return height, fmt.Errorf("instruction[%d]: %w", pc, err)
	}
	if height < pop {
		return height, fmt.Errorf("instruction[%d]: stack underflow: need %d value(s), have %d", pc, pop, height)
	}
	return height - pop + push, nil
}

func stackEffect(inst core.Instruction, m *module.Module) (int, int, error) {
	switch inst.Op {
	case core.OpNop, core.OpJump, core.OpPushTry, core.OpPopTry:
		return 0, 0, nil
	case core.OpConst, core.OpLoadLocal, core.OpLoadGlobal, core.OpMakeObject, core.OpImportFunc, core.OpImportCap, core.OpHostPoll, core.OpClosure:
		return 0, 1, nil
	case core.OpStoreLocal, core.OpStoreGlobal, core.OpJumpIfFalse, core.OpPop, core.OpThrow:
		return 1, 0, nil
	case core.OpHostCall:
		return int(inst.A) + 1, 1, nil
	case core.OpNeg, core.OpNot, core.OpTypeof:
		return 1, 1, nil
	case core.OpDup:
		return 1, 2, nil
	case core.OpAdd, core.OpSub, core.OpMul, core.OpDiv, core.OpMod,
		core.OpEq, core.OpNe, core.OpLt, core.OpGt, core.OpLe, core.OpGe,
		core.OpBitAnd, core.OpBitOr, core.OpBitXor, core.OpShl, core.OpShr,
		core.OpAnd, core.OpOr, core.OpIndex:
		return 2, 1, nil
	case core.OpMakeArray:
		return int(inst.A), 1, nil
	case core.OpClass:
		return classStackEffect(inst), 1, nil
	case core.OpGetProp:
		return 1, 1, nil
	case core.OpSetProp:
		return 2, 0, nil
	case core.OpCall:
		return callStackEffect(inst, m)
	case core.OpNewInstance:
		return int(inst.A) + 1, 1, nil
	case core.OpReturn:
		return 0, 0, nil
	default:
		return 0, 0, fmt.Errorf("unknown opcode %v", inst.Op)
	}
}

func callStackEffect(inst core.Instruction, m *module.Module) (int, int, error) {
	if inst.B > 0 {
		fnIndex := int(inst.A)
		if fnIndex >= len(m.Functions) {
			return 0, 0, fmt.Errorf("function index %d out of range (functions: %d)", inst.A, len(m.Functions))
		}
		target := m.Functions[fnIndex]
		if int(target.TypeIndex) >= len(m.Types) {
			return 0, 0, fmt.Errorf("call target function %d type index %d out of range", fnIndex, target.TypeIndex)
		}
		ft := m.Types[target.TypeIndex]
		argCount := int(inst.B)
		if argCount != len(ft.Params) {
			return 0, 0, fmt.Errorf("call target function %d expects %d argument(s), got %d", fnIndex, len(ft.Params), argCount)
		}
		return argCount, len(ft.Results), nil
	}

	argCount := int(inst.A)
	return argCount + 1, 1, nil
}

func classStackEffect(inst core.Instruction) int {
	privateFieldCount := int((inst.A >> 16) & 0xF)
	hasParent := int((inst.A >> 20) & 1)
	instanceMethodCount := int(inst.A & 0xF)
	staticMethodCount := int((inst.A >> 4) & 0xF)
	getterCount := int((inst.A >> 8) & 0xF)
	setterCount := int((inst.A >> 12) & 0xF)
	methodPairs := instanceMethodCount + staticMethodCount + getterCount + setterCount
	return hasParent + 1 + privateFieldCount + (methodPairs * 2)
}

func stackSuccessors(inst core.Instruction, pc int, codeLen int) []int {
	next := pc + 1
	switch inst.Op {
	case core.OpJump:
		return []int{int(inst.A)}
	case core.OpJumpIfFalse:
		if next < codeLen {
			return []int{int(inst.A), next}
		}
		return []int{int(inst.A)}
	case core.OpReturn, core.OpThrow:
		return nil
	default:
		if next < codeLen {
			return []int{next}
		}
		return nil
	}
}

func verifyConstantRefs(fn *module.Function) error {
	for i, c := range fn.Constants {
		if ft, ok := c.(*module.Function); ok {
			_ = ft
			_ = i
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

func verifyCapabilities(m *module.Module, opts VerifyOptions) error {
	allowed := make(map[module.CapabilityKind]bool, len(opts.AllowedCapabilities))
	for _, kind := range opts.AllowedCapabilities {
		allowed[kind] = true
	}

	for i, cap := range m.Capabilities {
		if cap.Kind != module.CapabilityFS && cap.Kind != module.CapabilityNetwork {
			return fmt.Errorf("capability[%d]: invalid kind %q", i, cap.Kind)
		}
		if opts.CapabilityAllowlistSet && !allowed[cap.Kind] {
			return fmt.Errorf("capability[%d]: kind %q is not allowed", i, cap.Kind)
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
	return op <= core.OpNewInstance
}

func verifyStackDepth(fn *module.Function) error {
	codeLen := len(fn.Code)
	if codeLen == 0 {
		return nil
	}

	depth := 0
	maxDepth := 0
	const maxAllowed = 1024

	for i, inst := range fn.Code {
		delta := stackDelta(inst)
		if delta < 0 && depth < -delta {
			return fmt.Errorf("instruction[%d]: stack underflow (need %d, have %d)", i, -delta, depth)
		}
		depth += delta
		if depth > maxDepth {
			maxDepth = depth
		}
		if maxDepth > maxAllowed {
			return fmt.Errorf("instruction[%d]: max stack depth %d exceeds limit %d", i, maxDepth, maxAllowed)
		}
	}

	return nil
}

func stackDelta(inst core.Instruction) int {
	switch inst.Op {
	case core.OpNop, core.OpJump, core.OpPushTry, core.OpPopTry,
		core.OpReturn:
		return 0

	case core.OpThrow:
		return -1

	case core.OpConst, core.OpLoadLocal, core.OpLoadGlobal,
		core.OpMakeObject, core.OpImportFunc, core.OpImportCap,
		core.OpHostPoll, core.OpDup, core.OpClosure:
		return 1

	case core.OpClass:
		return 1 - classStackEffect(inst)

	case core.OpStoreLocal, core.OpStoreGlobal, core.OpPop:
		return -1

	case core.OpAdd, core.OpSub, core.OpMul, core.OpDiv, core.OpMod,
		core.OpEq, core.OpNe, core.OpLt, core.OpGt, core.OpLe, core.OpGe,
		core.OpBitAnd, core.OpBitOr, core.OpBitXor, core.OpShl, core.OpShr,
		core.OpAnd, core.OpOr, core.OpIndex:
		return -1

	case core.OpNeg, core.OpNot, core.OpTypeof, core.OpObjectKeys:
		return 0

	case core.OpJumpIfFalse, core.OpJumpIfTrue:
		return -1

	case core.OpJumpIfNullish, core.OpJumpIfNotNullish:
		return 0

	case core.OpCall:
		if inst.B > 0 {
			return -int(inst.B) + 1
		}
		return -int(inst.A)

	case core.OpMakeArray:
		return -int(inst.A) + 1

	case core.OpGetProp:
		return 0

	case core.OpSetProp:
		return -2

	case core.OpHostCall:
		return -int(inst.A)

	case core.OpNewInstance:
		return -int(inst.A)

	default:
		return 0
	}
}

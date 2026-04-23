package module

import "iavm/pkg/core"

type Module struct {
	Magic        string
	Version      uint16
	Target       string
	ABIVersion   uint16
	FeatureFlags uint64

	Types        []core.FuncType
	Imports      []Import
	Functions    []Function
	Globals      []Global
	Exports      []Export
	DataSegments []DataSegment
	Capabilities []CapabilityDecl
	Custom       map[string][]byte

	// Constants holds module-level constants used by OpConst instructions.
	// When non-empty, OpConst references indices into this pool rather than
	// per-function Constants.
	Constants []any
}

type Function struct {
	Name         string
	TypeIndex    uint32
	Locals       []core.ValueKind
	Code         []core.Instruction
	Constants    []any
	MaxStack     uint32
	IsEntryPoint bool
	Captures     []uint32 // outer local indices captured by this closure
}

type Global struct {
	Name    string
	Mutable bool
	Type    core.ValueKind
	Value   any
}

type DataSegment struct {
	Name   string
	Offset uint32
	Data   []byte
}

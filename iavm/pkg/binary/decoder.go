package binary

import (
	"encoding/binary"
	"fmt"
	"iavm/pkg/core"
	"iavm/pkg/module"
	"math"
)

type decoder struct {
	data []byte
	pos  int
}

func DecodeModule(raw []byte) (*module.Module, error) {
	if len(raw) < 5 {
		return nil, fmt.Errorf("data too short")
	}

	d := &decoder{data: raw}
	mod := &module.Module{}

	// Magic (4 bytes)
	magic := d.readBytes(4)
	if string(magic) != "IAVM" {
		return nil, fmt.Errorf("invalid magic: %q", magic)
	}
	mod.Magic = string(magic)

	// Version (2 bytes)
	mod.Version = d.readUint16()

	// Target (length + string)
	targetLen := d.readUint16()
	mod.Target = string(d.readBytes(int(targetLen)))

	// ABIVersion (2 bytes)
	mod.ABIVersion = d.readUint16()

	// FeatureFlags (8 bytes)
	mod.FeatureFlags = d.readUint64()

	// Sections
	for d.pos < len(d.data) {
		sectionID := module.SectionID(d.readByte())
		size := d.readUint32()
		sectionData := d.readBytes(int(size))

		if err := decodeSection(mod, sectionID, sectionData); err != nil {
			return nil, err
		}
	}

	return mod, nil
}

func decodeSection(mod *module.Module, id module.SectionID, data []byte) error {
	d := &decoder{data: data}

	switch id {
	case module.SectionType:
		return decodeTypeSection(mod, d)
	case module.SectionImport:
		return decodeImportSection(mod, d)
	case module.SectionFunction:
		return decodeFunctionSection(mod, d)
	case module.SectionGlobal:
		return decodeGlobalSection(mod, d)
	case module.SectionExport:
		return decodeExportSection(mod, d)
	case module.SectionCode:
		return decodeCodeSection(mod, d)
	case module.SectionData:
		return decodeDataSection(mod, d)
	case module.SectionCapability:
		return decodeCapabilitySection(mod, d)
	case module.SectionConstant:
		return decodeConstantSection(mod, d)
	default:
		return nil // skip unknown sections
	}
}

func decodeTypeSection(mod *module.Module, d *decoder) error {
	count := d.readUint32()
	mod.Types = make([]core.FuncType, count)
	for i := 0; i < int(count); i++ {
		paramCount := d.readUint32()
		mod.Types[i].Params = make([]core.ValueKind, paramCount)
		for j := 0; j < int(paramCount); j++ {
			mod.Types[i].Params[j] = core.ValueKind(d.readByte())
		}
		resultCount := d.readUint32()
		mod.Types[i].Results = make([]core.ValueKind, resultCount)
		for j := 0; j < int(resultCount); j++ {
			mod.Types[i].Results[j] = core.ValueKind(d.readByte())
		}
	}
	return nil
}

func decodeImportSection(mod *module.Module, d *decoder) error {
	count := d.readUint32()
	mod.Imports = make([]module.Import, count)
	for i := 0; i < int(count); i++ {
		modLen := d.readUint16()
		mod.Imports[i].Module = string(d.readBytes(int(modLen)))
		nameLen := d.readUint16()
		mod.Imports[i].Name = string(d.readBytes(int(nameLen)))
		mod.Imports[i].Kind = module.ImportKind(d.readByte())
		mod.Imports[i].Type = d.readUint32()
	}
	return nil
}

func decodeFunctionSection(mod *module.Module, d *decoder) error {
	count := d.readUint32()
	mod.Functions = make([]module.Function, count)
	for i := 0; i < int(count); i++ {
		nameLen := d.readUint16()
		mod.Functions[i].Name = string(d.readBytes(int(nameLen)))
		mod.Functions[i].TypeIndex = d.readUint32()
		mod.Functions[i].MaxStack = uint32(d.readByte())
		mod.Functions[i].IsEntryPoint = d.readByte() == 1
		localCount := d.readUint32()
		mod.Functions[i].Locals = make([]core.ValueKind, localCount)
		for j := 0; j < int(localCount); j++ {
			mod.Functions[i].Locals[j] = core.ValueKind(d.readByte())
		}
		// Decode captures (upvalue indices) if remaining data
		if d.pos < len(d.data) {
			captureCount := d.readUint32()
			mod.Functions[i].Captures = make([]uint32, captureCount)
			for j := 0; j < int(captureCount); j++ {
				mod.Functions[i].Captures[j] = d.readUint32()
			}
			if mod.FeatureFlags&module.FeatureFlagFunctionThisBindings != 0 {
				mod.Functions[i].HasThis = d.readByte() == 1
				mod.Functions[i].ThisLocal = d.readUint32()
			}
		}
	}
	return nil
}

func decodeGlobalSection(mod *module.Module, d *decoder) error {
	count := d.readUint32()
	mod.Globals = make([]module.Global, count)
	for i := 0; i < int(count); i++ {
		nameLen := d.readUint16()
		mod.Globals[i].Name = string(d.readBytes(int(nameLen)))
		mod.Globals[i].Type = core.ValueKind(d.readByte())
		mod.Globals[i].Mutable = d.readByte() == 1
	}
	return nil
}

func decodeExportSection(mod *module.Module, d *decoder) error {
	count := d.readUint32()
	mod.Exports = make([]module.Export, count)
	for i := 0; i < int(count); i++ {
		nameLen := d.readUint16()
		mod.Exports[i].Name = string(d.readBytes(int(nameLen)))
		mod.Exports[i].Kind = module.ExportKind(d.readByte())
		mod.Exports[i].Index = d.readUint32()
	}
	return nil
}

func decodeCodeSection(mod *module.Module, d *decoder) error {
	count := d.readUint32()
	if len(mod.Functions) == 0 {
		mod.Functions = make([]module.Function, count)
	}
	for i := 0; i < int(count); i++ {
		if i >= len(mod.Functions) {
			mod.Functions = append(mod.Functions, module.Function{})
		}
		// Decode constants (legacy format only when no module-level constant pool)
		if len(mod.Constants) == 0 {
			constSize := d.readUint32()
			constData := d.readBytes(int(constSize))
			cd := &decoder{data: constData}
			constCount := cd.readUint32()
			mod.Functions[i].Constants = make([]any, constCount)
			for j := 0; j < int(constCount); j++ {
				mod.Functions[i].Constants[j] = cd.decodeConstant()
			}
		}

		// Decode instructions
		instSize := d.readUint32()
		instData := d.readBytes(int(instSize))
		icd := &decoder{data: instData}
		instCount := icd.readUint32()
		mod.Functions[i].Code = make([]core.Instruction, instCount)
		for j := 0; j < int(instCount); j++ {
			mod.Functions[i].Code[j].Op = core.OpCode(icd.readUint16())
			mod.Functions[i].Code[j].A = icd.readUint32()
			mod.Functions[i].Code[j].B = icd.readUint32()
			mod.Functions[i].Code[j].C = icd.readUint32()
		}
	}
	return nil
}

func (d *decoder) decodeConstant() any {
	if d.pos >= len(d.data) {
		return nil
	}
	kind := d.readByte()
	switch kind {
	case constNil:
		return nil
	case constBool:
		return d.readByte() == 1
	case constInt64:
		return int64(d.readUint64Signed())
	case constFloat64:
		return d.readFloat64()
	case constString:
		strLen := d.readUint32()
		return string(d.readBytes(int(strLen)))
	default:
		return nil
	}
}

func decodeConstantSection(mod *module.Module, d *decoder) error {
	count := d.readUint32()
	mod.Constants = make([]any, count)
	for i := 0; i < int(count); i++ {
		mod.Constants[i] = d.decodeConstant()
	}
	return nil
}

func decodeDataSection(mod *module.Module, d *decoder) error {
	count := d.readUint32()
	mod.DataSegments = make([]module.DataSegment, count)
	for i := 0; i < int(count); i++ {
		nameLen := d.readUint16()
		mod.DataSegments[i].Name = string(d.readBytes(int(nameLen)))
		mod.DataSegments[i].Offset = d.readUint32()
		dataSize := d.readUint32()
		mod.DataSegments[i].Data = d.readBytes(int(dataSize))
	}
	return nil
}

func decodeCapabilitySection(mod *module.Module, d *decoder) error {
	count := d.readUint32()
	mod.Capabilities = make([]module.CapabilityDecl, count)
	for i := 0; i < int(count); i++ {
		kindLen := d.readUint16()
		mod.Capabilities[i].Kind = module.CapabilityKind(string(d.readBytes(int(kindLen))))
		mod.Capabilities[i].Required = d.readByte() == 1
	}
	return nil
}

func (d *decoder) readByte() byte {
	if d.pos >= len(d.data) {
		return 0
	}
	b := d.data[d.pos]
	d.pos++
	return b
}

func (d *decoder) readBytes(n int) []byte {
	if d.pos+n > len(d.data) {
		return nil
	}
	data := d.data[d.pos : d.pos+n]
	d.pos += n
	return data
}

func (d *decoder) readUint16() uint16 {
	if d.pos+2 > len(d.data) {
		return 0
	}
	val := binary.LittleEndian.Uint16(d.data[d.pos:])
	d.pos += 2
	return val
}

func (d *decoder) readUint32() uint32 {
	if d.pos+4 > len(d.data) {
		return 0
	}
	val := binary.LittleEndian.Uint32(d.data[d.pos:])
	d.pos += 4
	return val
}

func (d *decoder) readUint64() uint64 {
	if d.pos+8 > len(d.data) {
		return 0
	}
	val := binary.LittleEndian.Uint64(d.data[d.pos:])
	d.pos += 8
	return val
}

func (d *decoder) readUint64Signed() int64 {
	if d.pos+8 > len(d.data) {
		return 0
	}
	val := int64(binary.LittleEndian.Uint64(d.data[d.pos:]))
	d.pos += 8
	return val
}

func (d *decoder) readFloat64() float64 {
	if d.pos+8 > len(d.data) {
		return 0
	}
	bits := binary.LittleEndian.Uint64(d.data[d.pos:])
	d.pos += 8
	return math.Float64frombits(bits)
}

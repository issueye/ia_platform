package binary

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"iavm/pkg/core"
	"iavm/pkg/module"
)

func EncodeModule(m *module.Module) ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("nil module")
	}

	var buf bytes.Buffer

	// Magic (4 bytes)
	if len(m.Magic) != 4 {
		return nil, fmt.Errorf("invalid magic: %q", m.Magic)
	}
	if m.Magic != "IAVM" {
		return nil, fmt.Errorf("invalid magic: %q, expected 'IAVM'", m.Magic)
	}
	buf.WriteString(m.Magic)

	// Version (2 bytes, little-endian)
	binary.Write(&buf, binary.LittleEndian, m.Version)

	// Target length + string
	binary.Write(&buf, binary.LittleEndian, uint16(len(m.Target)))
	buf.WriteString(m.Target)

	// ABIVersion (2 bytes)
	binary.Write(&buf, binary.LittleEndian, m.ABIVersion)

	// FeatureFlags (8 bytes)
	binary.Write(&buf, binary.LittleEndian, m.FeatureFlags)

	// Sections
	encodeSection(&buf, module.SectionType, encodeTypeSection(m))
	encodeSection(&buf, module.SectionImport, encodeImportSection(m))
	encodeSection(&buf, module.SectionFunction, encodeFunctionSection(m))
	encodeSection(&buf, module.SectionGlobal, encodeGlobalSection(m))
	encodeSection(&buf, module.SectionExport, encodeExportSection(m))
	encodeSection(&buf, module.SectionConstant, encodeConstantSection(m))
	encodeSection(&buf, module.SectionCode, encodeCodeSection(m))
	encodeSection(&buf, module.SectionData, encodeDataSection(m))
	encodeSection(&buf, module.SectionCapability, encodeCapabilitySection(m))

	return buf.Bytes(), nil
}

func encodeSection(buf *bytes.Buffer, id module.SectionID, data []byte) {
	if len(data) == 0 {
		return
	}
	buf.WriteByte(byte(id))
	binary.Write(buf, binary.LittleEndian, uint32(len(data)))
	buf.Write(data)
}

func encodeTypeSection(m *module.Module) []byte {
	if len(m.Types) == 0 {
		return nil
	}
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(len(m.Types)))
	for _, ft := range m.Types {
		binary.Write(&buf, binary.LittleEndian, uint32(len(ft.Params)))
		for _, p := range ft.Params {
			buf.WriteByte(byte(p))
		}
		binary.Write(&buf, binary.LittleEndian, uint32(len(ft.Results)))
		for _, r := range ft.Results {
			buf.WriteByte(byte(r))
		}
	}
	return buf.Bytes()
}

func encodeImportSection(m *module.Module) []byte {
	if len(m.Imports) == 0 {
		return nil
	}
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(len(m.Imports)))
	for _, imp := range m.Imports {
		binary.Write(&buf, binary.LittleEndian, uint16(len(imp.Module)))
		buf.WriteString(imp.Module)
		binary.Write(&buf, binary.LittleEndian, uint16(len(imp.Name)))
		buf.WriteString(imp.Name)
		buf.WriteByte(byte(imp.Kind))
		binary.Write(&buf, binary.LittleEndian, imp.Type)
	}
	return buf.Bytes()
}

func encodeFunctionSection(m *module.Module) []byte {
	if len(m.Functions) == 0 {
		return nil
	}
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(len(m.Functions)))
	for _, fn := range m.Functions {
		binary.Write(&buf, binary.LittleEndian, uint16(len(fn.Name)))
		buf.WriteString(fn.Name)
		binary.Write(&buf, binary.LittleEndian, fn.TypeIndex)
		buf.WriteByte(byte(fn.MaxStack))
		if fn.IsEntryPoint {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
		binary.Write(&buf, binary.LittleEndian, uint32(len(fn.Locals)))
		for _, l := range fn.Locals {
			buf.WriteByte(byte(l))
		}
	}
	return buf.Bytes()
}

func encodeGlobalSection(m *module.Module) []byte {
	if len(m.Globals) == 0 {
		return nil
	}
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(len(m.Globals)))
	for _, g := range m.Globals {
		binary.Write(&buf, binary.LittleEndian, uint16(len(g.Name)))
		buf.WriteString(g.Name)
		buf.WriteByte(byte(g.Type))
		if g.Mutable {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
	}
	return buf.Bytes()
}

func encodeExportSection(m *module.Module) []byte {
	if len(m.Exports) == 0 {
		return nil
	}
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(len(m.Exports)))
	for _, exp := range m.Exports {
		binary.Write(&buf, binary.LittleEndian, uint16(len(exp.Name)))
		buf.WriteString(exp.Name)
		buf.WriteByte(byte(exp.Kind))
		binary.Write(&buf, binary.LittleEndian, exp.Index)
	}
	return buf.Bytes()
}

func encodeCodeSection(m *module.Module) []byte {
	if len(m.Functions) == 0 {
		return nil
	}
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(len(m.Functions)))
	for _, fn := range m.Functions {
		if len(m.Constants) == 0 {
			// Legacy: encode per-function constants
			constData := encodeConstants(fn.Constants)
			binary.Write(&buf, binary.LittleEndian, uint32(len(constData)))
			buf.Write(constData)
		}

		// Encode instructions
		instData := encodeInstructions(fn.Code)
		binary.Write(&buf, binary.LittleEndian, uint32(len(instData)))
		buf.Write(instData)
	}
	return buf.Bytes()
}

func encodeConstantSection(m *module.Module) []byte {
	if len(m.Constants) == 0 {
		return nil
	}
	return encodeConstants(m.Constants)
}

const (
	constNil uint8 = iota
	constBool
	constInt64
	constFloat64
	constString
	constFuncRef
)

func encodeConstants(constants []any) []byte {
	if len(constants) == 0 {
		return nil
	}
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(len(constants)))
	for _, c := range constants {
		switch v := c.(type) {
		case nil:
			buf.WriteByte(constNil)
		case bool:
			buf.WriteByte(constBool)
			if v {
				buf.WriteByte(1)
			} else {
				buf.WriteByte(0)
			}
		case int:
			buf.WriteByte(constInt64)
			binary.Write(&buf, binary.LittleEndian, int64(v))
		case int64:
			buf.WriteByte(constInt64)
			binary.Write(&buf, binary.LittleEndian, v)
		case float64:
			buf.WriteByte(constFloat64)
			binary.Write(&buf, binary.LittleEndian, v)
		case string:
			buf.WriteByte(constString)
			binary.Write(&buf, binary.LittleEndian, uint32(len(v)))
			buf.WriteString(v)
		default:
			// Unknown type, encode as nil
			buf.WriteByte(constNil)
		}
	}
	return buf.Bytes()
}

func encodeInstructions(instructions []core.Instruction) []byte {
	if len(instructions) == 0 {
		return nil
	}
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(len(instructions)))
	for _, inst := range instructions {
		binary.Write(&buf, binary.LittleEndian, uint16(inst.Op))
		binary.Write(&buf, binary.LittleEndian, inst.A)
		binary.Write(&buf, binary.LittleEndian, inst.B)
		binary.Write(&buf, binary.LittleEndian, inst.C)
	}
	return buf.Bytes()
}

func encodeDataSection(m *module.Module) []byte {
	if len(m.DataSegments) == 0 {
		return nil
	}
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(len(m.DataSegments)))
	for _, ds := range m.DataSegments {
		binary.Write(&buf, binary.LittleEndian, uint16(len(ds.Name)))
		buf.WriteString(ds.Name)
		binary.Write(&buf, binary.LittleEndian, ds.Offset)
		binary.Write(&buf, binary.LittleEndian, uint32(len(ds.Data)))
		buf.Write(ds.Data)
	}
	return buf.Bytes()
}

func encodeCapabilitySection(m *module.Module) []byte {
	if len(m.Capabilities) == 0 {
		return nil
	}
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(len(m.Capabilities)))
	for _, cap := range m.Capabilities {
		binary.Write(&buf, binary.LittleEndian, uint16(len(cap.Kind)))
		buf.WriteString(string(cap.Kind))
		if cap.Required {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
	}
	return buf.Bytes()
}

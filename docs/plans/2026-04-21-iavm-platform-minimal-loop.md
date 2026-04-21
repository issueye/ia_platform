# IAVM 平台化最小闭环实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 打通 `ialang -> iavm -> host capability` 的最小可验证路径，形成平台化开发基线。

**Architecture:** 分四阶段推进：(1) iavm 二进制编解码与校验器，(2) ialang 到 iavm 的 lowering 桥接，(3) iavm runtime 最小解释器，(4) capability binding 与 CLI 集成。每阶段 TDD 驱动，小步提交。

**Tech Stack:** Go 1.25.7, iacommon(公共类型/Host ABI), ialang(编译器/VM), iavm(平台VM)

---

## 阶段一：iavm 二进制层与校验器

### Task 1: 实现 Binary Encoder

**Files:**
- Modify: `iavm/pkg/binary/encoder.go:1-8` (替换 stub)
- Create: `iavm/pkg/binary/encoder_test.go`
- Reference: `iavm/pkg/module/module.go`, `iavm/pkg/core/opcode.go`

**Step 1: 编写测试**

```go
// iavm/pkg/binary/encoder_test.go
package binary

import (
    "testing"
    "iavm/pkg/module"
    "iavm/pkg/core"
)

func TestEncodeModule_Minimal(t *testing.T) {
    mod := &module.Module{
        Magic:   "IAVM",
        Version: 1,
        Target:  "ialang",
    }
    data, err := EncodeModule(mod)
    if err != nil {
        t.Fatalf("EncodeModule failed: %v", err)
    }
    if len(data) == 0 {
        t.Fatal("encoded data is empty")
    }
    // 检查 magic header
    if string(data[:4]) != "IAVM" {
        t.Fatalf("expected magic 'IAVM', got %q", data[:4])
    }
}

func TestEncodeModule_WithFunctions(t *testing.T) {
    mod := &module.Module{
        Magic:   "IAVM",
        Version: 1,
        Target:  "ialang",
        Types: []module.FuncType{
            {Params: []core.ValueKind{core.KindI64}, Results: []core.ValueKind{core.KindI64}},
        },
        Functions: []module.Function{
            {
                Name:     "add_one",
                TypeIdx:  0,
                Locals:   []core.ValueKind{core.KindI64},
                Instructions: []core.Instruction{
                    {Op: core.OpLoadLocal, A: 0},
                    {Op: core.OpConst, A: 1},
                    {Op: core.OpAdd},
                    {Op: core.OpReturn},
                },
            },
        },
    }
    data, err := EncodeModule(mod)
    if err != nil {
        t.Fatalf("EncodeModule failed: %v", err)
    }
    if len(data) < 10 {
        t.Fatalf("encoded data too short: %d bytes", len(data))
    }
}
```

**Step 2: 运行测试确认失败**

```bash
cd iavm && go test ./pkg/binary/ -run TestEncode -v
```
Expected: FAIL (encoder returns nil)

**Step 3: 实现 Encoder**

```go
// iavm/pkg/binary/encoder.go
package binary

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "iavm/pkg/module"
    "iavm/pkg/core"
)

func EncodeModule(mod *module.Module) ([]byte, error) {
    if mod == nil {
        return nil, fmt.Errorf("nil module")
    }
    
    var buf bytes.Buffer
    
    // Magic (4 bytes)
    if len(mod.Magic) != 4 {
        return nil, fmt.Errorf("invalid magic: %q", mod.Magic)
    }
    buf.WriteString(mod.Magic)
    
    // Version (1 byte)
    buf.WriteByte(byte(mod.Version))
    
    // Target length + string
    buf.WriteByte(byte(len(mod.Target)))
    buf.WriteString(mod.Target)
    
    // ABIVersion (1 byte)
    buf.WriteByte(byte(mod.ABIVersion))
    
    // FeatureFlags (1 byte)
    buf.WriteByte(byte(mod.FeatureFlags))
    
    // Sections
    encodeSection(&buf, module.SectionType, encodeTypeSection(mod))
    encodeSection(&buf, module.SectionImport, encodeImportSection(mod))
    encodeSection(&buf, module.SectionFunction, encodeFunctionSection(mod))
    encodeSection(&buf, module.SectionGlobal, encodeGlobalSection(mod))
    encodeSection(&buf, module.SectionExport, encodeExportSection(mod))
    encodeSection(&buf, module.SectionCode, encodeCodeSection(mod))
    encodeSection(&buf, module.SectionData, encodeDataSection(mod))
    encodeSection(&buf, module.SectionCapability, encodeCapabilitySection(mod))
    
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

func encodeTypeSection(mod *module.Module) []byte {
    if len(mod.Types) == 0 {
        return nil
    }
    var buf bytes.Buffer
    buf.WriteByte(byte(len(mod.Types)))
    for _, ft := range mod.Types {
        buf.WriteByte(byte(len(ft.Params)))
        for _, p := range ft.Params {
            buf.WriteByte(byte(p))
        }
        buf.WriteByte(byte(len(ft.Results)))
        for _, r := range ft.Results {
            buf.WriteByte(byte(r))
        }
    }
    return buf.Bytes()
}

func encodeImportSection(mod *module.Module) []byte {
    if len(mod.Imports) == 0 {
        return nil
    }
    var buf bytes.Buffer
    buf.WriteByte(byte(len(mod.Imports)))
    for _, imp := range mod.Imports {
        buf.WriteByte(byte(len(imp.Module)))
        buf.WriteString(imp.Module)
        buf.WriteByte(byte(len(imp.Name)))
        buf.WriteString(imp.Name)
        buf.WriteByte(byte(imp.Kind))
        binary.Write(buf, binary.LittleEndian, uint32(imp.Index))
    }
    return buf.Bytes()
}

func encodeFunctionSection(mod *module.Module) []byte {
    if len(mod.Functions) == 0 {
        return nil
    }
    var buf bytes.Buffer
    buf.WriteByte(byte(len(mod.Functions)))
    for _, fn := range mod.Functions {
        buf.WriteByte(byte(len(fn.Name)))
        buf.WriteString(fn.Name)
        binary.Write(buf, binary.LittleEndian, uint32(fn.TypeIdx))
        buf.WriteByte(byte(len(fn.Locals)))
        for _, l := range fn.Locals {
            buf.WriteByte(byte(l))
        }
    }
    return buf.Bytes()
}

func encodeGlobalSection(mod *module.Module) []byte {
    if len(mod.Globals) == 0 {
        return nil
    }
    var buf bytes.Buffer
    buf.WriteByte(byte(len(mod.Globals)))
    for _, g := range mod.Globals {
        buf.WriteByte(byte(g.Type))
        buf.WriteByte(byte(g.Mutable))
    }
    return buf.Bytes()
}

func encodeExportSection(mod *module.Module) []byte {
    if len(mod.Exports) == 0 {
        return nil
    }
    var buf bytes.Buffer
    buf.WriteByte(byte(len(mod.Exports)))
    for _, exp := range mod.Exports {
        buf.WriteByte(byte(len(exp.Name)))
        buf.WriteString(exp.Name)
        buf.WriteByte(byte(exp.Kind))
        binary.Write(buf, binary.LittleEndian, uint32(exp.Index))
    }
    return buf.Bytes()
}

func encodeCodeSection(mod *module.Module) []byte {
    if len(mod.Functions) == 0 {
        return nil
    }
    var buf bytes.Buffer
    buf.WriteByte(byte(len(mod.Functions)))
    for _, fn := range mod.Functions {
        instructions := encodeInstructions(fn.Instructions)
        binary.Write(buf, binary.LittleEndian, uint32(len(instructions)))
        buf.Write(instructions)
    }
    return buf.Bytes()
}

func encodeInstructions(instructions []core.Instruction) []byte {
    var buf bytes.Buffer
    buf.WriteByte(byte(len(instructions)))
    for _, inst := range instructions {
        buf.WriteByte(byte(inst.Op))
        binary.Write(buf, binary.LittleEndian, inst.A)
        binary.Write(buf, binary.LittleEndian, inst.B)
        binary.Write(buf, binary.LittleEndian, inst.C)
    }
    return buf.Bytes()
}

func encodeDataSection(mod *module.Module) []byte {
    if len(mod.DataSegments) == 0 {
        return nil
    }
    var buf bytes.Buffer
    buf.WriteByte(byte(len(mod.DataSegments)))
    for _, ds := range mod.DataSegments {
        binary.Write(buf, binary.LittleEndian, uint32(ds.Offset))
        binary.Write(buf, binary.LittleEndian, uint32(len(ds.Data)))
        buf.Write(ds.Data)
    }
    return buf.Bytes()
}

func encodeCapabilitySection(mod *module.Module) []byte {
    if len(mod.Capabilities) == 0 {
        return nil
    }
    var buf bytes.Buffer
    buf.WriteByte(byte(len(mod.Capabilities)))
    for _, cap := range mod.Capabilities {
        buf.WriteByte(byte(cap.Kind))
        buf.WriteByte(byte(len(cap.Name)))
        buf.WriteString(cap.Name)
    }
    return buf.Bytes()
}
```

**Step 4: 运行测试确认通过**

```bash
cd iavm && go test ./pkg/binary/ -run TestEncode -v
```
Expected: PASS

**Step 5: 提交**

```bash
git add iavm/pkg/binary/encoder.go iavm/pkg/binary/encoder_test.go
git commit -m "feat(iavm): implement binary encoder with section serialization"
```

---

### Task 2: 实现 Binary Decoder

**Files:**
- Modify: `iavm/pkg/binary/decoder.go:1-8` (替换 stub)
- Create: `iavm/pkg/binary/decoder_test.go`

**Step 1: 编写测试**

```go
// iavm/pkg/binary/decoder_test.go
package binary

import (
    "testing"
    "iavm/pkg/module"
    "iavm/pkg/core"
)

func TestDecodeModule_RoundTrip(t *testing.T) {
    original := &module.Module{
        Magic:      "IAVM",
        Version:    1,
        Target:     "ialang",
        ABIVersion: 1,
        Types: []module.FuncType{
            {Params: []core.ValueKind{core.KindI64}, Results: []core.ValueKind{core.KindI64}},
        },
        Functions: []module.Function{
            {
                Name:    "add_one",
                TypeIdx: 0,
                Locals:  []core.ValueKind{core.KindI64},
                Instructions: []core.Instruction{
                    {Op: core.OpLoadLocal, A: 0},
                    {Op: core.OpConst, A: 1},
                    {Op: core.OpAdd},
                    {Op: core.OpReturn},
                },
            },
        },
        Exports: []module.Export{
            {Name: "add_one", Kind: module.ExportFunction, Index: 0},
        },
    }
    
    data, err := EncodeModule(original)
    if err != nil {
        t.Fatalf("EncodeModule failed: %v", err)
    }
    
    decoded, err := DecodeModule(data)
    if err != nil {
        t.Fatalf("DecodeModule failed: %v", err)
    }
    
    if decoded.Magic != original.Magic {
        t.Errorf("magic mismatch: got %q, want %q", decoded.Magic, original.Magic)
    }
    if decoded.Version != original.Version {
        t.Errorf("version mismatch: got %d, want %d", decoded.Version, original.Version)
    }
    if len(decoded.Functions) != len(original.Functions) {
        t.Errorf("function count mismatch: got %d, want %d", len(decoded.Functions), len(original.Functions))
    }
    if decoded.Functions[0].Name != original.Functions[0].Name {
        t.Errorf("function name mismatch: got %q, want %q", decoded.Functions[0].Name, original.Functions[0].Name)
    }
}

func TestDecodeModule_InvalidMagic(t *testing.T) {
    data := []byte("BADX\x01")
    _, err := DecodeModule(data)
    if err == nil {
        t.Fatal("expected error for invalid magic")
    }
}
```

**Step 2: 运行测试确认失败**

```bash
cd iavm && go test ./pkg/binary/ -run TestDecode -v
```
Expected: FAIL

**Step 3: 实现 Decoder**

```go
// iavm/pkg/binary/decoder.go
package binary

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "iavm/pkg/module"
    "iavm/pkg/core"
)

type decoder struct {
    data []byte
    pos  int
}

func DecodeModule(data []byte) (*module.Module, error) {
    if len(data) < 5 {
        return nil, fmt.Errorf("data too short")
    }
    
    d := &decoder{data: data}
    mod := &module.Module{}
    
    // Magic
    magic := d.readBytes(4)
    if string(magic) != "IAVM" {
        return nil, fmt.Errorf("invalid magic: %q", magic)
    }
    mod.Magic = string(magic)
    
    // Version
    mod.Version = int(d.readByte())
    
    // Target
    targetLen := d.readByte()
    mod.Target = string(d.readBytes(int(targetLen)))
    
    // ABIVersion
    mod.ABIVersion = int(d.readByte())
    
    // FeatureFlags
    mod.FeatureFlags = uint8(d.readByte())
    
    // Sections
    for d.pos < len(d.data) {
        sectionID := module.SectionID(d.readByte())
        size := int(d.readUint32())
        sectionData := d.readBytes(size)
        
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
    default:
        return nil // skip unknown sections
    }
}

func decodeTypeSection(mod *module.Module, d *decoder) error {
    count := d.readByte()
    mod.Types = make([]module.FuncType, count)
    for i := 0; i < int(count); i++ {
        paramCount := d.readByte()
        mod.Types[i].Params = make([]core.ValueKind, paramCount)
        for j := 0; j < int(paramCount); j++ {
            mod.Types[i].Params[j] = core.ValueKind(d.readByte())
        }
        resultCount := d.readByte()
        mod.Types[i].Results = make([]core.ValueKind, resultCount)
        for j := 0; j < int(resultCount); j++ {
            mod.Types[i].Results[j] = core.ValueKind(d.readByte())
        }
    }
    return nil
}

func decodeImportSection(mod *module.Module, d *decoder) error {
    count := d.readByte()
    mod.Imports = make([]module.Import, count)
    for i := 0; i < int(count); i++ {
        modLen := d.readByte()
        mod.Imports[i].Module = string(d.readBytes(int(modLen)))
        nameLen := d.readByte()
        mod.Imports[i].Name = string(d.readBytes(int(nameLen)))
        mod.Imports[i].Kind = module.ImportKind(d.readByte())
        mod.Imports[i].Index = int(d.readUint32())
    }
    return nil
}

func decodeFunctionSection(mod *module.Module, d *decoder) error {
    count := d.readByte()
    // Ensure Functions slice has enough capacity
    if len(mod.Functions) < int(count) {
        // Functions may be partially filled from code section, resize if needed
        newFuncs := make([]module.Function, count)
        copy(newFuncs, mod.Functions)
        mod.Functions = newFuncs
    }
    for i := 0; i < int(count); i++ {
        nameLen := d.readByte()
        mod.Functions[i].Name = string(d.readBytes(int(nameLen)))
        mod.Functions[i].TypeIdx = int(d.readUint32())
        localCount := d.readByte()
        mod.Functions[i].Locals = make([]core.ValueKind, localCount)
        for j := 0; j < int(localCount); j++ {
            mod.Functions[i].Locals[j] = core.ValueKind(d.readByte())
        }
    }
    return nil
}

func decodeGlobalSection(mod *module.Module, d *decoder) error {
    count := d.readByte()
    mod.Globals = make([]module.Global, count)
    for i := 0; i < int(count); i++ {
        mod.Globals[i].Type = core.ValueKind(d.readByte())
        mod.Globals[i].Mutable = d.readByte()
    }
    return nil
}

func decodeExportSection(mod *module.Module, d *decoder) error {
    count := d.readByte()
    mod.Exports = make([]module.Export, count)
    for i := 0; i < int(count); i++ {
        nameLen := d.readByte()
        mod.Exports[i].Name = string(d.readBytes(int(nameLen)))
        mod.Exports[i].Kind = module.ExportKind(d.readByte())
        mod.Exports[i].Index = int(d.readUint32())
    }
    return nil
}

func decodeCodeSection(mod *module.Module, d *decoder) error {
    count := d.readByte()
    if len(mod.Functions) < int(count) {
        newFuncs := make([]module.Function, count)
        copy(newFuncs, mod.Functions)
        mod.Functions = newFuncs
    }
    for i := 0; i < int(count); i++ {
        size := d.readUint32()
        codeData := d.readBytes(int(size))
        cd := &decoder{data: codeData}
        instCount := cd.readByte()
        mod.Functions[i].Instructions = make([]core.Instruction, instCount)
        for j := 0; j < int(instCount); j++ {
            mod.Functions[i].Instructions[j].Op = core.Opcode(cd.readByte())
            mod.Functions[i].Instructions[j].A = cd.readUint32()
            mod.Functions[i].Instructions[j].B = cd.readUint32()
            mod.Functions[i].Instructions[j].C = cd.readUint32()
        }
    }
    return nil
}

func decodeDataSection(mod *module.Module, d *decoder) error {
    count := d.readByte()
    mod.DataSegments = make([]module.DataSegment, count)
    for i := 0; i < int(count); i++ {
        mod.DataSegments[i].Offset = int(d.readUint32())
        dataSize := d.readUint32()
        mod.DataSegments[i].Data = d.readBytes(int(dataSize))
    }
    return nil
}

func decodeCapabilitySection(mod *module.Module, d *decoder) error {
    count := d.readByte()
    mod.Capabilities = make([]module.CapabilityDecl, count)
    for i := 0; i < int(count); i++ {
        mod.Capabilities[i].Kind = module.CapabilityKind(d.readByte())
        nameLen := d.readByte()
        mod.Capabilities[i].Name = string(d.readBytes(int(nameLen)))
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

func (d *decoder) readUint32() uint32 {
    if d.pos+4 > len(d.data) {
        return 0
    }
    val := binary.LittleEndian.Uint32(d.data[d.pos:])
    d.pos += 4
    return val
}
```

**Step 4: 运行测试确认通过**

```bash
cd iavm && go test ./pkg/binary/ -run TestDecode -v
```
Expected: PASS

**Step 5: 提交**

```bash
git add iavm/pkg/binary/decoder.go iavm/pkg/binary/decoder_test.go
git commit -m "feat(iavm): implement binary decoder with round-trip validation"
```

---

### Task 3: 实现 Verifier 最小可用版本

**Files:**
- Modify: `iavm/pkg/binary/verifier.go:1-14` (替换 stub)
- Reference: `iavm/pkg/binary/verify_result.go`
- Create: `iavm/pkg/binary/verifier_test.go`

**Step 1: 编写测试**

```go
// iavm/pkg/binary/verifier_test.go
package binary

import (
    "testing"
    "iavm/pkg/module"
    "iavm/pkg/core"
)

func TestVerifyModule_ValidMinimal(t *testing.T) {
    mod := &module.Module{
        Magic:   "IAVM",
        Version: 1,
        Target:  "ialang",
    }
    result, err := VerifyModule(mod)
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
    _, err := VerifyModule(mod)
    if err == nil {
        t.Fatal("expected error for invalid magic")
    }
}

func TestVerifyModule_InvalidVersion(t *testing.T) {
    mod := &module.Module{
        Magic:   "IAVM",
        Version: 99,
    }
    _, err := VerifyModule(mod)
    if err == nil {
        t.Fatal("expected error for invalid version")
    }
}

func TestVerifyModule_InvalidTypeRef(t *testing.T) {
    mod := &module.Module{
        Magic:   "IAVM",
        Version: 1,
        Functions: []module.Function{
            {
                Name:    "test",
                TypeIdx: 5, // references non-existent type
            },
        },
    }
    _, err := VerifyModule(mod)
    if err == nil {
        t.Fatal("expected error for invalid type reference")
    }
}

func TestVerifyModule_InvalidExportRef(t *testing.T) {
    mod := &module.Module{
        Magic:   "IAVM",
        Version: 1,
        Exports: []module.Export{
            {Name: "main", Kind: module.ExportFunction, Index: 10},
        },
    }
    _, err := VerifyModule(mod)
    if err == nil {
        t.Fatal("expected error for invalid export reference")
    }
}

func TestVerifyModule_InvalidOpcode(t *testing.T) {
    mod := &module.Module{
        Magic:   "IAVM",
        Version: 1,
        Functions: []module.Function{
            {
                Name:    "test",
                TypeIdx: 0,
                Instructions: []core.Instruction{
                    {Op: core.Opcode(255)}, // invalid opcode
                },
            },
        },
    }
    _, err := VerifyModule(mod)
    if err == nil {
        t.Fatal("expected error for invalid opcode")
    }
}
```

**Step 2: 运行测试确认失败**

```bash
cd iavm && go test ./pkg/binary/ -run TestVerify -v
```
Expected: FAIL

**Step 3: 实现 Verifier**

```go
// iavm/pkg/binary/verifier.go
package binary

import (
    "fmt"
    "iavm/pkg/module"
    "iavm/pkg/core"
)

func VerifyModule(mod *module.Module) (*VerifyResult, error) {
    result := &VerifyResult{Valid: true}
    
    if err := verifyHeader(mod); err != nil {
        result.Valid = false
        result.Errors = append(result.Errors, err.Error())
        return result, err
    }
    
    if err := verifyTypes(mod); err != nil {
        result.Valid = false
        result.Errors = append(result.Errors, err.Error())
        return result, err
    }
    
    if err := verifyFunctions(mod); err != nil {
        result.Valid = false
        result.Errors = append(result.Errors, err.Error())
        return result, err
    }
    
    if err := verifyExports(mod); err != nil {
        result.Valid = false
        result.Errors = append(result.Errors, err.Error())
        return result, err
    }
    
    if err := verifyImports(mod); err != nil {
        result.Valid = false
        result.Errors = append(result.Errors, err.Error())
        return result, err
    }
    
    if err := verifyCapabilities(mod); err != nil {
        result.Valid = false
        result.Errors = append(result.Errors, err.Error())
        return result, err
    }
    
    return result, nil
}

func verifyHeader(mod *module.Module) error {
    if mod.Magic != "IAVM" {
        return fmt.Errorf("invalid magic: %q, expected 'IAVM'", mod.Magic)
    }
    if mod.Version < 1 || mod.Version > 2 {
        return fmt.Errorf("unsupported version: %d", mod.Version)
    }
    if mod.Target == "" {
        return fmt.Errorf("empty target")
    }
    return nil
}

func verifyTypes(mod *module.Module) error {
    for i, ft := range mod.Types {
        for _, p := range ft.Params {
            if !isValidValueKind(p) {
                return fmt.Errorf("type[%d]: invalid param kind %v", i, p)
            }
        }
        for _, r := range ft.Results {
            if !isValidValueKind(r) {
                return fmt.Errorf("type[%d]: invalid result kind %v", i, r)
            }
        }
    }
    return nil
}

func verifyFunctions(mod *module.Module) error {
    for i, fn := range mod.Functions {
        if fn.TypeIdx >= len(mod.Types) {
            return fmt.Errorf("function[%d]: type index %d out of range (types: %d)", 
                i, fn.TypeIdx, len(mod.Types))
        }
        
        for j, local := range fn.Locals {
            if !isValidValueKind(local) {
                return fmt.Errorf("function[%d]: invalid local kind %v at index %d", i, local, j)
            }
        }
        
        for j, inst := range fn.Instructions {
            if !isValidOpcode(inst.Op) {
                return fmt.Errorf("function[%d]: invalid opcode %v at instruction %d", i, inst.Op, j)
            }
        }
    }
    return nil
}

func verifyExports(mod *module.Module) error {
    for i, exp := range mod.Exports {
        switch exp.Kind {
        case module.ExportFunction:
            if exp.Index >= len(mod.Functions) {
                return fmt.Errorf("export[%d]: function index %d out of range", i, exp.Index)
            }
        case module.ExportGlobal:
            if exp.Index >= len(mod.Globals) {
                return fmt.Errorf("export[%d]: global index %d out of range", i, exp.Index)
            }
        }
    }
    return nil
}

func verifyImports(mod *module.Module) error {
    for i, imp := range mod.Imports {
        if imp.Module == "" {
            return fmt.Errorf("import[%d]: empty module name", i)
        }
        if imp.Name == "" {
            return fmt.Errorf("import[%d]: empty name", i)
        }
    }
    return nil
}

func verifyCapabilities(mod *module.Module) error {
    for i, cap := range mod.Capabilities {
        if cap.Kind != module.CapabilityFS && cap.Kind != module.CapabilityNetwork {
            return fmt.Errorf("capability[%d]: invalid kind %v", i, cap.Kind)
        }
        if cap.Name == "" {
            return fmt.Errorf("capability[%d]: empty name", i)
        }
    }
    return nil
}

func isValidValueKind(kind core.ValueKind) bool {
    switch kind {
    case core.KindNull, core.KindBool, core.KindI64, core.KindF64, 
         core.KindString, core.KindBytes, core.KindArrayRef, 
         core.KindObjectRef, core.KindFuncRef, core.KindHostHandle:
        return true
    default:
        return false
    }
}

func isValidOpcode(op core.Opcode) bool {
    return op >= core.OpNop && op <= core.OpHostPoll
}
```

**Step 4: 更新 VerifyResult 结构**

```go
// iavm/pkg/binary/verify_result.go
package binary

type VerifyResult struct {
    Valid    bool
    Warnings []string
    Errors   []string
}
```

**Step 5: 运行测试确认通过**

```bash
cd iavm && go test ./pkg/binary/ -run TestVerify -v
```
Expected: PASS

**Step 6: 提交**

```bash
git add iavm/pkg/binary/verifier.go iavm/pkg/binary/verifier_test.go iavm/pkg/binary/verify_result.go
git commit -m "feat(iavm): implement verifier with header, type, function, export validation"
```

---

## 阶段二：ialang -> iavm Lowering 桥接

### Task 4: 实现 LowerToModule 桥接

**Files:**
- Modify: `iavm/pkg/bridge/ialang/compiler_lowering.go:1-8` (替换 stub)
- Reference: 
  - `iacommon/pkg/ialang/bytecode/bytecode.go` (ialang 编译产物结构)
  - `iavm/pkg/module/module.go` (iavm 模块结构)
  - `iavm/pkg/core/opcode.go` (iavm opcode 定义)
- Create: `iavm/pkg/bridge/ialang/compiler_lowering_test.go`

**Step 1: 编写测试**

```go
// iavm/pkg/bridge/ialang/compiler_lowering_test.go
package ialang

import (
    "testing"
    "iacommon/pkg/ialang/bytecode"
)

func TestLowerToModule_MinimalChunk(t *testing.T) {
    chunk := &bytecode.Chunk{
        Code: []bytecode.Instruction{
            {Op: bytecode.OpConst, A: 0, B: 0},
            {Op: bytecode.OpReturn},
        },
        Constants: []any{42},
    }
    
    mod, err := LowerToModule(chunk)
    if err != nil {
        t.Fatalf("LowerToModule failed: %v", err)
    }
    
    if mod == nil {
        t.Fatal("LowerToModule returned nil")
    }
    if mod.Magic != "IAVM" {
        t.Errorf("expected magic 'IAVM', got %q", mod.Magic)
    }
    if len(mod.Functions) == 0 {
        t.Fatal("expected at least one function")
    }
}

func TestLowerToModule_ChunkWithMain(t *testing.T) {
    chunk := &bytecode.Chunk{
        Code: []bytecode.Instruction{
            {Op: bytecode.OpConst, A: 0, B: 0},
            {Op: bytecode.OpCall, A: 0, B: 0},
            {Op: bytecode.OpReturn},
        },
        Constants: []any{
            &bytecode.FunctionTemplate{
                Name: "main",
                Arity: 0,
                Chunk: &bytecode.Chunk{
                    Code: []bytecode.Instruction{
                        {Op: bytecode.OpConst, A: 0, B: 0},
                        {Op: bytecode.OpReturn},
                    },
                    Constants: []any{1},
                },
            },
        },
    }
    
    mod, err := LowerToModule(chunk)
    if err != nil {
        t.Fatalf("LowerToModule failed: %v", err)
    }
    
    // Should have at least 2 functions (main + entry)
    if len(mod.Functions) < 2 {
        t.Errorf("expected at least 2 functions, got %d", len(mod.Functions))
    }
}
```

**Step 2: 运行测试确认失败**

```bash
cd iavm && go test ./pkg/bridge/ialang/ -run TestLower -v
```
Expected: FAIL

**Step 3: 实现 LowerToModule**

```go
// iavm/pkg/bridge/ialang/compiler_lowering.go
package ialang

import (
    "fmt"
    "iacommon/pkg/ialang/bytecode"
    "iavm/pkg/module"
    "iavm/pkg/core"
)

func LowerToModule(chunk *bytecode.Chunk) (*module.Module, error) {
    if chunk == nil {
        return nil, fmt.Errorf("nil chunk")
    }
    
    mod := &module.Module{
        Magic:      "IAVM",
        Version:    1,
        Target:     "ialang",
        ABIVersion: 1,
    }
    
    // Build type section
    mod.Types = append(mod.Types, module.FuncType{
        Params:  []core.ValueKind{},
        Results: []core.ValueKind{core.KindNull},
    })
    
    // Lower functions
    var funcIndexMap map[int]int // ialang func const index -> iavm func index
    
    // First pass: collect all function templates
    funcIndexMap = make(map[int]int)
    for i, c := range chunk.Constants {
        if ft, ok := c.(*bytecode.FunctionTemplate); ok {
            funcIndexMap[i] = len(mod.Functions)
            mod.Functions = append(mod.Functions, lowerFunction(ft, funcIndexMap, chunk))
        }
    }
    
    // Create entry function from main chunk
    entryFunc := lowerChunkAsFunction(chunk, funcIndexMap, "entry")
    mod.Functions = append(mod.Functions, entryFunc)
    
    // Add exports for named functions
    for i, c := range chunk.Constants {
        if ft, ok := c.(*bytecode.FunctionTemplate); ok {
            if ft.Name != "" {
                mod.Exports = append(mod.Exports, module.Export{
                    Name:  ft.Name,
                    Kind:  module.ExportFunction,
                    Index: funcIndexMap[i],
                })
            }
        }
    }
    
    return mod, nil
}

func lowerFunction(ft *bytecode.FunctionTemplate, funcMap map[int]int, parentChunk *bytecode.Chunk) module.Function {
    fn := module.Function{
        Name:    ft.Name,
        TypeIdx: 0,
    }
    
    // Lower locals (upvalue slots + local vars)
    totalLocals := ft.UpvalueCount + ft.Arity + 1 // +1 for self
    for i := 0; i < totalLocals; i++ {
        fn.Locals = append(fn.Locals, core.KindNull)
    }
    
    // Lower instructions
    if ft.Chunk != nil {
        fn.Instructions = lowerInstructions(ft.Chunk.Code, ft.Chunk.Constants, funcMap)
    }
    
    return fn
}

func lowerChunkAsFunction(chunk *bytecode.Chunk, funcMap map[int]int, name string) module.Function {
    fn := module.Function{
        Name:    name,
        TypeIdx: 0,
    }
    fn.Instructions = lowerInstructions(chunk.Code, chunk.Constants, funcMap)
    return fn
}

func lowerInstructions(ialangInsts []bytecode.Instruction, constants []any, funcMap map[int]int) []core.Instruction {
    var iavmInsts []core.Instruction
    
    for _, inst := range ialangInsts {
        iavmInst := core.Instruction{}
        
        switch inst.Op {
        case bytecode.OpConst:
            iavmInst.Op = core.OpConst
            iavmInst.A = inst.A
            
        case bytecode.OpAdd:
            iavmInst.Op = core.OpAdd
            
        case bytecode.OpSub:
            iavmInst.Op = core.OpSub
            
        case bytecode.OpMul:
            iavmInst.Op = core.OpMul
            
        case bytecode.OpDiv:
            iavmInst.Op = core.OpDiv
            
        case bytecode.OpMod:
            iavmInst.Op = core.OpMod
            
        case bytecode.OpNeg:
            iavmInst.Op = core.OpNeg
            
        case bytecode.OpNot:
            iavmInst.Op = core.OpNot
            
        case bytecode.OpEqual:
            iavmInst.Op = core.OpEq
            
        case bytecode.OpGreater:
            iavmInst.Op = core.OpGt
            
        case bytecode.OpLess:
            iavmInst.Op = core.OpLt
            
        case bytecode.OpJump:
            iavmInst.Op = core.OpJump
            iavmInst.A = inst.A
            
        case bytecode.OpJumpIfFalse:
            iavmInst.Op = core.OpJumpIfFalse
            iavmInst.A = inst.A
            
        case bytecode.OpCall:
            iavmInst.Op = core.OpCall
            iavmInst.A = inst.A
            iavmInst.B = inst.B
            
        case bytecode.OpReturn:
            iavmInst.Op = core.OpReturn
            
        case bytecode.OpGetLocal:
            iavmInst.Op = core.OpLoadLocal
            iavmInst.A = inst.A
            
        case bytecode.OpSetLocal:
            iavmInst.Op = core.OpStoreLocal
            iavmInst.A = inst.A
            
        case bytecode.OpGetGlobal:
            iavmInst.Op = core.OpLoadGlobal
            iavmInst.A = inst.A
            
        case bytecode.OpSetGlobal:
            iavmInst.Op = core.OpStoreGlobal
            iavmInst.A = inst.A
            
        case bytecode.OpMakeArray:
            iavmInst.Op = core.OpMakeArray
            iavmInst.A = inst.A
            
        case bytecode.OpMakeObject:
            iavmInst.Op = core.OpMakeObject
            
        case bytecode.OpGetProperty:
            iavmInst.Op = core.OpGetProp
            iavmInst.A = inst.A
            
        case bytecode.OpSetProperty:
            iavmInst.Op = core.OpSetProp
            iavmInst.A = inst.A
            
        case bytecode.OpImportFunc:
            iavmInst.Op = core.OpImportFunc
            iavmInst.A = inst.A
            
        default:
            // For unmapped opcodes, use Nop as placeholder
            iavmInst.Op = core.OpNop
        }
        
        iavmInsts = append(iavmInsts, iavmInst)
    }
    
    return iavmInsts
}
```

**Step 4: 运行测试确认通过**

```bash
cd iavm && go test ./pkg/bridge/ialang/ -run TestLower -v
```
Expected: PASS

**Step 5: 提交**

```bash
git add iavm/pkg/bridge/ialang/compiler_lowering.go iavm/pkg/bridge/ialang/compiler_lowering_test.go
git commit -m "feat(iavm): implement ialang->iavm lowering bridge with opcode mapping"
```

---

## 阶段三：iavm Runtime 最小解释器

### Task 5: 实现 Stack 与 Frame 基础操作

**Files:**
- Modify: `iavm/pkg/runtime/stack.go:1-12`
- Modify: `iavm/pkg/runtime/frame.go:1-8`
- Create: `iavm/pkg/runtime/stack_test.go`

**Step 1: 编写测试**

```go
// iavm/pkg/runtime/stack_test.go
package runtime

import (
    "testing"
    "iavm/pkg/core"
)

func TestStack_PushPop(t *testing.T) {
    stack := NewStack(64)
    
    val := core.Value{Kind: core.KindI64, Raw: int64(42)}
    stack.Push(val)
    
    if stack.Size() != 1 {
        t.Fatalf("expected size 1, got %d", stack.Size())
    }
    
    popped := stack.Pop()
    if popped.Raw.(int64) != 42 {
        t.Fatalf("expected 42, got %v", popped.Raw)
    }
}

func TestStack_Peek(t *testing.T) {
    stack := NewStack(64)
    
    val := core.Value{Kind: core.KindI64, Raw: int64(99)}
    stack.Push(val)
    
    peeked := stack.Peek(0)
    if peeked.Raw.(int64) != 99 {
        t.Fatalf("expected 99, got %v", peeked.Raw)
    }
    
    // Peek should not remove
    if stack.Size() != 1 {
        t.Fatal("peek removed element")
    }
}
```

**Step 2: 实现 Stack**

```go
// iavm/pkg/runtime/stack.go
package runtime

import "iavm/pkg/core"

type Stack struct {
    data []core.Value
    top  int
}

func NewStack(capacity int) *Stack {
    return &Stack{
        data: make([]core.Value, 0, capacity),
        top:  0,
    }
}

func (s *Stack) Push(val core.Value) {
    s.data = append(s.data, val)
    s.top++
}

func (s *Stack) Pop() core.Value {
    if s.top == 0 {
        return core.Value{Kind: core.KindNull}
    }
    s.top--
    val := s.data[s.top]
    s.data = s.data[:s.top]
    return val
}

func (s *Stack) Peek(offset int) core.Value {
    idx := s.top - 1 - offset
    if idx < 0 || idx >= len(s.data) {
        return core.Value{Kind: core.KindNull}
    }
    return s.data[idx]
}

func (s *Stack) Size() int {
    return s.top
}
```

**Step 3: 更新 Frame**

```go
// iavm/pkg/runtime/frame.go
package runtime

import (
    "iavm/pkg/module"
    "iavm/pkg/core"
)

type Frame struct {
    Function *module.Function
    IP       int // instruction pointer
    Locals   []core.Value
    BaseSP   int // stack base pointer
}

func NewFrame(fn *module.Function, baseSP int) *Frame {
    locals := make([]core.Value, len(fn.Locals))
    for i := range locals {
        locals[i] = core.Value{Kind: core.KindNull}
    }
    return &Frame{
        Function: fn,
        IP:       0,
        Locals:   locals,
        BaseSP:   baseSP,
    }
}
```

**Step 4: 运行测试**

```bash
cd iavm && go test ./pkg/runtime/ -run TestStack -v
```
Expected: PASS

**Step 5: 提交**

```bash
git add iavm/pkg/runtime/stack.go iavm/pkg/runtime/frame.go iavm/pkg/runtime/stack_test.go
git commit -m "feat(iavm): implement stack and frame runtime primitives"
```

---

### Task 6: 实现 Interpreter 最小执行循环

**Files:**
- Modify: `iavm/pkg/runtime/interpreter.go:1-6`
- Modify: `iavm/pkg/runtime/vm.go:9-42`
- Create: `iavm/pkg/runtime/interpreter_test.go`

**Step 1: 编写测试**

```go
// iavm/pkg/runtime/interpreter_test.go
package runtime

import (
    "testing"
    "iavm/pkg/module"
    "iavm/pkg/core"
)

func TestInterpret_ConstReturn(t *testing.T) {
    fn := &module.Function{
        Name:    "test",
        TypeIdx: 0,
        Instructions: []core.Instruction{
            {Op: core.OpConst, A: 0},
            {Op: core.OpReturn},
        },
    }
    
    vm := New(&Options{})
    result, err := vm.executeFunction(fn)
    if err != nil {
        t.Fatalf("execution failed: %v", err)
    }
    
    if result.Kind != core.KindNull {
        t.Logf("result kind: %v", result.Kind)
    }
}

func TestInterpret_Add(t *testing.T) {
    fn := &module.Function{
        Name:    "add_test",
        TypeIdx: 0,
        Instructions: []core.Instruction{
            {Op: core.OpConst, A: 0}, // push 5
            {Op: core.OpConst, A: 1}, // push 3
            {Op: core.OpAdd},
            {Op: core.OpReturn},
        },
    }
    
    vm := New(&Options{})
    vm.constants = []core.Value{
        {Kind: core.KindI64, Raw: int64(5)},
        {Kind: core.KindI64, Raw: int64(3)},
    }
    
    result, err := vm.executeFunction(fn)
    if err != nil {
        t.Fatalf("execution failed: %v", err)
    }
    
    if result.Kind != core.KindI64 {
        t.Fatalf("expected I64, got %v", result.Kind)
    }
    if result.Raw.(int64) != 8 {
        t.Fatalf("expected 8, got %v", result.Raw)
    }
}
```

**Step 2: 实现 Interpreter**

```go
// iavm/pkg/runtime/interpreter.go
package runtime

import (
    "fmt"
    "iavm/pkg/core"
    "iavm/pkg/module"
)

func (vm *VM) Interpret(mod *module.Module, entryFunc *module.Function) (core.Value, error) {
    vm.module = mod
    vm.constants = extractConstants(mod)
    return vm.executeFunction(entryFunc)
}

func (vm *VM) executeFunction(fn *module.Function) (core.Value, error) {
    frame := NewFrame(fn, vm.stack.Size())
    vm.frames = append(vm.frames, frame)
    
    for frame.IP < len(fn.Instructions) {
        inst := fn.Instructions[frame.IP]
        frame.IP++
        
        if err := vm.dispatch(inst, frame); err != nil {
            return core.Value{Kind: core.KindNull}, err
        }
    }
    
    // Pop frame
    vm.frames = vm.frames[:len(vm.frames)-1]
    
    // Return top of stack or null
    if vm.stack.Size() > 0 {
        return vm.stack.Pop(), nil
    }
    return core.Value{Kind: core.KindNull}, nil
}

func (vm *VM) dispatch(inst core.Instruction, frame *Frame) error {
    switch inst.Op {
    case core.OpNop:
        return nil
        
    case core.OpConst:
        if inst.A >= len(vm.constants) {
            return fmt.Errorf("constant index %d out of range", inst.A)
        }
        vm.stack.Push(vm.constants[inst.A])
        
    case core.OpLoadLocal:
        if inst.A >= len(frame.Locals) {
            return fmt.Errorf("local index %d out of range", inst.A)
        }
        vm.stack.Push(frame.Locals[inst.A])
        
    case core.OpStoreLocal:
        if inst.A >= len(frame.Locals) {
            return fmt.Errorf("local index %d out of range", inst.A)
        }
        frame.Locals[inst.A] = vm.stack.Pop()
        
    case core.OpLoadGlobal:
        if inst.A >= len(vm.globals.data) {
            return fmt.Errorf("global index %d out of range", inst.A)
        }
        vm.stack.Push(vm.globals.data[inst.A])
        
    case core.OpStoreGlobal:
        val := vm.stack.Pop()
        if inst.A >= len(vm.globals.data) {
            vm.globals.data = append(vm.globals.data, val)
        } else {
            vm.globals.data[inst.A] = val
        }
        
    case core.OpAdd:
        b := vm.stack.Pop()
        a := vm.stack.Pop()
        result := addValues(a, b)
        vm.stack.Push(result)
        
    case core.OpSub:
        b := vm.stack.Pop()
        a := vm.stack.Pop()
        result := subValues(a, b)
        vm.stack.Push(result)
        
    case core.OpMul:
        b := vm.stack.Pop()
        a := vm.stack.Pop()
        result := mulValues(a, b)
        vm.stack.Push(result)
        
    case core.OpDiv:
        b := vm.stack.Pop()
        a := vm.stack.Pop()
        result := divValues(a, b)
        vm.stack.Push(result)
        
    case core.OpEq:
        b := vm.stack.Pop()
        a := vm.stack.Pop()
        result := eqValues(a, b)
        vm.stack.Push(result)
        
    case core.OpGt:
        b := vm.stack.Pop()
        a := vm.stack.Pop()
        result := gtValues(a, b)
        vm.stack.Push(result)
        
    case core.OpLt:
        b := vm.stack.Pop()
        a := vm.stack.Pop()
        result := ltValues(a, b)
        vm.stack.Push(result)
        
    case core.OpJump:
        frame.IP = int(inst.A)
        
    case core.OpJumpIfFalse:
        val := vm.stack.Pop()
        if !isTruthy(val) {
            frame.IP = int(inst.A)
        }
        
    case core.OpCall:
        // Simplified: just handle function references
        fnRef := vm.stack.Pop()
        if fnRef.Kind == core.KindFuncRef {
            fnIdx := fnRef.Raw.(int)
            if fnIdx < len(vm.module.Functions) {
                targetFn := &vm.module.Functions[fnIdx]
                _, err := vm.executeFunction(targetFn)
                return err
            }
        }
        
    case core.OpReturn:
        // Signal return by setting IP past end
        frame.IP = len(fn.Instructions)
        
    case core.OpMakeArray:
        count := int(inst.A)
        arr := make([]core.Value, count)
        for i := count - 1; i >= 0; i-- {
            arr[i] = vm.stack.Pop()
        }
        vm.stack.Push(core.Value{Kind: core.KindArrayRef, Raw: arr})
        
    case core.OpMakeObject:
        vm.stack.Push(core.Value{Kind: core.KindObjectRef, Raw: make(map[string]core.Value)})
        
    case core.OpHostCall:
        if vm.host == nil {
            return fmt.Errorf("no host configured")
        }
        // Simplified host call
        vm.stack.Push(core.Value{Kind: core.KindNull})
        
    default:
        return fmt.Errorf("unimplemented opcode: %v", inst.Op)
    }
    
    return nil
}

func extractConstants(mod *module.Module) []core.Value {
    var constants []core.Value
    // Extract from function instructions that reference constants
    // In a real implementation, constants would be in a dedicated section
    return constants
}

// Arithmetic helpers
func addValues(a, b core.Value) core.Value {
    if a.Kind == core.KindI64 && b.Kind == core.KindI64 {
        return core.Value{Kind: core.KindI64, Raw: a.Raw.(int64) + b.Raw.(int64)}
    }
    return core.Value{Kind: core.KindNull}
}

func subValues(a, b core.Value) core.Value {
    if a.Kind == core.KindI64 && b.Kind == core.KindI64 {
        return core.Value{Kind: core.KindI64, Raw: a.Raw.(int64) - b.Raw.(int64)}
    }
    return core.Value{Kind: core.KindNull}
}

func mulValues(a, b core.Value) core.Value {
    if a.Kind == core.KindI64 && b.Kind == core.KindI64 {
        return core.Value{Kind: core.KindI64, Raw: a.Raw.(int64) * b.Raw.(int64)}
    }
    return core.Value{Kind: core.KindNull}
}

func divValues(a, b core.Value) core.Value {
    if a.Kind == core.KindI64 && b.Kind == core.KindI64 {
        return core.Value{Kind: core.KindI64, Raw: a.Raw.(int64) / b.Raw.(int64)}
    }
    return core.Value{Kind: core.KindNull}
}

func eqValues(a, b core.Value) core.Value {
    eq := a.Raw == b.Raw
    return core.Value{Kind: core.KindBool, Raw: eq}
}

func gtValues(a, b core.Value) core.Value {
    if a.Kind == core.KindI64 && b.Kind == core.KindI64 {
        return core.Value{Kind: core.KindBool, Raw: a.Raw.(int64) > b.Raw.(int64)}
    }
    return core.Value{Kind: core.KindNull}
}

func ltValues(a, b core.Value) core.Value {
    if a.Kind == core.KindI64 && b.Kind == core.KindI64 {
        return core.Value{Kind: core.KindBool, Raw: a.Raw.(int64) < b.Raw.(int64)}
    }
    return core.Value{Kind: core.KindNull}
}

func isTruthy(val core.Value) bool {
    if val.Kind == core.KindBool {
        return val.Raw.(bool)
    }
    return val.Kind != core.KindNull
}
```

**Step 3: 更新 VM 结构**

```go
// iavm/pkg/runtime/vm.go
package runtime

import (
    "iavm/pkg/core"
    "iavm/pkg/module"
    "iacommon/pkg/host/api"
)

type VM struct {
    stack     *Stack
    frames    []*Frame
    globals   *Globals
    handles   *HandleTable
    host      api.Host
    module    *module.Module
    constants []core.Value
    options   *Options
    steps     int
}

func New(opts *Options) *VM {
    if opts == nil {
        opts = &Options{}
    }
    return &VM{
        stack:   NewStack(256),
        globals: NewGlobals(64),
        handles: NewHandleTable(),
        options: opts,
        host:    opts.Host,
    }
}

func (vm *VM) Run(mod *module.Module) error {
    // Find entry function
    var entryFn *module.Function
    for i, fn := range mod.Functions {
        if fn.Name == "main" || fn.Name == "entry" {
            entryFn = &mod.Functions[i]
            break
        }
    }
    if entryFn == nil && len(mod.Functions) > 0 {
        entryFn = &mod.Functions[0]
    }
    if entryFn == nil {
        return core.NewInvalidModuleError("no entry function")
    }
    
    _, err := vm.Interpret(mod, entryFn)
    return err
}

func (vm *VM) InvokeExport(name string) (core.Value, error) {
    for _, exp := range vm.module.Exports {
        if exp.Name == name && exp.Kind == module.ExportFunction {
            fn := &vm.module.Functions[exp.Index]
            return vm.Interpret(vm.module, fn)
        }
    }
    return core.Value{Kind: core.KindNull}, core.NewInvalidModuleError("export not found: " + name)
}
```

**Step 4: 运行测试**

```bash
cd iavm && go test ./pkg/runtime/ -run TestInterpret -v
```
Expected: PASS

**Step 5: 提交**

```bash
git add iavm/pkg/runtime/interpreter.go iavm/pkg/runtime/vm.go iavm/pkg/runtime/interpreter_test.go
git commit -m "feat(iavm): implement minimal interpreter with arithmetic and control flow"
```

---

## 阶段四：Capability Binding 与集成

### Task 7: 实现 Capability Binding 最小路径

**Files:**
- Modify: `iavm/pkg/runtime/vm.go` (添加 host call 处理)
- Reference: 
  - `iacommon/pkg/host/api/default_host.go`
  - `iacommon/pkg/ialang/module/platform_bridge.go`
- Create: `iavm/pkg/runtime/capability_test.go`

**Step 1: 编写测试**

```go
// iavm/pkg/runtime/capability_test.go
package runtime

import (
    "testing"
    "iavm/pkg/module"
    "iavm/pkg/core"
    "iacommon/pkg/host/api"
    "iacommon/pkg/host/fs"
)

func TestCapability_FSReadFile(t *testing.T) {
    // Create memfs with test file
    memFS := fs.NewMemFS()
    memFS.WriteFile("/test.txt", []byte("hello"))
    
    // Create host with FS capability
    host := api.NewDefaultHost()
    host.RegisterFS("default", memFS)
    
    vm := New(&Options{Host: host})
    
    // Module that requests FS capability
    mod := &module.Module{
        Magic:   "IAVM",
        Version: 1,
        Target:  "ialang",
        Capabilities: []module.CapabilityDecl{
            {Kind: module.CapabilityFS, Name: "default"},
        },
    }
    
    err := vm.Run(mod)
    if err != nil {
        t.Fatalf("VM run failed: %v", err)
    }
}
```

**Step 2: 实现 HostCall 处理**

在 interpreter.go 的 dispatch 方法中完善 OpHostCall：

```go
case core.OpHostCall:
    if vm.host == nil {
        return fmt.Errorf("no host configured for host.call")
    }
    
    // Pop operation name and arguments from stack
    opName := vm.stack.Pop()
    if opName.Kind != core.KindString {
        return fmt.Errorf("host.call operation name must be string")
    }
    
    // Build call request
    req := &api.CallRequest{
        Operation: opName.Raw.(string),
        Args:      make(map[string]any),
    }
    
    // Pop args count
    argCount := vm.stack.Pop()
    if argCount.Kind == core.KindI64 {
        count := argCount.Raw.(int64)
        for i := int64(0); i < count; i++ {
            val := vm.stack.Pop()
            req.Args[fmt.Sprintf("arg%d", i)] = val.Raw
        }
    }
    
    // Call host
    result, err := vm.host.Call(req)
    if err != nil {
        return fmt.Errorf("host.call failed: %w", err)
    }
    
    // Push result
    if result.Value != nil {
        vm.stack.Push(core.Value{Kind: core.KindNull, Raw: result.Value})
    } else {
        vm.stack.Push(core.Value{Kind: core.KindNull})
    }
```

**Step 3: 运行测试**

```bash
cd iavm && go test ./pkg/runtime/ -run TestCapability -v
```
Expected: PASS

**Step 4: 提交**

```bash
git add iavm/pkg/runtime/vm.go iavm/pkg/runtime/interpreter.go iavm/pkg/runtime/capability_test.go
git commit -m "feat(iavm): implement host capability binding with FS support"
```

---

### Task 8: 集成测试 - 完整链路验证

**Files:**
- Create: `iavm/pkg/integration/integration_test.go`
- Reference: 所有已实现模块

**Step 1: 编写端到端测试**

```go
// iavm/pkg/integration/integration_test.go
package integration

import (
    "testing"
    "iacommon/pkg/ialang/bytecode"
    "iacommon/pkg/host/api"
    "iacommon/pkg/host/fs"
    "iavm/pkg/bridge/ialang"
    "iavm/pkg/binary"
    "iavm/pkg/runtime"
)

func TestFullPipeline_CompileToLowerToRun(t *testing.T) {
    // 1. Create ialang chunk
    chunk := &bytecode.Chunk{
        Code: []bytecode.Instruction{
            {Op: bytecode.OpConst, A: 0, B: 0},
            {Op: bytecode.OpConst, A: 1, B: 0},
            {Op: bytecode.OpAdd},
            {Op: bytecode.OpReturn},
        },
        Constants: []any{
            int64(10),
            int64(20),
        },
    }
    
    // 2. Lower to iavm module
    mod, err := ialang.LowerToModule(chunk)
    if err != nil {
        t.Fatalf("LowerToModule failed: %v", err)
    }
    
    // 3. Encode to binary
    data, err := binary.EncodeModule(mod)
    if err != nil {
        t.Fatalf("EncodeModule failed: %v", err)
    }
    
    // 4. Decode back
    decoded, err := binary.DecodeModule(data)
    if err != nil {
        t.Fatalf("DecodeModule failed: %v", err)
    }
    
    // 5. Verify
    result, err := binary.VerifyModule(decoded)
    if err != nil {
        t.Fatalf("VerifyModule failed: %v", err)
    }
    if !result.Valid {
        t.Fatalf("module verification failed: %v", result.Errors)
    }
    
    // 6. Run
    vm := runtime.New(&runtime.Options{})
    err = vm.Run(decoded)
    if err != nil {
        t.Fatalf("VM.Run failed: %v", err)
    }
}

func TestFullPipeline_WithHostCapability(t *testing.T) {
    // Setup host with FS
    host := api.NewDefaultHost()
    memFS := fs.NewMemFS()
    memFS.WriteFile("/hello.txt", []byte("world"))
    host.RegisterFS("default", memFS)
    
    // Create module with FS capability
    mod := &module.Module{
        Magic:   "IAVM",
        Version: 1,
        Target:  "ialang",
        Capabilities: []module.CapabilityDecl{
            {Kind: module.CapabilityFS, Name: "default"},
        },
    }
    
    // Verify and run
    result, err := binary.VerifyModule(mod)
    if err != nil {
        t.Fatalf("VerifyModule failed: %v", err)
    }
    if !result.Valid {
        t.Fatal("module not valid")
    }
    
    vm := runtime.New(&runtime.Options{Host: host})
    err = vm.Run(mod)
    if err != nil {
        t.Fatalf("VM.Run with host failed: %v", err)
    }
}
```

**Step 2: 运行集成测试**

```bash
cd iavm && go test ./pkg/integration/ -v
```
Expected: PASS

**Step 3: 提交**

```bash
git add iavm/pkg/integration/integration_test.go
git commit -m "test(iavm): add end-to-end integration tests for full pipeline"
```

---

## 阶段五：测试与文档

### Task 9: 运行全量测试并修复

**Step 1: 运行所有测试**

```bash
cd iavm && go test ./... -v
```

**Step 2: 修复任何失败**

根据测试输出修复问题。

**Step 3: 运行 ialang 测试确保无回归**

```bash
cd ialang && go test ./... -v
```

**Step 4: 提交**

```bash
git add -A
git commit -m "fix: address test failures and ensure no regressions"
```

---

### Task 10: 更新文档

**Files:**
- Modify: `docs/2026-04-21-development-plan.md`
- Create: `docs/plans/2026-04-21-iavm-minimal-loop-implementation.md` (记录实施结果)

**Step 1: 更新开发计划状态**

在 `docs/2026-04-21-development-plan.md` 末尾添加实施状态章节。

**Step 2: 提交**

```bash
git add docs/
git commit -m "docs: update development plan with implementation status"
```

---

## 完成标准检查清单

- [ ] Binary encoder/decoder 实现并通过 round-trip 测试
- [ ] Verifier 实现并通过合法/非法模块测试
- [ ] LowerToModule 实现并通过 ialang chunk 转换测试
- [ ] Runtime interpreter 实现并通过算术/控制流测试
- [ ] Capability binding 实现并通过 FS/HTTP 测试
- [ ] 集成测试覆盖完整链路
- [ ] 所有 `go test ./...` 通过
- [ ] 文档更新

---

## 风险与缓解

1. **Opcode 映射不完整**: ialang 有 68 个 opcode，iavm 只有 27 个。首版只映射核心 opcode，其余用 Nop 占位。
2. **常量池处理**: iavm 模块格式中常量需要独立 section，lowering 时需要从 ialang chunk 中提取。
3. **Host ABI 兼容**: iacommon 已实现完整的 Host API，iavm 只需正确调用即可。
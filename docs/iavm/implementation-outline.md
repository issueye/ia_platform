# IAVM 可实施目录结构与接口草案

## 1. 文档目的

本文是 `iavm/wasm-like-vm-platform-design.md` 的实施细化版，目标是把总体方案进一步落到：

1. 可直接创建目录的工程结构
2. 可直接落代码的 Go 接口草案
3. 第一阶段实现时的模块边界
4. `ialang -> iavm` 的桥接接入点

本版本优先考虑：

- 尽量复用现有 `ialang` 的 lexer / parser / AST / compiler / runtime 资产
- 首版只把 **文件系统** 与 **网络访问** 从当前 runtime builtin 中抽离成平台 Host ABI
- 首版先做 **解释执行 + verifier + capability binding**，不引入 JIT

---

## 2. 推荐目录结构

建议将 `iavm` 组织为以下目录：

```text
iavm/
  docs/
    wasm-like-vm-platform-design.md
    implementation-outline.md
    module-format.md
    verifier.md
    abi-fs.md
    abi-network.md
  pkg/
    core/
      opcode.go
      instruction.go
      value.go
      types.go
      errors.go
    module/
      module.go
      section.go
      capability.go
      import_export.go
      manifest.go
    binary/
      encoder.go
      decoder.go
      verifier.go
      verify_result.go
    runtime/
      vm.go
      frame.go
      stack.go
      globals.go
      interpreter.go
      handles.go
      resources.go
      options.go
    host/
      api/
        host.go
        capability.go
        context.go
      fs/
        provider.go
        localfs.go
        memfs.go
        pathmap.go
      network/
        provider.go
        http_provider.go
        socket_provider.go
        policy.go
    bridge/
      ialang/
        compiler_lowering.go
        builtin_fs.go
        builtin_http.go
        module_loader.go
    sdk/
      fs/
        fs.go
      http/
        http.go
```

---

## 3. 各目录职责说明

## 3.1 `pkg/core`

放平台最稳定的核心定义：

- opcode
- instruction
- value kind
- 通用错误码
- 函数签名与类型描述

要求：

- 不依赖宿主实现
- 不依赖 `ialang`
- 尽量只包含平台最小内核模型

### 建议文件

#### `pkg/core/opcode.go`

定义平台 opcode：

```go
package core

type OpCode uint16

const (
	OpNop OpCode = iota
	OpConst
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpNeg
	OpNot
	OpEq
	OpNe
	OpLt
	OpGt
	OpLe
	OpGe
	OpJump
	OpJumpIfFalse
	OpCall
	OpReturn
	OpLoadLocal
	OpStoreLocal
	OpLoadGlobal
	OpStoreGlobal
	OpMakeArray
	OpMakeObject
	OpGetProp
	OpSetProp
	OpImportFunc
	OpImportCap
	OpHostCall
	OpHostPoll
)
```

#### `pkg/core/instruction.go`

```go
package core

type Instruction struct {
	Op OpCode
	A  uint32
	B  uint32
	C  uint32
}
```

说明：

- 比当前 `ialang/pkg/lang/bytecode/bytecode.go:76` 的 `Instruction{Op, A, B}` 稍微宽一点
- 目的是减少后续频繁改格式
- 首版解释器仍可只使用其中一部分字段

#### `pkg/core/value.go`

```go
package core

type ValueKind uint8

const (
	ValueNull ValueKind = iota
	ValueBool
	ValueI64
	ValueF64
	ValueString
	ValueBytes
	ValueArrayRef
	ValueObjectRef
	ValueFuncRef
	ValueHostHandle
)

type Value struct {
	Kind ValueKind
	Raw  any
}
```

---

## 3.2 `pkg/module`

定义 IAVM 模块内存表示，供：

- lowering 阶段输出
- verifier 输入
- binary encoder/decoder 使用
- runtime loader 使用

### 建议文件

#### `pkg/module/module.go`

```go
package module

import "iavm/pkg/core"

type Module struct {
	Magic        string
	Version      uint16
	Target       string
	ABIVersion   uint16
	FeatureFlags uint64

	Types        []FuncType
	Imports      []Import
	Functions    []Function
	Globals      []Global
	Exports      []Export
	DataSegments []DataSegment
	Capabilities []CapabilityDecl
	Custom       map[string][]byte
}

type Function struct {
	Name         string
	TypeIndex    uint32
	Locals       []core.ValueKind
	Code         []core.Instruction
	Constants    []any
	MaxStack     uint32
	IsEntryPoint bool
}
```

#### `pkg/module/import_export.go`

```go
package module

type ImportKind uint8

const (
	ImportFunction ImportKind = iota
	ImportGlobal
	ImportCapability
)

type Import struct {
	Module string
	Name   string
	Kind   ImportKind
	Type   uint32
}

type ExportKind uint8

const (
	ExportFunction ExportKind = iota
	ExportGlobal
)

type Export struct {
	Name  string
	Kind  ExportKind
	Index uint32
}
```

#### `pkg/module/capability.go`

```go
package module

type CapabilityKind string

const (
	CapabilityFS      CapabilityKind = "fs"
	CapabilityNetwork CapabilityKind = "network"
)

type CapabilityDecl struct {
	Kind     CapabilityKind
	Required bool
	Config   map[string]any
}
```

说明：

- 首版 `Config` 直接使用 `map[string]any`，降低演进成本
- 二版再拆成强类型结构

---

## 3.3 `pkg/binary`

负责编解码与验证。

### 建议文件

#### `pkg/binary/encoder.go`

```go
package binary

import "iavm/pkg/module"

func EncodeModule(m *module.Module) ([]byte, error) {
	return nil, nil
}
```

#### `pkg/binary/decoder.go`

```go
package binary

import "iavm/pkg/module"

func DecodeModule(raw []byte) (*module.Module, error) {
	return nil, nil
}
```

#### `pkg/binary/verifier.go`

```go
package binary

import "iavm/pkg/module"

type VerifyOptions struct {
	RequireEntry bool
	AllowCustom  bool
}

type VerifyResult struct {
	Warnings []string
}

func VerifyModule(m *module.Module, opts VerifyOptions) (*VerifyResult, error) {
	return nil, nil
}
```

### 验证器首版职责

1. 模块头检查
2. 函数类型索引检查
3. 导入导出索引检查
4. 跳转目标检查
5. 常量索引检查
6. capability 声明结构检查
7. `host.call` 是否只能指向声明过的 capability

---

## 3.4 `pkg/runtime`

这里是执行器与资源管理层。

### 建议文件

#### `pkg/runtime/options.go`

```go
package runtime

import (
	"time"
	"iavm/pkg/host/api"
)

type Options struct {
	MaxSteps      int64
	MaxMemory     int64
	MaxDuration   time.Duration
	Host          api.Host
	EnableTracing bool
}
```

#### `pkg/runtime/vm.go`

```go
package runtime

import (
	"iavm/pkg/module"
)

type VM struct {
	mod       *module.Module
	options   Options
	stack     []any
	globals   map[string]any
	functions []CompiledFunction
	handles   *HandleTable
	startedAt int64
	stepCount int64
}

func New(mod *module.Module, opts Options) (*VM, error) {
	return nil, nil
}

func (vm *VM) Run() error {
	return nil
}

func (vm *VM) InvokeExport(name string, args ...any) (any, error) {
	return nil, nil
}
```

#### `pkg/runtime/frame.go`

```go
package runtime

type Frame struct {
	FunctionIndex uint32
	IP            uint32
	Locals        []any
	BasePointer   uint32
}
```

#### `pkg/runtime/handles.go`

```go
package runtime

type HandleKind uint8

const (
	HandleFile HandleKind = iota
	HandleSocket
	HandleListener
	HandleHTTPStream
)

type HandleEntry struct {
	ID    uint64
	Kind  HandleKind
	Value any
}

type HandleTable struct {
	nextID  uint64
	entries map[uint64]HandleEntry
}

func NewHandleTable() *HandleTable {
	return &HandleTable{entries: map[uint64]HandleEntry{}}
}
```

#### `pkg/runtime/resources.go`

职责：

- 注册/关闭句柄
- VM 退出时统一清理资源
- 限制最大打开文件数/连接数

---

## 3.5 `pkg/host/api`

这是最关键的宿主接口层，平台边界应该定义在这里。

### 建议文件

#### `pkg/host/api/host.go`

```go
package api

import "context"

type Host interface {
	AcquireCapability(ctx context.Context, req AcquireRequest) (CapabilityInstance, error)
	ReleaseCapability(ctx context.Context, capID string) error
	Call(ctx context.Context, req CallRequest) (CallResult, error)
	Poll(ctx context.Context, handleID uint64) (PollResult, error)
}
```

#### `pkg/host/api/capability.go`

```go
package api

type CapabilityKind string

const (
	CapabilityFS      CapabilityKind = "fs"
	CapabilityNetwork CapabilityKind = "network"
)

type AcquireRequest struct {
	Kind   CapabilityKind
	Config map[string]any
}

type CapabilityInstance struct {
	ID      string
	Kind    CapabilityKind
	Rights  []string
	Meta    map[string]any
}
```

#### `pkg/host/api/context.go`

```go
package api

type CallRequest struct {
	CapabilityID string
	Operation    string
	Args         map[string]any
}

type CallResult struct {
	Value map[string]any
}

type PollResult struct {
	Done  bool
	Value map[string]any
	Error string
}
```

说明：

- 首版建议使用 `map[string]any` 作为跨 ABI 参数承载
- 这样实现成本低，适合先稳定边界
- 等 ABI 稳定后，再把高频操作改为强类型结构体

---

## 3.6 `pkg/host/fs`

这里落文件系统能力的 provider。

### 建议文件

#### `pkg/host/fs/provider.go`

```go
package fs

import "context"

type Provider interface {
	Open(ctx context.Context, path string, opts OpenOptions) (FileHandle, error)
	ReadFile(ctx context.Context, path string) ([]byte, error)
	WriteFile(ctx context.Context, path string, data []byte, opts WriteOptions) error
	AppendFile(ctx context.Context, path string, data []byte) error
	ReadDir(ctx context.Context, path string) ([]DirEntry, error)
	Stat(ctx context.Context, path string) (FileInfo, error)
	Mkdir(ctx context.Context, path string, opts MkdirOptions) error
	Remove(ctx context.Context, path string, opts RemoveOptions) error
	Rename(ctx context.Context, oldPath, newPath string) error
}

type OpenOptions struct {
	Read   bool
	Write  bool
	Create bool
	Trunc  bool
	Append bool
}

type WriteOptions struct {
	Create bool
	Trunc  bool
}

type MkdirOptions struct {
	Recursive bool
}

type RemoveOptions struct {
	Recursive bool
}

type FileHandle interface {
	Read(ctx context.Context, p []byte) (int, error)
	Write(ctx context.Context, p []byte) (int, error)
	Seek(ctx context.Context, offset int64, whence int) (int64, error)
	Close(ctx context.Context) error
}

type FileInfo struct {
	Name    string
	Size    int64
	Mode    string
	IsDir   bool
	ModUnix int64
}

type DirEntry struct {
	Name  string
	IsDir bool
}
```

#### `pkg/host/fs/pathmap.go`

```go
package fs

type Preopen struct {
	VirtualPath string
	RealPath    string
	ReadOnly    bool
}

type PathMapper interface {
	Resolve(virtualPath string) (realPath string, matchedPreopen Preopen, err error)
}
```

#### `pkg/host/fs/localfs.go`

职责：

- 通过 `PathMapper` 做虚拟路径到真实路径映射
- 做路径清洗与越界检查
- 最终调用宿主 `os` 包

#### `pkg/host/fs/memfs.go`

职责：

- 提供纯内存文件系统
- 便于测试 verifier / runtime / sdk

---

## 3.7 `pkg/host/network`

定义网络访问层。

### 建议文件

#### `pkg/host/network/provider.go`

```go
package network

import "context"

type Provider interface {
	HTTPFetch(ctx context.Context, req HTTPRequest) (*HTTPResponse, error)
	Dial(ctx context.Context, endpoint Endpoint, opts DialOptions) (SocketHandle, error)
	Listen(ctx context.Context, endpoint Endpoint, opts ListenOptions) (ListenerHandle, error)
}

type HTTPRequest struct {
	Method    string
	URL       string
	Headers   map[string]string
	Body      []byte
	TimeoutMS int64
}

type HTTPResponse struct {
	Status  int
	Headers map[string]string
	Body    []byte
}

type Endpoint struct {
	Network string
	Host    string
	Port    int
}

type DialOptions struct {
	TimeoutMS int64
}

type ListenOptions struct {
	Backlog int
}

type SocketHandle interface {
	Send(ctx context.Context, data []byte) (int, error)
	Recv(ctx context.Context, size int) ([]byte, error)
	Close(ctx context.Context) error
}

type ListenerHandle interface {
	Accept(ctx context.Context) (SocketHandle, error)
	Close(ctx context.Context) error
}
```

#### `pkg/host/network/policy.go`

```go
package network

type Policy struct {
	Rights             []string
	AllowHosts         []string
	AllowPorts         []int
	AllowSchemes       []string
	AllowCIDRs         []string
	MaxConnections     int
	MaxInflightRequest int
	MaxBytesPerRequest int64
}
```

#### `pkg/host/network/http_provider.go`

职责：

- 基于 Go `net/http` 实现 `HTTPFetch`
- 在进入真实请求前检查 `Policy`

#### `pkg/host/network/socket_provider.go`

职责：

- 基于 Go `net` 实现 `Dial/Listen`
- 在建立连接前检查 `Policy`

---

## 3.8 `pkg/bridge/ialang`

这是连接现有 `ialang` 与未来 `iavm` 的桥。

### 建议文件

#### `pkg/bridge/ialang/compiler_lowering.go`

职责：

- 输入：`ialang` AST 或现有 bytecode chunk
- 输出：`iavm/pkg/module.Module`

建议第一阶段优先走：

```text
ialang AST -> 复用现有编译逻辑 -> 生成过渡 Module
```

而不是立即重写全量编译器。

#### `pkg/bridge/ialang/builtin_fs.go`

职责：

- 暴露给 ialang 的 `@platform/fs` 模块包装
- 对外 API 类似当前 `fs`，对内改为 `host.call(capID, op, args)`

#### `pkg/bridge/ialang/builtin_http.go`

职责：

- 暴露给 ialang 的 `@platform/http` 模块包装
- 对接 network capability

#### `pkg/bridge/ialang/module_loader.go`

职责：

- 把 `ialang` 当前模块解析逻辑适配到 `iavm` 模块加载器

---

## 4. 文件系统 ABI 草案

建议 FS 操作在 Host ABI 层先统一为：

```go
Operation string
```

操作名建议如下：

- `fs.read_file`
- `fs.write_file`
- `fs.append_file`
- `fs.read_dir`
- `fs.stat`
- `fs.mkdir`
- `fs.remove`
- `fs.rename`
- `fs.open`
- `fs.read`
- `fs.write`
- `fs.seek`
- `fs.close`

### 4.1 参数示例

#### `fs.read_file`

```json
{
  "path": "/workspace/a.txt"
}
```

返回：

```json
{
  "data": "base64-or-utf8",
  "encoding": "utf-8"
}
```

#### `fs.write_file`

```json
{
  "path": "/workspace/out.txt",
  "data": "hello",
  "encoding": "utf-8",
  "create": true,
  "trunc": true
}
```

#### `fs.read_dir`

```json
{
  "path": "/workspace"
}
```

返回：

```json
{
  "entries": [
    {"name": "a.txt", "is_dir": false},
    {"name": "subdir", "is_dir": true}
  ]
}
```

---

## 5. 网络 ABI 草案

建议网络操作名如下：

- `network.http_fetch`
- `network.dial`
- `network.listen`
- `network.accept`
- `network.send`
- `network.recv`
- `network.close`

### 5.1 参数示例

#### `network.http_fetch`

```json
{
  "method": "GET",
  "url": "https://api.example.com/data",
  "headers": {
    "accept": "application/json"
  },
  "timeout_ms": 5000
}
```

返回：

```json
{
  "status": 200,
  "headers": {
    "content-type": "application/json"
  },
  "body": "..."
}
```

#### `network.dial`

```json
{
  "network": "tcp",
  "host": "example.com",
  "port": 443,
  "timeout_ms": 3000
}
```

返回：

```json
{
  "handle_id": 1001
}
```

---

## 6. 统一 Host 实现建议

建议提供一个默认宿主实现：

#### `pkg/host/api/default_host.go`

```go
package api

type DefaultHost struct {
	FSProvider      any
	NetworkProvider any
}
```

更合适的最终版接口建议是：

```go
package api

import (
	"context"
	hostfs "iavm/pkg/host/fs"
	hostnet "iavm/pkg/host/network"
)

type DefaultHost struct {
	FS      hostfs.Provider
	Network hostnet.Provider
}

func (h *DefaultHost) AcquireCapability(ctx context.Context, req AcquireRequest) (CapabilityInstance, error) {
	return CapabilityInstance{}, nil
}

func (h *DefaultHost) ReleaseCapability(ctx context.Context, capID string) error {
	return nil
}

func (h *DefaultHost) Call(ctx context.Context, req CallRequest) (CallResult, error) {
	return CallResult{}, nil
}

func (h *DefaultHost) Poll(ctx context.Context, handleID uint64) (PollResult, error) {
	return PollResult{}, nil
}
```

说明：

- `AcquireCapability` 负责把 module capability declaration 与宿主配置绑定
- `Call` 负责 operation dispatch
- `Poll` 负责异步句柄/流式句柄

---

## 7. `ialang` 集成策略

## 7.1 现有关键接入点

建议从这些位置开始接：

- 编译入口：`ialang/cmd/ialang/run.go:45`
- 运行入口：`ialang/cmd/ialang/run.go:77`
- 当前字节码结构：`ialang/pkg/lang/bytecode/bytecode.go:76`
- 当前 VM：`ialang/pkg/lang/runtime/vm/vm.go:88`
- 当前 FS builtin：`ialang/pkg/lang/runtime/builtin/fs.go:8`
- 当前 HTTP builtin：`ialang/pkg/lang/runtime/builtin/http.go:96`
- 当前 builtin 注册：`ialang/pkg/lang/runtime/builtin/registry.go:8`

## 7.2 首阶段最小接法

### 路线 A：保留 `ialang` VM，先替换 builtin 宿主调用

做法：

1. 保留现有 `ialang` 字节码与 VM
2. 增加 `iavm/host` 能力层
3. 将 `builtin/fs.go`、`builtin/http.go` 改成调 `iavm/host`

优点：

- 改动最小
- 最快得到 capability-based FS/Network

缺点：

- 还没有完整 `.iavm` 模块格式

### 路线 B：生成 IAVM Module，但运行期先兼容现有 VM 执行模型

做法：

1. 引入 `iavm/pkg/module`
2. 从 `ialang compiler` 输出过渡 Module
3. 编写 IAVM interpreter
4. 内建模块通过 bridge 对接 host API

优点：

- 更接近最终目标

缺点：

- 工作量更大

### 建议

先做 **A + 局部铺路 B**：

- 本轮先把能力边界抽出来
- 模块格式和 verifier 同步建壳
- 真正切运行时放到下一阶段

---

## 8. 第一阶段实现清单

建议按以下顺序推进：

### P1. 定义平台核心结构

1. `pkg/core/opcode.go`
2. `pkg/core/instruction.go`
3. `pkg/module/module.go`
4. `pkg/module/capability.go`

### P2. 定义 Host API

1. `pkg/host/api/host.go`
2. `pkg/host/fs/provider.go`
3. `pkg/host/network/provider.go`

### P3. 提供宿主实现

1. `pkg/host/fs/pathmap.go`
2. `pkg/host/fs/localfs.go`
3. `pkg/host/fs/memfs.go`
4. `pkg/host/network/http_provider.go`
5. `pkg/host/network/socket_provider.go`

### P4. 接入 ialang builtin

1. 把当前 `fs` builtin 改为 capability-backed
2. 把当前 `http` builtin 改为 capability-backed
3. 在模块注册处保留兼容命名，但内部不再直连 `os` / `net/http`

### P5. 建 verifier 壳子

1. `pkg/binary/verifier.go`
2. 加入 capability 声明校验
3. 加入指令与索引基础校验

---

## 9. 风险与取舍

## 9.1 最大风险

当前 `ialang` 是动态语言运行时，若过早追求强类型 verifier，会拖慢整体进度。

### 建议取舍

首版 verifier 只做：

- 结构完整性
- 索引合法性
- capability 合法性
- 控制流合法性

先不要试图做完整静态类型证明。

## 9.2 FS / Network API 过早定死

风险在于未来 `ialang` 标准库体验可能变化。

### 建议取舍

- 先稳定 Host ABI
- 语言层 SDK 作为适配层，允许变动

即：

```text
用户代码依赖 sdk API
sdk API 调用 host ABI
host ABI 尽量稳定
```

---

## 10. 结论

对于当前仓库，最可实施的落地方式是：

1. 在 `iavm/pkg/host` 先建立 **文件系统/网络 provider + capability API**
2. 在 `iavm/pkg/module` 与 `iavm/pkg/binary` 建立 **模块与 verifier 壳结构**
3. 在 `iavm/pkg/bridge/ialang` 中建立 **与现有 ialang builtin/runtime 的桥接层**
4. 第一阶段优先让 `fs/http` 脱离对 `os` / `net/http` 的直接依赖

这样能以最小风险，把现有 `ialang` 从“语言运行时”推向“平台运行时”的第一步。

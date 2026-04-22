# IAVM / ialang 后续开发计划（2026-04-22）

## 1. 当前项目状态

### 1.1 仓库结构

当前项目是一个 Go workspace，包含三个核心模块：

```text
ia_platform/
├── iacommon/   # 公共 Host ABI、ialang 公共字节码/包格式/运行时类型
├── ialang/     # 语言前端、编译器、运行时、CLI、内建库、工具链
├── iavm/       # 类 WASM 的平台 VM、模块格式、二进制编码、运行时、桥接层
└── docs/       # 设计文档与阶段计划
```

Workspace 配置：`go.work` 引入 `./iacommon`、`./ialang`、`./iavm`。

### 1.2 已落地能力

| 方向 | 当前状态 | 关键文件 |
|---|---|---|
| ialang 语言链路 | 已具备源码解析、编译、VM 执行、模块加载、内建库和 CLI | `ialang/cmd/ialang/run.go`, `ialang/pkg/lang/*` |
| 公共层抽离 | Host ABI、FS/Network provider、ialang bytecode/packagefile 已放入 `iacommon` | `iacommon/pkg/host/*`, `iacommon/pkg/ialang/*` |
| IAVM 模块结构 | 已定义 Module/Function/Import/Export/Capability/常量池等核心结构 | `iavm/pkg/module/module.go`, `iavm/pkg/module/import_export.go`, `iavm/pkg/module/capability.go` |
| 二进制格式 | Encoder/Decoder 已支持模块序列化与反序列化 | `iavm/pkg/binary/encoder.go`, `iavm/pkg/binary/decoder.go` |
| Verifier | 已包含 header/type/function/export/import/global/capability/entry/control-flow/constant-ref 等校验 | `iavm/pkg/binary/verifier.go` |
| ialang → IAVM lowering | 已从占位推进到可转换 `*bytecode.Chunk`，含函数、全局、常量池和部分 opcode 映射 | `iavm/pkg/bridge/ialang/compiler_lowering.go` |
| IAVM Runtime | 已具备解释器、栈/帧、基础算术/比较/控制流/函数调用/对象数组/host call/builtin | `iavm/pkg/runtime/interpreter.go`, `iavm/pkg/runtime/vm.go`, `iavm/pkg/runtime/builtins.go` |
| Capability | 已有 `OpImportCap` / `OpHostCall` 与 `DefaultHost` 的最小绑定 | `iavm/pkg/runtime/interpreter.go`, `iacommon/pkg/host/api/default_host.go` |
| CLI 集成 | 已提供 `build-iavm` 与 `run-iavm` 路径 | `ialang/cmd/ialang/build_iavm.go`, `ialang/cmd/ialang/run_iavm.go` |
| 测试基线 | `iacommon`、`ialang`、`iavm` 子模块测试通过 | 见本文第 1.4 节 |

### 1.3 与昨日计划相比的变化

昨日文档中的“平台桥接最小闭环”已基本落地：

- `LowerToModule` 不再是空实现，当前输入明确为 `*bytecode.Chunk`。
- IAVM verifier 已从结构校验扩展到控制流与常量引用检查。
- IAVM runtime 已能运行模块入口函数，并能执行基础 opcode、builtin 与 host capability 调用。
- CLI 已具备 `build-iavm` / `run-iavm` 命令路径。
- 模块级常量池已加入 `module.Module.Constants`，lowering 后可去重并重写 `OpConst` 引用。

这意味着下一阶段重点应从“打通最小链路”切换到“语义完整性、模块化能力、稳定性和开发体验”。

### 1.4 当前测试结果

在 workspace 根目录直接执行 `go test ./...` 会失败：

```text
pattern ./...: directory prefix . does not contain modules listed in go.work or their selected dependencies
```

当前可用测试方式是按子模块执行：

| 命令 | 结果 |
|---|---|
| `go test ./iavm/...` | PASS |
| `go test ./iacommon/...` | PASS |
| `go test ./ialang/...` | PASS |

建议后续补一个统一脚本或 CI 任务，避免开发者在 workspace 根目录误用 `go test ./...`。

---

## 2. 当前主要缺口

### 2.1 Opcode 与语言语义覆盖仍不完整

`iavm/pkg/core/opcode.go` 当前 IAVM opcode 已包含基础算术、比较、控制流、函数、局部/全局变量、对象数组、host call、异常相关占位和 `OpIndex`。但从 ialang bytecode 映射角度看，仍存在降级为 `OpNop` 或语义不完整的类别：

| 类别 | 缺口 | 影响 |
|---|---|---|
| 模块系统 | `ImportNamespace`、`ImportDynamic`、`ExportName`、`ExportAs`、`ExportDefault`、`ExportAll` | 无法完整运行依赖多文件/包导入导出的 ialang 程序 |
| Closure/upvalue | `OpClosure` 仍未形成真正的闭包捕获模型 | 内层函数访问外层局部变量场景不可靠 |
| 类与继承 | `Class`、`New`、`Super`、`SuperCall` 等未完整实现 | 面向对象语法无法进入 IAVM 稳定路径 |
| 异步 | `Await` 未实现 | 与现有 ialang async runtime 断层 |
| spread/nullish | spread array/object/call 与 nullish jump 未覆盖 | 现代语法与部分标准库模式受限 |
| 真值/跳转优化 | `Truthy` / `JumpIfTrue` 仍可通过组合指令表达，但缺少独立 opcode | 性能与 verifier 栈效应分析不够清晰 |

### 2.2 Verifier 仍需从“索引合法”走向“栈效应和语义合法”

当前 verifier 已能检查 header、类型、函数索引、局部变量、跳转目标、常量引用等基础规则。但后续平台化需要更强约束：

- 指令栈效应检查：每条指令 pop/push 数量、栈下溢、分支合流栈高度一致。
- 函数调用签名检查：`OpCall` 与 `FuncType` 参数/返回值匹配。
- Host capability 检查：模块声明、导入能力、调用 operation 之间的一致性。
- 资源上限检查：最大栈深、最大 locals、最大函数数、最大常量池大小。
- 异常控制流检查：`OpPushTry` / `OpPopTry` / `OpThrow` 的合法嵌套。

### 2.3 Runtime 返回值、错误和宿主能力模型需要收敛

当前 IAVM runtime 已可执行，但还需要进一步明确：

- `VM.Run()` 是否返回入口函数结果，还是只通过 stdout/host side effect 表达结果。
- builtin 错误是返回 `null`、抛出 VM 错误，还是进入 ialang 异常体系。
- `OpHostCall` 的参数与返回值如何稳定编码，避免直接依赖松散 `map[string]any`。
- `run-iavm` 中存在未使用的 `simpleHost` 辅助类型，可在确认无测试依赖后清理。
- `builtinPrint` 直接使用 `fmt.Print`，后续应考虑输出注入到 VM options 或 host，以便测试和嵌入式运行控制输出。

### 2.4 CLI 和开发体验仍偏最小可用

当前 `build-iavm` / `run-iavm` 已具备基本链路，但还缺少平台化开发所需的工具能力：

- `verify-iavm`：只验证模块，不执行。
- `inspect-iavm`：输出模块 header、types、imports、exports、capabilities、函数和常量池摘要。
- `build-iavm --dump`：输出 lowering 后的 module/IR，便于定位映射错误。
- `run-iavm --cap-config`：由配置控制 FS/Network capability，而不是固定 `MemFSProvider` / `HTTPProvider`。
- 更明确的错误输出：decode / verify / runtime 阶段分层打印。

### 2.5 文档与测试需要随实现刷新

现有 `docs/2026-04-21-development-plan.md` 与 `docs/2026-04-21/development-plan.md` 描述了从占位到最小闭环的转变，但部分内容已经被实现推进覆盖。后续应维护：

- IAVM opcode 映射矩阵。
- ialang bytecode → IAVM opcode 的语义差异表。
- Capability operation 规范。
- CLI 使用指南和示例。
- Verifier 规则清单。

---

## 3. 后续开发路线

### Phase 1：稳定当前最小闭环

目标：确保 `ialang source -> IAVM module -> verify -> run` 对基础语言子集稳定可用。

#### P0.1 建立端到端样例矩阵

建议新增或完善以下 e2e 场景：

| 场景 | 覆盖内容 |
|---|---|
| 基础表达式 | 常量、算术、比较、逻辑 |
| 条件分支 | `if/else`、嵌套条件、真假值 |
| 循环 | `while` / `for`、break/continue 如已支持 |
| 函数调用 | 参数、返回值、嵌套调用、递归基础场景 |
| 数组与对象 | literal、property get/set、index |
| builtin | `print`、`len`、`typeof`、`str`、`int`、`float` |
| capability | FS read/write、HTTP basic request |
| 错误路径 | 非法常量索引、非法跳转、缺失入口、host call 失败 |

完成标准：

- 每类场景至少有一个 `build-iavm + run-iavm` 或等价 Go 集成测试。
- 失败信息能定位到 decode / verify / lowering / runtime 阶段。

#### P0.2 统一测试入口

建议新增脚本或 Make target：

```sh
go test ./iacommon/...
go test ./iavm/...
go test ./ialang/...
```

完成标准：

- 开发者无需记忆 workspace 根目录 `go test ./...` 的限制。
- CI 与本地测试命令一致。

#### P0.3 清理最小闭环遗留代码

建议检查并处理：

- `ialang/cmd/ialang/run_iavm.go` 中的 `simpleHost` 是否仍有保留必要。
- `LowerToModule` 中重复调用 `collectGlobalNames(chunk)` 的逻辑是否可简化。
- `lowerChunkAsFunction` 如未被使用，确认是否保留为测试辅助或删除。
- Runtime builtin 输出是否需要通过 options 注入 writer。

完成标准：

- 不引入行为变化的前提下，删除无用代码。
- 保持 `go test ./iacommon/...`、`go test ./iavm/...`、`go test ./ialang/...` 全部通过。

---

### Phase 2：补齐 verifier 语义校验

目标：让 IAVM 模块在运行前具备更强安全性和可诊断性。

#### P1.1 指令栈效应表

为所有 opcode 定义静态栈效应：

| Opcode 类别 | 校验内容 |
|---|---|
| 常量/变量 | `OpConst` push 1，`OpLoad*` push 1，`OpStore*` pop 1 |
| 算术/比较 | 二元运算 pop 2 push 1，一元运算 pop 1 push 1 |
| 控制流 | jump target 合法，分支合流栈高度一致 |
| 调用 | 参数数量、函数引用、返回值数量 |
| 对象/数组 | array 元素数量、property name 常量类型 |
| host call | capability handle、operation、参数布局 |
| 异常 | try handler target、throw 栈输入、handler 栈状态 |

完成标准：

- verifier 能检测栈下溢。
- verifier 能检测基础分支合流栈高度不一致。
- verifier 测试覆盖每类失败原因。

#### P1.2 资源限制校验

建议新增 VerifyOptions 字段：

- `MaxFunctions`
- `MaxConstants`
- `MaxCodeSizePerFunction`
- `MaxLocalsPerFunction`
- `MaxStackPerFunction`
- `AllowedCapabilities`

完成标准：

- 可在 CLI 或嵌入式调用中开启保守限制。
- 资源超限返回明确 verifier error，而不是运行时失败。

#### P1.3 Capability 声明与调用一致性

建议建立 capability operation registry：

| Capability | Operation 示例 |
|---|---|
| FS | `fs.read_file`, `fs.write_file`, `fs.exists`, `fs.list_dir` |
| Network | `http.request`, `http.get`, `http.post` |

完成标准：

- 模块声明的 capability 与实际 `OpImportCap` / `OpHostCall` 可关联。
- 未声明或未授权 operation 在 verifier 或 runtime 初始化阶段失败。

---

### Phase 3：提升 ialang → IAVM 语义覆盖

目标：减少 lowering 中降级为 `OpNop` 的语义，让更大比例 ialang 程序可在 IAVM 路径运行。

#### P2.1 模块系统 lowering

建议优先支持静态 import/export：

1. `ImportName`
2. `ImportNamespace`
3. `ExportName`
4. `ExportDefault`
5. `ExportAll`

暂缓 `ImportDynamic`，直到异步和模块加载策略明确。

完成标准：

- 多文件 ialang 程序可 build 成 IAVM 模块或模块集合。
- `run-iavm` 能解析入口模块依赖。
- export 表与 verifier 检查一致。

#### P2.2 Closure/upvalue

建议设计显式 closure representation：

- 函数引用：function index。
- 捕获环境：upvalue 数组或对象。
- 指令补充：如 `OpMakeClosure`、`OpLoadUpvalue`、`OpStoreUpvalue`。

完成标准：

- 内层函数可读取外层局部变量。
- 闭包逃逸后仍能访问捕获值。
- verifier 能检查 upvalue 索引合法。

#### P2.3 对象、类和构造语义

建议分两步推进：

1. 先补完整对象属性语义：`ObjectKeys`、动态 property、method call 的 this/self 约定。
2. 再支持 `Class`、`New`、`Super`、继承链。

完成标准：

- 普通对象模型稳定。
- 类语法 lowering 不再降级为 `OpNop`。
- 构造函数、方法调用、继承至少具备基础 e2e 覆盖。

#### P2.4 异步与 host poll

已有 `OpHostPoll`，后续可与 ialang async runtime 对齐：

- 明确 promise/future 在 IAVM 的值表示。
- `OpAwait` 与 `OpHostPoll` 的协作模型。
- host async handle 生命周期。

完成标准：

- `await http.request(...)` 或等价场景可在 IAVM 中表达。
- runtime 可暂停/恢复，或至少以同步兼容模式执行。

---

### Phase 4：CLI、SDK 与平台化能力

目标：把 IAVM 从内部实验路径推进为可开发、可诊断、可嵌入的平台层。

#### P3.1 CLI 子命令完善

建议命令：

```text
ialang build-iavm <entry.ia> [-o app.iavm] [--dump-ir]
ialang verify-iavm <app.iavm> [--strict]
ialang inspect-iavm <app.iavm> [--json]
ialang run-iavm <app.iavm> [--cap-config caps.toml]
```

完成标准：

- build / verify / inspect / run 生命周期清晰分离。
- 错误输出对用户可读，对 CI 可机器解析。

#### P3.2 Capability 配置文件

建议支持最小 TOML 配置：

```toml
[fs]
mode = "mem" # mem | readonly | sandbox
root = "."

[network]
enabled = true
allow_hosts = ["example.com"]
```

完成标准：

- 默认运行保持安全保守。
- 用户可显式开启 FS/Network 能力。
- Host provider 选择不写死在 `run-iavm` 中。

#### P3.3 SDK 文档和样例

建议补充：

- `iavm/pkg/sdk/fs` 使用示例。
- `iavm/pkg/sdk/http` 使用示例。
- 嵌入式 Go 程序如何加载、验证、运行 IAVM 模块。

完成标准：

- 外部 Go 程序可不依赖 `ialang/cmd` 直接嵌入 IAVM。
- capability provider 的扩展点明确。

---

## 4. 推荐优先级排序

| 优先级 | 任务 | 原因 |
|---|---|---|
| P0 | 端到端测试矩阵 | 当前最小闭环已形成，首先要防回归 |
| P0 | 统一测试入口 | 解决 workspace 根目录测试命令易踩坑问题 |
| P0 | 清理无用/临时代码 | 避免实验代码固化进平台层 |
| P1 | Verifier 栈效应检查 | 平台 VM 的安全边界核心 |
| P1 | Capability operation 规范 | 防止 host ABI 与 builtin 体系继续漂移 |
| P2 | 模块系统 lowering | 提升真实项目可运行比例 |
| P2 | Closure/upvalue | 函数语义完整性的关键缺口 |
| P3 | CLI inspect/verify/cap-config | 提升开发体验和可运维性 |
| P3 | 异步与类系统 | 依赖前面语义模型稳定后推进 |

---

## 5. 建议下一批可执行任务

### Task 1：新增统一测试入口

输出物：

- `Makefile`、脚本或 CI 任务中的统一测试命令。
- 文档中说明 workspace 根目录直接 `go test ./...` 的限制。

验收：

- 一条命令完成 `iacommon`、`iavm`、`ialang` 全量测试。

实施状态（2026-04-22）：

- 已新增根目录 `Makefile`，提供 `test`、`test-iacommon`、`test-iavm`、`test-ialang` 目标。
- 已新增 `scripts/test-all.sh`，用于无 `make` 环境的一键测试。
- 本地环境缺少 `make`，因此使用 `bash scripts/test-all.sh` 完成验证。
- 验证结果：`go test ./iacommon/...`、`go test ./iavm/...`、`go test ./ialang/...` 全部通过。

### Task 2：补 IAVM 端到端测试矩阵

输出物：

- `iavm/pkg/integration` 或 `ialang/tests/e2e` 中的 build/run IAVM 用例。
- 覆盖基础表达式、控制流、函数、数组对象、builtin、capability。

验收：

- 任意 lowering 改动若破坏基础语义，测试能失败。

实施状态（2026-04-22）：

- 已在 `iavm/pkg/integration/integration_test.go` 中新增 `runIalangChunkPipeline` 测试辅助函数，统一执行 lowering、encode、decode、verify、runtime run 和结果读取。
- 已新增控制流分支、数组索引、字符串索引、对象属性读写、builtin `len` 的完整 pipeline 测试。
- 测试暴露并修复了 runtime 在模块级常量池启用后，`OpGetProp` / `OpSetProp` 仍读取函数局部常量的问题。
- 已在 `iavm/pkg/runtime/interpreter.go` 中新增统一常量读取逻辑，`OpConst`、`OpGetProp`、`OpSetProp` 均兼容模块级常量池和函数级常量。
- 验证结果：`go test ./iavm/...` 与 `bash scripts/test-all.sh` 全部通过。

### Task 3：设计并实现 verifier 栈效应检查

输出物：

- opcode 栈效应表。
- verifier 中的栈高度分析。
- 栈下溢、分支栈高度不一致、非法调用参数等失败测试。

验收：

- verifier 能阻止明显非法但索引合法的模块进入 runtime。

### Task 4：梳理并固化 capability operation 规范

输出物：

- FS/Network operation 命名、参数、返回值规范。
- runtime host call 与 `DefaultHost` 的映射测试。

验收：

- FS/HTTP 最小能力不再依赖散落的 `map[string]any` 约定。

### Task 5：模块系统 lowering 设计草案

输出物：

- import/export 映射方案。
- 单模块、多模块和包格式的关系说明。
- `run-iavm` 加载依赖模块的策略。

验收：

- 能明确下一步代码改动范围，不再直接扩大 runtime 复杂度。

---

## 6. 风险与缓解

| 风险 | 影响 | 缓解 |
|---|---|---|
| lowering 与 ialang bytecode 继续耦合 | ialang 编译器变化会破坏 IAVM | 增加映射矩阵和 table-driven tests |
| verifier 不做栈效应 | 非法模块进入 runtime，错误定位困难 | Phase 2 优先实现栈分析 |
| host ABI 使用 `map[string]any` 约定扩散 | 后续 capability 难维护、难兼容 | 建立 operation schema 和 typed adapter |
| 模块系统过早复杂化 | run-iavm 加载策略返工 | 先支持静态 import/export，动态 import 暂缓 |
| 异步/类/闭包同时推进 | 多个语义模型互相牵制 | 先 closure，再模块，再 async/class |
| 文档滞后 | 后续维护者误判实现状态 | 每个阶段完成后更新映射表和 CLI 文档 |

---

## 7. 当前结论

截至 2026-04-22，项目已经从“设计和骨架阶段”进入“最小平台闭环已可验证”的阶段。下一步不宜继续盲目扩展 opcode，而应优先稳定端到端测试、增强 verifier、安全收敛 capability ABI，并逐步补齐模块系统与 closure 语义。

推荐立即推进的工作顺序：

1. 统一测试入口。
2. IAVM 端到端测试矩阵。
3. Verifier 栈效应检查。
4. Capability operation 规范。
5. 模块系统 lowering 方案。

---

*生成日期：2026-04-22*  
*文档版本：v1.0*

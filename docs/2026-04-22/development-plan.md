# IA Platform 今日开发计划（2026-04-22）

## 1. 当前项目情况

### 1.1 项目结构

当前仓库是一个 Go workspace，根目录 `go.work` 引入三个模块：

| 模块 | 角色 | 当前重点 |
|---|---|---|
| `iacommon` | 公共 Host ABI、FS/Network provider、ialang 公共结构 | Host capability 边界已形成，FS/Network 默认宿主已有实现 |
| `ialang` | 语言前端、编译器、运行时、CLI、标准库与工具链 | 原有语言能力完整度较高，正在接入 IAVM build/run/verify/inspect 路径 |
| `iavm` | 类 WASM 平台 VM、模块格式、二进制编解码、verifier、runtime、ialang bridge | 最小平台闭环已推进较多，但当前源码存在冲突标记导致编译失败 |

根目录已有统一测试入口：

- `Makefile`: `test` / `test-iacommon` / `test-iavm` / `test-ialang`
- `scripts/test-all.sh`: 顺序执行三个子模块测试

注意：本机当前环境没有 `bash` 命令，无法直接运行 `bash scripts/test-all.sh`，需使用 `go test ./iacommon/...`、`go test ./iavm/...`、`go test ./ialang/...` 分别验证，或在具备 bash 的环境使用脚本。

### 1.2 已完成能力

根据现有文档、提交记录与源码状态，项目已经完成以下关键能力：

- `iacommon` 已抽出 Host API、FS provider、Network provider、ialang bytecode/packagefile 等公共层。
- `iavm` 已具备 Module、Function、Import/Export、Capability、常量池等核心模块结构。
- `iavm/pkg/binary` 已有 Encoder、Decoder、Verifier，并已推进到控制流、常量引用、栈效应和资源限制方向。
- `iavm/pkg/runtime` 已具备解释器、栈/帧、基础算术、比较、控制流、函数调用、对象/数组、builtin、host call、try-catch 异常恢复等能力。
- `iavm/pkg/bridge/ialang` 已能从 `*bytecode.Chunk` lowering 到 IAVM Module，并支持一批核心 opcode 映射。
- `ialang/cmd/ialang` 已出现 `build-iavm`、`run-iavm`、`verify-iavm`、`inspect-iavm` 相关命令实现。
- 最近开发重点已经从“打通最小闭环”转向“恢复稳定基线、补齐模块语义、固化 capability 规范”。

### 1.3 当前测试结果

本次检查在仓库根目录分别执行了三个子模块测试。

| 命令 | 结果 | 说明 |
|---|---|---|
| `go test ./iacommon/...` | PASS | 公共层测试通过 |
| `go test ./iavm/...` | FAIL | `verifier.go`、`verifier_test.go`、`interpreter.go` 存在 Git 冲突标记，导致编译失败 |
| `go test ./ialang/...` | FAIL | `ialang/cmd/ialang` 依赖 `iavm/pkg/binary` 和 `iavm/pkg/runtime`，因此随 IAVM 编译失败 |

当前最直接的阻塞不是功能缺失，而是源码中残留冲突标记：

| 文件 | 问题 |
|---|---|
| `iavm/pkg/binary/verifier.go` | `verifyFunctions` 中 `verifyStackDepth` 与 `verifyStackEffects` 两套实现冲突未解决 |
| `iavm/pkg/binary/verifier_test.go` | 栈深度/资源限制相关测试存在大段冲突内容 |
| `iavm/pkg/runtime/interpreter.go` | `OpGetProp` / `OpSetProp` 的常量读取逻辑存在冲突内容 |

这些冲突会阻断 IAVM 和 ialang CLI 构建，必须作为今日 P0 处理。

### 1.4 当前主要缺口

#### P0 阻塞：恢复可编译基线

源码中残留冲突标记，导致 `iavm` 不能编译，进而影响 `ialang/cmd/ialang`。

应优先确认并合并：

- verifier 使用新的 `verifyStackEffects(i, &fn, m)`，同时保留资源限制相关校验。
- runtime 属性读写使用统一 `vm.constantAt(frame, inst.A)`，兼容模块级常量池和函数级常量。
- verifier 测试保留栈下溢、分支合流栈高度、直接调用参数、MaxStack/资源限制等有效用例。

#### P1：Capability operation 规范仍需固化

`iacommon/pkg/host/api/default_host.go` 已支持一批操作：

- FS: `fs.read_file`、`fs.write_file`、`fs.append_file`、`fs.read_dir`、`fs.stat`、`fs.mkdir`、`fs.remove`、`fs.rename`
- Network: `network.http_fetch`

但当前仍以 `map[string]any` 作为 ABI 参数和返回值承载。短期可以接受，但需要文档化每个 operation 的参数、返回值、错误语义和权限要求，否则 runtime、SDK、CLI 配置会继续各自扩散约定。

#### P1：模块导入导出语义不完整

当前 lowering 中仍有一批 opcode 降级为 `OpNop`：

- `OpImportNamespace`
- `OpImportDynamic`
- `OpExportName`
- `OpExportAs`
- `OpExportDefault`
- `OpExportAll`
- `OpSuper`
- `OpSuperCall`
- `OpSpreadArray`
- `OpSpreadObject`
- `OpSpreadCall`
- `OpAwait`

其中今天最适合推进的是静态模块导出/导入的最小子集：

- `OpExportAs`
- `OpExportDefault` 的简单表达式导出
- `OpImportNamespace` 的设计与最小测试

#### P2：Closure upvalue 尚未完整支持

`OpClosure` 当前更接近函数引用加载，尚未形成真正的 lexical capture/upvalue 模型。ialang 原运行时支持闭包，但 IAVM 路径还没有完整捕获环境表示。

这项工作牵涉 compiler upvalue 分析、IAVM opcode、runtime 闭包对象和 verifier 索引校验，工作量较大，今天建议只完成设计草案或 spike，不作为当天必须落地的主线。

#### P2：类、继承、异步、spread 仍是后续语义大块

`class/new/super`、`await`、spread array/object/call 都还没有进入稳定 IAVM 路径。它们依赖对象模型、调用约定、异步 host poll 和闭包模型进一步收敛，不建议在恢复基线前并行扩大范围。

---

## 2. 今日目标

今日目标按“先止血、再固化、最后小步扩展”排序：

1. 恢复 IAVM 与 ialang CLI 的可编译测试基线。
2. 固化当前 capability operation 规范，避免 host ABI 继续隐式扩散。
3. 推进模块导出最小语义，优先处理低风险的 `OpExportAs` 和 `OpExportDefault` 简单场景。
4. 为下一阶段 closure/upvalue 和模块加载策略补充设计说明。

今日不建议把类系统、异步和 spread 作为主线开发任务。

---

## 3. 今日开发计划

### Task 1：清理冲突标记并恢复测试基线（P0）

目标：让 `iavm` 和 `ialang/cmd/ialang` 恢复可编译状态。

工作项：

- 修复 `iavm/pkg/binary/verifier.go` 中 `verifyStackDepth` / `verifyStackEffects` 的冲突，保留更完整的栈效应分析路径。
- 修复 `iavm/pkg/binary/verifier_test.go` 中冲突测试，保留当前阶段应覆盖的 verifier 失败场景。
- 修复 `iavm/pkg/runtime/interpreter.go` 中 `OpGetProp` / `OpSetProp` 常量读取冲突，统一使用 `vm.constantAt`。
- 执行 `gofmt`。
- 分别运行：
  - `go test ./iavm/...`
  - `go test ./ialang/...`
  - `go test ./iacommon/...`

验收标准：

- 仓库中不再存在 `<<<<<<<`、`=======`、`>>>>>>>` 冲突标记。
- `go test ./iavm/...` 通过。
- `go test ./ialang/...` 中 `cmd/ialang` 能正常编译。

预估工作量：0.5-1 小时。

执行状态（2026-04-22）：

- 已清理 `iavm/pkg/binary/verifier.go`、`iavm/pkg/binary/verifier_test.go`、`iavm/pkg/runtime/interpreter.go` 中残留的 Git 冲突标记。
- Verifier 保留 `verifyStackEffects(i, &fn, m)` 路径，继续覆盖栈下溢、分支合流栈高度、直接调用参数和 MaxStack 场景。
- Runtime 属性读写统一使用 `vm.constantAt(frame, inst.A)` 获取属性名常量，兼容模块级常量池与函数级常量。
- 已执行 `gofmt`。
- 验证结果：`go test ./iavm/...`、`go test ./ialang/...`、`go test ./iacommon/...` 全部通过。

### Task 2：补充 capability operation 规范文档（P1）

目标：把现有 Host ABI 从“代码里的隐式约定”固化成可维护文档。

建议新增或更新文档：

- `docs/iavm/abi-capabilities.md`

文档内容：

- capability 分类：`fs`、`network`
- operation 命名规范
- 每个 FS operation 的参数、返回值、错误类型
- `network.http_fetch` 的参数、返回值、超时和策略约束
- `DefaultHost` 当前支持范围与未支持范围
- `map[string]any` 临时承载策略与未来 typed adapter 方向

验收标准：

- 文档能覆盖 `iacommon/pkg/host/api/default_host.go` 中所有已实现 operation。
- 后续 runtime/SDK/CLI 读取 capability 配置时有明确依据。

预估工作量：1-1.5 小时。

执行状态（2026-04-22）：

- 已新增 `docs/iavm/abi-capabilities.md`。
- 文档覆盖当前 `DefaultHost` 已支持的 FS operations：`fs.read_file`、`fs.write_file`、`fs.append_file`、`fs.read_dir`、`fs.stat`、`fs.mkdir`、`fs.remove`、`fs.rename`。
- 文档覆盖当前 Network operation：`network.http_fetch`。
- 文档记录了当前尚未暴露的 handle/socket 类 operation，以及 `map[string]any` 作为临时 ABI 承载层的演进规则。

### Task 3：完善 `OpExportAs` lowering 与测试（P1）

目标：补齐 `export { local as alias }` 的最小 IAVM 导出语义。

工作项：

- 阅读 ialang bytecode 中 `OpExportAs` 的 `A/B` 字段含义。
- 在 `LowerToModule` 的 export 提取阶段识别 `OpExportAs`。
- 将 local name 映射到全局 index，alias 作为 `module.Export.Name`。
- 新增 bridge 或 integration 测试，覆盖：
  - `export { value as answer }`
  - alias 指向不存在的本地名时的行为

验收标准：

- `OpExportAs` 不再只在 lowering 指令流中降级为 `OpNop`，而是在 module exports 表中体现别名导出。
- `go test ./iavm/...` 通过。

预估工作量：1-2 小时。

执行状态（2026-04-22）：

- 已在 `LowerToModule` 中提取 `OpExportAs`，将 `export { local as alias }` 写入 `module.Export{Name: alias, Kind: ExportGlobal}`。
- 已新增 `buildGlobals`，根据 lowering 阶段收集的全局名填充 `mod.Globals`，使 `ExportGlobal` 能通过 verifier 的全局索引检查。
- 已新增 bridge 单元测试覆盖普通全局导出和别名导出。
- 验证结果：`go test ./iavm/pkg/bridge/ialang ./iavm/pkg/integration ./iavm/pkg/binary` 通过。

### Task 4：设计 `OpExportDefault` 最小语义（P1）

目标：明确 default export 在 IAVM module 中如何表达，先支持简单表达式或已命名绑定。

建议方案：

- 对 `export default <expr>`，优先 lowering 成名为 `default` 的 `ExportGlobal`。
- 如果 compiler 已生成临时全局，则导出该临时全局。
- 如果当前 bytecode 无法稳定追踪表达式值，则先文档化限制，并添加 verifier/lowering warning 设计。

工作项：

- 阅读 ialang compiler 对 `export default` 生成的 bytecode。
- 写一个最小 spike 测试，打印 chunk 指令或直接断言 lowering 结果。
- 决定是否当天实现简单场景，复杂匿名 class/function 留到后续。

验收标准：

- 至少形成明确设计结论。
- 若实现，则 `export default 42` 或等价简单场景在 module exports 中出现 `default`。

预估工作量：1-2 小时。

执行状态（2026-04-22）：

- 已采用最小语义：顶层 `OpExportDefault` lowering 为 `OpStoreGlobal default`，并在 module exports 中登记 `ExportGlobal("default")`。
- 该语义与 ialang VM 当前行为一致：`OpExportDefault` 消费栈顶值并写入默认导出。
- 已新增 bridge 单元测试覆盖 `export default <expr>` 的 lowering 结果。
- 复杂匿名 class/function 的完整语义仍依赖 IAVM class/closure 模型，后续继续扩展。

### Task 5：维护计划与状态记录（P2）

目标：保持开发节奏可追踪。

工作项：

- 更新 `docs/2026-04-22/follow-up-plan.md` 或追加状态说明。
- 如果 Task 1-4 有代码落地，在文档中记录完成状态、测试命令和结果。
- 把未完成项转入下一日计划，而不是留在含混状态。

验收标准：

- 今日结束时能从文档看出：
  - 当前基线是否恢复。
  - 哪些 IAVM 语义已推进。
  - 下一步优先级是什么。

预估工作量：0.5 小时。

---

## 4. 今日推荐排期

| 时间段 | 任务 | 输出 |
|---|---|---|
| 第 1 段 | Task 1：清理冲突并恢复测试基线 | IAVM/ialang 可编译，测试结果明确 |
| 第 2 段 | Task 2：capability operation 规范文档 | `abi-capabilities.md` |
| 第 3 段 | Task 3：`OpExportAs` lowering | 代码与测试 |
| 第 4 段 | Task 4：`OpExportDefault` spike 或实现 | 设计结论或最小实现 |
| 收尾 | Task 5：同步计划状态 | 文档更新 |

如果当天时间不足，优先保证 Task 1 和 Task 2 完成。Task 3/4 可以顺延，但不要在测试基线失败时继续扩大功能面。

---

## 5. 风险与处理

| 风险 | 影响 | 今日处理方式 |
|---|---|---|
| 冲突标记说明已有变更合并不完整 | 当前无法编译，所有功能开发都会受阻 | P0 先解决，不跳过 |
| verifier 两套栈校验逻辑取舍不清 | 可能误删资源限制或降低安全性 | 以 `verifyStackEffects` 为主，确保资源限制测试保留 |
| module-level constants 与 function constants 混用 | 属性读写、常量引用在 runtime 中行为不一致 | runtime 统一通过 `constantAt` 获取 |
| capability ABI 继续依赖散落 map key | SDK/CLI/runtime 后续难兼容 | 今天先文档化 operation schema |
| default export 语义牵涉匿名函数/类 | 容易把任务扩大 | 只处理简单场景，复杂场景记录限制 |

---

## 6. 今日完成标准

最低完成标准：

- [x] 清理所有 Git 冲突标记。
- [x] `go test ./iavm/...` 通过。
- [x] `go test ./ialang/...` 至少恢复到不因 IAVM 编译错误失败。
- [x] 新增 capability operation 规范文档。

理想完成标准：

- [x] `go test ./iacommon/...`、`go test ./iavm/...`、`go test ./ialang/...` 全部通过。
- [x] `OpExportAs` lowering 与测试完成。
- [x] `OpExportDefault` 简单语义有设计结论或最小实现。
- [x] 今日计划文档与后续计划文档状态一致。

---

## 7. 结论

截至 2026-04-22 当前收尾检查点，IA Platform 的方向已经从 IAVM 最小闭环进入平台稳定化阶段。今日已清理残留冲突标记，恢复 `iavm`、`ialang`、`iacommon` 全量测试基线，并完成 capability ABI 文档、`OpExportAs` lowering、简单 `OpExportDefault` lowering。

下一步建议继续推进模块导入侧语义，优先处理 `OpImportNamespace` 的对象绑定与测试；同时补充 `run-iavm --cap-config` 的配置设计，使 capability 文档进入 CLI 可用路径。

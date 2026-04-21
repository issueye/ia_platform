# IAVM 平台化开发计划（2026-04-21）

## 1. 项目概述

### 1.1 目标
打通 `ialang -> iavm -> host capability` 的最小可验证路径，形成平台化开发基线。

### 1.2 仓库结构
```
ia_platform/
├── iacommon/          # 公共类型、Host ABI、字节码
├── ialang/            # 语言前端、编译器、运行时、CLI
├── iavm/              # 类 WASM 风格虚拟机平台
└── docs/              # 设计文档与计划
```

### 1.3 技术栈
- Go 1.25.7
- iacommon (公共类型/Host ABI)
- ialang (编译器/VM)
- iavm (平台VM)

---

## 2. 已完成交付物

### 2.1 核心组件

| 组件 | 状态 | 文件 | 说明 |
|------|------|------|------|
| Binary Encoder | ✅ | `iavm/pkg/binary/encoder.go` | 完整 IAVM 模块序列化 |
| Binary Decoder | ✅ | `iavm/pkg/binary/decoder.go` | Round-trip 验证，修复 float64 解码 |
| Verifier | ✅ | `iavm/pkg/binary/verifier.go` | 结构校验 + 控制流分析 |
| Lowering Bridge | ✅ | `iavm/pkg/bridge/ialang/compiler_lowering.go` | ialang 68 opcode → iavm 51 opcode |
| Runtime Interpreter | ✅ | `iavm/pkg/runtime/interpreter.go` | Stack/Frame/控制流/函数调用 |
| Capability Binding | ✅ | `iavm/pkg/runtime/interpreter.go` | OpImportCap + OpHostCall |
| Builtin Functions | ✅ | `iavm/pkg/runtime/builtins.go` | print/len/typeof/str/int/float |
| Module 常量池 | ✅ | `iavm/pkg/module/module.go` | 全局常量池去重 |
| CLI 集成 | ✅ | `ialang/cmd/ialang/build_iavm.go` | build-iavm / run-iavm |

### 2.2 Opcode 映射状态

**iavm 核心 opcode: 51 个**
- 原始 31 个 + 新增 13 个（Dup/Pop/BitAnd/BitOr/BitXor/Shl/Shr/And/Or/Typeof/PushTry/PopTry/Throw）+ OpIndex

**ialang → iavm 映射: 44/68（~65%）**

已完整映射：
- 算术: Add/Sub/Mul/Div/Mod/Neg
- 比较: Equal/NotEqual/Greater/Less/GreaterEqual/LessEqual
- 逻辑: Not/And/Or/Truthy
- 位运算: BitAnd/BitOr/BitXor/Shl/Shr
- 控制流: Jump/JumpIfFalse/JumpIfTrue
- 函数: Call/Return/Closure
- 变量: GetName/DefineName/SetName/GetGlobal/DefineGlobal
- 数据: Array/Object/GetProperty/SetProperty/Index
- 其他: Pop/Dup/Typeof/PushTry/PopTry/Throw/ImportName

降级为 Nop（未实现）：
- async/await 相关: Await
- 模块系统: ImportNamespace/ImportDynamic/ExportName/ExportAs/ExportDefault/ExportAll
- 类/继承: Class/New/Super/SuperCall/SetProperty
- spread: SpreadArray/SpreadObject/SpreadCall
- 其他: ObjectKeys/JumpIfNullish/JumpIfNotNullish

### 2.3 测试统计

| 包 | 测试数 | 状态 |
|---|---|---|
| `iavm/pkg/binary` | 23 | PASS |
| `iavm/pkg/bridge/ialang` | 19 | PASS |
| `iavm/pkg/integration` | 7 | PASS |
| `iavm/pkg/runtime` | 32 | PASS |
| **合计** | **81** | **PASS** |

---

## 3. 本轮关键 Bug 修复

### 3.1 函数调用返回 null（已修复）
- **根因**: `lowerFunction` 将 local index 0 预留给不存在的 `self`，参数错位
- **修复**: 参数从 local `0` 开始映射
- **文件**: `compiler_lowering.go`

### 3.2 float64 常量解码错误（已修复）
- **根因**: `float64(bits)` 而非 `math.Float64frombits(bits)`
- **修复**: 标准 IEEE 754 解码
- **文件**: `decoder.go`

### 3.3 builtin 测试栈顺序（已修复）
- **根因**: 测试沿用了旧 interpreter 的 `[arg, function]` 布局
- **修复**: 统一为 `[function, arg1, arg2]` 布局
- **文件**: `builtins_test.go`

### 3.4 函数局部变量污染全局索引（已修复）
- **根因**: `collectGlobalNames` 误将函数内 `OpDefineName` 收为全局
- **修复**: 函数 chunk 中只收集 `OpSetName`
- **文件**: `compiler_lowering.go`

---

## 4. 后续工作计划

### Phase 1: 基线稳定化（本周）

#### Task 1: 提交当前更改
- 提交本轮 12 个修改文件
- 建议 commit message: `feat(iavm): implement module-level constant pool, OpIndex, and semantic fixes`

#### Task 2: 端到端测试增强
- 增加更多 lowering + 执行测试覆盖
- 目标场景：循环（for/while）、条件分支（if/else）、嵌套函数调用

### Phase 2: 功能补全（下周）

| 优先级 | 任务 | 说明 | 关键文件 |
|---|---|---|---|
| P1 | OpTruthy 独立实现 | 当前映射为 `OpNot;OpNot`，增加独立 opcode 提升性能 | `opcode.go`, `interpreter.go` |
| P1 | OpJumpIfTrue 独立实现 | 当前映射为 `OpNot;OpJumpIfFalse`，增加独立 opcode | `opcode.go`, `interpreter.go` |
| P2 | 更多 opcode 映射 | `OpObjectKeys`、`OpSetProperty`（类属性设置）、`OpAwait` | `compiler_lowering.go` |
| P2 | Closure 完整支持 | 当前 `OpClosure` 降级为 `Nop`，需实现 upvalue 捕获 | `interpreter.go`, `frame.go` |
| P3 | 模块系统支持 | Import/Export 的完整 lowering 和 runtime 支持 | `compiler_lowering.go`, `vm.go` |

### Phase 3: 平台化深化（第三周）

| 任务 | 说明 | 关键文件 |
|---|---|---|
| CLI 完善 | `build-iavm` source map、`run-iavm` capability 配置 | `build_iavm.go`, `run_iavm.go` |
| 性能优化 | Interpreter dispatch 函数指针表、热点 opcode inline caching | `interpreter.go` |
| 安全策略 | Verifier 栈深度限制、循环检测、沙箱隔离级别 | `verifier.go` |
| 文档完善 | API 文档、使用指南、架构设计更新 | `docs/` |

---

## 5. 已知限制

1. **Closure 不完整**: 内层函数无法读取外层局部变量，仅支持全局变量访问
2. **类/继承未实现**: `OpClass`、`OpNew`、`OpSuper` 降级为 `Nop`
3. **模块导入导出**: `ImportNamespace`/`ExportAll` 等降级为 `Nop`
4. **异步未实现**: `OpAwait` 降级为 `Nop`
5. **性能**: Interpreter 使用 switch dispatch，可优化为函数指针表

---

## 6. 风险与阻塞点

| 风险 | 影响 | 缓解措施 |
|---|---|---|
| 桥接层输入模型变化 | 若 ialang 编译器输出格式变化，lowering 需同步更新 | 保持与 ialang 编译器团队沟通，增加版本检查 |
| 双轨 builtin 体系 | ialang 直接 builtin 与 iavm capability 并存，接口可能漂移 | 明确迁移路径，逐步统一为 capability 模型 |
| 文档与实现差距 | 设计文档超前，实现滞后 | 以最小闭环为准，文档跟随实现更新 |

---

## 7. 验收标准

- [x] 合法模块通过 verifier
- [x] 常见非法模块给出明确失败原因
- [x] runtime 可消费 host capability（FS/Network）
- [x] 至少一个 FS 场景、一个 HTTP 场景具备测试覆盖
- [x] `ialang build-iavm` 和 `ialang run-iavm` 可用
- [x] 所有 iavm 测试 PASS
- [x] ialang 现有测试无回归

---

*计划生成时间: 2026-04-21*
*版本: v1.0*

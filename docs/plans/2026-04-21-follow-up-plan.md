# IAVM 平台化后续开发计划（2026-04-21 续）

## 1. 当前状态（截至本轮完成）

### 1.1 已完成工作

| 任务 | 状态 | 说明 |
|---|---|---|
| 修复 ialang/pkg/lang 测试构建失败 | ✅ | 新建 helpers_test.go，补齐 compileTestSource/itoaTest/createTestVM/runVMTestNoError/runVMTestExpectError |
| OpTruthy 独立 opcode | ✅ | 新增 OpTruthy，替代 OpNot;OpNot 模拟，更新 lowering 映射和测试 |
| OpJumpIfTrue 独立 opcode | ✅ | 新增 OpJumpIfTrue，替代 OpNot;OpJumpIfFalse 模拟，更新 lowering 映射和测试 |

### 1.2 测试状态

所有测试通过，无回归：
- `ialang/...`：全部 PASS
- `iavm/...`：全部 PASS（binary 23 测试, bridge 19 测试, integration 7 测试, runtime 32 测试）
- `iacommon/...`：全部 PASS

---

## 2. 后续工作计划

### Phase 2A: Opcode 映射补全（本轮）

#### P1: 实现 OpObjectKeys
**目标**：补全 `OpObjectKeys` 的 lowering 映射和 runtime 实现，支持 `for...in` 循环 lowering 后的执行。

**理由**：ialang 编译器对 `for (k in obj)` 会生成 `OpObjectKeys` + 迭代逻辑，当前降级为 Nop 导致 for-in 在 iavm 中无法工作。

**关键文件**：
- `iavm/pkg/core/opcode.go` — 确认 OpObjectKeys 是否已定义（若未定义则添加）
- `iavm/pkg/runtime/interpreter.go` — 实现 OpObjectKeys dispatch
- `iavm/pkg/bridge/ialang/compiler_lowering.go` — 解除 Nop 降级，建立映射
- `iavm/pkg/binary/verifier.go` — 增加跳转目标校验中的 OpJumpIfTrue 分支

**验收标准**：
- [ ] lowering 映射将 ialang `OpObjectKeys` 映射到 iavm 对应 opcode
- [ ] interpreter 能正确弹出 object 并压入 key 数组
- [ ] 新增 interpreter 单元测试覆盖 OpObjectKeys
- [ ] verifier 控制流分析包含 OpJumpIfTrue

#### P1: 实现 JumpIfNullish / JumpIfNotNullish
**目标**：补全可选链操作符 (`?.`) 相关的跳转 opcode。

**理由**：ialang 支持 `a?.b?.c` 可选链语法，编译器会生成 `OpJumpIfNullish` 或 `OpJumpIfNotNullish`。

**关键文件**：
- `iavm/pkg/core/opcode.go` — 添加 OpJumpIfNullish, OpJumpIfNotNullish
- `iavm/pkg/runtime/interpreter.go` — 实现 dispatch 逻辑
- `iavm/pkg/bridge/ialang/compiler_lowering.go` — 建立映射
- `iavm/pkg/binary/verifier.go` — 控制流校验包含新 opcode

**验收标准**：
- [ ] 新增两个 opcode 定义
- [ ] interpreter 正确实现 nullish 判断语义（null/undefined 视为 nullish，iavm 中只有 null）
- [ ] lowering 解除 Nop 降级
- [ ] 新增 interpreter 单元测试

### Phase 2B: Verifier 安全增强（本轮）

#### P1: 栈深度分析
**目标**：在 verifier 中增加每个函数的栈深度边界检查，防止恶意模块通过过度 push 导致栈溢出。

**理由**：平台化安全基线。当前 verifier 只检查 jump target 和 local index，没有栈操作安全分析。

**关键文件**：
- `iavm/pkg/binary/verifier.go` — 新增 `verifyStackDepth` 函数
- `iavm/pkg/binary/verifier_test.go` — 新增测试

**实现思路**：
1. 对每条指令分析其对栈的影响（push +N, pop -M, 净变化 delta）
2. 模拟执行每个基本块，记录最大栈深度和最小栈深度（不能为负）
3. 报告栈欠流（underflow）和栈溢出（overflow）风险

**验收标准**：
- [ ] 合法模块通过栈深度验证
- [ ] 栈欠流模块（如连续 pop）被检测并报告错误
- [ ] 栈深度超限模块（如无限循环 push）被检测

### Phase 2C: 集成测试增强（本轮）

#### P2: 端到端 lowering + 执行测试
**目标**：增加更多从 ialang chunk → iavm module → execute 的完整链路测试。

**场景**：
- [ ] 循环 lowering + 执行（while/for）
- [ ] 条件分支 lowering + 执行（if/else）
- [ ] 嵌套函数调用 lowering + 执行
- [ ] 对象属性访问 lowering + 执行

**关键文件**：
- `iavm/pkg/integration/integration_test.go`

---

## 3. 本轮任务优先级

| 优先级 | 任务 | 预计工作量 | 阻塞关系 |
|---|---|---|---|
| P1 | OpObjectKeys 实现 | 小 | 无 |
| P1 | JumpIfNullish/NotNullish 实现 | 小 | 无 |
| P1 | Verifier 栈深度分析 | 中 | 无 |
| P2 | 端到端集成测试增强 | 中 | 依赖 ObjectKeys |

---

## 4. 已知限制（本轮不改变）

1. **Closure 不完整**：仍保持当前降级策略，本轮不实现 upvalue 捕获
2. **类/继承未实现**：`OpClass`、`OpNew`、`OpSuper` 保持降级
3. **模块导入导出**：`ImportNamespace`/`ExportAll` 等保持降级
4. **异步未实现**：`OpAwait` 保持降级
5. **Spread 操作**：保持降级

---

## 5. 验收标准

- [ ] OpObjectKeys lowering + runtime + 测试全部完成
- [ ] JumpIfNullish / JumpIfNotNullish lowering + runtime + 测试全部完成
- [ ] Verifier 栈深度分析实现并附带测试
- [ ] 所有 iavm 测试 PASS
- [ ] 所有 ialang 测试 PASS（无回归）

---

*计划生成时间: 2026-04-21*
*版本: v1.1*

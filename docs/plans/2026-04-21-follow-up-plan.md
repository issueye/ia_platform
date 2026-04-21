# IAVM 平台化后续开发计划（2026-04-21 续）

## 1. 当前状态（本轮全部完成）

### 1.1 已完成工作

| 任务 | 状态 | 说明 |
|---|---|---|
| 修复 ialang/pkg/lang 测试构建失败 | ✅ | 新建 helpers_test.go，补齐 compileTestSource/itoaTest/createTestVM/runVMTestNoError/runVMTestExpectError |
| OpTruthy 独立 opcode | ✅ | 新增 OpTruthy，替代 OpNot;OpNot 模拟，更新 lowering 映射和测试 |
| OpJumpIfTrue 独立 opcode | ✅ | 新增 OpJumpIfTrue，替代 OpNot;OpJumpIfFalse 模拟，更新 lowering 映射和测试 |
| OpObjectKeys 实现 | ✅ | 新增 OpObjectKeys opcode，interpreter 实现，lowering 映射，测试覆盖 |
| JumpIfNullish/NotNullish 实现 | ✅ | 新增两个 opcode，interpreter 实现（peek+conditional pop），lowering 映射，测试覆盖 |
| Verifier 栈深度分析 | ✅ | 新增 verifyStackDepth，检测栈欠流和溢出，附带测试 |
| 集成测试增强 | ✅ | 新增 IfElse、WhileLoop、ObjectPropertyAccess 端到端测试 |
| Interpreter 常量回退 | ✅ | OpGetProp/OpSetProp 支持 module-level 常量回退（配合 buildModuleConstantPool） |
| Try-Catch 异常恢复 | ✅ | 新增 OpPushTry/OpPopTry/OpThrow，interpreter 实现 try-catch 机制 |
| OpClosure 实现 | ✅ | 新增 OpClosure opcode，支持函数表达式 lowering 和 runtime |
| 全局变量导出 | ✅ | LowerToModule 提取 OpExportName 到 module.Exports（ExportGlobal） |

### 1.2 测试状态

所有测试通过，无回归：
- `ialang/...`：全部 PASS
- `iavm/...`：全部 PASS（binary 25 测试, bridge 19 测试, integration 11 测试, runtime 38 测试）
- `iacommon/...`：全部 PASS

---

## 2. 已完成工作详情

### Phase 2A: Opcode 映射补全 ✅

| 任务 | 状态 | 关键文件 |
|---|---|---|
| OpObjectKeys | ✅ | `iavm/pkg/core/opcode.go`, `interpreter.go`, `compiler_lowering.go` |
| JumpIfNullish/NotNullish | ✅ | `iavm/pkg/core/opcode.go`, `interpreter.go`, `compiler_lowering.go`, `verifier.go` |

**Opcode 映射更新**：ialang → iavm 映射从 44/68 提升到 47/68（~69%）

### Phase 2B: Verifier 安全增强 ✅

| 任务 | 状态 | 关键文件 |
|---|---|---|
| 栈深度分析 | ✅ | `iavm/pkg/binary/verifier.go`, `verifier_test.go` |

**实现方案**：线性扫描每条指令，计算 stackDelta，检测 underflow（深度<0）和 overflow（深度>1024）。

### Phase 2C: 集成测试增强 ✅

| 场景 | 状态 | 说明 |
|---|---|---|
| IfElse | ✅ | 条件分支 lowering + 执行 |
| WhileLoop | ✅ | while 循环 lowering + 执行（向后跳转） |
| ObjectPropertyAccess | ✅ | 对象创建/属性设置/属性获取 lowering + 执行 |

**Bug 修复**：`interpreter.go` 中 `OpGetProp`/`OpSetProp` 增加对 `module.Constants` 的回退支持（当 `buildModuleConstantPool` 清空函数级常量后）。

---

## 3. 本轮任务优先级（回顾）

| 优先级 | 任务 | 预计工作量 | 实际完成 |
|---|---|---|---|
| P1 | OpObjectKeys 实现 | 小 | ✅ |
| P1 | JumpIfNullish/NotNullish 实现 | 小 | ✅ |
| P1 | Verifier 栈深度分析 | 中 | ✅ |
| P2 | 端到端集成测试增强 | 中 | ✅ |

---

## 4. 已知限制（本轮未改变）

1. **Closure upvalue 不完整**：OpClosure 仅加载函数引用，未实现真正的局部变量 upvalue 捕获
2. **类/继承未实现**：`OpClass`、`OpNew`、`OpSuper` 保持降级
3. **模块导入导出不完整**：`ImportNamespace`/`ExportAll`/`ExportAs`/`ExportDefault` 保持降级
4. **异步未实现**：`OpAwait` 保持降级
5. **Spread 操作**：保持降级
6. **嵌套函数调用**：lowerFunction 中函数引用处理仅限 entry 函数，内层函数通过 globalNames 访问（非函数引用）

---

## 5. 验收标准

- [x] OpObjectKeys lowering + runtime + 测试全部完成
- [x] JumpIfNullish / JumpIfNotNullish lowering + runtime + 测试全部完成
- [x] Verifier 栈深度分析实现并附带测试
- [x] 端到端集成测试增强（IfElse/WhileLoop/ObjectPropertyAccess/TryCatch/FunctionExpression）
- [x] Try-Catch 异常恢复机制实现
- [x] OpClosure 函数表达式支持
- [x] 全局变量导出（ExportGlobal）提取
- [x] 所有 iavm 测试 PASS
- [x] 所有 ialang 测试 PASS（无回归）

---

## 6. 提交记录（本轮）

1. `fix(ialang): add missing test helper functions for lang package`
2. `feat(iavm): add OpObjectKeys, OpJumpIfNullish, OpJumpIfNotNullish opcodes with verifier stack depth analysis`
3. `docs: add follow-up development plan for iavm platform`
4. `feat(iavm): add integration tests for if/else, while loop, object property access`
5. `feat(iavm): implement try-catch exception recovery with PushTry/PopTry/Throw`
6. `feat(iavm): add OpClosure support for function expressions`
7. `feat(iavm): extract global variable exports from OpExportName in lowering`

---

*计划生成时间: 2026-04-21*
*版本: v1.2*
*状态: 全部完成（含后续轮次追加）*

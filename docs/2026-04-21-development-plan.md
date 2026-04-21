# 今日开发计划（2026-04-21）

## 1. 项目现状分析

### 1.1 仓库结构
当前仓库是一个 Go workspace，多模块组织如下：

- `iacommon`：公共类型、Host ABI、文件系统/网络能力抽象、通用字节码与包格式
- `ialang`：语言前端、编译器、运行时、CLI、内建模块、测试与工具链
- `iavm`：类 WASM 风格的虚拟机平台实现，目前处于骨架与桥接预研阶段
- `docs/iavm`：IAVM 平台设计与实施草案

工作区配置见 `go.work:1`。

### 1.2 技术栈
- 主语言：Go 1.25.7（`go.work:1`、`ialang/go.mod:3`、`iacommon/go.mod:3`、`iavm/go.mod:3`）
- 运行与网络依赖：`gorilla/websocket`、`net/http`
- 数据访问相关：SQLite / MySQL / PostgreSQL / MSSQL 驱动
- 配置与序列化：TOML / YAML
- 调度与异步：`robfig/cron`、自定义 async runtime / goroutine pool
- 编辑器生态：`ialang/tools/vscode-ialang`

### 1.3 当前核心执行链路
`ialang` 已具备“源码 -> 词法/语法分析 -> 编译 -> VM 执行”的完整闭环：

1. CLI 入口：`ialang/cmd/ialang/main.go:5`
2. 运行命令主流程：`ialang/cmd/ialang/run.go:16`、`ialang/cmd/ialang/run.go:45`、`ialang/cmd/ialang/run.go:77`
3. 词法分析器：`ialang/pkg/lang/frontend/lexer.go:5`
4. 模块加载与导入解析：`ialang/pkg/lang/module_loader.go:12`、`ialang/pkg/lang/module_loader.go:50`
5. 当前 VM 入口与执行循环：`ialang/pkg/lang/runtime/vm/vm.go:9`、`ialang/pkg/lang/runtime/vm/vm.go:14`、`ialang/pkg/lang/runtime/vm/vm.go:69`、`ialang/pkg/lang/runtime/vm/vm.go:106`

### 1.4 平台化进展
IAVM 方向已有较清晰设计文档，但代码实现仍偏早期：

- 设计目标与平台定位：`docs/iavm/wasm-like-vm-platform-design.md`
- 实施分层与首版范围：`docs/iavm/implementation-outline.md:15`、`docs/iavm/implementation-outline.md:16`
- `ialang -> iavm` lowering 当前仅为占位：`iavm/pkg/bridge/ialang/compiler_lowering.go:5`
- IAVM runtime 当前仍是骨架：`iavm/pkg/runtime/vm.go:9`
- Host 能力抽象已开始落地，`DefaultHost` 已具备 capability acquire/call 机制：`iacommon/pkg/host/api/default_host.go:21`、`iacommon/pkg/host/api/default_host.go:30`、`iacommon/pkg/host/api/default_host.go:70`

### 1.5 近期改动信号
最近几次提交高度集中在“公共类型抽离”和“运行时/字节码公共化”：

- `68e0548`：将 `rttypes.Value` 替换为 `common.Value`
- `8943726`：将公共类型移动到 `iacommon`
- `5ad5d8b`：将 `bytecode` 和 `packagefile` 移动到 `iacommon/pkg/ialang`

这说明当前项目的主线不是继续堆叠语言特性，而是为 `ialang` 与 `iavm` 共享基础设施做准备。

---

## 2. 今日开发目标建议

结合当前代码状态，今日最适合推进的方向应聚焦在“平台桥接最小闭环”，而不是继续扩展语言表层能力。

### 建议总目标
打通 `ialang -> iavm` 的最小可验证路径，形成后续平台化开发基线。

### 今日建议优先级

#### P0：明确并收敛 lowering 输入输出模型
目标：定义 `ialang` 编译产物如何稳定映射到 `iavm/module.Module`。

建议动作：
- 梳理当前 `ialang` 编译结果中的最小必要字段
- 对齐 `iacommon/pkg/ialang/bytecode` 与 `iavm/pkg/core` / `iavm/pkg/module` 的映射关系
- 给 `LowerToModule` 增加明确输入类型与输出约束

原因：`iavm/pkg/bridge/ialang/compiler_lowering.go:5` 目前仍是空实现，这是桥接工作的第一阻塞点。

#### P1：补齐 IAVM verifier 最小规则集
目标：让模块在进入 runtime 前具备基本可验证性。

建议动作：
- 先实现模块头、函数表、导入导出、capability 声明、常量/指令索引范围检查
- 暂不追求完整控制流证明，先实现“结构合法性校验”
- 为 verifier 增加失败原因模型，便于 CLI 与测试输出

原因：设计文档已明确 verifier 是首版核心之一，见 `docs/iavm/implementation-outline.md:337`、`docs/iavm/implementation-outline.md:358`。

#### P1：把 Host ABI 与 runtime builtin 的边界进一步拉直
目标：把现有 `ialang` 中直接触达宿主的 builtin 能力，逐步切换到 capability 驱动模型。

建议动作：
- 先聚焦 FS / HTTP 两类能力
- 对照 `iacommon/pkg/host/api/default_host.go` 已有接口，统一 operation naming 与参数载荷格式
- 明确 `builtin_fs.go` / `builtin_http.go` 到 IAVM host call 的迁移路径

原因：实施文档明确首版只抽离文件系统与网络能力，见 `docs/iavm/implementation-outline.md:15`。

#### P2：建立桥接与平台的回归测试基线
目标：避免后续平台演进时桥接层反复回归。

建议动作：
- 为 lowering 增加 table-driven tests
- 为 verifier 增加合法/非法模块样例测试
- 为 host capability binding 增加最小集成测试

原因：当前项目测试密度高，CI 已默认执行 `go test ./...`，见 `ialang/.github/workflows/ci.yml:1`。

---

## 3. 今日开发计划

## 任务 1：梳理 `ialang` 到 `iavm` 的桥接契约
**目标**：形成一份最小 mapping 清单，明确当前编译产物如何落到 IAVM 模块结构。

**输出物**：
- lowering 输入结构说明
- opcode / 常量池 / locals / exports 映射表
- 未决差异清单

**重点关注文件**：
- `ialang/cmd/ialang/run.go:45`
- `ialang/pkg/lang/module_loader.go:54`
- `ialang/pkg/lang/runtime/vm/vm.go:69`
- `iavm/pkg/bridge/ialang/compiler_lowering.go:5`

**完成标准**：
- `LowerToModule` 的输入输出边界不再模糊
- 能列出“现有 chunk -> iavm module”最小字段映射

## 任务 2：实现 verifier 最小可用版本
**目标**：实现模块结构校验，而不是直接执行未经约束的模块。

**输出物**：
- `iavm/pkg/binary/verifier.go` 最小实现
- 验证结果结构体与错误分类
- 对应单元测试

**校验范围建议**：
- header/version 合法性
- function/type/import/export 基本引用关系
- capability 声明结构
- 指令引用范围
- entry/export 基础一致性

**完成标准**：
- 合法模块通过
- 常见非法模块能给出明确失败原因

## 任务 3：打通 capability binding 的最小运行路径
**目标**：让 runtime 至少能在 FS / HTTP 场景中完成 capability acquire 与 host.call 调用。

**输出物**：
- capability acquire/bind 流程
- FS/HTTP operation 映射实现
- 最小集成测试

**重点关注文件**：
- `iacommon/pkg/host/api/default_host.go:21`
- `iacommon/pkg/host/api/default_host.go:30`
- `iacommon/pkg/host/api/default_host.go:70`
- `iavm/pkg/runtime/vm.go:9`

**完成标准**：
- runtime 可消费 host capability
- 至少一个 FS 场景、一个 HTTP 场景具备测试覆盖

## 任务 4：为 CLI / 工具链预留验证入口
**目标**：为后续命令行验证、编译与调试提供统一接口。

**输出物**：
- verifier / lowering 的调用入口设计
- CLI 子命令或内部调用点草案
- 与现有 `run` 流程的衔接方案

**重点关注文件**：
- `ialang/cmd/ialang/main.go:5`
- `ialang/cmd/ialang/run.go:16`

**完成标准**：
- 明确未来是扩展现有 CLI，还是新增 IAVM 专用入口
- 不破坏当前 `ialang run` 链路

---

## 4. 风险与阻塞点

### 风险 1：桥接层输入模型尚未收敛
`LowerToModule` 仍为空实现，说明 `ialang` 当前编译结果与 IAVM 模块格式之间还没有稳定契约。若不先收敛这层映射，后续 verifier/runtime 实现会反复返工。

### 风险 2：IAVM runtime 仍偏骨架
`iavm/pkg/runtime/vm.go:9` 只有基本结构，没有真正执行能力，因此今天不宜把目标定为“完整运行 IAVM 程序”，更适合先完成 verifier 与 capability binding 的基础设施。

### 风险 3：宿主能力边界仍有双轨并存
`ialang` 现有 builtin 体系仍可直接走本地运行时，而 `iacommon` / `iavm` 已开始引入 capability/host API。若今天同时大规模改动两条路线，容易出现接口漂移。

### 风险 4：文档与实现存在领先差
`docs/iavm/implementation-outline.md` 已有较细实施草案，但实现明显未跟上。开发时应以“最小闭环”为准，避免一次性追全文档范围。

---

## 5. 今日建议交付物

建议今天以“4 个可检查交付物”为收口：

1. `ialang -> iavm` mapping 草案
2. verifier 最小实现与测试
3. FS / HTTP capability binding 最小闭环
4. 后续 CLI 接入点说明

如果今日资源有限，建议优先完成前两项；这是后续平台化演进的地基。

---

## 6. 结论

当前仓库已经完成了 `ialang` 运行时基础能力建设，并正在把公共类型、字节码与宿主能力抽到共享层。结合最近提交与现有代码状态，今天最有价值的推进方向不是新增语言语法，而是补上 `ialang -> iavm -> host capability` 这条平台化链路中的关键缺口。

建议今日开发围绕“桥接契约、模块校验、Host ABI 最小闭环”展开，这样既能承接近期重构成果，也能为后续 IAVM 真正落地建立稳定基线。

# 类 WASM 的字节码虚拟机平台设计方案

## 1. 文档目标

本文面向当前 `ialang` 仓库，设计一个“类 WASM”的字节码虚拟机平台，目标是：

1. 以 **可验证、可移植、可沙箱化** 的字节码格式作为运行载体。
2. 允许使用 **ialang** 编写应用程序，并编译到该平台上运行。
3. 将宿主能力从语言运行时中剥离，重点抽象出：
   - 文件系统访问系统
   - 网络访问系统
4. 为后续实现 `iavm` 子项目提供结构化蓝图。

---

## 2. 当前仓库现状与设计依据

当前 `ialang` 已具备从源码到字节码再到 VM 执行的基本闭环：

- 源码读取、词法/语法分析、编译、执行入口位于 `ialang/cmd/ialang/run.go:16`、`ialang/cmd/ialang/run.go:45`、`ialang/cmd/ialang/run.go:77`
- 词法分析器位于 `ialang/pkg/lang/frontend/lexer.go:16`
- 语法分析器位于 `ialang/pkg/lang/frontend/parser.go:104`
- AST 定义位于 `ialang/pkg/lang/ast/ast.go:19`
- 当前字节码模型位于 `ialang/pkg/lang/bytecode/bytecode.go:3`
- 当前 VM 主循环位于 `ialang/pkg/lang/runtime/vm/vm.go:9`、`ialang/pkg/lang/runtime/vm/vm.go:88`
- 当前沙箱策略位于 `ialang/pkg/lang/runtime/sandbox.go:9`
- 当前内建模块注册位于 `ialang/pkg/lang/runtime/builtin/registry.go:8`
- 当前文件系统内建模块直接绑定宿主 OS，位于 `ialang/pkg/lang/runtime/builtin/fs.go:8`
- 当前 HTTP/网络能力直接绑定 Go `net/http`，位于 `ialang/pkg/lang/runtime/builtin/http.go:96`

### 2.1 现状总结

当前 `ialang` 实际上已经是一个：

- TS-like 语法前端
- 自定义 AST
- 自定义字节码编译器
- 基于栈机的解释执行 VM
- 带有模块系统、异步运行时、沙箱策略、内建模块注册机制的语言运行时

### 2.2 当前限制

虽然已有 VM 和字节码，但它还不是“平台级 VM”，主要问题是：

1. **字节码仍然偏语言内部实现**，缺少稳定的模块封装格式与版本协议。
2. **宿主能力耦合过深**：`fs`、`http` 直接调用 Go 标准库，不利于跨宿主、跨环境部署。
3. **沙箱控制粒度偏粗**：当前是布尔开关与模块白名单，缺少 capability/handle 级授权。
4. **平台 ABI 不独立**：语言层内建模块 API 与宿主实现绑定，不利于未来接入浏览器、边缘运行时、嵌入式宿主。
5. **缺少模块验证层**：没有类似 WASM verifier 的加载前约束检查。

因此，`iavm` 的核心职责不是简单“再写一个 VM”，而是把现有 `ialang VM` 进一步平台化。

---

## 3. 平台总体定位

`iavm` 定位为：

> 一个面向 `ialang` 及未来其他前端语言的、具备模块格式、验证器、宿主能力抽象、沙箱与可移植 ABI 的字节码虚拟机平台。

它与当前 `ialang` 的关系如下：

- `ialang`：语言前端、语义编译层、标准库映射层
- `iavm`：模块格式、指令集、验证器、执行器、宿主接口、权限系统

换句话说：

- `ialang` 负责“把程序表达出来”
- `iavm` 负责“把程序安全、稳定、可控地跑起来”

---

## 4. 设计目标

## 4.1 核心目标

1. **类 WASM 模块化**
   - 统一二进制模块格式
   - 可导入/导出函数、内存、宿主能力
   - 可做静态校验

2. **宿主能力显式化**
   - 文件系统和网络能力必须通过导入能力获得
   - 默认无权限
   - 能力实例由宿主注入

3. **兼容 ialang 编译链**
   - 允许复用现有 lexer/parser/AST/compiler 基础设施
   - 先做“现有字节码平台化”，再逐步演进为更强验证模型

4. **强沙箱**
   - 指令步数限制
   - 时间限制
   - 内存限制
   - FS/Network capability 限制
   - 模块导入白名单

5. **跨宿主可移植**
   - 本地 CLI 宿主
   - Server 宿主
   - 未来浏览器/Worker/边缘宿主

## 4.2 非目标

1. 首版不追求完整兼容 WASM 二进制格式。
2. 首版不强行引入 LLVM/JIT。
3. 首版不要求支持多语言前端。
4. 首版不直接替换现有所有 `ialang` 内建模块。

---

## 5. 总体架构

建议 `iavm` 拆为 6 层：

```text
+------------------------------------------------+
|               ialang 应用源码                  |
+------------------------------------------------+
|      ialang 前端（Lexer / Parser / AST）       |
+------------------------------------------------+
|   ialang -> IAVM IR / Bytecode Lowering        |
+------------------------------------------------+
|    IAVM Module Format + Verifier + Loader      |
+------------------------------------------------+
|        IAVM Runtime / Interpreter Core         |
+------------------------------------------------+
| Host ABI: FS / Network / Clock / Process       |
+------------------------------------------------+
```

### 5.1 编译时组件

1. **Frontend**：复用 `ialang/pkg/lang/frontend`
2. **AST**：复用 `ialang/pkg/lang/ast`
3. **Lowering/Compiler**：从 AST 编译为 `IAVM Module`
4. **Verifier**：对模块结构、导入、导出、控制流、常量池、能力声明做静态校验
5. **Packager**：生成 `.iavm` 二进制或 `.iavmpkg` 包

### 5.2 运行时组件

1. **Loader**：加载模块、解析 section、链接 imports
2. **Capability Binder**：由宿主注入文件系统/网络等能力实例
3. **Interpreter**：执行字节码
4. **Resource Manager**：句柄、内存页、socket、文件描述符、异步任务管理
5. **Security Manager**：步数/时间/内存/权限检查

---

## 6. 模块格式设计

建议定义 `IAVM Binary Module`，扩展名暂定为 `.iavm`。

## 6.1 文件结构

```text
+-----------------------------+
| Magic: IAVM                 |
| Version                     |
| Header Flags                |
+-----------------------------+
| Type Section                |
| Import Section              |
| Function Section            |
| Table Section (optional)    |
| Memory Section (optional)   |
| Global Section              |
| Export Section              |
| Code Section                |
| Data Section                |
| Capability Section          |
| Custom Section              |
+-----------------------------+
```

## 6.2 Header 关键字段

- `magic = "IAVM"`
- `version = 1`
- `target = ialang`
- `abi_version`
- `feature_flags`

## 6.3 Capability Section

这是区别于当前 `ialang` 字节码的重要增强，记录模块声明需要的宿主能力，例如：

```json
{
  "fs": {
    "required": true,
    "rights": ["read", "write", "list"],
    "preopens": ["/workspace", "/tmp"]
  },
  "network": {
    "required": true,
    "rights": ["tcp_connect", "http_fetch"],
    "allow_hosts": ["api.example.com:443"],
    "allow_schemes": ["https"]
  }
}
```

平台加载时先检查 capability，再完成模块实例化。

---

## 7. 指令集设计

## 7.1 设计原则

1. 保持与当前 `ialang` 栈机模型兼容，降低迁移成本。
2. 增加“平台级”能力指令，而不是将所有宿主访问写死在内建模块里。
3. 指令集稳定、语义清晰、支持验证。

## 7.2 指令分层

### A. 通用执行指令

- `const`
- `add/sub/mul/div/mod`
- `eq/ne/lt/gt/le/ge`
- `and/or/not`
- `jump/jump_if/jump_if_not`
- `call/return`
- `load_local/store_local`
- `load_global/store_global`
- `make_array/make_object`
- `get_prop/set_prop`

### B. 模块与链接指令

- `import_func`
- `import_cap`
- `export_func`
- `export_global`

### C. 宿主能力指令

- `cap.acquire`
- `cap.release`
- `host.call`
- `host.poll`

### D. 文件系统与网络系统调用指令

可选两种路线：

#### 路线 1：统一 host.call

所有 FS/Network 操作都走：

```text
host.call <capability-kind> <operation-id>
```

优点：
- ABI 易扩展
- 指令集膨胀小

缺点：
- 验证器难做得极强类型化

#### 路线 2：专用 sys 指令

例如：

- `fs.open`
- `fs.read`
- `fs.write`
- `net.connect`
- `net.send`
- `net.recv`
- `http.fetch`

优点：
- 可读性与可验证性更强

缺点：
- 指令集更大

### 7.3 建议

首版采用 **统一 `host.call` + 操作码表**，二版再按热点能力拆分专用指令。

---

## 8. 值模型与内存模型

`ialang` 当前值模型偏动态对象系统，因此 `iavm` 不建议直接照抄 WASM 的纯数值栈，而应采用 **混合值模型**。

## 8.1 值类型

建议首版支持：

- `i32`
- `i64`
- `f64`
- `bool`
- `string`
- `bytes`
- `array_ref`
- `object_ref`
- `func_ref`
- `host_handle`
- `null`

## 8.2 内存模型

建议采用双层内存：

1. **VM Value Stack**：用于表达式求值与调用栈
2. **Linear Heap / Object Heap**：用于对象、数组、字符串、bytes、句柄元信息

### 设计理由

- 纯 WASM 线性内存更适合低级语言
- `ialang` 更接近 JS/TS，对象与字符串频繁出现
- 句柄化对象可避免把宿主资源直接暴露为普通对象

---

## 9. 文件系统访问系统设计

这是本设计的重点之一。

### 9.1 设计原则

1. **默认无文件系统权限**
2. **路径不直接等于宿主真实路径权限**
3. **以 capability + preopen + handle 模型访问**
4. **编程模型对 ialang 友好**
5. **支持替换后端**：本地磁盘、内存文件系统、只读资源包、远程对象存储映射

### 9.2 核心抽象

定义三层抽象：

#### 1) FileSystemProvider

代表宿主提供的某种文件系统实现。

```text
interface FileSystemProvider {
  open(path, opts) -> FileHandle
  stat(path) -> FileInfo
  readDir(path) -> DirEntries
  mkdir(path, opts)
  remove(path, opts)
  rename(oldPath, newPath)
}
```

#### 2) FileCapability

表示某个模块在运行期得到的文件系统访问授权。

```text
FileCapability {
  id
  provider
  rights: read|write|list|create|delete|rename|stat
  preopens: ["/workspace", "/tmp"]
  readonly: bool
  quotaBytes
}
```

#### 3) FileHandle

对具体打开文件的句柄抽象。

```text
FileHandle {
  handleId
  path
  mode
  cursor
  closed
}
```

### 9.3 路径模型

建议引入 **虚拟路径空间**：

- 模块内只能看到虚拟根，如 `/workspace/app.txt`
- 宿主负责把 `/workspace` 映射到真实目录
- 模块永远不直接知道真实绝对路径

例如：

```text
虚拟路径: /workspace/data/config.json
宿主映射: E:/code/issueye/project/data/config.json
```

这样可以天然避免路径穿越越权，并支持跨平台宿主。

### 9.4 preopen 机制

借鉴 WASI：

宿主在创建 VM 实例时注入可见目录：

```json
{
  "preopens": {
    "/workspace": "E:/code/issueye/project",
    "/tmp": "C:/Temp/iavm"
  }
}
```

模块若访问 `/etc/passwd` 或 `../` 越界路径，应在路径规范化后被拒绝。

### 9.5 权限模型

建议细分为：

- `fs.read`
- `fs.write`
- `fs.list`
- `fs.create`
- `fs.delete`
- `fs.rename`
- `fs.stat`
- `fs.watch`（二期）

示例：

- 只读配置模块：`read + stat`
- 生成构建产物模块：`read + write + create + list`
- 清理模块：额外需要 `delete`

### 9.6 文件系统 ABI

建议提供以下最小操作集：

- `fs.open(path, mode)` -> `handle`
- `fs.read(handle, size)` -> `bytes`
- `fs.write(handle, bytes)` -> `n`
- `fs.seek(handle, offset, whence)` -> `pos`
- `fs.close(handle)`
- `fs.stat(path)` -> `FileInfo`
- `fs.read_dir(path)` -> `DirEntry[]`
- `fs.mkdir(path, recursive)`
- `fs.remove(path, recursive)`
- `fs.rename(old, new)`

### 9.7 ialang 暴露方式

建议把当前 `fs` 模块改造成 capability-backed 模块，而不是直接调用 `os`。

理想使用方式：

```ts
import { fs } from "@platform/fs";

let text = fs.readFile("/workspace/a.txt");
fs.writeFile("/workspace/out.txt", text + "\n");
```

语言层 API 可以保持高层易用，但底层必须变为：

```text
ialang fs API -> IAVM host fs ABI -> Host FileSystemProvider
```

### 9.8 可替换后端

建议至少支持：

1. **LocalFSProvider**：映射宿主本地目录
2. **MemFSProvider**：内存文件系统，适合测试和沙箱
3. **BundleFSProvider**：打包资源只读文件系统
4. **OverlayFSProvider**：只读底层 + 可写上层

### 9.9 与当前仓库的衔接

当前 `ialang/pkg/lang/runtime/builtin/fs.go:8` 直接调用 `os.ReadFile`、`os.WriteFile`、`os.ReadDir` 等。后续应改造为：

```text
builtin/fs.go
   -> 调用 runtime capability manager
   -> capability manager 调用 provider
   -> provider 访问真实宿主资源
```

这样 `builtin/fs.go` 只保留语言层包装，不再承担宿主实现细节。

---

## 10. 网络访问系统设计

这是本设计的第二个重点。

### 10.1 设计原则

1. **默认无网络权限**
2. **按协议、主机、端口、方向细分授权**
3. **区分高层 HTTP 能力与底层 socket 能力**
4. **异步优先，兼容同步封装**
5. **支持替换后端**：本地 TCP/HTTP、代理层、Mock 网络层

### 10.2 核心抽象

#### 1) NetworkProvider

```text
interface NetworkProvider {
  dial(endpoint, opts) -> SocketHandle
  listen(endpoint, opts) -> ListenerHandle
  httpFetch(req) -> HttpResponse
}
```

#### 2) NetworkCapability

```text
NetworkCapability {
  id
  rights: tcp_connect|tcp_listen|udp_bind|http_fetch|http_serve|ws_connect
  allowHosts
  allowPorts
  allowSchemes
  allowCIDRs
  egressOnly
}
```

#### 3) Host Handles

- `SocketHandle`
- `ListenerHandle`
- `HttpStreamHandle`

句柄统一纳入 VM 资源表管理。

### 10.3 分层模型

建议网络访问分为两层：

#### A. 高层 HTTP 能力

适合业务脚本，大多数 ialang 应用首先使用它。

最小 ABI：

- `http.fetch(req)` -> `resp`
- `http.stream(req)` -> `streamHandle`
- `http.serve(spec)` -> `serverHandle`（二期）

#### B. 低层 Socket 能力

适合网关、代理、协议桥接、RPC、IAX、IPC over socket 等场景。

最小 ABI：

- `net.connect(addr, opts)` -> `socket`
- `net.listen(addr, opts)` -> `listener`
- `net.accept(listener)` -> `socket`
- `net.send(socket, bytes)`
- `net.recv(socket, size)` -> `bytes`
- `net.close(handle)`

### 10.4 权限模型

建议最小权限集合：

- `network.http.fetch`
- `network.http.serve`
- `network.tcp.connect`
- `network.tcp.listen`
- `network.udp.bind`
- `network.ws.connect`
- `network.resolve_dns`

建议配合细粒度限制：

- `allow_hosts`
- `allow_ports`
- `allow_schemes`
- `allow_cidrs`
- `max_connections`
- `max_inflight_requests`
- `max_bytes_per_request`

### 10.5 URL/Endpoint 白名单

示例：

```json
{
  "network": {
    "rights": ["http_fetch"],
    "allow_hosts": ["api.github.com", "example.com"],
    "allow_ports": [443],
    "allow_schemes": ["https"]
  }
}
```

这样即便脚本里写了：

```ts
http.client.get("http://127.0.0.1:2375/")
```

也应在 capability 检查阶段被拒绝。

### 10.6 网络 ABI

#### HTTP ABI

请求结构：

```json
{
  "method": "GET",
  "url": "https://api.example.com/data",
  "headers": {"Authorization": "Bearer ..."},
  "body": "",
  "timeout_ms": 5000
}
```

响应结构：

```json
{
  "status": 200,
  "headers": {"content-type": "application/json"},
  "body": "...",
  "trailers": {}
}
```

#### Socket ABI

```text
connect(endpoint, opts) -> socketHandle
send(socketHandle, bytes) -> written
recv(socketHandle, size) -> bytes
shutdown(socketHandle, how)
close(socketHandle)
```

### 10.7 ialang 暴露方式

当前 `ialang/pkg/lang/runtime/builtin/http.go:96` 直接构造 `http.Client` 并开放 `client.get/post/request` 与 `server.proxy/forward/serve`。

建议未来拆成：

- `@platform/http`：平台高层 HTTP API
- `@platform/net`：平台低层 socket API
- `@platform/ws`：websocket API

而现有 `http` 内建模块变成这些 capability 的语言包装层。

### 10.8 服务端能力的特殊处理

`http.server.serve`、`proxy`、`forward` 这类能力本质上属于 **监听型权限**，风险高于出站请求。

建议：

1. 与 `http_fetch` 分离授权
2. 需要显式声明监听地址白名单
3. 默认禁止绑定 `0.0.0.0`
4. 如需开放外网监听，必须宿主额外确认

### 10.9 与当前仓库的衔接

当前 `http` 模块同时覆盖 client/server/proxy/forward，功能很强，但平台化后应拆分为：

- 语言层 API
- Host ABI
- Provider 实现
- Capability 权限检查

即：

```text
builtin/http.go
   -> platform network capability layer
   -> provider/http, provider/socket
   -> Go net/http / net 等真实实现
```

---

## 11. 宿主能力注入模型

建议宿主创建实例时显式注入能力，而不是让模块直接拿到系统全局对象。

示例：

```go
runtime := iavm.NewRuntime(iavm.RuntimeOptions{
    MaxSteps: 100000,
    MaxDuration: 5 * time.Second,
    FileCapabilities: []iavm.FileCapabilityConfig{
        {
            MountPoint: "/workspace",
            RealPath:   "E:/code/issueye/project",
            Rights:     []string{"read", "write", "list", "create"},
        },
    },
    NetworkCapabilities: []iavm.NetworkCapabilityConfig{
        {
            Rights:       []string{"http_fetch"},
            AllowHosts:   []string{"api.example.com"},
            AllowPorts:   []int{443},
            AllowSchemes: []string{"https"},
        },
    },
})
```

模块实例创建时：

1. Loader 读取 capability section
2. Host 校验声明与运行时配置是否匹配
3. 生成 capability table
4. 模块通过 capability ID 访问宿主能力

---

## 12. 验证器设计

类 WASM 平台要有 verifier，首版至少做以下检查：

### 12.1 结构检查

- magic/version 是否正确
- section 顺序是否正确
- section 长度是否合法
- 常量池索引是否合法

### 12.2 代码检查

- 跳转目标是否合法
- 栈深不会出现负数
- call 参数个数匹配
- import/export 名称合法
- 未声明 capability 的 host call 不允许出现

### 12.3 安全检查

- 模块声明的 FS/Network 权限是否超出宿主允许范围
- 是否请求被禁用功能
- 是否存在不支持的 opcode

---

## 13. 与当前 ialang 字节码的迁移方案

建议采用“两阶段迁移”。

## 13.1 阶段一：平台化现有字节码

复用当前：

- AST
- 编译器主框架
- 栈机执行模型
- VM 值对象系统

新增：

- `iavm` 模块格式
- verifier
- capability section
- host ABI
- FS/Network provider 层

此阶段本质上是：

> 让当前 `ialang bytecode + vm` 变成一个真正的平台运行时。

## 13.2 阶段二：类型化与更强验证

后续可逐步加入：

- 更严格的局部变量槽类型信息
- 更规范的导入/导出签名
- 更稳定的 ABI versioning
- 更强的异步调度模型
- 可选 JIT / AOT

---

## 14. 目录结构建议

建议在当前仓库中把 `iavm` 演进为独立平台目录：

```text
iavm/
  README.md
  docs/
    architecture.md
    abi-fs.md
    abi-network.md
    module-format.md
  pkg/
    abi/
      fs.go
      network.go
      capability.go
    binary/
      module.go
      encoder.go
      decoder.go
      verifier.go
    runtime/
      vm.go
      frame.go
      stack.go
      handles.go
      resources.go
    host/
      fs/
        provider.go
        localfs.go
        memfs.go
      network/
        provider.go
        http_provider.go
        socket_provider.go
    bridge/
      ialang/
        compiler_lowering.go
        builtins_fs.go
        builtins_http.go
```

---

## 15. ialang 侧编译与运行流程

建议未来执行链路为：

```text
ialang source
  -> lexer
  -> parser
  -> AST
  -> ialang semantic compile
  -> iavm module
  -> verifier
  -> runtime instantiate
  -> bind FS/Network capabilities
  -> run
```

### 15.1 CLI 形态建议

可以扩展出两个命令：

```bash
ialang build app.ia -o app.iavm
ialang run app.iavm
```

也可以保留脚本直跑：

```bash
ialang run app.ia
```

内部路径为：

```text
.ia -> compile to in-memory iavm module -> run in iavm runtime
```

---

## 16. 示例：一个带文件与网络权限的模块

### 16.1 ialang 源码

```ts
import { fs } from "@platform/fs";
import { http } from "@platform/http";

function main() {
  let token = fs.readFile("/workspace/token.txt").trim();
  let resp = http.get("https://api.example.com/data", {
    headers: {
      Authorization: "Bearer " + token
    }
  });
  fs.writeFile("/workspace/result.json", resp.body);
}
```

### 16.2 对应 capability 声明

```json
{
  "fs": {
    "rights": ["read", "write"],
    "preopens": ["/workspace"]
  },
  "network": {
    "rights": ["http_fetch"],
    "allow_hosts": ["api.example.com"],
    "allow_ports": [443],
    "allow_schemes": ["https"]
  }
}
```

### 16.3 运行期效果

- 可以读写 `/workspace`
- 只能访问 `https://api.example.com:443`
- 无法写宿主任意路径
- 无法访问内网地址或未授权域名

---

## 17. 安全模型

## 17.1 默认拒绝

- 无 capability 不可访问 FS/Network
- 无导入声明不可调用 host ABI

## 17.2 最小权限原则

- 权限细分到操作级别
- 配置细分到目录、主机、端口、协议

## 17.3 资源配额

建议对以下资源设置上限：

- 最大步数
- 最大执行时长
- 最大内存
- 最大打开文件数
- 最大 socket 数
- 单请求最大 body
- 单模块最大并发异步任务数

## 17.4 审计能力

建议运行时记录 capability 访问日志：

- 访问时间
- 模块名
- capability id
- 操作类型
- 目标路径/目标主机
- 结果/错误码

这样后续可做调试、审计与计费。

---

## 18. 实施建议

## 18.1 第一阶段

1. 在 `iavm` 中定义模块格式与 verifier
2. 从当前 `ialang/pkg/lang/bytecode` 映射到 `iavm` 格式
3. 实现 capability table
4. 抽象 FS provider
5. 抽象 HTTP provider
6. 用桥接层替换当前 `builtin/fs.go` 和 `builtin/http.go` 的直接宿主调用

## 18.2 第二阶段

1. 拆分 `http` 与 `net` 平台 API
2. 引入 preopen VFS
3. 引入 handle table 和资源回收
4. 增强 verifier 的栈/签名检查
5. 增加 `.iavm` 文件编解码器

## 18.3 第三阶段

1. 加入多模块链接
2. 加入缓存与 AOT
3. 加入远程/边缘宿主适配
4. 支持更多 provider（MemFS、OverlayFS、MockNet）

---

## 19. 结论

基于当前仓库，最合理的路线不是推翻 `ialang` 现有 VM，而是：

1. **保留现有前端与栈机执行优势**
2. **将字节码升级为稳定模块格式**
3. **将文件系统和网络访问改造成 capability-based host ABI**
4. **在 `iavm` 中沉淀 verifier、loader、provider、resource manager**

其中，文件系统与网络访问系统的关键不是“提供更多 API”，而是：

- 将宿主实现与语言 API 解耦
- 将权限从布尔开关升级为显式能力模型
- 将资源访问变为可验证、可限制、可替换、可审计

这一步完成后，`ialang` 才会从“一个语言运行时”真正演进为“一个可承载应用的平台”。

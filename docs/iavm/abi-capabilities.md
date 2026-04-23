# IAVM Capability ABI 规范

## 1. 目标

本文档记录当前 IAVM Host Capability ABI 的稳定约定，覆盖 `iacommon/pkg/host/api` 与 `DefaultHost` 已实现的 operation。

当前 ABI 仍采用 `map[string]any` 作为参数和返回值承载格式，目的是先稳定模块、runtime、SDK 与 CLI 之间的边界。`0.0.5`/`0.0.6` 已为高频 FS / Network operation 接入第一轮 typed adapter，同时保留 `map[string]any` 作为兼容承载层。

## 2. Host 接口

Host 由 `iacommon/pkg/host/api.Host` 定义：

```go
type Host interface {
    AcquireCapability(ctx context.Context, req AcquireRequest) (CapabilityInstance, error)
    ReleaseCapability(ctx context.Context, capID string) error
    Call(ctx context.Context, req CallRequest) (CallResult, error)
    Poll(ctx context.Context, handleID uint64) (PollResult, error)
}
```

### 2.1 Capability 类型

当前支持两类 capability：

| Kind | 字符串 | 用途 |
|---|---|---|
| `CapabilityFS` | `fs` | 文件系统读写、目录、元信息操作 |
| `CapabilityNetwork` | `network` | HTTP 请求，后续可扩展 socket/listen |

### 2.2 AcquireRequest

```go
type AcquireRequest struct {
    Kind   CapabilityKind
    Config map[string]any
}
```

`DefaultHost` 当前会从 `Config["rights"]` 读取字符串数组并写入 `CapabilityInstance.Rights`，同时将 `Config` 克隆到 `CapabilityInstance.Meta`。`run-iavm --cap-config <file.toml>` 已可把外部 TOML 配置注入到 `AcquireRequest.Config`，并同时用于构建本地 FS preopen 与 Network HTTP policy。

### 2.3 CallRequest

```go
type CallRequest struct {
    CapabilityID string
    Operation    string
    Args         map[string]any
}
```

调用规则：

- `CapabilityID` 必须来自成功的 `AcquireCapability`。
- `Operation` 使用本文档定义的稳定字符串。
- `Args` 使用 snake_case key；为了兼容早期调用，少量字段也接受 camelCase。
- 错误通过 Go `error` 返回；成功结果通过 `CallResult.Value` 返回。

## 3. FS Capability

FS capability kind 为 `fs`，由 `host/fs.Provider` 执行底层操作。

### 3.1 `fs.read_file`

读取文件全部内容。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 虚拟路径或 provider 接受的路径 |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| `data` | `[]byte` | 文件内容 |

错误：

- `ErrInvalidCallArgs`: 缺少 `path` 或类型错误。
- provider 返回的读文件错误。

### 3.2 `fs.write_file`

写入文件内容。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 目标路径 |
| `data` | `[]byte` 或 `string` | 是 | 写入内容 |
| `create` | `bool` | 否 | 是否允许创建文件 |
| `trunc` | `bool` | 否 | 是否截断已有文件 |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| 无 | 空 map | 成功时返回空结果 |

错误：

- `ErrInvalidCallArgs`: 缺少参数或类型错误。
- provider 返回的写文件错误。

### 3.3 `fs.append_file`

追加文件内容。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 目标路径 |
| `data` | `[]byte` 或 `string` | 是 | 追加内容 |

返回：空 map。

### 3.4 `fs.read_dir`

读取目录条目。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 目录路径 |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| `entries` | `[]fs.DirEntry` | 目录条目，包含 `Name` 与 `IsDir` |

### 3.5 `fs.stat`

读取文件或目录元信息。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 文件或目录路径 |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| `info` | `fs.FileInfo` | 包含 `Name`、`Size`、`Mode`、`IsDir`、`ModUnix` |

### 3.6 `fs.mkdir`

创建目录。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 目录路径 |
| `recursive` | `bool` | 否 | 是否递归创建 |

返回：空 map。

### 3.7 `fs.remove`

删除文件或目录。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 目标路径 |
| `recursive` | `bool` | 否 | 是否递归删除目录 |

返回：空 map。

### 3.8 `fs.rename`

重命名或移动路径。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `old_path` 或 `oldPath` | `string` | 是 | 原路径 |
| `new_path` 或 `newPath` | `string` | 是 | 新路径 |

返回：空 map。

### 3.9 `fs.open`

打开文件并返回宿主句柄。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 目标路径 |
| `read` | `bool` | 否 | 读模式 |
| `write` | `bool` | 否 | 写模式 |
| `create` | `bool` | 否 | 允许创建 |
| `trunc` | `bool` | 否 | 截断已有文件 |
| `append` | `bool` | 否 | 追加写 |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| `handle` | `uint64` | 文件句柄 ID |

### 3.10 `fs.read`

基于句柄读取文件内容。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `handle` | number | 是 | 文件句柄 ID |
| `size` | number | 是 | 期望读取字节数 |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| `data` | `[]byte` | 读取到的数据 |
| `n` | `int64` | 实际读取字节数 |
| `eof` | `bool` | 是否到达 EOF |

### 3.11 `fs.write`

基于句柄写入文件内容。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `handle` | number | 是 | 文件句柄 ID |
| `data` | `[]byte` 或 `string` | 是 | 写入内容 |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| `n` | `int64` | 实际写入字节数 |

### 3.12 `fs.seek`

调整文件句柄偏移。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `handle` | number | 是 | 文件句柄 ID |
| `offset` | number | 是 | 偏移量 |
| `whence` | number | 否 | 对齐基准，默认 0 |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| `offset` | `int64` | 调整后的绝对偏移 |

### 3.13 `fs.close`

关闭并释放文件句柄。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `handle` | number | 是 | 文件句柄 ID |

返回：空 map。

### 3.14 `host.poll` / `OpHostPoll`

`OpHostPoll` 当前已进入 async-ready 最小稳定范围：runtime 会从栈上弹出 handle ID，调用 `Host.Poll(handleID)`，并把结果封装为 Promise 值压栈。

语义约定如下：

- 若 `PollResult.Done == true`，Promise 立即 resolve 为 poll 结果对象
- 若 `PollResult.Done == false`，`await host.poll(handle)` 会让 VM 进入 suspension
- `ResumeSuspension()` 会再次执行 `Host.Poll(handle)`；当结果变为 done 后继续解释执行
- `WaitSuspension(ctx)` 会优先调用宿主可选 `Wait(handle)` 能力，wait 完成后再恢复执行
- 若宿主未实现 `Wait(handle)`，runtime 会回退到按 `WaitInterval` 轮询 `Host.Poll(handle)`
- `RunUntilSettled(ctx)` 会统一执行 `run -> wait -> resume` 闭环，直到模块完成或 context 结束
- host capability acquire/call/poll 当前与 settled 主循环共享同一执行 context
- `Options.MaxDuration` 可为整个 settled 执行过程建立统一 deadline
- `Options.HostTimeout` 当前作用于单次 `AcquireCapability` / `Call` / `Poll`
- `Options.WaitTimeout` 当前作用于单次 `Wait(handle)`
- `Options.RetryCount` / `Options.RetryBackoff` 当前作用于 `poll/wait` 的超时重试
- `Options.RetryMultiplier` / `Options.RetryMaxBackoff` / `Options.RetryJitter` 可定义 retry 的退避曲线与抖动策略
- `Options.RetryCallOps` 可显式声明允许自动重试的 `host.call` operation allowlist
- capability `Config.host_timeout_ms` / `Config.wait_timeout_ms` 可覆盖默认 operation timeout
- capability `Config.retry_count` / `Config.retry_backoff_ms` 可覆盖默认 retry/backoff
- capability `Config.retry_multiplier` / `Config.retry_backoff_max_ms` / `Config.retry_jitter` 可覆盖默认 backoff 曲线与抖动策略
- capability `Config.retry_call_ops` 可覆盖默认 `host.call` retry allowlist
- pending promise 会保留触发时的 capability timeout profile，用于后续恢复路径
- pending promise 也会保留触发时的 retry/backoff profile，用于后续恢复路径
- 当前 wakeup 模型仍是最小实现：宿主只需保证 wait 最终返回 done 或 context 结束，不要求主动事件推送协议

当前 Promise resolve 后的 poll 结果对象包含以下字段：

| Key | 类型 | 说明 |
|---|---|---|
| `done` | `bool` | poll 是否已完成 |
| `ready` | `bool` | 资源当前可继续执行 |
| `handle` | `uint64` | 被轮询的句柄 ID |
| `error` | `string` | 错误文本；无错误时为空 |

当前 `DefaultHost.Poll` 对已打开的文件、socket、listener 句柄都返回同步 ready 结果；因此 `host.poll` 已具备统一 ABI，但 backpressure 仍停留在最小语义层，不保证真实事件驱动或公平调度。

补充约定：

- `Host.Poll(handle)` 负责“查询当前状态”
- `Host.Wait(handle)` 若实现，则负责“阻塞直到值得再次恢复”
- runtime 负责 settled 主循环；CLI `run-iavm` 当前直接复用 runtime 入口
- 当 context 结束时，runtime 当前返回标准 context 错误，如 `context deadline exceeded`
- operation timeout 与总 deadline 并存时，先到期者生效
- capability profile 优先级高于默认 timeout option
- retry/backoff 目前仅覆盖可安全重试的 `poll/wait`，不自动重试 `host.call`
- `host.call` 当前仅对显式 allowlist 中的 operation 开启超时重试；未列入 allowlist 的调用即使配置了 retry/backoff 也只执行一次
- 默认 backoff 仍为确定性退避；仅当 `RetryJitter` 或 `Config.retry_jitter` 大于 `0` 时，runtime 才会在 base backoff 周围加入有界随机抖动
- 当前 jitter 使用对称区间策略：`factor = 1 - jitter + 2 * jitter * random`，并继续受 `retry_backoff_max_ms` 上限约束
- runtime 当前会在两类错误上触发 retry：`context deadline exceeded`，以及宿主通过 `iacommon/pkg/host/api.MarkRetryable(err)` 显式标记的 retryable error

## 4. Network Capability

Network capability kind 为 `network`。当前 `DefaultHost` 已稳定暴露 HTTP 请求，并已接入第一轮 handle-based socket ABI：`network.dial/listen/accept/send/recv/close`。其中：

- client 路径已通过 `dial -> host.poll -> await -> send/recv -> close` 回归
- server 路径已通过 `listen -> host.poll -> await -> accept -> recv/send -> close` 回归
- 这些路径当前使用最小 poll/wakeup 语义，不提供独立事件循环或背压队列

### 4.1 `network.http_fetch`

发起一次 HTTP 请求。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `url` | `string` | 是 | 目标 URL，必须包含 scheme 与 host |
| `method` | `string` | 否 | HTTP 方法；空值由 provider 决定默认行为 |
| `headers` | `map[string]string` 或 `map[string]any` | 否 | 请求头，所有值必须为字符串 |
| `body` | `[]byte` 或 `string` | 否 | 请求体 |
| `timeout_ms` 或 `timeoutMS` | number | 否 | 超时时间，单位毫秒 |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| `status` | `int` | HTTP 状态码 |
| `headers` | `map[string]string` | 响应头 |
| `body` | `[]byte` | 响应体 |

错误：

- `ErrInvalidCallArgs`: 参数缺失或类型错误。
- network provider 返回的请求错误。
- policy 校验错误，例如 scheme、host、port、请求体大小不允许。

### 4.2 Network policy

`host/network.Policy` 支持以下字段：

| 字段 | 用途 |
|---|---|
| `Rights` | 预留权限集合 |
| `AllowHosts` | 允许的 hostname 列表 |
| `AllowPorts` | 允许的端口列表 |
| `AllowSchemes` | 允许的 scheme 列表，如 `http`、`https` |
| `AllowCIDRs` | 允许的 IP CIDR 列表 |
| `MaxConnections` | 预留连接数限制 |
| `MaxInflightRequest` | 预留并发请求限制 |
| `MaxBytesPerRequest` | 单请求 body 最大字节数 |

当前 `ValidateHTTPRequest` 会检查 URL、scheme、host、port 和请求体大小。

### 4.3 `network.dial`

建立 socket 连接并返回句柄。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `network` | `string` | 否 | 默认 `tcp` |
| `host` | `string` | 是 | 目标主机 |
| `port` | number | 是 | 目标端口 |
| `timeout_ms` 或 `timeoutMS` | number | 否 | 连接超时毫秒数 |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| `handle` | `uint64` | socket 句柄 ID |

### 4.4 `network.listen`

建立监听句柄。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `network` | `string` | 否 | 默认 `tcp` |
| `host` | `string` | 是 | 监听主机 |
| `port` | number | 是 | 监听端口 |
| `backlog` | number | 否 | backlog 大小 |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| `handle` | `uint64` | listener 句柄 ID |

### 4.5 `network.accept`

从 listener 句柄接收新连接。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `handle` | number | 是 | listener 句柄 ID |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| `handle` | `uint64` | 新 socket 句柄 ID |

### 4.6 `network.send`

向 socket 句柄发送数据。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `handle` | number | 是 | socket 句柄 ID |
| `data` | `[]byte` 或 `string` | 是 | 发送内容 |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| `n` | `int64` | 实际发送字节数 |

### 4.7 `network.recv`

从 socket 句柄接收数据。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `handle` | number | 是 | socket 句柄 ID |
| `size` | number | 否 | 读取上限，默认 4096 |

返回：

| Key | 类型 | 说明 |
|---|---|---|
| `data` | `[]byte` | 接收数据 |
| `n` | `int64` | 实际接收字节数 |

### 4.8 `network.close`

关闭 socket 或 listener 句柄。

参数：

| Key | 类型 | 必填 | 说明 |
|---|---|---|---|
| `handle` | number | 是 | socket 或 listener 句柄 ID |

返回：空 map。

## 5. 错误语义

当前错误通过 Go `error` 返回，调用方应按阶段处理：

| 错误 | 来源 | 说明 |
|---|---|---|
| `ErrCapabilityNotFound` | `DefaultHost.lookupCapability` | capability ID 未注册或已释放 |
| `ErrProviderUnavailable` | `DefaultHost.newCapabilityInstance` / `Call` | 对应 provider 未配置 |
| `ErrInvalidCallArgs` | 参数解析 | 缺少参数或参数类型不符合 ABI |
| `ErrPollNotSupported` | `Host.Poll` 实现 | 某些自定义宿主不支持 poll |
| `ErrCapabilityUnsupported` | capability 或 operation 分发 | capability kind 或 operation 未支持 |

runtime 可以先将这些错误包装成运行时错误；后续若接入 ialang 结构化异常，应保留原始错误分类。

补充约定：

- 宿主若希望某个错误被 runtime 视为“可安全重试”，可通过 `api.MarkRetryable(err)` 包装后返回
- runtime 可通过 `api.IsRetryableError(err)` 识别该标记，并把它与 timeout retry 规则统一处理
- `MarkRetryable` 只表达“该错误类别允许重试”，不自动绕过 `host.call` 的 allowlist 约束
- `DefaultHost` 当前会对网络 I/O 路径上的瞬时 `net.Error`（如 timeout / temporary）自动应用该标记；参数校验、policy 拒绝、capability 不存在等错误仍保持非 retryable
- 对 `network.http_fetch`，`DefaultHost` 还支持 capability 级 `retry_http_statuses` / `retryHTTPStatuses` 配置；只有命中该显式状态码列表时，HTTP 响应才会被映射为 retryable error，未配置时仍按普通响应返回

## 5.1 IAVM `OpHostCall` 参数约定

当前 IAVM runtime 对 `OpHostCall` 采用以下栈约定：

- `inst.A = 参数个数`
- 栈布局为：`[..., arg1, arg2, ..., operation]`
- runtime 先弹出 `operation`，再按 `inst.A` 弹出参数

参数编码规则：

- 当 `inst.A == 1` 且唯一参数是对象时，该对象会直接转换成 `CallRequest.Args`
- 否则 runtime 会构造：
  - `args`: 保持原顺序的参数数组
  - `arg0`, `arg1`, ...: 位置参数镜像

建议当前 ABI 调用优先使用“单对象参数”模式，这样可以直接对接 `fs.read_file`、`fs.write_file`、`network.http_fetch` 等已稳定 operation 的命名字段。

补充约定：

- `OpImportCap` 的 capability kind 常量可来自模块级常量池或函数级常量池，与 `OpConst` 的常量解析规则保持一致。
- `OpImportCap` 会把模块 `CapabilityDecl.Config` 原样传入 `AcquireRequest.Config`。
- `OpHostCall` 会绑定到最近一次成功执行 `OpImportCap` 导入的 capability。

## 5.2 `run-iavm --cap-config` TOML 约定

当前 CLI 已支持以下最小 TOML 结构：

```toml
[fs]
rights = ["read", "write"]

[[fs.preopens]]
virtual_path = "/workspace"
real_path = "C:/tmp/workspace"
read_only = false

[network]
rights = ["http"]
allow_hosts = ["example.com"]
allow_schemes = ["https"]
allow_ports = [443]
allow_cidrs = ["10.0.0.0/8"]
max_connections = 8
max_inflight_request = 8
max_bytes_per_request = 1048576
```

说明：

- `fs.preopens` 会用于构造 `host/fs.LocalFSProvider`。
- 未配置 `fs.preopens` 时，`run-iavm` 默认仍使用内存文件系统 provider。
- `network` 配置会映射到 `host/network.Policy`，用于 `network.http_fetch` 请求校验。

## 6. 版本与演进规则

短期 ABI 规则：

- operation 字符串一旦进入本文档，应视为稳定。
- 新 operation 只能追加，避免改变现有参数含义。
- 参数 key 优先使用 snake_case。
- 返回值中的结构体可先保持 Go 结构，跨进程或二进制边界稳定前再定义 JSON schema。
- `map[string]any` 是临时承载层，不是鼓励业务代码直接依赖的长期 SDK 形态。

后续建议：

1. 为每个 operation 增加 table-driven host API 测试。
2. FS / Network typed request/response adapter 已接入 `DefaultHost`，后续再把更多 operation 纳入同一模式。
3. verifier 已对静态 `OpHostCall` operation 名称执行 capability 一致性校验；动态 operation 名称仍保留运行时处理。
4. 为 `run-iavm --cap-config` 设计 TOML 配置，并映射到 `AcquireRequest.Config`。

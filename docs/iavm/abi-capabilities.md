# IAVM Capability ABI 规范

## 1. 目标

本文档记录当前 IAVM Host Capability ABI 的稳定约定，覆盖 `iacommon/pkg/host/api` 与 `DefaultHost` 已实现的 operation。

当前 ABI 仍采用 `map[string]any` 作为参数和返回值承载格式，目的是先稳定模块、runtime、SDK 与 CLI 之间的边界。后续可以在不改变 operation 名称的前提下，为高频操作增加强类型 adapter。

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

`DefaultHost` 当前会从 `Config["rights"]` 读取字符串数组并写入 `CapabilityInstance.Rights`，同时将 `Config` 克隆到 `CapabilityInstance.Meta`。具体权限检查还没有在 `DefaultHost.Call` 层强制执行，短期由 provider 或调用方策略承担。

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

### 3.9 当前未接入 FS operation

`host/fs.Provider` 已定义 `Open`、`FileHandle.Read`、`Write`、`Seek`、`Close`，但 `DefaultHost.Call` 当前尚未暴露以下 operation：

- `fs.open`
- `fs.read`
- `fs.write`
- `fs.seek`
- `fs.close`

这些操作需要先明确 handle 生命周期、资源上限和 `Poll` 语义，再进入 ABI 稳定范围。

## 4. Network Capability

Network capability kind 为 `network`，当前 `DefaultHost` 只稳定暴露 HTTP 请求。

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

### 4.3 当前未接入 Network operation

`host/network.Provider` 已定义 socket/listen 能力，但 `DefaultHost.Call` 尚未暴露：

- `network.dial`
- `network.listen`
- `network.accept`
- `network.send`
- `network.recv`
- `network.close`

这些操作依赖 handle table、异步 poll 和连接资源管理，后续应和 `OpHostPoll` 一起设计。

## 5. 错误语义

当前错误通过 Go `error` 返回，调用方应按阶段处理：

| 错误 | 来源 | 说明 |
|---|---|---|
| `ErrCapabilityNotFound` | `DefaultHost.lookupCapability` | capability ID 未注册或已释放 |
| `ErrProviderUnavailable` | `DefaultHost.newCapabilityInstance` / `Call` | 对应 provider 未配置 |
| `ErrInvalidCallArgs` | 参数解析 | 缺少参数或参数类型不符合 ABI |
| `ErrPollNotSupported` | `DefaultHost.Poll` | 当前默认宿主不支持 poll |
| `ErrCapabilityUnsupported` | capability 或 operation 分发 | capability kind 或 operation 未支持 |

runtime 可以先将这些错误包装成运行时错误；后续若接入 ialang 结构化异常，应保留原始错误分类。

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

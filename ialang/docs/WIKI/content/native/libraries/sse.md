# SSE

- 上级目录：[网络类库](network)
- 导入：`import * as sse from "sse";`

## 模块定位

`sse` 适合单向事件推送；如果你的代理需要把上游状态、发布事件或任务进度广播给前端，SSE 会比轮询更直接。

## 接口说明

- 核心入口：`server.serve`、`client.connect`
- `server.serve()` 会返回带 `send()` / `sendAsync()` / `close()` 的事件源对象。
- `client.connect()` 用于订阅事件流，返回带 `recv()` / `recvAsync()` / `close()` 的客户端对象。
- SSE 适合文本事件流；如果需要双向通信，应优先考虑 `websocket`。

## 参数要点

### `sse.server.serve([options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `addr` | `string` | 监听地址，默认 `127.0.0.1:0` |
| `path` | `string` | 事件路径，默认 `/events`，必须以 `/` 开头 |
| `headers` | `object` | 追加响应头 |

### `sse.client.connect(url, [options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `url` | `string` | SSE URL |
| `options.headers` | `object` | 请求头 |
| `options.timeoutMs` | `number` | 请求超时，默认 `15000` |

### 服务端/客户端对象方法

服务端对象：

- `send(data, [event])`
- `sendAsync(data, [event])`
- `close()`

客户端对象：

- `recv()`
- `recvAsync()`
- `close()`

## 返回值

### 服务端对象

| 字段 | 说明 |
|---|---|
| `addr` | 实际监听地址 |
| `path` | 事件路径 |
| `url` | 完整访问 URL |

`send()` / `sendAsync()` 返回成功投递到的客户端数量。

### 客户端 `recv()` 返回值

| 字段 | 类型 | 说明 |
|---|---|---|
| `event` | `string` | 事件名 |
| `data` | `string` | 事件数据 |
| `id` | `string` | 事件 ID |
| `retry` | `number` | 重连建议毫秒数 |

`close()` 返回 `true`。

## 最小示例

```javascript
import * as sse from "sse";

function main() {
  let server = sse.server.serve({
    addr: "127.0.0.1:0",
    path: "/events"
  });
  server.send("deploy", "build-finished");
  print(server.url);
  server.close();
}
```

## 代理/网关场景示例

```javascript
import * as sse from "sse";

let feed = sse.server.serve({
  addr: "127.0.0.1:0",
  path: "/events"
});

feed.send("gateway", "route-reloaded");
print(feed.url);
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


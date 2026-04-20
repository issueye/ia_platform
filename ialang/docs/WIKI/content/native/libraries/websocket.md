# WebSocket

- 上级目录：[网络类库](network)
- 导入：`import * as websocket from "websocket";`

## 模块定位

`websocket` 适合处理双向长连接场景，例如代理会话、实时通知、交互式控制面板或桥接内部服务。

## 接口说明

- 核心入口：`server.serve`、`client.connect`、连接对象 `send/recv/close`
- `client.connect/connectAsync` 用于建立连接；`server.serve/serveAsync` 用于启动 WebSocket 服务。
- 这类接口更适合消息代理和实时转发，不适合一次性请求响应。
- 当前服务端实现偏轻量，核心行为是 `welcome` 欢迎消息和 `echo` 回显。

## 参数要点

### `websocket.client.connect(url, [options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `url` | `string` | WebSocket URL |
| `options.headers` | `object` | 握手请求头 |
| `options.timeoutMs` | `number` | 握手超时，默认 `15000` |

### `websocket.server.serve([options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `addr` | `string` | 监听地址，默认 `127.0.0.1:0` |
| `path` | `string` | 路径，默认 `/`，必须以 `/` 开头 |
| `echo` | `bool` | 是否回显收到的消息，默认 `true` |
| `welcome` | `string` | 建连成功后发送的欢迎消息 |

### 连接对象方法

- `send(message)`
- `recv()`
- `close()`
- `sendAsync(message)`
- `recvAsync()`

## 返回值

### 服务对象

| 字段 | 说明 |
|---|---|
| `addr` | 实际监听地址 |
| `path` | 服务路径 |
| `url` | 完整 WebSocket URL |
| `close()` | 关闭服务并返回 `true` |

### 客户端连接对象

| 入口 | 返回值 |
|---|---|
| `send()` | `true` |
| `recv()` | 字符串消息 |
| `close()` | `true` |
| `sendAsync()/recvAsync()` | 异步任务句柄 |

## 最小示例

```javascript
import * as websocket from "websocket";

function main() {
  let s = websocket.server.serve({
    addr: "127.0.0.1:0",
    path: "/chat",
    echo: true,
    welcome: "ready"
  });
  let c = websocket.client.connect(s.url);
  print(c.recv());
  c.send("ping");
  print(c.recv());
  c.close();
  s.close();
}
```

## 代理/网关场景示例

```javascript
import { client } from "websocket";

let upstream = client.connect("ws://127.0.0.1:8080/events", {
  timeoutMs: 3000
});

upstream.send("subscribe:proxy-status");
let frame = upstream.recv();
print(frame);
upstream.close();
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


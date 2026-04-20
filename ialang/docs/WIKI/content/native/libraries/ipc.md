# IPC

- 上级目录：[网络类库](network)
- 导入：`import * as ipc from "ipc";`

## 模块定位

`ipc` 面向本机进程间通信，适合把网关主进程和本地 sidecar、控制器、管理面板拆成多个脚本。

## 接口说明

- 核心入口：`server.listen`、`client.connect`、`accept`、连接对象 `send/recv/call/reply`
- `server.listen` 和 `client.connect` 都只允许 loopback 地址，这是模块本身的安全边界。
- 连接对象常见方法是 `send/recv/call/reply/close`，`call` 会发送 request 并等待 response。
- `buildRequest` / `buildResponse` 可以在你需要自己控制消息包结构时使用。

## 参数要点

### `ipc.server.listen([options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `addr` | `string` | 默认 `127.0.0.1:0`，必须是 loopback 地址 |

### `ipc.client.connect(addr, [options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `addr` | `string` | 必填，必须是 loopback 地址 |
| `timeoutMs` | `number` | 拨号超时，默认 `5000`，必须大于 0 |

### 连接对象方法

| 方法 | 参数 | 说明 |
|---|---|---|
| `send(value)` | 任意可 JSON 编码的值 | 发送一条消息 |
| `recv()` | 无 | 读取一条消息 |
| `call(method, payload, [options])` | 方法名、载荷、可选 `id` | 发送 request 并等待 response |
| `reply(request, ok, payload, [error])` | 原请求对象、是否成功、返回载荷、可选错误信息 | 回复一次 request |
| `close()` | 无 | 关闭连接 |

### 消息构造辅助

- `buildRequest(method, payload, [options])`
- `buildResponse(requestId, ok, payload, [error])`

如果你不需要自己维护消息格式，通常直接使用 `call()` 和 `reply()` 就够了。

## 返回值

### 服务对象

| 字段 | 类型 | 说明 |
|---|---|---|
| `network` | `string` | 当前实现固定为 `tcp` |
| `addr` | `string` | 实际监听地址 |
| `accept()` | `function` | 接受一个连接并返回连接对象 |
| `acceptAsync()` | `function` | 异步接受连接 |
| `close()` | `function` | 关闭监听 |

### 连接对象

连接对象本身不暴露额外状态字段，主要通过方法工作；`recv()` 返回的是反序列化后的对象或数组，`call()` 返回的是响应对象。

## 最小示例

```javascript
import * as ipc from "ipc";
import * as json from "json";

function main() {
  let server = ipc.server.listen({ addr: "127.0.0.1:0" });
  let clientConn = ipc.client.connect(server.addr);
  let peer = server.accept();
  clientConn.send({ type: "ping" });
  let req = peer.recv();
  peer.send({ type: "pong", echo: req });
  let resp = clientConn.recv();
  print(json.stringify(resp, true));
  clientConn.close();
  peer.close();
  server.close();
}
```

## 代理/网关场景示例

```javascript
import { server, client } from "ipc";

let control = server.listen({ addr: "127.0.0.1:0" });
let sidecar = client.connect(control.addr, { timeoutMs: 1000 });
let peer = control.accept();

sidecar.send({ type: "reload-routes" });
print(peer.recv().type);
peer.send({ ok: true });
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


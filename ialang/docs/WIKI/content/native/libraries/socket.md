# Socket

- 上级目录：[网络类库](network)
- 导入：`import * as socket from "socket";`

## 模块定位

`socket` 暴露更底层的 TCP / UDP 能力，适合写协议适配器、探活器、专线桥接或非 HTTP 代理。

## 接口说明

- 核心入口：`server.listen`、`client.connect`、`udp.bind`、`udp.connect`
- `socket` 提供更底层的 TCP/UDP 能力，适合协议适配器、探活器、专线桥接或非 HTTP 代理。
- 模块同时提供同步和异步版本：`listenAsync`、`connectAsync`、`bindAsync`、`udp.connectAsync`。
- 底层 socket 接口更灵活，但需要你自己处理协议边界和连接生命周期。

## 参数要点

### TCP 服务端与客户端

| 入口 | 参数 | 说明 |
|---|---|---|
| `socket.server.listen([options])` | 可选对象 | 启动 TCP 监听 |
| `socket.client.connect(addr, [options])` | 地址 + 可选对象 | 建立 TCP 连接 |

常见 `options`：

| 字段 | 类型 | 说明 |
|---|---|---|
| `network` | `string` | TCP 支持 `tcp`、`tcp4`、`tcp6` |
| `addr` | `string` | 监听地址，默认 `127.0.0.1:0` |
| `timeoutMs` | `number` | 客户端连接超时，默认 `5000` |

### UDP

| 入口 | 参数 | 说明 |
|---|---|---|
| `socket.udp.bind([options])` | 可选对象 | 绑定 UDP 端点 |
| `socket.udp.connect(addr, [options])` | 地址 + 可选对象 | 建立 UDP 连接 |

UDP 的 `network` 支持 `udp`、`udp4`、`udp6`。

### 连接对象方法

TCP 连接对象：

- `write(data)` / `writeAsync(data)`
- `read([size])` / `readAsync([size])`
- `send(data)` / `sendAsync(data)`
- `recv([size])` / `recvAsync([size])`
- `localAddr()`
- `remoteAddr()`
- `close()`

UDP 端点对象：

- `sendTo(data, addr)` / `sendToAsync(data, addr)`
- `recvFrom([size])` / `recvFromAsync([size])`
- `localAddr()`
- `close()`

## 返回值

### TCP 服务对象

| 字段/方法 | 说明 |
|---|---|
| `network` | 网络类型 |
| `addr` | 实际监听地址 |
| `accept()` / `acceptAsync()` | 接收新连接 |
| `close()` | 关闭监听器 |

### TCP 连接对象

| 入口 | 返回值 |
|---|---|
| `write()` | 写入字节数 |
| `send()` | `true` |
| `read()/recv()` | 字符串 |
| `localAddr()/remoteAddr()` | 地址字符串 |
| `close()` | `true` |

### UDP 端点对象

`recvFrom()` 返回：

- `data`
- `addr`

其余常见返回值：

- `sendTo()` 返回写入字节数
- `close()` 返回 `true`

## 最小示例

```javascript
import * as socket from "socket";

function main() {
  let listener = socket.server.listen({
    network: "tcp",
    addr: "127.0.0.1:0"
  });
  print(listener.addr);
  listener.close();
}
```

## 代理/网关场景示例

```javascript
import * as socket from "socket";

let listener = socket.server.listen({
  network: "tcp",
  addr: "127.0.0.1:7000"
});

print(listener.addr);
let udp = socket.udp.bind({ addr: "127.0.0.1:0" });
print(udp.addr);
udp.close();
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


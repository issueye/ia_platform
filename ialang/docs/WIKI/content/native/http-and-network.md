# HTTP 与网络

这一组模块覆盖 HTTP 客户端与服务端、WebSocket、SSE、底层 socket、进程间通信和 `ialang` 自定义交互协议。下面每个库都给一个最小示例。

如果你要继续看单库级别的详细接口说明，直接进入 [网络类库目录页](libraries/network)。

## http

适合发请求、起测试服务、做代理和转发。

### 示例：GET 请求

```javascript
import { client } from "http";

let resp = client.get("https://httpbin.org/get", {
  timeoutMs: 3000,
  headers: {
    "X-Demo": "ialang"
  }
});

print(resp.statusCode);
print(resp.body);
```

### 示例：启动代理

```javascript
import { server } from "http";

let proxy = server.proxy({
  addr: "127.0.0.1:8080",
  target: "http://127.0.0.1:9000",
  requestMutations: {
    setHeaders: { "x-proxy": "ialang" }
  }
});

print(proxy.addr);
```

## express

适合快速写 REST API 和中间件式 Web 服务。

### 示例：创建一个 JSON 接口

```javascript
import { express, logger, recovery } from "express";

let app = express();
app.use(recovery());
app.use(logger());

app.get("/health", (ctx) => {
  ctx.json({ ok: true, service: "demo-api" });
});

let server = app.listen("127.0.0.1:3000");
print(server.addr);
```

## websocket

适合双向消息通信。

### 示例：连接 WebSocket 服务端

```javascript
import { client } from "websocket";

let ws = client.connect("ws://127.0.0.1:8080/chat", {
  timeoutMs: 3000
});

ws.send("hello");
let message = ws.recv();
print(message);
ws.close();
```

## sse

适合服务端单向推送。

### 示例：启动 SSE 服务并推送一条事件

```javascript
import { server } from "sse";

let s = server.serve({
  addr: "127.0.0.1:0",
  path: "/events"
});

s.send("build finished", "deploy");
print(s.url);
```

## socket

适合 TCP / UDP 级别的底层网络通信。

### 示例：启动 TCP 监听

```javascript
import { server } from "socket";

let listener = server.listen({
  network: "tcp",
  addr: "127.0.0.1:0"
});

print(listener.addr);
```

### 示例：绑定 UDP 端点

```javascript
import { udp } from "socket";

let endpoint = udp.bind({
  network: "udp",
  addr: "127.0.0.1:0"
});

print(endpoint.localAddr());
endpoint.close();
```

## ipc

适合本机进程间 request/reply 或消息收发。

### 示例：本机请求回复

```javascript
import { server, client } from "ipc";

let s = server.listen({ addr: "127.0.0.1:0" });
let c = client.connect(s.addr);
let peer = s.accept();

c.send({ type: "ping" });
let req = peer.recv();
peer.send({ type: "pong", echo: req });

print(c.recv().type);
```

## iax

适合在连接对象之上定义更稳定的业务协议。

### 示例：构造一次业务调用

```javascript
import { buildRequest, version } from "iax";

let req = buildRequest("inventory", "reserve", {
  sku: "SKU-42",
  qty: 2
});

print(version());
print(req.service);
print(req.action);
```

### 示例：和 ipc 组合使用

```javascript
import { server, client } from "ipc";
import { callAsync, receive, reply } from "iax";

let s = server.listen({ addr: "127.0.0.1:0" });
let c = client.connect(s.addr);
let peer = s.accept();

let pending = callAsync(c, "user", "profile", { id: 7 });
let req = receive(peer);
reply(peer, req, true, { name: "alice" });

let resp = await pending;
print(resp.ok);
```

## net

适合地址解析、DNS 查询和网段判断。

### 示例：解析地址和网段

```javascript
import { parseHostPort, lookupIP, containsCIDR } from "net";

let hp = parseHostPort("127.0.0.1:8080");
print(hp.host);
print(hp.port);

let ips = lookupIP("localhost");
print(ips);

print(containsCIDR("192.168.1.0/24", "192.168.1.15"));
```

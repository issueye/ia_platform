# HTTP

- 上级目录：[网络类库](network)
- 导入：`import * as http from "http";`

## 模块定位

`http` 同时覆盖 HTTP 客户端、静态响应服务、反向代理和请求转发，是写网关脚本时最核心的网络模块。

## 接口说明

- 核心入口：`client.request`、`client.get`、`client.post`、`server.serve`、`server.proxy`、`server.forward`
- `client.request/get/post` 采用 `url, [options]` 形式；常见 options 包括 `method`、`headers`、`body`、`contentType`、`timeoutMs`、`proxy`。
- `client.stream` 返回流对象，常见字段是 `ok`、`statusCode`、`headers`，以及 `recv/recvAsync/close`。
- `server.proxy` 与 `server.forward` 都要求 `options.target`，并支持 `requestMutations` / `responseMutations` 改写方法、路径、查询串、头和响应体。
- `server.serve/proxy/forward` 返回的服务对象至少包含 `addr` 和 `close()`。

## 参数要点

### `http.client.request(url, [options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `method` | `string` | 默认按调用入口决定；`request` 默认 `GET`，会自动转大写 |
| `body` | `string` | 请求体；`GET` 或空字符串时不发送正文 |
| `contentType` | `string` | 请求体存在且未显式设置头时作为 `Content-Type` |
| `headers` | `object` | 额外请求头 |
| `timeoutMs` | `number` | 超时时间，必须大于 0 |
| `chunkSize` | `number` | 仅 `client.stream` 使用，控制每次 `recv()` 读取块大小 |
| `proxy` | `string` | 客户端代理地址，必须带 scheme 和 host，例如 `http://127.0.0.1:7890` |

### `http.server.serve([options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `addr` | `string` | 默认 `127.0.0.1:0` |
| `statusCode` | `number` | 默认 `200`，范围必须在 `100-599` |
| `body` | `string` | 默认 `ok` |
| `contentType` | `string` | 默认 `text/plain; charset=utf-8` |
| `headers` | `object` | 静态响应头 |

### `http.server.proxy([options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `addr` | `string` | 监听地址，默认 `127.0.0.1:0` |
| `target` | `string` | 必填，上游目标地址，必须带 scheme 和 host |
| `stripPrefix` | `string` | 进入代理前先从原始路径去掉此前缀 |
| `preserveHost` | `bool` | 是否保留原始 `Host` |
| `headers` | `object` | 代理前静态追加/覆盖的请求头 |
| `requestMutations` | `object/function` | 请求改写器，可静态配置也可动态生成 |
| `responseMutations` | `object/function` | 响应改写器，可静态配置也可动态生成 |

### `http.server.forward([options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `addr` | `string` | 监听地址，默认 `127.0.0.1:0` |
| `target` | `string` | 必填，上游目标地址 |
| `keepPath` | `bool` | 默认 `true`，是否保留原始请求路径 |
| `path` | `string` | 显式指定转发路径；设置后优先级高于 `keepPath` |
| `preserveHost` | `bool` | 是否保留原始 `Host` |
| `headers` | `object` | 转发前静态追加/覆盖的请求头 |
| `timeoutMs` | `number` | 上游请求超时，必须大于 0 |
| `requestMutations` | `object/function` | 请求改写器 |
| `responseMutations` | `object/function` | 响应改写器 |

### `requestMutations` / `responseMutations`

请求改写对象常用字段：

- `method`
- `path`
- `appendPath`
- `setQuery`
- `removeQuery`
- `setHeaders`
- `removeHeaders`
- `body`

响应改写对象常用字段：

- `statusCode`
- `setHeaders`
- `removeHeaders`
- `body`

其中 `body` 支持直接替换整个请求体或响应体；动态改写函数则适合按实时请求内容生成改写结果。

## 返回值

### 客户端响应对象

| 字段 | 类型 | 说明 |
|---|---|---|
| `ok` | `bool` | `2xx` 时为 `true` |
| `status` | `string` | 完整状态文本，例如 `200 OK` |
| `statusCode` | `number` | 数字状态码 |
| `body` | `string` | 响应正文 |
| `headers` | `object` | 响应头 |

### 流式响应对象

除了 `ok/status/statusCode/headers` 外，还提供：

- `recv()`：读取一块数据，返回 `{ chunk, done }`
- `recvAsync()`：异步读取一块数据
- `close()`：主动关闭响应体

### 服务对象

`server.serve()`、`server.proxy()`、`server.forward()` 都会返回：

- `addr`：实际监听地址
- `close()`：关闭服务

## 最小示例

```javascript
import * as http from "http";

function main() {
  let s = http.server.serve({
    addr: "127.0.0.1:0",
    statusCode: 200,
    body: "ok"
  });
  let resp = http.client.request("http://" + s.addr, { method: "GET" });
  print(resp.statusCode);
  print(resp.body);
  s.close();
}
```

## 代理/网关场景示例

```javascript
import { server } from "http";

let upstream = server.serve({
  addr: "127.0.0.1:9000",
  body: "origin"
});

let gateway = server.proxy({
  addr: "127.0.0.1:8080",
  target: "http://" + upstream.addr,
  requestMutations: {
    setHeaders: { "x-gateway": "ialang" },
    appendPath: "/v1"
  },
  responseMutations: {
    setHeaders: { "x-proxied": "true" }
  }
});

print(gateway.addr);
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


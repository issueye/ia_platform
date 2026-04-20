# HTTP 与网络

本文汇总 `http`、`websocket`、`sse`、`socket`、`ipc`、`iax` 等网络相关模块。

## http

统一响应对象常见字段：

- `ok: bool`
- `status: string`
- `statusCode: number`
- `body: string`
- `headers: object<string, string>`

### http.client

常用函数：

- `request(url, [options])`
- `get(url, [options])`
- `post(url, [options])`
- `stream(url, [options])`
- `requestAsync(url, [options])`
- `getAsync(url, [options])`
- `postAsync(url, [options])`
- `streamAsync(url, [options])`

### http.server

常用函数：

- `serve([options])`
- `serveAsync([options])`
- `proxy([options])`
- `proxyAsync([options])`
- `forward([options])`
- `forwardAsync([options])`

`proxy` 与 `forward` 支持 `requestMutations` / `responseMutations` 做请求和响应改写。

## websocket

### client

- `connect(url, [options])`
- `connectAsync(url, [options])`

### server

- `serve([options])`
- `serveAsync([options])`

## sse

### server

- `serve([options])`
- `serveAsync([options])`

### client

- `connect(url, [options])`
- `connectAsync(url, [options])`

## ipc

用于本机进程间通信。

- `ipc.server.listen([options])`
- `ipc.client.connect(addr, [options])`
- 连接对象常见方法：`send` / `recv` / `call` / `reply`

## iax

`iax` 是 `ialang` 程序间交互协议封装，常与 `ipc` 配合使用。

常用函数：

- `iax.version()`
- `iax.buildRequest(service, action, payload, [options])`
- `iax.call(conn, service, action, payload, [options])`
- `iax.buildEvent(topic, payload, [options])`
- `iax.publish(conn, topic, payload, [options])`
- `iax.subscribe(conn, [topics], [options])`
- `iax.receive(conn, [options])`
- `iax.reply(conn, recvResult, ok, data, [error])`
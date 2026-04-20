# 网络类库

这一页覆盖和通信、服务、传输相关的原生库。每个单库页都已经内嵌完整示例代码。

## http

- 文档页：[HTTP](http)
- 导入：`import * as http from "http";`
- 常用入口：`client.request`、`client.get`、`client.post`、`server.serve`、`server.proxy`、`server.forward`

## express

- 文档页：[Express](express)
- 导入：`import { express, logger, recovery } from "express";`
- 常用入口：`express()`、`app.use(...)`、`app.get(...)`、`app.post(...)`、`app.listen(...)`

## websocket

- 文档页：[WebSocket](websocket)
- 导入：`import * as websocket from "websocket";`
- 常用入口：`server.serve`、`client.connect`、连接对象 `send/recv/close`

## sse

- 文档页：[SSE](sse)
- 导入：`import * as sse from "sse";`
- 常用入口：`server.serve`、`server.send`、`client.connect`

## socket

- 文档页：[Socket](socket)
- 导入：`import * as socket from "socket";`
- 常用入口：`server.listen`、`client.connect`、`udp.bind`

## ipc

- 文档页：[IPC](ipc)
- 导入：`import * as ipc from "ipc";`
- 常用入口：`server.listen`、`client.connect`、`accept`、连接对象 `send/recv/call/reply`

## iax

- 文档页：[IAX](iax)
- 导入：`import * as iax from "iax";`
- 别名导入：`interaction`
- 常用入口：`version`、`buildRequest`、`call/callAsync`、`receive`、`reply`

## net

- 文档页：[Net](net)
- 导入：`import * as net from "net";`
- 常用入口：`parseHostPort`、`joinHostPort`、`lookupIP`、`containsCIDR`

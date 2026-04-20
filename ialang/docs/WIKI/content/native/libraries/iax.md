# IAX

- 上级目录：[网络类库](network)
- 导入：`import * as iax from "iax";`
- 别名导入：`interaction`

## 模块定位

`iax` 在连接之上提供更稳定的消息封装，适合把“代理控制命令”“服务调用”“事件通知”统一成带 service/action 的协议。

## 接口说明

- 核心入口：`buildRequest`、`buildEvent`、`call/callAsync`、`publish/publishAsync`、`receive/reply`
- `iax` 不负责底层传输，它要求连接对象至少提供 `send()` 和 `recv()`，通常与 `ipc` 一起使用。
- 它把原始连接升级成“请求/响应 + 事件广播”的统一信封格式，适合代理、网关、控制面和服务编排脚本。
- 当前实现同时提供持久化与回放能力，适合把事件总线做成“本地落盘 + 启动恢复”的模式。

## 参数要点

### `iax.buildRequest(service, action, payload, [options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `service` | `string` | 必填，服务名，不能为空 |
| `action` | `string` | 必填，动作名，不能为空 |
| `payload` | `any` | 请求数据 |
| `options.from` | `string` | 来源标识，默认 `ialang-app` |
| `options.requestId` | `string` | 可覆盖自动生成的请求 ID |
| `options.timestampMs` | `number` | 可显式指定毫秒时间戳 |
| `options.traceId` | `string` | 链路追踪 ID；未提供时按 `from-requestId` 生成 |
| `options.routeMode` | `string` | 路由格式，支持 `dot`、`slash`、`colon`、`express` |
| `options.route` | `string` | 直接覆盖自动生成路由 |
| `options.routeTemplate` | `string` | 仅 `express` 模式使用，如 `/:service/:action` |
| `options.routePrefix` | `string` | 仅 `express` 模式使用，给路径加统一前缀 |
| `options.routeMethod` | `string` | 仅 `express` 模式使用，默认 `POST` |

### `iax.buildEvent(topic, payload, [options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `topic` | `string` | 必填，事件主题，不能为空 |
| `payload` | `any` | 事件负载 |
| `options.from` | `string` | 来源标识，默认 `ialang-app` |
| `options.eventId` | `string` | 可覆盖自动生成的事件 ID |
| `options.timestampMs` | `number` | 可显式指定毫秒时间戳 |
| `options.traceId` | `string` | 链路追踪 ID；未提供时按 `from-eventId` 生成 |

### `iax.call(conn, service, action, payload, [options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `conn` | `object` | 连接对象，至少需要 `send()` 和 `recv()` |
| `service` / `action` | `string` | RPC 目标 |
| `payload` | `any` | 请求数据 |
| `options.requestOptions` | `object` | 透传给 `buildRequest()` 的信封选项 |
| `options.callOptions` | `object` | 透传给底层 `conn.call()` 的调用选项 |
| `options.route` | `string` | 覆盖最终 IPC method |

`iax.callAsync(...)` 与 `call()` 参数一致，但返回异步任务句柄。

### `iax.publish(conn, topic, payload, [options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `conn` | `object` | 事件发送连接 |
| `topic` | `string` | 事件主题 |
| `payload` | `any` | 事件数据 |
| `options.from/eventId/timestampMs/traceId` |  | 与 `buildEvent()` 一致 |
| `options.persist` | `bool/object` | 是否为此次发布启用持久化；也可传 `{ enabled, path }` |
| `options.persistPath` | `string` | 覆盖持久化目录 |

`iax.publishAsync(...)` 与 `publish()` 参数一致，但以异步方式发送。

### `iax.configurePersistence(options)` / `iax.loadEvents([options])` / `iax.replay(conn, [options])`

| 入口 | 关键字段 | 说明 |
|---|---|---|
| `configurePersistence` | `enabled`、`path` | 配置全局持久化；`enabled=true` 时必须提供 `path` |
| `loadEvents` | `path`、`topic/topics`、`sinceMs`、`limit` | 从 LevelDB 读取事件并按时间排序 |
| `replay` | `path`、`topic/topics`、`sinceMs`、`limit` | 将已持久化事件重新推送到连接 |

### `iax.subscribe(conn, [topics], [options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `conn` | `object` | 事件接收连接 |
| `topics` | `string/array` | 可省略；默认 `["*"]`，支持 `*` 通配 |
| `options.strictProtocol` | `bool` | 默认 `true`，要求事件信封中的 `protocol` 为 `iax/1` |

### `iax.receive(conn, [options])` / `iax.reply(conn, receiveResult, ok, data, [error])`

| 入口 | 字段 | 说明 |
|---|---|---|
| `receive` | `options.requireProtocol` | 默认 `true`，要求请求信封带 `protocol=iax/1` |
| `reply` | `receiveResult` | 必须传 `iax.receive()` 的返回对象 |
| `reply` | `ok` | 是否处理成功 |
| `reply` | `data` | 响应业务数据 |
| `reply` | `error` | 当 `ok=false` 时可自定义错误消息 |

`iax.receiveAsync(...)` 与 `iax.replyAsync(...)` 分别提供异步接收和异步回复版本。

## 返回值

### 请求/事件信封

`buildRequest()` 返回对象常见字段：

- `protocol`
- `from`
- `service`
- `action`
- `routeMode`
- `route`
- `requestId`
- `traceId`
- `timestampMs`
- `payload`

`buildEvent()` 返回对象常见字段：

- `protocol`
- `type`
- `topic`
- `eventId`
- `traceId`
- `from`
- `timestampMs`
- `payload`

### `iax.call()` 返回值

| 字段 | 类型 | 说明 |
|---|---|---|
| `ok` | `bool` | 远端是否成功 |
| `code` | `string` | `OK`、`REMOTE_ERROR`、`BAD_PROTOCOL` 等 |
| `message` | `string` | 错误说明或空字符串 |
| `data` | `any` | 远端 `reply()` 里的业务数据 |
| `envelope` | `object` | 远端返回的 IAX 响应信封 |
| `response` | `object` | 底层 IPC 响应对象 |

### `iax.publish()` / `iax.replay()` / `iax.getPersistence()` 返回值

| 入口 | 返回值 |
|---|---|
| `publish()` | `{ ok, topic, event, sentAt }` |
| `replay()` | `{ ok, count }` |
| `getPersistence()` | `{ enabled, path }` |
| `loadEvents()` | 事件数组，每项是 `buildEvent()` 结构 |

### `iax.subscribe()` 返回值

订阅对象包含：

- `topics`：最终生效的主题数组
- `next()`：阻塞读取下一条匹配事件
- `nextAsync()`：异步读取下一条匹配事件
- `close()`：关闭订阅句柄

`next()` / `nextAsync()` 返回对象常见字段：

- `ok`
- `topic`
- `event`
- `message`
- `payload`
- `eventId`

### `iax.receive()` 返回值

| 字段 | 类型 | 说明 |
|---|---|---|
| `ok` | `bool` | 是否成功识别为合法 IAX 请求 |
| `code` | `string` | `OK`、`BAD_PROTOCOL`、`BAD_IPC_PAYLOAD` 等 |
| `message` | `string` | 错误说明 |
| `request` | `object` | 原始 IPC 请求对象 |
| `envelope` | `object` | 解析出的 IAX 请求信封 |
| `service` | `string` | 服务名 |
| `action` | `string` | 动作名 |
| `payload` | `any` | 业务负载 |
| `route` | `string` | 当前请求路由 |
| `routeMethod` | `string` | `express` 路由时可用 |
| `routePath` | `string` | `express` 路由时可用 |

## 最小示例

```javascript
import * as iax from "iax";

function main() {
  let req = iax.buildRequest("inventory", "reserve", {
    sku: "SKU-42",
    qty: 2
  }, {
    from: "gateway-admin",
    routeMode: "express",
    routePrefix: "/rpc",
    routeMethod: "POST"
  });

  let evt = iax.buildEvent("inventory.updated", {
    sku: "SKU-42",
    stock: 18
  });

  print(iax.version());
  print(req.service);
  print(req.route);
  print(evt.topic);
}
```

## 代理/网关场景示例

```javascript
import { server, client } from "ipc";
import * as iax from "iax";

let svc = server.listen({ addr: "127.0.0.1:0" });
let gateway = client.connect(svc.addr);
let upstream = svc.accept();

iax.configurePersistence({
  enabled: true,
  path: "./data/iax-events"
});

let pending = iax.callAsync(gateway, "proxy", "reload", {
  scope: "routes",
  source: "edge-gateway"
}, {
  requestOptions: {
    from: "gateway-control",
    routeMode: "express",
    routePrefix: "/rpc"
  }
});

let req = iax.receive(upstream);
if (!req.ok) {
  iax.reply(upstream, req, false, null, req.message);
} else {
  iax.publish(upstream, "proxy.audit", {
    service: req.service,
    action: req.action,
    traceId: req.envelope.traceId
  }, {
    persist: true
  });

  iax.reply(upstream, req, true, {
    applied: true,
    route: req.route
  });
}

let result = await pending;
print(result.ok);
print(result.data.route);
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


# Express

- 上级目录：[网络类库](network)
- 导入：`import { express, logger, recovery } from "express";`

## 模块定位

`express` 提供了更高层的 Web 应用抽象，适合把代理、管理接口、健康检查和内部控制面放到一个脚本里。

## 接口说明

- 核心入口：`express()`、`app.use(...)`、`app.get(...)`、`app.post(...)`、`app.listen(...)`
- 模块入口通常是 `express()`、`router()`、`logger()`、`recovery()`、`cors()`。
- `app` / `router` / `group` 都支持 `get/post/put/delete/patch` 这类路由方法；`app.listen` 返回带 `addr` 和 `close()` 的服务对象。
- 处理函数接收 `ctx`，常用 `ctx.param/query/body/header` 读取请求，`ctx.json/send/status/set/redirect/cookie` 组织响应。

## 参数要点

### 模块级入口

| 入口 | 参数 | 说明 |
|---|---|---|
| `express()` | 无 | 创建应用对象 |
| `router()` | 无 | 创建独立路由器 |
| `logger()` | 无 | 创建日志中间件 |
| `recovery()` | 无 | 创建异常恢复中间件 |
| `cors([options])` | `object` | 仅会读取字符串值，常用于构造 CORS 头 |

### 应用对象

| 方法 | 参数 | 说明 |
|---|---|---|
| `app.use(...middlewares)` | 中间件对象列表 | 挂载中间件 |
| `app.get/post/put/delete/patch/options/all(path, ...handlers)` | 路径 + 处理器 | 注册路由 |
| `app.group(prefix, ...middlewares)` | 前缀 + 中间件 | 创建路由组 |
| `app.static(prefix, root)` | 路径前缀 + 根目录 | 托管静态文件 |
| `app.listen(addr)` | 监听地址 | 返回服务对象 |

### `ctx` 常用方法

请求读取：

- `ctx.param(name)`
- `ctx.query(name)`
- `ctx.body()`
- `ctx.header(name)`
- `ctx.method()`
- `ctx.path()`
- `ctx.ip()`

响应输出：

- `ctx.json(data, [statusCode])`
- `ctx.send(data)`
- `ctx.status(code)`
- `ctx.set(name, value)`
- `ctx.redirect(url, [statusCode])`
- `ctx.cookie(name, value)`

控制流：

- `ctx.next()`
- `ctx.end()`

## 返回值

### `app.listen(addr)` 返回对象

| 字段 | 类型 | 说明 |
|---|---|---|
| `addr` | `string` | 实际监听地址 |
| `close()` | `function` | 关闭当前服务 |

### 中间件对象

`logger()`、`recovery()`、`cors()` 返回的是可被 `use()` 接收的中间件对象；文档中通常直接把它们当作中间件参数使用，不需要手动读取内部字段。

## 最小示例

```javascript
import { express, logger, recovery } from "express";

function main() {
  let app = express();
  app.use(recovery());
  app.use(logger());
  app["get"]("/health", function(ctx) {
    ctx.json({ ok: true });
  });
  let server = app.listen("127.0.0.1:0");
  print(server.addr);
  server.close();
}
```

## 代理/网关场景示例

```javascript
import { express, logger, recovery } from "express";

let app = express();
app.use(recovery());
app.use(logger());

app.get("/proxy/health", (ctx) => {
  ctx.set("x-gateway", "ialang-express");
  ctx.json({ ok: true, role: "edge-proxy" });
});

let admin = app.listen("127.0.0.1:3000");
print(admin.addr);
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


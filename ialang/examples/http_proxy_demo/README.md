# HTTP Proxy Demo

演示 `http.server.proxy` 的对象式改写能力（包含路由改写）：

- 请求改写：`method/path/query/header`
- 响应改写：`statusCode/header/body`

## 运行

```bash
go run ./cmd/ialang run examples/http_proxy_demo/main.ia
```

路由改写示例：

```bash
go run ./cmd/ialang run examples/http_proxy_demo/route_rewrite.ia
```

## 你会看到

- 启动一个 upstream（静态 `http.server.serve`）服务
- 启动一个 proxy 服务
- 发起一次客户端请求到 proxy
- proxy 先按 `requestMutations` 改写请求后转发到 upstream
- 然后按 `responseMutations` 改写状态码、响应头与响应 body

输出里重点关注：

- `statusCode`（应为 `202`）
- 响应头 `X-Proxy-Mode: object`
- 响应 body：`response rewritten by proxy`

`route_rewrite.ia` 输出里重点关注：

- `statusCode`（应为 `202`）
- 响应头 `X-Route-Rewrite: enabled`
- 响应 body 中 `hit: "healthz"`（证明路由已从 `/api/original` 改写到 `/upstream/healthz`）

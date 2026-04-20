# @agent/sdk

- 上级目录：[工具类库](tools)
- 导入：`import { llm, tool, memory } from "@agent/sdk";`

## 模块定位

`@agent/sdk` 适合把“模型调用、工具调度、记忆读取”封装到同一层里，常见于意图路由、代理编排和带上下文的请求分发。

当前 `ialang` runtime 里的 `@agent/sdk` 是 **mock 实现**：它能验证脚本调用流程、异步等待和错误处理，但不会真的连接外部 LLM、工具平台或持久化记忆服务。

## 接口说明

- 核心入口：`llm.chat`、`llm.chatAsync`、`tool.call`、`memory.get`
- `llm.chat(prompt)` 立即返回字符串，格式是 `"[mock-llm] " + prompt`
- `llm.chatAsync(prompt)` 返回异步任务，`await` 后得到 `"[mock-llm-async] " + prompt`
- `tool.call(name, ...args)` 至少要求 1 个参数，当前返回 `"[mock-tool] called " + name`
- `memory.get(key)` 返回 `"[mock-memory] " + key`

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `llm.chat(prompt)` | `string` | 同步 prompt 文本 |
| `llm.chatAsync(prompt)` | `string` | 异步 prompt 文本，返回任务句柄 |
| `tool.call(name, ...args)` | `string`, 可选附加参数 | 当前实现只要求首个参数存在，附加参数可用于预留调用位 |
| `memory.get(key)` | `string` | 读取记忆项的键 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `llm.chat()` | 字符串，前缀为 `[mock-llm] ` |
| `llm.chatAsync()` | 异步任务句柄；`await` 后得到字符串 |
| `tool.call()` | 字符串，前缀为 `[mock-tool] called ` |
| `memory.get()` | 字符串，前缀为 `[mock-memory] ` |

## 最小示例

```javascript
import { llm, tool, memory } from "@agent/sdk";

async function main() {
  let summary = llm.chat("summarize: release pipeline");
  print(summary);

  let asyncReply = await llm.chatAsync("classify: /api/orders");
  print(asyncReply);

  let toolResult = tool.call("gateway.route", "/api/orders", "GET");
  print(toolResult);

  let cached = memory.get("last_route");
  print(cached);
}
```

## 代理/网关场景示例

```javascript
import { llm, tool, memory } from "@agent/sdk";
import * as json from "json";

async function main() {
  let inbound = {
    path: "/gateway/payments/refund",
    method: "POST",
    body: "{\"refund_id\":\"rf_1001\"}"
  };

  let policy = await llm.chatAsync(
    "route this request: " + json.stringify(inbound)
  );
  print(policy);

  let routeResult = tool.call("proxy.dispatch", inbound.path, inbound.method);
  print(routeResult);

  let lastDecision = memory.get("gateway:last-decision");
  print(lastDecision);
}
```

## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。

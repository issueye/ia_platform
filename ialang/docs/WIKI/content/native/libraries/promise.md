# Promise

- 上级目录：[工具类库](tools)
- 导入：`import { all, race, allSettled } from "Promise";`

## 模块定位

`Promise` 负责并发编排，适合在代理里并行拉多个上游、同时执行检查或聚合多个异步结果。

## 接口说明

- 核心入口：`all`、`race`、`allSettled`
- `Promise` 负责并发编排，适合并行拉多个上游、同时执行检查或聚合多个异步结果。
- 传入数组里既可以放 promise/awaitable，也可以放普通值；普通值会被当作已完成结果处理。
- 和 `http.*Async`、`ipc.*Async`、`timer.sleepAsync` 组合时最常见。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `Promise.all(items)` | 数组 | 全部成功后返回结果数组 |
| `Promise.race(items)` | 数组 | 返回最先完成的结果 |
| `Promise.allSettled(items)` | 数组 | 等所有任务结束后返回结果数组 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `all()` | Promise/Awaitable |
| `race()` | Promise/Awaitable |
| `allSettled()` | Promise/Awaitable |

使用时通常配合 `await`：

- `await Promise.all([...])`
- `await Promise.race([...])`
- `await Promise.allSettled([...])`

## 最小示例

```javascript
import { all, race, allSettled } from "Promise";

async function task1() {
  return "a";
}

async function task2() {
  return "b";
}

async function main() {
  print(await all([task1(), task2()]));
  print(await race([task1(), task2()]));
  print(await allSettled([task1(), task2()]));
  print(await all(["static", task1()]));
}
```

## 代理/网关场景示例

```javascript
import { all, race } from "Promise";
import * as timer from "timer";

async function primary() {
  await timer.sleepAsync(20);
  return "primary";
}

async function backup() {
  await timer.sleepAsync(5);
  return "backup";
}

print(await all([primary(), backup()]));
print(await race([primary(), backup()]));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


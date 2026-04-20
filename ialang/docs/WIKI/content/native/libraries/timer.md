# Timer

- 上级目录：[系统类库](system)
- 导入：`import * as timer from "timer";`
- 兼容 plain 名：`setTimeout`

## 模块定位

`timer` 负责延时、周期任务和 cron 调度，适合做路由热更新、上游探活、证书轮换和定时清理。

## 接口说明

- 核心入口：`setTimeout`、`setInterval`、`sleep/sleepAsync`、`cron`、`removeJob`
- `timer` 负责调度，`time` 负责读时间；前者是执行控制，后者是时间数据。
- 所有回调都不接收参数，定时 ID 由运行时分配，适合做热更新、探活、清理和轮询。
- 如果你需要可取消的周期任务，优先用 `setInterval()`；`every()` 更像“永不停止的异步循环”。

## 参数要点

### `timer.setTimeout(callback, delayMs)` / `timer.setInterval(callback, intervalMs)`

| 字段 | 类型 | 说明 |
|---|---|---|
| `callback` | `function` | 必填，无参回调 |
| `delayMs` / `intervalMs` | `number` | 毫秒数 |

### `timer.clearTimeout(id)` / `timer.clearInterval(id)`

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `number` | `setTimeout()` / `setInterval()` 返回的任务 ID |

### `timer.sleep(ms)` / `timer.sleepAsync(ms)`

| 字段 | 类型 | 说明 |
|---|---|---|
| `ms` | `number` | 等待毫秒数 |

`sleep()` 会阻塞当前执行流，`sleepAsync()` 返回异步任务句柄，更适合在脚本主流程里 `await`。

### `timer.defer(callback, delayMs)` / `timer.every(callback, intervalMs)`

| 入口 | 说明 |
|---|---|
| `defer()` | 延迟后执行一次，并返回异步任务句柄 |
| `every()` | 每隔一段时间永久执行一次回调；当前接口不返回可取消 ID |

### `timer.cron(expression, callback)` / `timer.removeJob(id)`

| 字段 | 类型 | 说明 |
|---|---|---|
| `expression` | `string` | Cron 表达式；当前实现使用标准 5 段格式：`分 时 日 月 周` |
| `callback` | `function` | 到点执行的回调 |
| `id` | `number` | `cron()` 返回的任务 ID，用于 `removeJob()` |

## 返回值

| 入口 | 返回值 |
|---|---|
| `setTimeout()` | 定时器 ID |
| `setInterval()` | 周期任务 ID |
| `clearTimeout()` | `true` |
| `clearInterval()` | `true` |
| `sleep()` | `true` |
| `sleepAsync()` | 异步任务句柄 |
| `defer()` | 异步任务句柄 |
| `every()` | 异步任务句柄 |
| `cron()` | Cron 任务 ID |
| `removeJob()` | `true` |

## 最小示例

```javascript
import * as timer from "timer";

function main() {
  let timeoutId = timer.setTimeout(() => {
    print("timeout fired");
  }, 50);

  let intervalId = timer.setInterval(() => {
    print("interval tick");
  }, 20);

  timer.sleep(80);
  timer.clearInterval(intervalId);
  timer.clearTimeout(timeoutId);
}
```

## 代理/网关场景示例

```javascript
import * as timer from "timer";

let healthCheckId = timer.setInterval(() => {
  print("check upstream health");
}, 5000);

let reloadJob = timer.cron("*/1 * * * *", () => {
  print("reload route table");
});

await timer.sleepAsync(12000);
timer.clearInterval(healthCheckId);
timer.removeJob(reloadJob);
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


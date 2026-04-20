# Pool

- 上级目录：[系统类库](system)
- 导入：`import * as pool from "pool";`

## 模块定位

`pool` 提供协程池 / 任务池抽象，适合限制后台探活、批量预热、审计写入等并发任务的上限。

## 接口说明

- 核心入口：`submit`、`submitWithRetry`、`getStats`、`createPool`、`shutdown`
- `pool` 同时提供“默认全局池”和“自建池对象”两套用法，适合约束后台工作负载。
- 这类池更适合配置刷新、批量预热、异步审计写入，而不是替代请求主链路。
- 返回的任务对象可继续交给异步编排逻辑处理。

## 参数要点

### 模块级函数

| 入口 | 参数 | 说明 |
|---|---|---|
| `pool.submit(callback)` | 无参函数 | 提交到默认池执行 |
| `pool.submitWithRetry(callback, [maxRetries])` | 函数 + 可选重试次数 | 提交到默认池并失败重试 |
| `pool.getStats()` | 无 | 读取全局池统计信息 |
| `pool.createPool([options])` | 可选对象 | 创建独立池 |
| `pool.shutdown([timeoutMs])` | 可选毫秒数 | 关闭默认池 |

`createPool()` 的 `options` 常见字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `minWorkers` | `number` | 最小 worker 数 |
| `maxWorkers` | `number` | 最大 worker 数 |
| `queueSize` | `number` | 队列长度 |
| `maxRetries` | `number` | 最大重试次数 |

### 池对象方法

| 方法 | 参数 | 说明 |
|---|---|---|
| `submit(callback)` | 无参函数 | 提交任务到该池 |
| `getStats()` | 无 | 获取该池统计信息 |
| `shutdown([timeoutMs])` | 可选毫秒数 | 关闭该池 |

## 返回值

### 模块级 `getStats()`

返回对象字段：

- `totalSubmitted`
- `totalCompleted`
- `totalFailed`
- `totalRejected`
- `totalPools`
- `activePools`
- `totalWorkers`
- `activeWorkers`
- `queuedTasks`

### 池对象 `getStats()`

返回对象字段：

- `activeWorkers`
- `idleWorkers`
- `totalWorkers`
- `queuedTasks`
- `completedTasks`
- `failedTasks`
- `rejectedTasks`
- `maxConcurrency`
- `currentLoad`

### 其他入口

| 入口 | 返回值 |
|---|---|
| `submit()` | 任务对象 |
| `submitWithRetry()` | 任务对象 |
| `createPool()` | 池对象 |
| `shutdown()` | `bool` |

## 最小示例

```javascript
import * as pool from "pool";

function main() {
  let p = pool.createPool({
    minWorkers: 1,
    maxWorkers: 2,
    queueSize: 8
  });
  let task = p.submit(() => {
    return "warmup-done";
  });
  let stats = p.getStats();
  print(stats.totalWorkers);
  print(stats.queuedTasks);
  print(task != null);
  p.shutdown(2000);
}
```

## 代理/网关场景示例

```javascript
import * as pool from "pool";

let p = pool.createPool({
  minWorkers: 1,
  maxWorkers: 4,
  queueSize: 32
});

print(p.getStats().totalWorkers);
let globalTask = pool.submit(() => {
  return "refresh-routes";
});
print(globalTask != null);
p.shutdown(1000);
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


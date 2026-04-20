# Time

- 上级目录：[系统类库](system)
- 导入：`import * as time from "time";`

## 模块定位

`time` 提供时间戳和时间字符串工具，适合给代理请求打时间标签、生成审计记录或写缓存过期时间。

## 接口说明

- 核心入口：`nowUnix`、`nowUnixMilli`、`nowISO`、`parseISO`、`sleep`、`sleepAsync`
- `time` 同时提供“读取当前时间”“解析 ISO 时间”“简单等待”三类能力，但不负责定时任务编排。
- 如果你需要周期调度或 Cron，应转向 `timer`。
- 和 `log`、`uuid` 组合时，常见做法是手动附加时间字段。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `time.nowUnix()` | 无 | 返回当前 Unix 秒时间戳 |
| `time.nowUnixMilli()` | 无 | 返回当前 Unix 毫秒时间戳 |
| `time.nowISO()` | 无 | 返回当前 RFC3339Nano 字符串 |
| `time.parseISO(text)` | ISO 时间字符串 | 解析为 Unix 毫秒时间戳 |
| `time.sleep(ms)` | 非负整数 | 阻塞等待指定毫秒数 |
| `time.sleepAsync(ms)` | 非负整数 | 异步等待指定毫秒数 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `nowUnix()` | `number` |
| `nowUnixMilli()` | `number` |
| `nowISO()` | 字符串 |
| `parseISO()` | `number` |
| `sleep()` | `true` |
| `sleepAsync()` | 异步任务句柄 |

## 最小示例

```javascript
import * as time from "time";

function main() {
  print(time.nowUnix());
  print(time.nowUnixMilli());
  print(time.nowISO());
  print(time.parseISO("2026-04-17T11:00:00Z"));
}
```

## 代理/网关场景示例

```javascript
import * as time from "time";

print(time.nowUnix());
print(time.nowUnixMilli());
print(time.nowISO());
await time.sleepAsync(10);
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


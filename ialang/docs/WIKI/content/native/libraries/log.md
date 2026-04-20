# Log

- 上级目录：[工具类库](tools)
- 导入：`import * as log from "log";`

## 模块定位

`log` 适合输出结构化访问日志、错误日志和审计事件，是代理脚本里最常用的观测模块之一。

## 接口说明

- 核心入口：`debug`、`info`、`warn`、`error`、`log`、`with`、`setLevel`、`setJSON`、`setOutputPath`
- `log` 基于结构化字段输出，默认级别是 `info`，默认输出到 `stdout`。
- 如果网关要接入外部日志系统，优先切到 JSON 输出并统一字段名。
- `with()` 可以生成带固定上下文字段的子 logger，适合请求级或模块级日志。

## 参数要点

### 日志输出

| 入口 | 参数 | 说明 |
|---|---|---|
| `log.debug(message, [fields])` | 字符串 + 可选对象 | 输出 debug 日志 |
| `log.info(message, [fields])` | 字符串 + 可选对象 | 输出 info 日志 |
| `log.warn(message, [fields])` | 字符串 + 可选对象 | 输出 warn 日志 |
| `log.error(message, [fields])` | 字符串 + 可选对象 | 输出 error 日志 |
| `log.log(level, message, [fields])` | 级别 + 字符串 + 可选对象 | 自定义级别输出 |

### 配置与上下文

| 入口 | 参数 | 说明 |
|---|---|---|
| `log.with(fields)` | 字段对象 | 生成带固定字段的子 logger |
| `log.setLevel(level)` | 字符串或数字 | 支持 `debug/info/warn/error` |
| `log.getLevel()` | 无 | 获取当前级别 |
| `log.setJSON(enabled)` | `bool` | 是否切到 JSON 输出 |
| `log.isJSON()` | 无 | 当前是否为 JSON 输出 |
| `log.setOutputPath(path)` | 字符串 | 设置输出文件路径；`stdout` 或空串表示标准输出 |
| `log.getOutputPath()` | 无 | 获取当前输出路径 |

`fields` 必须是对象；对象和数组会递归转成结构化值。

## 返回值

| 入口 | 返回值 |
|---|---|
| `debug/info/warn/error/log()` | `true` |
| `with()` | 新的 logger 对象 |
| `setLevel()` | 当前级别字符串 |
| `getLevel()` | 当前级别字符串 |
| `setJSON()` | `bool` |
| `isJSON()` | `bool` |
| `setOutputPath()` | 实际设置的路径字符串 |
| `getOutputPath()` | 路径字符串，未设置时为 `stdout` |

## 最小示例

```javascript
import * as log from "log";

function main() {
  log.setJSON(true);
  log.setLevel("info");
  log.info("job started", {
    service: "wiki-demo",
    batch: 3
  });
}
```

## 代理/网关场景示例

```javascript
import * as log from "log";
import * as uuid from "uuid";

log.setJSON(true);
let requestLog = log.with({
  traceId: uuid.v4(),
  component: "gateway"
});

requestLog.info("proxy request", {
  route: "/orders",
  upstream: "billing",
  status: 200
});
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


# UUID

- 上级目录：[数据类库](data)
- 导入：`import * as uuid from "uuid";`

## 模块定位

`uuid` 适合生成请求 ID、链路 ID 和任务 ID，是网关观测和审计里很常见的工具。

## 接口说明

- 核心入口：`v4`、`isValid`
- 当前实现专注于 UUID v4：生成的是随机 v4，校验也是按 v4 格式检查。
- 适合生成 `traceId`、`x-request-id`、任务 ID 和审计事件 ID。
- 常与 `log`、`time` 一起出现在访问日志链路里。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `uuid.v4()` | 无 | 生成随机 UUID v4 |
| `uuid.isValid(text)` | 字符串 | 检查字符串是否为合法 UUID v4 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `v4()` | UUID 字符串 |
| `isValid()` | `bool` |

## 最小示例

```javascript
import * as uuid from "uuid";

function main() {
  let id = uuid.v4();
  print(id);
  print(uuid.isValid(id));
}
```

## 代理/网关场景示例

```javascript
import * as uuid from "uuid";

let traceId = uuid.v4();
print(traceId);
print(uuid.isValid(traceId));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


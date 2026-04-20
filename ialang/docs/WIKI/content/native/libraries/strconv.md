# Strconv

- 上级目录：[工具类库](tools)
- 导入：`import * as strconv from "strconv";`

## 模块定位

`strconv` 负责字符串与数字、布尔值之间的显式转换，适合处理环境变量、查询参数和配置项。

## 接口说明

- 核心入口：`atoi`、`itoa`、`parseFloat`、`formatFloat`、`parseBool`、`formatBool`
- `strconv` 负责显式类型转换，比隐式转换更适合处理外部输入。
- 代理脚本里常见于 `timeoutMs`、`port`、`sampleRate`、`enabled` 这类配置解析。
- 和 `os`、`process`、`url` 组合时最常见。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `strconv.atoi(text)` | 字符串 | 解析十进制整数 |
| `strconv.itoa(number)` | 整数 | 把整数转成字符串 |
| `strconv.parseFloat(text)` | 字符串 | 解析浮点数 |
| `strconv.formatFloat(number, [precision])` | 浮点数 + 可选精度 | 按十进制文本输出 |
| `strconv.parseBool(text)` | 字符串 | 解析布尔值 |
| `strconv.formatBool(value)` | `bool` | 把布尔值转成 `true/false` |

## 返回值

| 入口 | 返回值 |
|---|---|
| `atoi()` | `number` |
| `itoa()` | 字符串 |
| `parseFloat()` | `number` |
| `formatFloat()` | 字符串 |
| `parseBool()` | `bool` |
| `formatBool()` | 字符串 |

## 最小示例

```javascript
import * as strconv from "strconv";

function main() {
  print(strconv.atoi("42"));
  print(strconv.formatFloat(3.14159, 2));
  print(strconv.parseBool("true"));
  print(strconv.formatBool(false));
}
```

## 代理/网关场景示例

```javascript
import * as strconv from "strconv";

print(strconv.atoi("8080"));
print(strconv.parseBool("true"));
print(strconv.formatFloat(0.125, 3));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


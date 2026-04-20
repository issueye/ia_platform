# String

- 上级目录：[工具类库](tools)
- 导入：`import * as strmod from "string";`

## 模块定位

`string` 适合做请求路径、头、路由键和日志文本的清洗，是代理脚本里高频出现的基础工具。

## 接口说明

- 核心入口：`split`、`join`、`parseInt`、`parseFloat`、`fromCodePoint`、`trim`、`replace`、`toLowerCase`、`toUpperCase`
- `string` 是最基础的文本处理模块，适合做路由键拆分、大小写归一化、前后空白清理和标识拼接。
- 对外部输入做文本规整时，通常会和 `regexp`、`strconv`、`url` 配合。
- 当前实现里这些能力以模块函数形式提供，适合在脚本里明确写出处理步骤。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `string.split(text, sep)` | 字符串 + 分隔符 | 按分隔符拆分为数组 |
| `string.join(arr, [sep])` | 数组 + 可选分隔符 | 把数组元素拼接成字符串 |
| `string.parseInt(text)` | 字符串 | 解析成整数；失败时返回 `0` |
| `string.parseFloat(text)` | 字符串 | 解析成浮点数；失败时返回 `0` |
| `string.fromCodePoint(code)` | 数字 | 把 Unicode code point 转成单字符字符串 |
| `string.trim(text)` | 字符串 | 去掉首尾空白 |
| `string.replace(text, old, next)` | 3 个字符串 | 全量替换子串 |
| `string.toLowerCase(text)` | 字符串 | 转小写 |
| `string.toUpperCase(text)` | 字符串 | 转大写 |
| `string.startsWith(text, prefix)` | 字符串 + 前缀 | 是否以前缀开头 |
| `string.endsWith(text, suffix)` | 字符串 + 后缀 | 是否以后缀结尾 |
| `string.contains(text, substr)` | 字符串 + 子串 | 是否包含子串 |
| `string.indexOf(text, substr)` | 字符串 + 子串 | 返回子串位置，未命中为 `-1` |
| `string.length(text)` | 字符串 | 返回字节长度 |
| `string.repeat(text, count)` | 字符串 + 数字 | 重复拼接若干次 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `split()` | 字符串数组 |
| `join()` | 字符串 |
| `parseInt()` | `number` |
| `parseFloat()` | `number` |
| `fromCodePoint()` | 字符串 |
| `trim()` | 字符串 |
| `replace()` | 字符串 |
| `toLowerCase()` | 字符串 |
| `toUpperCase()` | 字符串 |
| `startsWith()` | `bool` |
| `endsWith()` | `bool` |
| `contains()` | `bool` |
| `indexOf()` | `number` |
| `length()` | `number` |
| `repeat()` | 字符串 |

## 最小示例

```javascript
import * as strmod from "string";

function main() {
  let text = "  hello,ialang  ";
  print(strmod.toUpperCase(strmod.trim(text)));
  print(strmod.split("a,b,c", ","));
  print(strmod.join(["gw", "orders"], ":"));
}
```

## 代理/网关场景示例

```javascript
import * as strmod from "string";

let route = "/api/v1/orders";
let normalized = strmod.toLowerCase(strmod.trim(route));
print(normalized);
print(strmod.split(normalized, "/"));
print(strmod.startsWith(normalized, "/api/"));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


# Bytes

- 上级目录：[数据类库](data)
- 导入：`import * as bytes from "bytes";`

## 模块定位

`bytes` 适合在文本和二进制表示之间切换，常用于网关缓存、签名和内容传输。

## 接口说明

- 核心入口：`fromString`、`toString`、`fromBase64`、`toBase64`、`concat`、`slice`、`length`
- `bytes` 的“字节串”在 ialang 里表现为 `array<number>`，每一项范围是 `0-255`。
- 适合在文本、Base64 和原始字节数组之间切换，再交给 `hash`、`hex`、`encoding` 等模块处理。
- 纯文本 JSON 场景通常不需要它；涉及原始负载、分片、签名时更适合引入。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `bytes.fromString(text)` | 字符串 | 把字符串转成字节数组 |
| `bytes.toString(buf)` | `array<number>` | 把字节数组还原成字符串 |
| `bytes.fromBase64(text)` | Base64 字符串 | Base64 解码为字节数组 |
| `bytes.toBase64(buf)` | `array<number>` | 字节数组编码为 Base64 |
| `bytes.concat(...bufs)` | 多个字节数组 | 顺序拼接多个字节数组 |
| `bytes.slice(buf, start, [end])` | 字节数组 + 起止索引 | 截取子数组；越界会自动收敛 |
| `bytes.length(buf)` | 字节数组 | 返回长度 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `fromString()` | `array<number>` |
| `toString()` | 字符串 |
| `fromBase64()` | `array<number>` |
| `toBase64()` | Base64 字符串 |
| `concat()` | `array<number>` |
| `slice()` | `array<number>` |
| `length()` | `number` |

## 最小示例

```javascript
import * as bytes from "bytes";

function main() {
  let data = bytes.fromString("hello");
  print(bytes.toBase64(data));
  print(bytes.toString(data));
  print(bytes.length(data));
}
```

## 代理/网关场景示例

```javascript
import * as bytes from "bytes";

let header = bytes.fromString("gw:");
let token = bytes.fromString("secret-token");
let raw = bytes.concat(header, token);

print(bytes.toBase64(raw));
print(bytes.toString(bytes.slice(raw, 0, 3)));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


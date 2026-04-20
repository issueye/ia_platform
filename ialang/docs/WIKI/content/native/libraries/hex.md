# Hex

- 上级目录：[数据类库](data)
- 导入：`import * as hex from "hex";`

## 模块定位

`hex` 适合把摘要、二进制片段和调试信息转成稳定可读的十六进制字符串。

## 接口说明

- 核心入口：`encode`、`decode`、`encodeBytes`、`decodeBytes`
- `hex` 同时支持“字符串 <-> 十六进制文本”和“字节数组 <-> 十六进制文本”两条链路。
- 常见于签名值展示、追踪标识、调试日志和二进制协议桥接。
- 和 `hash`、`bytes` 搭配最常见。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `hex.encode(text)` | 字符串 | 把普通文本编码为十六进制字符串 |
| `hex.decode(text)` | 十六进制字符串 | 把十六进制字符串解码回普通文本 |
| `hex.encodeBytes(buf)` | `array<number>` | 把字节数组编码为十六进制字符串 |
| `hex.decodeBytes(text)` | 十六进制字符串 | 把十六进制字符串解码成字节数组 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `encode()` | 十六进制字符串 |
| `decode()` | 普通字符串 |
| `encodeBytes()` | 十六进制字符串 |
| `decodeBytes()` | `array<number>` |

## 最小示例

```javascript
import * as hex from "hex";

function main() {
  let encoded = hex.encode("ialang");
  print(encoded);
  print(hex.decode(encoded));
  print(hex.encodeBytes([1, 2, 255]));
}
```

## 代理/网关场景示例

```javascript
import * as hex from "hex";

let trace = hex.encode("gw-001");
print(trace);
print(hex.decode(trace));
print(hex.decodeBytes("0102ff"));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


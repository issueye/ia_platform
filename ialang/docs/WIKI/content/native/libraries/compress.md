# Compress

- 上级目录：[数据类库](data)
- 导入：`import * as compress from "compress";`

## 模块定位

`compress` 适合处理压缩和解压场景，例如缓存压缩、响应归档和离线传输。

## 接口说明

- 核心入口：`gzipCompress`、`gzipDecompress`、`zlibCompress`、`zlibDecompress`
- 当前实现的压缩结果不是原始二进制，而是“压缩后二进制再做 Base64 编码”的字符串。
- 这类工具更适合缓存、归档、状态快照和链路外存储，不适合直接拿来充当 HTTP 响应体，除非你再自行处理编码头。
- 和 `encoding`、`json`、`fs` 一起使用时很常见。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `compress.gzipCompress(text)` | 字符串 | 进行 gzip 压缩并返回 Base64 文本 |
| `compress.gzipDecompress(text)` | Base64 字符串 | 先 Base64 解码，再做 gzip 解压 |
| `compress.zlibCompress(text)` | 字符串 | 进行 zlib 压缩并返回 Base64 文本 |
| `compress.zlibDecompress(text)` | Base64 字符串 | 先 Base64 解码，再做 zlib 解压 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `gzipCompress()` | Base64 字符串 |
| `gzipDecompress()` | 解压后的字符串 |
| `zlibCompress()` | Base64 字符串 |
| `zlibDecompress()` | 解压后的字符串 |

## 最小示例

```javascript
import * as compress from "compress";

function main() {
  let gz = compress.gzipCompress("hello");
  print(gz);
  print(compress.gzipDecompress(gz));
}
```

## 代理/网关场景示例

```javascript
import * as compress from "compress";
import * as json from "json";

let snapshot = json.stringify({
  service: "gateway",
  routes: ["/orders", "/invoice"],
  updatedAt: 1713326400000
}, true);

let zipped = compress.gzipCompress(snapshot);
print(zipped);
print(compress.gzipDecompress(zipped));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


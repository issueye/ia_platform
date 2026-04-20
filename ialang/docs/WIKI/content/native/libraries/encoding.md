# Encoding

- 上级目录：[数据类库](data)
- 导入：`import * as encoding from "encoding";`

## 模块定位

`encoding` 适合处理 Base64、URL 编码等场景，常见于代理签名、头转发和令牌处理。

## 接口说明

- 核心入口：`base64Encode`、`base64Decode`、`urlEncode`、`urlDecode`
- 这是轻量编码工具模块，适合处理认证头、查询串、回调地址和跨系统文本传输。
- 当前实现全部以字符串为输入和输出，不提供二进制缓冲区对象。
- 与 `url`、`hash`、`http` 组合时，能覆盖常见网关编码需求。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `encoding.base64Encode(text)` | 字符串 | 将文本编码为标准 Base64 |
| `encoding.base64Decode(text)` | Base64 字符串 | 解码回普通文本 |
| `encoding.urlEncode(text)` | 字符串 | 按查询串规则进行 URL 编码 |
| `encoding.urlDecode(text)` | 已编码字符串 | 进行 URL 解码 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `base64Encode()` | Base64 字符串 |
| `base64Decode()` | 解码后的字符串 |
| `urlEncode()` | URL 编码后的字符串 |
| `urlDecode()` | URL 解码后的字符串 |

## 最小示例

```javascript
import * as encoding from "encoding";

function main() {
  let b64 = encoding.base64Encode("hello");
  print(b64);
  print(encoding.base64Decode(b64));
  print(encoding.urlEncode("a b+c"));
}
```

## 代理/网关场景示例

```javascript
import * as encoding from "encoding";

let basicAuth = "Basic " + encoding.base64Encode("gateway:secret");
let callback = "https://gw.internal/callback?next=" + encoding.urlEncode("/api/orders?id=1");

print(basicAuth);
print(callback);
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


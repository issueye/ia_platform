# Hash

- 上级目录：[数据类库](data)
- 导入：`import * as hash from "hash";`

## 模块定位

`hash` 提供更多摘要算法选择，适合做内容校验、缓存键和兼容旧系统的摘要输出。

## 接口说明

- 核心入口：`sha1`、`sha256`、`sha512`、`crc32`、`fnv32a`、`fnv64a`
- 适合生成内容摘要、缓存键、路由签名和兼容旧系统要求的校验值。
- `sha*` 返回十六进制字符串，`crc32/fnv32a/fnv64a` 返回数字。
- 如果你要做安全签名或 HMAC，应优先使用更明确的密码学模块；这里更偏“摘要与散列工具”。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `hash.sha1(text)` | 字符串 | 生成 SHA-1 十六进制摘要 |
| `hash.sha256(text)` | 字符串 | 生成 SHA-256 十六进制摘要 |
| `hash.sha512(text)` | 字符串 | 生成 SHA-512 十六进制摘要 |
| `hash.crc32(text)` | 字符串 | 生成 CRC32 校验值 |
| `hash.fnv32a(text)` | 字符串 | 生成 FNV-1a 32 位散列值 |
| `hash.fnv64a(text)` | 字符串 | 生成 FNV-1a 64 位散列值 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `sha1()` | 十六进制字符串 |
| `sha256()` | 十六进制字符串 |
| `sha512()` | 十六进制字符串 |
| `crc32()` | `number` |
| `fnv32a()` | `number` |
| `fnv64a()` | `number` |

## 最小示例

```javascript
import * as hash from "hash";

function main() {
  print(hash.sha1("demo"));
  print(hash.sha256("demo"));
  print(hash.sha512("demo"));
  print(hash.crc32("demo"));
}
```

## 代理/网关场景示例

```javascript
import * as hash from "hash";

let routeKey = "tenant=a|route=/orders|target=billing";
print(hash.sha256(routeKey));
print(hash.fnv64a(routeKey));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


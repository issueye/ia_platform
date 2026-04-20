# HMAC

- 上级目录：[数据类库](data)
- 导入：`import * as hmac from "hmac";`

## 模块定位

`hmac` 用于给请求体、回调通知和代理转发内容生成签名或做验签，适合接第三方 API、Webhook 和网关鉴权场景。

## 接口说明

- 核心入口：`sha1`、`sha256`、`sha512`、`verifySha256`
- 三个签名函数都要求 `(key, data)` 两个字符串参数，返回十六进制字符串
- `verifySha256(key, data, signatureHex)` 用常量时间比较校验签名
- 当前只内置 `verifySha256` 校验接口；如果你使用 `sha1/sha512`，需要自行比对结果

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `hmac.sha1(key, data)` | `string`, `string` | 生成 HMAC-SHA1 十六进制摘要 |
| `hmac.sha256(key, data)` | `string`, `string` | 生成 HMAC-SHA256 十六进制摘要 |
| `hmac.sha512(key, data)` | `string`, `string` | 生成 HMAC-SHA512 十六进制摘要 |
| `hmac.verifySha256(key, data, signatureHex)` | `string`, `string`, `string` | 验证给定签名是否匹配 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `sha1()` | 十六进制字符串 |
| `sha256()` | 十六进制字符串 |
| `sha512()` | 十六进制字符串 |
| `verifySha256()` | `bool` |

说明：

- `verifySha256()` 在签名不是合法十六进制字符串时直接返回 `false`
- 签名函数不附带 Base64 编码，需要时可配合 `encoding` 模块自行转换

## 最小示例

```javascript
import * as hmac from "hmac";

function main() {
  let key = "demo-secret";
  let payload = "order=1001&amount=88";

  let sig1 = hmac.sha1(key, payload);
  let sig256 = hmac.sha256(key, payload);
  let sig512 = hmac.sha512(key, payload);

  print(sig1);
  print(sig256);
  print(sig512);
  print(hmac.verifySha256(key, payload, sig256));
}
```

## 代理/网关场景示例

```javascript
import * as hmac from "hmac";
import * as time from "time";
import * as strconv from "strconv";

function main() {
  let secret = "proxy-upstream-secret";
  let ts = strconv.itoa(time.nowUnix());
  let canonical = "POST\n/api/payments/refund\n" + ts + "\n" + "{\"refund_id\":\"rf_2001\"}";

  let signature = hmac.sha256(secret, canonical);
  print(signature);

  let trusted = hmac.verifySha256(secret, canonical, signature);
  print(trusted);
}
```

## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。

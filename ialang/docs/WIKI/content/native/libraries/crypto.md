# Crypto

- 上级目录：[数据类库](data)
- 导入：`import * as crypto from "crypto";`

## 模块定位

`crypto` 适合做常见摘要和轻量安全处理，常见于签名、请求校验和内容指纹。

## 接口说明

- 核心入口：`sha256`、`md5`
- 当前实现只提供两个最常用的摘要入口，输入都是字符串，输出都是十六进制字符串。
- 如果只是生成摘要值，直接用这一层更省事；如果要做真正签名，通常会和 `hmac` 联用。
- 安全相关能力要区分“摘要”与“签名”，不要混用。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `crypto.sha256(text)` | 字符串 | 生成 SHA-256 十六进制摘要 |
| `crypto.md5(text)` | 字符串 | 生成 MD5 十六进制摘要 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `sha256()` | 十六进制字符串 |
| `md5()` | 十六进制字符串 |

## 最小示例

```javascript
import * as crypto from "crypto";

function main() {
  print(crypto.sha256("secret"));
  print(crypto.md5("secret"));
}
```

## 代理/网关场景示例

```javascript
import * as crypto from "crypto";

let signatureBase = "POST:/orders:payload";
print(crypto.sha256(signatureBase));
print(crypto.md5("health-check"));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


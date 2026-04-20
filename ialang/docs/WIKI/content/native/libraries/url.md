# URL

- 上级目录：[数据类库](data)
- 导入：`import * as url from "url";`

## 模块定位

`url` 负责 URL 解析和查询串处理，适合在代理里做路径重写、参数补全和跳转目标构造。

## 接口说明

- 核心入口：`parse`、`escape`、`unescape`、`queryEncode`、`queryDecode`
- `url` 负责 URL 拆解、转义和查询串编解码，适合做路径改写、回调地址拼装和上游参数透传。
- `queryEncode()` 适合从对象生成查询串，`queryDecode()` 适合把现有查询串转回对象再做修改。
- 通常与 `http`、`encoding`、`regexp` 配合使用。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `url.parse(raw)` | URL 字符串 | 解析完整 URL |
| `url.escape(text)` | 字符串 | 按查询串规则转义 |
| `url.unescape(text)` | 字符串 | 反转义查询串文本 |
| `url.queryEncode(obj)` | 对象 | 把对象编码成查询串；值需可转为字符串 |
| `url.queryDecode(query)` | 查询串文本 | 把 `a=1&b=2` 解成对象 |

## 返回值

### `url.parse()` 返回值

| 字段 | 类型 | 说明 |
|---|---|---|
| `scheme` | `string` | 协议，如 `https` |
| `host` | `string` | 主机名与端口 |
| `path` | `string` | 路径 |
| `query` | `string` | 原始查询串，不含 `?` |
| `fragment` | `string` | 片段，不含 `#` |
| `opaque` | `string` | 非透明 URL 的原始值 |
| `userinfo` | `string/null` | 用户信息 |

### 其他入口返回值

| 入口 | 返回值 |
|---|---|
| `escape()` | 转义后的字符串 |
| `unescape()` | 解码后的字符串 |
| `queryEncode()` | 编码后的查询串 |
| `queryDecode()` | 对象；重复参数会变成数组 |

## 最小示例

```javascript
import * as url from "url";

function main() {
  let parsed = url.parse("https://example.com/api?q=ialang#top");
  print(parsed.host);
  print(url.queryEncode({ q: "ialang", page: "1" }));
  print(url.queryDecode("q=ialang&page=1"));
}
```

## 代理/网关场景示例

```javascript
import { parse, queryEncode } from "url";

let u = parse("https://api.internal/orders?tenant=a");
print(u.host);
print(queryEncode({ tenant: "b", trace: "gw-001" }));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


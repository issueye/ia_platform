# JSON

- 上级目录：[数据类库](data)
- 导入：`import * as json from "json";`

## 模块定位

`json` 是处理 HTTP 负载和配置文件最常见的模块之一，适合解析上游响应、序列化转发结果和保存本地配置。

## 接口说明

- 核心入口：`parse`、`fromFile`、`stringify`、`saveToFile`、`valid`
- `json` 负责“字符串 <-> 运行时对象”的转换，是处理配置、HTTP 负载和审计数据的基础模块。
- `stringify(value, true)` 与 `saveToFile(value, path, true)` 会输出带缩进的 pretty JSON。
- 解析失败会直接报错；如果输入可能不可信，先用 `valid()` 做快速校验。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `json.parse(text)` | JSON 字符串 | 解析为运行时对象、数组、数字、布尔或 `null` |
| `json.fromFile(path)` | 文件路径 | 读取并解析 JSON 文件 |
| `json.stringify(value, [pretty])` | 任意值 + 可选布尔 | 序列化为 JSON 字符串 |
| `json.saveToFile(value, path, [pretty])` | 任意值 + 路径 + 可选布尔 | 序列化后写入文件 |
| `json.valid(text)` | JSON 字符串 | 仅检查格式是否合法 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `parse()` | 解析后的运行时值 |
| `fromFile()` | 解析后的运行时值 |
| `stringify()` | JSON 字符串 |
| `saveToFile()` | `true` |
| `valid()` | `bool` |

## 最小示例

```javascript
import * as json from "json";

function main() {
  let raw = `{"name":"ialang","ok":true,"ports":[8080,8081]}`;
  if (!json.valid(raw)) {
    print("invalid json");
    return;
  }

  let obj = json.parse(raw);
  print(obj.name);
  print(obj.ports[0]);
  print(json.stringify(obj, true));
}
```

## 代理/网关场景示例

```javascript
import * as json from "json";

let upstreamBody = `{
  "service": "billing",
  "ok": true,
  "routes": ["/invoice", "/refund"]
}`;

let payload = json.parse(upstreamBody);
payload.gateway = "edge-01";
payload.traceId = "gw-001";

json.saveToFile(payload, "./runtime/last-upstream.json", true);
print(json.stringify(payload, true));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


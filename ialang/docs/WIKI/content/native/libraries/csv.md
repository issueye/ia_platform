# CSV

- 上级目录：[数据类库](data)
- 导入：`import * as csv from "csv";`

## 模块定位

`csv` 适合导入导出批量路由数据、审计报表和离线节点清单。

## 接口说明

- 核心入口：`parse`、`stringify`
- `csv` 当前是轻量二维数组接口，适合导入导出批量路由、节点清单和审计报表。
- `parse()` 与 `stringify()` 都支持可选 `delimiter`，默认分隔符是英文逗号。
- 如果要进一步按列名处理，通常会先把首行当 header，再自行转成对象数组。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `csv.parse(text, [options])` | CSV 文本 + 可选对象 | 解析成二维数组 |
| `csv.stringify(rows, [options])` | 二维数组 + 可选对象 | 序列化成 CSV 文本 |
| `options.delimiter` | 单字符字符串 | 自定义分隔符，例如 `;` 或 `\t` |

## 返回值

| 入口 | 返回值 |
|---|---|
| `parse()` | 二维数组，每行是 `array<string>` |
| `stringify()` | CSV 字符串 |

## 最小示例

```javascript
import * as csv from "csv";

function main() {
  let rows = csv.parse(`name,age
alice,18
bob,20
`);
  print(rows[1][0]);
  print(csv.stringify([["id", "name"], ["1", "alice"]]));
  print(csv.parse("name;age\nalice;18\n", { delimiter: ";" })[1][1]);
}
```

## 代理/网关场景示例

```javascript
import * as csv from "csv";

let rows = csv.parse("service,target\norders,10.0.0.8:9000\n");
print(rows[1][0]);
print(csv.stringify([["route", "status"], ["orders", "ok"]]));
print(csv.stringify([["service", "target"], ["billing", "10.0.0.9:9000"]], {
  delimiter: ";"
}));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


# YAML

- 上级目录：[数据类库](data)
- 导入：`import * as yaml from "yaml";`

## 模块定位

`yaml` 适合写人类可读的配置文件，例如网关路由表、上游列表和功能开关。

## 接口说明

- 核心入口：`parse`、`fromFile`、`stringify`、`saveToFile`
- 兼容别名：`load` = `parse`，`dump` = `stringify`
- 典型流程是 `parse` 配置，再把对象传给其他模块；`stringify` 适合导出配置快照或默认模板。
- 如果脚本需要热加载配置，通常配合 `fs.readFile` 和 `timer`。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `yaml.parse(text)` | YAML 字符串 | 解析为运行时值 |
| `yaml.fromFile(path)` | 文件路径 | 读取并解析 YAML |
| `yaml.stringify(value)` | 任意运行时值 | 序列化为 YAML |
| `yaml.saveToFile(value, path)` | 值 + 路径 | 序列化并写入文件 |
| `yaml.load(text)` | 同 `parse()` | 别名 |
| `yaml.dump(value)` | 同 `stringify()` | 别名 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `parse()/fromFile()/load()` | 解析后的运行时值 |
| `stringify()/dump()` | 字符串 |
| `saveToFile()` | `true` |

## 最小示例

```javascript
import * as yaml from "yaml";

function main() {
  let cfg = yaml.parse(`server:
  host: localhost
  port: 8080
`);
  print(cfg.server.host);
  print(yaml.stringify({ app: { name: "demo" } }));
}
```

## 代理/网关场景示例

```javascript
import * as yaml from "yaml";

let cfg = yaml.parse("upstream:\n  host: localhost\n  port: 8080\n");
print(cfg.upstream.host);
print(yaml.stringify({ route: { prefix: "/api" } }));
print(yaml.dump({ audit: { enabled: true } }));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


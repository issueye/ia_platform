# TOML

- 上级目录：[数据类库](data)
- 导入：`import * as toml from "toml";`

## 模块定位

`toml` 适合写结构清晰的工程配置，例如代理启动参数、静态节点表和插件配置。

## 接口说明

- 核心入口：`parse`、`fromFile`、`stringify`、`saveToFile`
- 兼容别名：`encode` = `stringify`，`decode` = `parse`
- 如果配置本身更接近节和键值表，TOML 会更稳定，适合代理启动参数和静态节点表。
- 需要注意：`toml.stringify()` 和 `saveToFile()` 当前要求传入顶层对象。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `toml.parse(text)` | TOML 字符串 | 解析为运行时对象 |
| `toml.fromFile(path)` | 文件路径 | 读取并解析 TOML |
| `toml.stringify(obj)` | 顶层对象 | 序列化为 TOML 字符串 |
| `toml.saveToFile(obj, path)` | 顶层对象 + 路径 | 序列化并写入文件 |
| `toml.encode(obj)` | 同 `stringify()` | 别名 |
| `toml.decode(text)` | 同 `parse()` | 别名 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `parse()/fromFile()/decode()` | 对象 |
| `stringify()/encode()` | 字符串 |
| `saveToFile()` | `true` |

## 最小示例

```javascript
import * as toml from "toml";

function main() {
  let cfg = toml.parse(`[server]
host = "localhost"
port = 3000
`);
  print(cfg.server.port);
  print(toml.stringify({ title: "demo", version: "1.0.0" }));
}
```

## 代理/网关场景示例

```javascript
import * as toml from "toml";

let cfg = toml.parse("[proxy]\nport = 8080\nmode = \"edge\"\n");
print(cfg.proxy.port);
print(toml.stringify({ proxy: { mode: "internal" } }));
print(toml.encode({ upstream: { host: "127.0.0.1", port: 9000 } }));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


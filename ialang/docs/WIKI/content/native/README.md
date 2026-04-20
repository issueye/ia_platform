# 原生库概览

`ialang` 的原生库由运行时直接注册，不需要额外安装包。当前注册表除了基础模块，还包含一批面向服务端脚本的扩展模块，例如 `express`、`db`、`yaml`、`toml`、`asset`、`timer`、`pool`、`orm`。

## 导入方式

同一个模块通常支持三种导入写法：

```javascript
import { readFile } from "fs";
import { readFile } from "@std/fs";
import { readFile } from "@stdlib/fs";
```

例外：

- `timer` 额外保留了 `setTimeout` 的 plain 兼容名
- `exec` 还支持 `os/exec`
- `db` 与 `database`、`asset` 与 `bundle`、`iax` 与 `interaction` 是别名

## 按类别浏览

### HTTP 与网络

- `http`
- `websocket`
- `sse`
- `socket`
- `ipc`
- `iax`
- `net`
- `express`

见：[HTTP 与网络](http-and-network)

### 系统与运行时

- `fs`
- `path`
- `os`
- `process`
- `signal`
- `exec`
- `time`
- `timer`
- `pool`
- `asset` / `bundle`

见：[系统与运行时](system)

### 数据与工具

- `json`、`yaml`、`toml`、`xml`、`csv`
- `math`、`string`、`array`、`sort`、`set`
- `bytes`、`encoding`、`hex`、`mime`
- `crypto`、`hash`、`hmac`、`uuid`、`regexp`、`url`
- `rand`、`strconv`
- `db` / `database`、`orm`
- `Promise`、`log`、`@agent/sdk`

见：[数据与工具](data-and-utils)

## 深入到库级别

如果你不想只看分类总览，而是想按单个库打开说明与示例，请从这里进入：

- [原生库目录](libraries)
- [网络类库](libraries/network)
- [系统类库](libraries/system)
- [数据类库](libraries/data)
- [工具类库](libraries/tools)

## 使用建议

- 写脚本工具时，先看 `fs`、`path`、`os`、`process`、`exec`
- 写 HTTP 服务或网关时，先看 `http`、`express`、`websocket`、`sse`
- 做进程间通信时，优先看 `ipc`，需要更高层协议时再看 `iax`
- 做配置与数据处理时，优先看 `json`、`yaml`、`toml`、`csv`
- 做数据库访问时，直接看 `db` / `database` 与 `orm`

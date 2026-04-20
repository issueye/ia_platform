# FS

- 上级目录：[系统类库](system)
- 导入：`import * as fs from "fs";`

## 模块定位

`fs` 是最常用的系统模块之一，常用于读取代理配置、落盘日志、缓存上游响应或生成运行时状态文件。

## 接口说明

- 核心入口：`readFile`、`writeFile`、`appendFile`、`exists`、`mkdir`、`readDir`、`stat`
- 同步接口适合启动阶段或一次性脚本；`readFileAsync/writeFileAsync/appendFileAsync` 适合后台流程。
- 这个模块提供的都是基础文件能力，没有文件句柄抽象，适合配置读取、日志落盘和状态快照。
- 文件系统访问是否可用，取决于宿主 VM 的 `AllowFS` 配置。

## 参数要点

### 文件读写

| 入口 | 参数 | 说明 |
|---|---|---|
| `fs.readFile(path)` | 文件路径 | 读取整个文件并返回字符串 |
| `fs.writeFile(path, content)` | 文件路径 + 字符串 | 覆盖写入整个文件 |
| `fs.appendFile(path, content)` | 文件路径 + 字符串 | 追加写入；文件不存在时会自动创建 |

### 文件与目录状态

| 入口 | 参数 | 说明 |
|---|---|---|
| `fs.exists(path)` | 路径 | 检查路径是否存在 |
| `fs.mkdir(path, [recursive])` | 路径 + 可选布尔 | `recursive=true` 时递归创建目录 |
| `fs.readDir(path)` | 目录路径 | 读取目录项名称数组 |
| `fs.stat(path)` | 路径 | 读取文件或目录的元信息 |

### 异步接口

| 入口 | 参数 | 说明 |
|---|---|---|
| `fs.readFileAsync(path)` | 同 `readFile()` | 返回异步任务句柄 |
| `fs.writeFileAsync(path, content)` | 同 `writeFile()` | 返回异步任务句柄 |
| `fs.appendFileAsync(path, content)` | 同 `appendFile()` | 返回异步任务句柄 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `readFile()` | 文件内容字符串 |
| `writeFile()` | `true` |
| `appendFile()` | `true` |
| `exists()` | `bool` |
| `mkdir()` | `true` |
| `readDir()` | 文件名数组 |
| `readFileAsync()` | 异步任务句柄 |
| `writeFileAsync()` | 异步任务句柄 |
| `appendFileAsync()` | 异步任务句柄 |

### `fs.stat()` 返回值

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | `string` | 文件名或目录名 |
| `isDir` | `bool` | 是否目录 |
| `size` | `number` | 文件大小 |
| `mode` | `string` | 权限/模式文本 |
| `modTimeUnix` | `number` | 最后修改时间 Unix 秒时间戳 |

## 最小示例

```javascript
import * as fs from "fs";

function main() {
  fs.mkdir("./.tmp/wiki_examples/fs", true);
  fs.writeFile("./.tmp/wiki_examples/fs/demo.txt", "hello");
  print(fs.exists("./.tmp/wiki_examples/fs/demo.txt"));
  print(fs.readFile("./.tmp/wiki_examples/fs/demo.txt"));
  print(fs.stat("./.tmp/wiki_examples/fs/demo.txt").size);
}
```

## 代理/网关场景示例

```javascript
import * as fs from "fs";
import * as json from "json";

let routes = json.parse(fs.readFile("./config/routes.json"));
print(routes.defaultTarget);

fs.appendFile("./logs/gateway.log", "routes loaded\n");
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


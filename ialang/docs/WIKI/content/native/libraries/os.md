# OS

- 上级目录：[系统类库](system)
- 导入：`import * as os from "os";`

## 模块定位

`os` 让脚本感知运行平台、系统目录和环境变量，适合给代理脚本做启动配置和环境差异处理。

## 接口说明

- 核心入口：`platform`、`arch`、`hostname`、`cwd`、`tmpDir/tempDir`、`userDir`、`dataDir`、`configDir`、`cacheDir`
- `os` 负责平台识别、目录约定和环境变量读写，适合做跨平台启动配置。
- 读取环境变量时，通常和 `process.args()` 一起决定目标端口、上游地址或日志级别。
- `userDir`、`dataDir`、`configDir`、`cacheDir` 适合构造默认配置目录和缓存目录。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `os.platform()` | 无 | 返回当前平台，如 `windows`、`linux` |
| `os.arch()` | 无 | 返回架构，如 `amd64` |
| `os.hostname()` | 无 | 主机名 |
| `os.cwd()` | 无 | 当前工作目录 |
| `os.tmpDir()` / `os.tempDir()` | 无 | 临时目录 |
| `os.userDir()` | 无 | 用户主目录 |
| `os.dataDir()` | 无 | 用户数据目录 |
| `os.configDir()` | 无 | 用户配置目录 |
| `os.cacheDir()` | 无 | 用户缓存目录 |
| `os.getEnv(key)` | 环境变量名 | 读取单个环境变量 |
| `os.setEnv(key, value)` | 2 个字符串 | 设置环境变量 |
| `os.env()` | 无 | 读取全部环境变量对象 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `platform()` | 字符串 |
| `arch()` | 字符串 |
| `hostname()` | 字符串 |
| `cwd()` | 字符串 |
| `tmpDir()/tempDir()` | 字符串 |
| `userDir()` | 字符串 |
| `dataDir()` | 字符串 |
| `configDir()` | 字符串 |
| `cacheDir()` | 字符串 |
| `getEnv()` | 字符串或 `nil` |
| `setEnv()` | `true` |
| `env()` | 环境变量对象 |

## 最小示例

```javascript
import * as os from "os";

function main() {
  print(os.platform());
  print(os.arch());
  print(os.userDir());
  print(os.tempDir());
  print(os.dataDir());
}
```

## 代理/网关场景示例

```javascript
import * as os from "os";

print(os.platform());
print(os.arch());
print(os.getEnv("GATEWAY_TARGET"));
print(os.tempDir());
print(os.cacheDir());
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


# Process

- 上级目录：[系统类库](system)
- 导入：`import * as process from "process";`

## 模块定位

`process` 负责读取当前脚本进程的参数和上下文，适合做 CLI 代理、守护脚本或多模式启动入口。

## 接口说明

- 核心入口：`pid`、`ppid`、`args`、`cwd`、`chdir`、`getEnv`、`setEnv`、`env`、`exit`
- `process` 更关注当前运行实例，而不是整台机器，适合 CLI 代理、守护脚本和多模式入口。
- 如果脚本需要支持 `--config`、`--port`、`--mode` 这类参数，通常先从这里取值。
- 当前嵌入式运行时里 `process.exit()` 被禁用，调用时会返回错误而不会真正退出。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `process.pid()` | 无 | 当前进程 PID |
| `process.ppid()` | 无 | 父进程 PID |
| `process.args()` | 无 | 启动参数数组 |
| `process.cwd()` | 无 | 当前工作目录 |
| `process.chdir(path)` | 路径字符串 | 切换当前工作目录 |
| `process.getEnv(key)` | 环境变量名 | 读取单个环境变量 |
| `process.setEnv(key, value)` | 2 个字符串 | 设置环境变量 |
| `process.env()` | 无 | 读取当前环境变量对象 |
| `process.exit([code])` | 可选退出码 | 当前运行时里会报错，不会退出 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `pid()` | `number` |
| `ppid()` | `number` |
| `args()` | 字符串数组 |
| `cwd()` | 字符串 |
| `chdir()` | `true` |
| `getEnv()` | 字符串或 `nil` |
| `setEnv()` | `true` |
| `env()` | 环境变量对象 |
| `exit()` | 当前实现返回错误 |

## 最小示例

```javascript
import * as process from "process";

function main() {
  print(process.pid());
  print(process.ppid());
  print(process.cwd());
  print(process.args());
  print(process.getEnv("PATH"));
}
```

## 代理/网关场景示例

```javascript
import * as process from "process";

print(process.pid());
print(process.cwd());
print(process.args());
process.setEnv("GATEWAY_MODE", "edge");
print(process.getEnv("GATEWAY_MODE"));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


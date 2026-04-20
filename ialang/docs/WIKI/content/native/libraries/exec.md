# Exec

- 上级目录：[系统类库](system)
- 导入：`import * as exec from "exec";`
- 别名导入：`os/exec`

## 模块定位

`exec` 用于调用外部命令，适合把代理脚本和现有系统工具、证书生成器、探活命令或部署命令串起来。

## 接口说明

- 核心入口：`run`、`runAsync`、`start`、`lookPath`、`which`
- `run()` 适合“一次执行并收集输出”，`start()` 适合启动长时间运行的子进程。
- 支持工作目录、环境变量、标准输入、超时、shell 模式和输出继承。
- 外部进程能力是否可用，取决于宿主 VM 的 `AllowProcess` 配置。

## 参数要点

### `exec.run(command, [options])` / `exec.runAsync(command, [options])` / `exec.start(command, [options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `command` | `string` | 要执行的命令名或命令行 |
| `options.args` | `array` | 命令参数数组 |
| `options.cwd` | `string` | 工作目录 |
| `options.env` | `object` | 追加或覆盖环境变量 |
| `options.stdin` | `string` | 作为标准输入传给子进程 |
| `options.timeoutMs` | `number` | 超时毫秒数，必须大于 0 |
| `options.shell` | `bool` | 是否通过 shell 执行；Windows 下使用 `cmd /C` |
| `options.inheritOutput` | `bool` | 是否直接把输出写到当前进程标准输出/错误 |

### `exec.lookPath(name)` / `exec.which(name)`

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | `string` | 可执行文件名 |

## 返回值

### `exec.run()` / `exec.runAsync()` / `exec.child.wait()`

| 字段 | 类型 | 说明 |
|---|---|---|
| `ok` | `bool` | 是否以成功状态结束 |
| `code` | `number` | 退出码；启动失败时可能为 `-1` |
| `stdout` | `string` | 标准输出 |
| `stderr` | `string` | 标准错误 |
| `combined` | `string` | `stdout + stderr` |
| `timedOut` | `bool` | 是否因超时结束 |
| `error` | `string/null` | 错误文本 |

### `exec.start()` 返回值

子进程对象包含：

- `pid()`：返回子进程 PID
- `wait()`：等待进程结束并返回结果对象
- `waitAsync()`：异步等待
- `kill()`：强制终止进程
- `isRunning()`：是否仍在运行

### `lookPath()` / `which()` 返回值

- 找到可执行文件时返回路径字符串
- 找不到时返回 `nil`

## 最小示例

```javascript
import * as exec from "exec";

function main() {
  let result = exec.run("go", {
    args: ["version"],
    timeoutMs: 5000
  });
  print(result.ok);
  print(result.code);
  print(result.stdout);
}
```

## 代理/网关场景示例

```javascript
import * as exec from "exec";

let health = exec.run("curl", {
  args: ["-s", "http://127.0.0.1:8080/health"],
  timeoutMs: 2000
});

print(health.ok);
print(health.stdout);

let child = exec.start("curl", {
  args: ["-N", "http://127.0.0.1:8080/events"]
});
print(child.isRunning());
child.kill();
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


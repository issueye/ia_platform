# Signal

- 上级目录：[系统类库](system)
- 导入：`import * as signal from "signal";`

## 模块定位

`signal` 适合处理优雅退出和重载场景，例如代理进程收到 `SIGINT` 或 `SIGTERM` 后关闭监听并清理资源。

## 接口说明

- 核心入口：`notify`、`ignore`、`reset`，以及订阅对象 `recv/stop`
- `notify()` 接受单个信号名或信号数组，返回订阅对象；通常通过 `recv()` 阻塞等待，再调用 `stop()` 结束订阅。
- 这种接口很适合和 `http.server.close()`、`ipc.close()`、`pool.shutdown()` 搭配，做优雅退出。
- 信号行为受平台影响，跨平台脚本要注意 Windows 和 Unix 的差异。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `signal.notify(signals)` | `string/array` | 订阅一个或多个信号 |
| `signal.ignore([signals])` | 可选 `string/array` | 忽略指定信号；省略参数时对所有信号生效 |
| `signal.reset([signals])` | 可选 `string/array` | 恢复指定信号的默认行为；省略参数时对所有信号生效 |

常见常量：

- `signal.SIGINT`
- 以及当前平台支持的其他信号常量，如 `SIGTERM`

## 返回值

| 入口 | 返回值 |
|---|---|
| `notify()` | 订阅对象 |
| `ignore()` | `true` |
| `reset()` | `true` |

### 订阅对象

订阅对象包含：

- `recv()`：阻塞等待下一个信号，返回信号名字符串
- `stop()`：停止订阅并返回 `true`

## 最小示例

```javascript
import * as signal from "signal";

function main() {
  let sub = signal.notify(signal.SIGINT);
  print("signal subscription ready");
  sub.stop();
}
```

## 代理/网关场景示例

```javascript
import * as signal from "signal";

let sub = signal.notify(signal.SIGINT);
print("gateway waiting for shutdown signal");

let name = sub.recv();
print(name);
sub.stop();
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


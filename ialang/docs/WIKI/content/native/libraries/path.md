# Path

- 上级目录：[系统类库](system)
- 导入：`import * as path from "path";`

## 模块定位

`path` 负责路径拼装和拆解，适合把日志目录、缓存目录、证书目录和静态资源路径统一管理。

## 接口说明

- 核心入口：`join`、`base`、`dir`、`ext`、`clean`、`abs`
- 这一层基于宿主系统路径规则工作，适合同时兼容 Windows 和 Unix 风格的部署环境。
- 推荐把路径计算集中在启动阶段，避免在业务逻辑里散落字符串拼接。
- 通常与 `fs` 一起使用，用于构造可靠的配置、缓存和输出路径。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `path.join(...parts)` | 一个或多个路径片段 | 按系统规则拼接路径 |
| `path.base(pathText)` | 路径字符串 | 取文件名或最后一级目录名 |
| `path.dir(pathText)` | 路径字符串 | 取上级目录 |
| `path.ext(pathText)` | 路径字符串 | 取扩展名，含前导 `.` |
| `path.clean(pathText)` | 路径字符串 | 清理多余分隔符与 `.`、`..` |
| `path.abs(pathText)` | 路径字符串 | 转成绝对路径 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `join()` | 字符串 |
| `base()` | 字符串 |
| `dir()` | 字符串 |
| `ext()` | 字符串 |
| `clean()` | 字符串 |
| `abs()` | 字符串 |

## 最小示例

```javascript
import * as path from "path";

function main() {
  let p = path.join("logs", "2026", "app.log");
  print(p);
  print(path.base(p));
  print(path.ext(p));
  print(path.dir(p));
  print(path.clean("./logs/../logs/2026/app.log"));
}
```

## 代理/网关场景示例

```javascript
import * as path from "path";

let logFile = path.join(".tmp", "gateway", "access.log");
print(logFile);
print(path.base(logFile));
print(path.ext(logFile));
print(path.abs(logFile));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


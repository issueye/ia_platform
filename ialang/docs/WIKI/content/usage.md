# 命令与项目结构

本文聚焦 `ialang` 的 CLI 使用、项目入口规则、格式化和打包。

## 1. 运行脚本

```bash
go run ./ialang/cmd/ialang run path/to/app.ia
```

传参：

```bash
go run ./ialang/cmd/ialang run path/to/app.ia arg1 arg2 arg3
```

## 2. 检查语法

```bash
go run ./ialang/cmd/ialang check
```

```bash
go run ./ialang/cmd/ialang check path/to/file.ia
```

```bash
go run ./ialang/cmd/ialang check ./path/to/project
```

入口解析规则：

1. 如果目录下存在 `pkg.toml`，优先读取其中的 `entry`
2. 否则回退到 `main.ia`

## 3. 格式化

```bash
go run ./ialang/cmd/ialang fmt path/to/file.ia
```

```bash
go run ./ialang/cmd/ialang fmt ./path/to/dir
```

补充说明：

- `fmt` 当前只处理 `.ia` 文件
- 对目录执行时会递归扫描
- 会跳过常见目录，例如 `node_modules`、`.git`、`.tmp`

## 4. 初始化项目

```bash
go run ./ialang/cmd/ialang init myapp
```

## 5. 打包与运行包文件

构建字节码包：

```bash
go run ./ialang/cmd/ialang build examples/package_demo/main.ia -o app.iapkg
```

运行字节码包：

```bash
go run ./ialang/cmd/ialang run-pkg app.iapkg
```

## 6. 模块组织建议

### 本地模块

```javascript
import { greet } from "./modules/utils/index";
```

如果希望从项目根目录开始写绝对导入，可以在 `pkg.toml` 里配置根别名：

```toml
name = "myapp"
version = "0.1.0"
entry = "main.ia"

[imports]
root_alias = "@"
```

配置后可以这样导入：

```javascript
import { greet } from "@/modules/utils/index";
```

当前打包器会递归收集通过相对路径、项目根导入和已配置别名导入的本地模块。

### 内置模块

```javascript
import { sqrt } from "math";
import { fromFile } from "json";
import { info } from "log";
```

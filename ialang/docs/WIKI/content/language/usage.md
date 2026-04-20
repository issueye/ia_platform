# 命令与项目结构

这一页聚焦 `ialang` 的 CLI 工作流：运行、检查、格式化、初始化、打包和生成独立可执行文件。

## 在哪里执行命令

如果你位于当前仓库根目录：

```bash
go run ./ialang/cmd/ialang <command> ...
```

如果你已经进入 [`ialang`](/E:/code/issueye/ialang/ialang) 子目录：

```bash
go run ./cmd/ialang <command> ...
```

下文默认以“仓库根目录”写法为主。

## 运行脚本

```bash
go run ./ialang/cmd/ialang run ./ialang/examples/hello.ia
```

脚本加载完成后，如果定义了 `main()`，运行器会自动调用它。

传递脚本参数：

```bash
go run ./ialang/cmd/ialang run ./ialang/examples/hello.ia arg1 arg2 arg3
```

## 语法检查

检查当前项目：

```bash
go run ./ialang/cmd/ialang check
```

检查指定入口文件：

```bash
go run ./ialang/cmd/ialang check ./ialang/examples/hello.ia
```

检查指定项目目录：

```bash
go run ./ialang/cmd/ialang check ./ialang/examples/package_demo
```

成功时会输出：

```text
syntax check passed: <module-count> module(s), entry=<entry-path>
```

入口解析规则：

1. 目录下存在 `pkg.toml` 时，优先读取其中的 `entry`
2. 否则回退到 `main.ia`

## 格式化源码

格式化单文件：

```bash
go run ./ialang/cmd/ialang fmt ./ialang/examples/hello.ia
```

格式化目录：

```bash
go run ./ialang/cmd/ialang fmt ./ialang/examples
```

当前 `fmt` 的行为：

- 只处理 `.ia` 文件
- 目录模式下递归扫描
- 会跳过常见目录，例如 `.git`、`.tmp`、`node_modules`

## 初始化项目

```bash
go run ./ialang/cmd/ialang init myapp
cd myapp
```

默认生成结构：

```text
myapp/
├── .gitignore
├── README.md
├── main.ia
├── pkg.toml
├── config/
│   └── app.json
└── modules/
    ├── pkg/
    │   └── index.ia
    └── utils/
        └── index.ia
```

`init` 还会尝试在目标目录执行 `git init`；如果 `.git` 已存在，则不会重复创建。

默认生成的 `pkg.toml` 会包含项目根别名配置：

```toml
[imports]
root_alias = "@"
```

## 本地模块和内置模块

本地模块：

```javascript
import { greet } from "./modules/utils/index";
```

项目根绝对导入：

```javascript
import { greet } from "@/modules/utils/index";
```

内置模块：

```javascript
import { sqrt } from "math";
import { fromFile } from "json";
import { info } from "log";
```

## 构建字节码包

```bash
go run ./ialang/cmd/ialang build ./ialang/examples/package_demo/main.ia -o app.iapkg
```

运行包文件：

```bash
go run ./ialang/cmd/ialang run-pkg app.iapkg
```

`run-pkg` 也支持继续传参数：

```bash
go run ./ialang/cmd/ialang run-pkg app.iapkg arg1 arg2
```

## 生成独立可执行文件

```bash
go run ./ialang/cmd/ialang build-bin ./ialang/examples/package_demo/main.ia -o package_demo.exe
```

当前实现方式是把 `.iapkg` 数据追加到 `ialang` 可执行文件尾部，因此有两个注意点：

- 输出文件不能覆盖当前正在运行的 `ialang` 可执行文件
- Windows 常见输出名是 `.exe`，其他平台可以不带扩展名

当前打包器会递归收集通过相对路径、项目根导入和已配置别名导入的本地模块；内置模块由运行时提供。

## 常见工作流

单脚本开发：

```bash
go run ./ialang/cmd/ialang run ./ialang/examples/hello.ia
go run ./ialang/cmd/ialang check ./ialang/examples/hello.ia
go run ./ialang/cmd/ialang fmt ./ialang/examples/hello.ia
```

小项目开发：

```bash
go run ./ialang/cmd/ialang init myapp
cd myapp
go run ../ialang/cmd/ialang fmt
go run ../ialang/cmd/ialang check
go run ../ialang/cmd/ialang run main.ia
```

分发产物：

```bash
go run ./ialang/cmd/ialang build ./ialang/examples/package_demo/main.ia -o app.iapkg
go run ./ialang/cmd/ialang run-pkg app.iapkg
go run ./ialang/cmd/ialang build-bin ./ialang/examples/package_demo/main.ia -o package_demo.exe
```

## 对应示例

如果你不想只看命令格式，而是想直接运行已经整理好的 demo：

- [示例总览](../examples)
- [原生库概览](../native)
- [原生库目录](../native/libraries)

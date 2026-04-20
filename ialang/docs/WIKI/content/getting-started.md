# 快速开始

## 前置条件

- Go 1.25.x 或兼容版本
- 当前仓库根目录可执行 `go run ./ialang/cmd/ialang ...`
- 如果进入 [`ialang`](/E:/code/issueye/ialang/ialang) 子目录，则使用 `go run ./cmd/ialang ...`

## 先跑通一个脚本

在仓库根目录执行：

```bash
go run ./ialang/cmd/ialang run ./ialang/examples/hello.ia
```

如果脚本里定义了 `main()`，运行器会在加载完成后自动调用它。

跑通 `hello.ia` 之后，建议继续看 [示例总览](examples)，那里已经把原生库主题示例和单库示例整理好了。

## 检查语法

```bash
go run ./ialang/cmd/ialang check ./ialang/examples/hello.ia
```

成功时会看到类似输出：

```text
syntax check passed: <module-count> module(s), entry=<entry-path>
```

## 格式化源码

格式化单文件：

```bash
go run ./ialang/cmd/ialang fmt ./ialang/examples/hello.ia
```

格式化整个示例目录：

```bash
go run ./ialang/cmd/ialang fmt ./ialang/examples
```

## 初始化一个项目

```bash
go run ./ialang/cmd/ialang init myapp
cd myapp
```

生成后的结构通常是：

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

默认生成的 `pkg.toml` 会包含：

```toml
[imports]
root_alias = "@"
```

这样项目内可以直接使用：

```javascript
import { greet } from "@/modules/utils/index";
```

## 打包与运行产物

构建 `.iapkg`：

```bash
go run ./ialang/cmd/ialang build ./ialang/examples/package_demo/main.ia -o app.iapkg
```

运行 `.iapkg`：

```bash
go run ./ialang/cmd/ialang run-pkg app.iapkg
```

生成独立可执行文件：

```bash
go run ./ialang/cmd/ialang build-bin ./ialang/examples/package_demo/main.ia -o package_demo.exe
```

## 常用命令速查

| 命令 | 作用 |
|---|---|
| `run <file.ia>` | 运行入口脚本 |
| `check [path]` | 检查单文件、当前项目或指定项目目录 |
| `fmt [path]` | 格式化 `.ia` 文件或目录 |
| `init [dir]` | 初始化新项目 |
| `build <entry> -o app.iapkg` | 构建字节码包 |
| `run-pkg <file.iapkg>` | 运行字节码包 |
| `build-bin <entry> -o app.exe` | 生成独立可执行文件 |

## 下一步

- 想继续看真实代码：打开 [示例总览](examples)
- 想看语法和项目结构：打开 [语言概览](language) 和 [命令与项目结构](language/usage)
- 想直接看内置库：打开 [原生库概览](native)

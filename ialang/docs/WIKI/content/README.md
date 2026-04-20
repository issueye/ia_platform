# ialang Wiki

这是一个基于 `mdwiki` 的 `ialang` 文档站，覆盖两部分：

- `ialang` 语言与 CLI 的实际使用方式
- `ialang` 原生库 / 内置模块的导入方式与常用 API

`mdwiki` 首页使用的是 [index](index)。这个 `README.md` 保留为内容目录入口说明，正文与首页保持一致。

## 文档入口

- [快速开始](getting-started)
- [示例总览](examples)
- [语言概览](language)
- [命令与项目结构](language/usage)
- [原生库概览](native)
- [原生库目录](native/libraries)
- [网络类库目录](native/libraries/network)
- [系统类库目录](native/libraries/system)
- [数据类库目录](native/libraries/data)
- [工具类库目录](native/libraries/tools)

## 这套文档解决什么问题

如果你当前就在本仓库根目录 `E:\code\issueye\ialang` 下工作，最常见的困惑通常是：

- 文档站在根目录，但真正的 `ialang` 源码在 [`ialang`](/E:/code/issueye/ialang/ialang)
- 从仓库根目录运行 CLI 时，命令应写成 `go run ./ialang/cmd/ialang ...`
- 从 [`ialang`](/E:/code/issueye/ialang/ialang) 子目录内部运行时，命令则是 `go run ./cmd/ialang ...`
- 项目本地模块除了相对路径，也支持在 `pkg.toml` 里配置 `[imports].root_alias` 后使用 `@/...` 这类项目根导入
- 原生库既支持 `"fs"`，也支持 `"@std/fs"` 与 `"@stdlib/fs"` 这样的别名导入
- 部分模块还有兼容别名，例如 `database`、`bundle`、`interaction`、`os/exec`

本文档按这个工作场景组织，优先给出能直接复制执行的命令和示例。

## 怎么选阅读路径

- 想先跑通命令：从 [快速开始](getting-started) 开始
- 想直接找可运行的示例：先看 [示例总览](examples)
- 想确认语法和模块系统：看 [语言概览](language) 与 [命令与项目结构](language/usage)
- 只想按类别浏览内置库：看 [原生库概览](native)
- 想直接打开某一个库的说明和示例：从 [原生库目录](native/libraries) 进入

## 示例代码入口

文档里已经提供了两层示例说明：

- 按主题聚合的示例导读：见 [示例总览](examples)
- 按库拆分的详细代码：见 [原生库目录](native/libraries)

如果你要逐个看原生库，优先从单库页面里的代码块直接开始。

## 主要资料来源

文档内容基于当前仓库实现整理，核心来源包括：

- [`ialang/README.md`](/E:/code/issueye/ialang/ialang/README.md)
- [`ialang/docs/usage-guide.md`](/E:/code/issueye/ialang/ialang/docs/usage-guide.md)
- [`ialang/docs/2026-04-07/NATIVE_MODULES.md`](/E:/code/issueye/ialang/ialang/docs/2026-04-07/NATIVE_MODULES.md)
- [`ialang/docs/2026-04-08/README.md`](/E:/code/issueye/ialang/ialang/docs/2026-04-08/README.md)
- [`ialang/pkg/lang/runtime/builtin/registry.go`](/E:/code/issueye/ialang/ialang/pkg/lang/runtime/builtin/registry.go)

## 用 mdwiki 浏览

在仓库根目录执行：

```powershell
.\mdwiki\mdwiki.exe -config .\ialang_wiki\config.json
```

如果你要用源码方式启动，则先进入 `mdwiki` 子目录：

```powershell
cd .\mdwiki
go run . -config ..\ialang_wiki\config.json -content ..\ialang_wiki\content
```

然后访问：

```text
http://localhost:8081
```

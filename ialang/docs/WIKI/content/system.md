# 系统与进程

本文汇总与操作系统、文件、进程和运行环境相关的模块。

## fs

同步函数：

- `readFile(path)`
- `writeFile(path, content)`
- `appendFile(path, content)`
- `exists(path)`
- `mkdir(path, [recursive])`
- `readDir(path)`
- `stat(path)`

异步函数：

- `readFileAsync(path)`
- `writeFileAsync(path, content)`
- `appendFileAsync(path, content)`

## path

常用于拼接和处理路径。

## os

用于获取系统相关信息。

## process

用于处理当前进程相关信息和参数。

## exec

执行外部命令：

- `run(command, [options])`
- `runAsync(command, [options])`
- `lookPath(name)` / `which(name)`

## signal

处理系统信号。

## timer

定时与调度能力。

## pool

协程池 / 任务池相关能力。

## 沙箱限制

`ialang` 运行时支持若干安全开关：

- `AllowFS`
- `AllowNetwork`
- `AllowProcess`
- `AllowedModules`
- `MaxSteps`
- `MaxDuration`
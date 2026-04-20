# 系统类库

这一页覆盖文件系统、宿主运行时、进程、调度与资源打包相关的原生库。每个单库页都已经内嵌完整示例代码。

## fs

- 文档页：[FS](fs)
- 导入：`import * as fs from "fs";`
- 常用入口：`readFile`、`writeFile`、`appendFile`、`mkdir`、`readDir`、`stat`

## path

- 文档页：[Path](path)
- 导入：`import * as path from "path";`
- 常用入口：`join`、`base`、`dir`、`ext`、`clean`、`abs`

## os

- 文档页：[OS](os)
- 导入：`import * as os from "os";`
- 常用入口：`platform`、`arch`、`cwd`、`userDir`、`configDir`、`cacheDir`

## process

- 文档页：[Process](process)
- 导入：`import * as process from "process";`
- 常用入口：`pid`、`ppid`、`args`、`cwd`、`chdir`

## signal

- 文档页：[Signal](signal)
- 导入：`import * as signal from "signal";`
- 常用入口：`notify`、`ignore`、`reset`、订阅对象 `recv/stop`

## exec

- 文档页：[Exec](exec)
- 导入：`import * as exec from "exec";`
- 别名导入：`os/exec`
- 常用入口：`run`、`runAsync`、`start`、`lookPath`、`which`

## time

- 文档页：[Time](time)
- 导入：`import * as time from "time";`
- 常用入口：`nowUnix`、`nowUnixMilli`、`nowISO`、`sleep`、`sleepAsync`

## timer

- 文档页：[Timer](timer)
- 导入：`import * as timer from "timer";`
- 兼容 plain 名：`setTimeout`
- 常用入口：`setTimeout`、`setInterval`、`sleepAsync`、`cron`、`removeJob`

## pool

- 文档页：[Pool](pool)
- 导入：`import * as pool from "pool";`
- 常用入口：`createPool`、`getStats`、`shutdown`

## asset / bundle

- 文档页：[Asset](asset)
- 导入：`import * as asset from "asset";`
- 别名导入：`bundle`
- 常用入口：`bundle`、`minify`、`hash`、`version`、`analyze`

# 系统与运行时

这一页聚焦文件系统、路径、操作系统、进程、外部命令、定时任务和运行时资源管理。下面每个库都给一个最小示例。

如果你要继续看单库级别的详细接口说明，直接进入 [系统类库目录页](libraries/system)。

## fs

### 示例：读写文件

```javascript
import { writeFile, readFile, exists } from "fs";

writeFile("./tmp/demo.txt", "hello ialang");
let text = readFile("./tmp/demo.txt");

print(exists("./tmp/demo.txt"));
print(text);
```

## path

### 示例：拼接和分析路径

```javascript
import { join, base, ext, dir } from "path";

let p = join("logs", "2026", "app.log");
print(p);
print(base(p));
print(ext(p));
print(dir(p));
```

## os

### 示例：读取系统目录和环境变量

```javascript
import { platform, arch, userDir, tempDir, getEnv } from "os";

print(platform());
print(arch());
print(userDir());
print(tempDir());
print(getEnv("PATH"));
```

## process

### 示例：读取当前进程上下文

```javascript
import { pid, ppid, args, cwd } from "process";

print(pid());
print(ppid());
print(cwd());
print(args());
```

## signal

### 示例：订阅中断信号

```javascript
import { notify, SIGINT } from "signal";

let sub = notify([SIGINT]);
print("waiting for SIGINT...");

let name = sub.recv();
print(name);
sub.stop();
```

## exec

### 示例：执行外部命令

```javascript
import { run } from "exec";

let result = run("git", {
  args: ["status", "--short"],
  cwd: ".",
  timeoutMs: 5000
});

print(result.ok);
print(result.stdout);
```

## time

### 示例：获取时间戳

```javascript
import { nowUnix, nowUnixMilli, nowISO } from "time";

print(nowUnix());
print(nowUnixMilli());
print(nowISO());
```

## timer

### 示例：注册一次延时任务

```javascript
import { setTimeout, clearTimeout } from "timer";

function onDone() {
  print("timer fired");
}

let id = setTimeout(onDone, 200);
clearTimeout(id);
```

### 示例：等待异步延时

```javascript
import { sleepAsync } from "timer";

await sleepAsync(100);
print("awake");
```

## pool

### 示例：创建池并查看统计

```javascript
import { createPool } from "pool";

let p = createPool({
  minWorkers: 1,
  maxWorkers: 2,
  queueSize: 8
});

let stats = p.getStats();

print(stats.totalWorkers);
print(stats.queuedTasks);
p.shutdown(2000);
```

## asset / bundle

### 示例：生成内容哈希

```javascript
import { hash, minify } from "asset";

let minified = minify("body { color: red; }", "css");
let hashed = hash(minified.content);

print(minified.content);
print(hashed.hash);
```

### 示例：分析资源目录

```javascript
import { analyze } from "bundle";

let stats = analyze("./assets");
print(stats.fileCount);
print(stats.totalSize);
```

## 运行时限制

`ialang` 运行时支持沙箱配置，常见开关包括：

- `AllowFS`
- `AllowNetwork`
- `AllowProcess`
- `AllowedModules`
- `MaxSteps`
- `MaxDuration`

如果你在宿主侧嵌入 VM，这些开关决定脚本是否可以访问文件、网络或外部进程。

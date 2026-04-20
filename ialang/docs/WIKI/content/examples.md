# 示例总览

这一页不再把你引到外部示例文件，而是直接给出三类最常见的文档内示例阅读路径。

## 1. 网络侧示例

适合先看请求、服务、代理和本机通信：

- [HTTP 与网络](native/http-and-network)
- [网络类库目录](native/libraries/network)

典型代码形态：

```javascript
import { server } from "http";

let upstream = server.serve({
  addr: "127.0.0.1:9000",
  body: "origin"
});

let gateway = server.proxy({
  addr: "127.0.0.1:8080",
  target: "http://" + upstream.addr,
  requestMutations: {
    setHeaders: { "x-gateway": "ialang" }
  }
});

print(gateway.addr);
```

## 2. 系统侧示例

适合先看文件、进程、命令执行和定时调度：

- [系统与运行时](native/system)
- [系统类库目录](native/libraries/system)

典型代码形态：

```javascript
import * as fs from "fs";
import * as timer from "timer";

fs.mkdir("./.tmp/wiki", true);
fs.writeFile("./.tmp/wiki/status.txt", "ready");

await timer.sleepAsync(50);
print(fs.readFile("./.tmp/wiki/status.txt"));
```

## 3. 数据侧示例

适合先看配置、编码、安全、数据库和异步编排：

- [数据与工具](native/data-and-utils)
- [数据类库目录](native/libraries/data)
- [工具类库目录](native/libraries/tools)

典型代码形态：

```javascript
import * as json from "json";
import { db } from "db";

let payload = json.parse(`{"service":"billing","ok":true}`);
let conn = db.sqlite(":memory:");
conn.exec("CREATE TABLE logs (service TEXT, ok INTEGER)");
conn.exec("INSERT INTO logs (service, ok) VALUES (?, ?)", [payload.service, 1]);

print(conn.query("SELECT * FROM logs")[0].service);
conn.close();
```

## 建议阅读顺序

- 先看 [快速开始](getting-started)
- 再按主题看 [HTTP 与网络](native/http-and-network)、[系统与运行时](native/system)、[数据与工具](native/data-and-utils)
- 最后进入 [原生库目录](native/libraries) 看单库详细页和内嵌代码块

## 使用提醒

- 这套文档里的代码块现在就是主要示例来源
- 如果你只想快速理解某个库，直接打开对应单库页即可
- 需要成组理解模块配合方式时，优先看三张主题总览页

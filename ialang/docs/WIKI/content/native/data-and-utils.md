# 数据与工具

这一组模块覆盖数据格式、集合与字符串处理、编码、安全、数据库和一些高层工具能力。下面每个库都给一个最小示例。

如果你要继续看单库级别的详细接口说明，可以分别进入 [数据类库目录页](libraries/data) 和 [工具类库目录页](libraries/tools)。

## 数值、字符串与集合

### math

```javascript
import { sqrt, pow, PI } from "math";

print(sqrt(16));
print(pow(2, 8));
print(PI);
```

### string

```javascript
import { split } from "string";

let text = "  hello,ialang  ";
print(text.trim().toUpperCase());
print(split("a,b,c", ","));
```

### array

```javascript
import { range } from "array";

let nums = range(1, 5);
nums.push(9);

print(nums.join("-"));
print(nums.at(-1));
```

### sort

```javascript
import { asc, unique } from "sort";

print(asc([5, 2, 9, 2]));
print(unique([1, 1, 2, 3, 3]));
```

### set

```javascript
import { union, intersect, diff } from "set";

print(union([1, 2], [2, 3]));
print(intersect([1, 2, 3], [2, 4]));
print(diff([1, 2, 3], [2]));
```

### strconv

```javascript
import { atoi, formatFloat, parseBool } from "strconv";

print(atoi("42"));
print(formatFloat(3.14159, 2));
print(parseBool("true"));
```

### rand

```javascript
import { int, float, pick, string } from "rand";

print(int(100));
print(float(1, 10));
print(pick(["red", "green", "blue"]));
print(string(8));
```

## 日志与异步工具

### log

```javascript
import { info, setJSON } from "log";

setJSON(true);
info("job started", { service: "sync-worker", batch: 3 });
```

### Promise

```javascript
import { all, race, allSettled } from "Promise";

async function task1() { return "a"; }
async function task2() { return "b"; }

print(await all([task1(), task2()]));
print(await race([task1(), task2()]));
print(await allSettled([task1(), task2()]));
```

### @agent/sdk

```javascript
import { llm, tool, memory } from "@agent/sdk";

print(llm.chat("summarize the build plan"));
print(tool.call("deploy.preview", "web"));
print(memory.get("last_release"));
```

## 数据格式与编码

### json

```javascript
import { parse, stringify } from "json";

let obj = parse("{\"name\":\"ialang\",\"ok\":true}");
print(obj.name);
print(stringify(obj, true));
```

### yaml

```javascript
import { parse, stringify } from "yaml";

let cfg = parse("server:\n  host: localhost\n  port: 8080\n");
print(cfg.server.host);
print(stringify({ app: { name: "demo" } }));
```

### toml

```javascript
import { parse, stringify } from "toml";

let cfg = parse("[server]\nhost = \"localhost\"\nport = 3000\n");
print(cfg.server.port);
print(stringify({ title: "demo", version: "1.0.0" }));
```

### xml

```javascript
import { parse, stringify, valid } from "xml";

let doc = parse("<user><name>alice</name></user>");
print(doc.name);
print(valid("<root></root>"));
print(stringify({ name: "note", attrs: {}, text: "hello", children: [] }));
```

### csv

```javascript
import { parse, stringify } from "csv";

let rows = parse("name,age\nalice,18\nbob,20\n");
print(rows[1][0]);
print(stringify([["id", "name"], ["1", "alice"]]));
```

### encoding

```javascript
import { base64Encode, base64Decode, urlEncode } from "encoding";

let b64 = base64Encode("hello");
print(b64);
print(base64Decode(b64));
print(urlEncode("a b+c"));
```

### hex

```javascript
import { encode, decode } from "hex";

let encoded = encode("ialang");
print(encoded);
print(decode(encoded));
```

### bytes

```javascript
import { fromString, toBase64, toString } from "bytes";

let data = fromString("hello");
print(toBase64(data));
print(toString(data));
```

### mime

```javascript
import { typeByExt, extByType, detectByPath } from "mime";

print(typeByExt(".json"));
print(extByType("text/html"));
print(detectByPath("./public/app.js"));
```

## 安全与标识

### crypto

```javascript
import { sha256, md5 } from "crypto";

print(sha256("secret"));
print(md5("secret"));
```

### hash

```javascript
import { sha1, sha512, crc32 } from "hash";

print(sha1("demo"));
print(sha512("demo"));
print(crc32("demo"));
```

### hmac

```javascript
import { sha256, verifySha256 } from "hmac";

let sig = sha256("key-1", "payload");
print(sig);
print(verifySha256("key-1", "payload", sig));
```

### uuid

```javascript
import { v4, isValid } from "uuid";

let id = v4();
print(id);
print(isValid(id));
```

### regexp

```javascript
import { compile, findAllSubmatch } from "regexp";

let re = compile("(\\w+)-(\\d+)");
print(re.find("item-42"));
print(findAllSubmatch("(\\w+)-(\\d+)", "item-42 task-7"));
```

### url

```javascript
import { parse, queryEncode, queryDecode } from "url";

let u = parse("https://example.com/api?q=ialang#top");
print(u.host);
print(queryEncode({ q: "ialang", page: "1" }));
print(queryDecode("q=ialang&page=1"));
```

## 数据库相关

### db / database

```javascript
import { db } from "db";

let conn = db.sqlite(":memory:");
conn.exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)");
conn.exec("INSERT INTO users (name) VALUES (?)", ["alice"]);

let rows = conn.query("SELECT * FROM users");
print(rows[0].name);
conn.close();
```

### orm

```javascript
import { db } from "db";
import { init, DataTypes, defineModel, QueryBuilder, buildQuery } from "orm";

let conn = db.sqlite(":memory:");
let orm = init(conn);
let User = defineModel("User", {
  name: { type: DataTypes.STRING, notNull: true },
  age: { type: DataTypes.INTEGER }
});

let qb = QueryBuilder(User).where({ name: "alice" }).order("age", "DESC").limit(10);
let query = buildQuery(qb);

print(query.sql);
print(query.params);
```

## 推荐组合

- 配置文件处理：`fs` + `json` / `yaml` / `toml`
- 文本清洗：`string` + `regexp` + `strconv`
- 安全签名：`crypto` / `hash` / `hmac`
- 数据导入导出：`csv` + `json` + `bytes`
- 服务端脚本日志：`log` + `time`

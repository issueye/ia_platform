# DB

- 上级目录：[数据类库](data)
- 导入：`import { db } from "db";`
- 别名导入：`database`

## 模块定位

`db` / `database` 提供直接的数据库连接和查询能力，适合把路由规则、审计日志、租户配置和限流状态持久化。

## 接口说明

- 核心入口：`sqlite`、`postgres`、`mysql`、`sqlserver`、`query`、`exec`
- 连接入口包括 `connect/sqlite/mysql/postgres/sqlserver`；数据库对象常见方法是 `query/queryOne/exec/transaction/prepare/ping/close`。
- `exec` 返回 `affectedRows/insertId`；`transaction` 会把事务对象传给回调。
- SQLite 最适合本地演示和单机脚本；生产代理通常会接独立数据库。

## 参数要点

### 建立连接

| 入口 | 参数 | 说明 |
|---|---|---|
| `db.connect(driver, [dsn])` | 驱动名 + 连接串 | 通用入口 |
| `db.sqlite(path)` | 文件路径或 `:memory:` | SQLite 便捷入口 |
| `db.mysql(dsn)` | DSN | MySQL 便捷入口 |
| `db.postgres(dsn)` | DSN | PostgreSQL 便捷入口 |
| `db.sqlserver(dsn)` | DSN | SQL Server 便捷入口 |

### 数据库对象方法

| 方法 | 参数 | 说明 |
|---|---|---|
| `query(sql, [params])` | SQL + 参数数组 | 返回结果行数组 |
| `queryOne(sql, [params])` | SQL + 参数数组 | 当前实现较简化，适合谨慎使用 |
| `exec(sql, [params])` | SQL + 参数数组 | 执行写操作 |
| `execute(sql, [params])` | 同 `exec` | `exec` 的别名 |
| `transaction(callback)` | 回调函数 | 创建事务并把事务对象传给回调 |
| `prepare(sql)` | SQL | 返回预处理语句对象 |
| `ping()` | 无 | 检查连接可用性 |
| `close()` | 无 | 关闭连接 |

### 事务对象

事务回调收到的对象常见方法：

- `query(sql, [params])`
- `exec(sql, [params])`
- `commit()`
- `rollback()`

### 预处理语句对象

`prepare()` 返回的对象常见方法：

- `query([params])`
- `exec([params])`
- `close()`

## 返回值

### `query()` 返回值

返回数组，每一项是按列名组织的对象，例如：

```javascript
[
  { id: 1, name: "alice" }
]
```

### `exec()` 返回值

| 字段 | 类型 | 说明 |
|---|---|---|
| `affectedRows` | `number` | 受影响行数 |
| `insertId` | `number` | 最后插入 ID；不支持时可能为 `0` |

### `transaction()` 返回值

| 字段 | 类型 | 说明 |
|---|---|---|
| `committed` | `bool` | 是否已提交 |
| `result` | `any` | 回调返回值 |

## 最小示例

```javascript
import { db } from "db";

function main() {
  let conn = db.sqlite(":memory:");
  conn.exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)");
  conn.exec("INSERT INTO users (name) VALUES (?)", ["alice"]);
  let rows = conn.query("SELECT * FROM users");
  print(rows[0].name);
  conn.close();
}
```

## 代理/网关场景示例

```javascript
import { db } from "db";

let conn = db.sqlite(":memory:");
conn.exec("CREATE TABLE routes (path TEXT, target TEXT)");
conn.exec("INSERT INTO routes (path, target) VALUES (?, ?)", ["/orders", "billing"]);
print(conn.query("SELECT * FROM routes")[0].target);
conn.close();
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


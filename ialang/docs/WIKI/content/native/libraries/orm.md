# ORM

- 上级目录：[数据类库](data)
- 导入：`import * as orm from "orm";`

## 模块定位

`orm` 适合在脚本层定义模型、拼装查询和统一管理数据访问逻辑，适合复杂一些的配置中心或审计系统。

## 接口说明

- 核心入口：`init`、`defineModel`、`createModel/updateModel/deleteModel`、`QueryBuilder`、`buildQuery`
- `orm` 当前更偏“SQL 生成器 + 模型描述器”，而不是完整 Active Record。
- 如果只是执行少量 SQL，直接用 `db` 更直接；如果你想把模型定义、命名约定和查询拼装收敛起来，再引入 `orm`。
- 需要注意：当前 `association` 和 `include` 相关接口主要是占位结构，`buildQuery()` 只会生成基础查询，不会自动生成 JOIN。

## 参数要点

### `orm.init(database)`

| 字段 | 类型 | 说明 |
|---|---|---|
| `database` | `object` | 通常来自 `db.sqlite()`、`db.postgres()` 等连接对象 |

### `orm.defineModel(name, fields, [options])`

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | `string` | 模型名，例如 `RouteRule` |
| `fields` | `object` | 字段定义；值可为简单类型字符串，也可为详细对象 |
| `options.tableName` | `string` | 覆盖默认表名 |
| `options.timestamps` | `bool` | 默认 `true`，自动补 `createdAt` / `updatedAt` |

字段定义对象常见属性：

| 字段 | 类型 | 说明 |
|---|---|---|
| `type` | `string` | 推荐来自 `orm.DataTypes.*` |
| `primaryKey` | `bool` | 是否主键 |
| `autoIncrement` | `bool` | 是否自增 |
| `notNull` | `bool` | 是否非空 |
| `defaultValue` | `string/number` | 默认值 |
| `unique` | `bool` | 是否唯一 |

补充规则：

- 未显式声明 `id` 时，会自动添加自增主键 `id`
- `timestamps=true` 时，会自动追加 `createdAt` 和 `updatedAt`
- 默认表名按模型名转蛇形并做简单复数化，例如 `RoutePolicy -> route_policies`

### `orm.createModel(model, data)` / `orm.updateModel(model, data, conditions)` / `orm.deleteModel(model, conditions)`

| 入口 | 说明 |
|---|---|
| `createModel()` | 生成 `INSERT` 语句与参数，不直接执行 |
| `updateModel()` | 生成 `UPDATE` 语句与参数，不直接执行 |
| `deleteModel()` | 生成 `DELETE` 语句与参数，不直接执行 |

### `orm.QueryBuilder(model)`

查询构建器当前支持的方法：

- `where(object)`
- `order(field, [direction])`
- `limit(number)`
- `offset(number)`
- `include(value)`
- `buildQuery()`

其中：

- `where()` 会把对象字段转成 `snake_case = ?`
- `order()` 默认方向为 `ASC`
- `include()` 当前只记录状态，`buildQuery()` 不会据此生成联表 SQL

### 其他入口

| 入口 | 说明 |
|---|---|
| `orm.DataTypes` | 内建类型映射，如 `STRING`、`INTEGER`、`JSON` |
| `orm.migrationHelper(database)` | 返回 `{ createTable(), sync() }` 的轻量迁移辅助对象 |
| `orm.underscore(text)` | 驼峰转蛇形 |
| `orm.camelize(text)` | 蛇形转驼峰 |

## 返回值

### `orm.init()` 返回值

返回 ORM 运行时对象，常见字段：

- `DataTypes`
- `database`
- `models`
- `associations`

### `orm.defineModel()` 返回值

模型对象常见字段：

- `name`
- `tableName`
- `fields`
- `fieldList`
- `timestamps`
- `database`
- `associations`

并附带占位关联方法：

- `hasOne()`
- `belongsTo()`
- `hasMany()`
- `belongsToMany()`

### `createModel()` / `updateModel()` / `deleteModel()` / `buildQuery()` 返回值

| 字段 | 类型 | 说明 |
|---|---|---|
| `sql` | `string` | 生成的 SQL 语句 |
| `params` | `array` | 占位参数数组 |

### `migrationHelper()` 返回值

返回轻量对象：

- `createTable(model)`
- `sync()`

当前实现主要用于接口统一和脚本占位，返回值较简化。

## 最小示例

```javascript
import { db } from "db";
import * as orm from "orm";

function main() {
  let conn = db.sqlite(":memory:");
  let runtimeORM = orm.init(conn);
  print(runtimeORM != null);

  let User = orm.defineModel("User", {
    name: { type: orm.DataTypes.STRING, notNull: true },
    age: { type: orm.DataTypes.INTEGER, defaultValue: 0 }
  });

  let insertStmt = orm.createModel(User, {
    name: "alice",
    age: 18
  });
  print(insertStmt.sql);

  let qb = orm.QueryBuilder(User)
    .where({ name: "alice" })
    .order("age", "DESC")
    .limit(10);
  let query = orm.buildQuery(qb);
  print(query.sql);
  print(query.params);
  conn.close();
}
```

## 代理/网关场景示例

```javascript
import { db } from "db";
import * as orm from "orm";

let conn = db.sqlite(":memory:");
let runtimeORM = orm.init(conn);

let Route = orm.defineModel("RouteRule", {
  path: { type: orm.DataTypes.STRING, notNull: true, unique: true },
  target: { type: orm.DataTypes.STRING, notNull: true },
  enabled: { type: orm.DataTypes.BOOLEAN, defaultValue: 1 }
});

let insertRoute = orm.createModel(Route, {
  path: "/orders",
  target: "billing",
  enabled: 1
});

let selectRoute = orm.buildQuery(
  orm.QueryBuilder(Route)
    .where({ enabled: 1 })
    .order("path", "ASC")
    .limit(20)
);

print(runtimeORM.database != null);
print(insertRoute.sql);
print(selectRoute.sql);
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


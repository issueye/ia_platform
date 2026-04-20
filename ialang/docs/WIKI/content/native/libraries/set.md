# Set

- 上级目录：[工具类库](tools)
- 导入：`import * as setlib from "set";`

## 模块定位

`set` 适合处理标签集合、权限集合、允许路径集合和去重后的节点集合。

## 接口说明

- 核心入口：`union`、`intersect`、`diff`、`has`
- `set` 用数组表达集合，适合处理标签集合、权限集合、允许路径集合和去重后的节点集合。
- 当前实现按“类型 + 值”做去重键，因此适合标量值集合处理。
- 这类函数通常与 `array` 共同组成配置预处理链路。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `set.union(a, b)` | 2 个数组 | 并集；保留首次出现顺序 |
| `set.intersect(a, b)` | 2 个数组 | 交集；结果顺序按第一个数组 |
| `set.diff(a, b)` | 2 个数组 | 差集；返回在 `a` 中但不在 `b` 中的项 |
| `set.has(arr, value)` | 数组 + 值 | 集合中是否存在该值 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `union()` | 数组 |
| `intersect()` | 数组 |
| `diff()` | 数组 |
| `has()` | `bool` |

## 最小示例

```javascript
import * as setlib from "set";

function main() {
  print(setlib.union([1, 2], [2, 3]));
  print(setlib.intersect([1, 2, 3], [2, 4]));
  print(setlib.diff([1, 2, 3], [2]));
}
```

## 代理/网关场景示例

```javascript
import * as setlib from "set";

let required = ["trace", "auth"];
let enabled = ["auth", "cache"];
print(setlib.union(required, enabled));
print(setlib.intersect(required, enabled));
print(setlib.has(enabled, "cache"));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


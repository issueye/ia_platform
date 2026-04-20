# Array

- 上级目录：[工具类库](tools)
- 导入：`import * as array from "array";`

## 模块定位

`array` 适合处理上游列表、路由表、节点池和批量任务结果。

## 接口说明

- 核心入口：`range`、`concat`、`from`、`of`、`map`、`filter`、`find`、`reduce`、`slice`、`flat`
- `array` 是列表整形模块，适合处理上游列表、路由表、节点池和批量任务结果。
- 当前实现既有构造类函数，也有回调类函数，适合把“筛选、映射、聚合”写成显式流水线。
- 和 `sort`、`set` 搭配时，可以快速做去重、切片、排序和分页。

## 参数要点

### 构造与检测

| 入口 | 参数 | 说明 |
|---|---|---|
| `array.concat(...items)` | 多个数组或普通值 | 顺序拼接；普通值会直接追加 |
| `array.range(end)` | 终点 | 生成 `[0, end)` |
| `array.range(start, end, [step])` | 起点、终点、可选步长 | 支持正负步长 |
| `array.from(value)` | 数组或字符串 | 数组会复制；字符串会拆成字符数组 |
| `array.isArray(value)` | 任意值 | 是否数组 |
| `array.of(...items)` | 任意值 | 直接构造数组 |

### 回调类函数

这些回调统一接收：

- `item`
- `index`
- `array`

| 入口 | 参数 | 说明 |
|---|---|---|
| `array.map(arr, callback)` | 数组 + 回调 | 映射为新数组 |
| `array.filter(arr, callback)` | 数组 + 回调 | 过滤出 truthy 项 |
| `array.find(arr, callback)` | 数组 + 回调 | 返回首个命中项 |
| `array.findIndex(arr, callback)` | 数组 + 回调 | 返回首个命中索引 |
| `array.forEach(arr, callback)` | 数组 + 回调 | 遍历执行 |
| `array.some(arr, callback)` | 数组 + 回调 | 任一命中则 `true` |
| `array.every(arr, callback)` | 数组 + 回调 | 全部命中则 `true` |
| `array.reduce(arr, callback, [initial])` | 数组 + 回调 + 可选初值 | 聚合结果 |
| `array.flatMap(arr, callback)` | 数组 + 回调 | 映射后扁平一层 |

### 查询与变换

| 入口 | 参数 | 说明 |
|---|---|---|
| `array.includes(arr, value)` | 数组 + 值 | 是否包含 |
| `array.indexOf(arr, value)` | 数组 + 值 | 首次出现位置 |
| `array.lastIndexOf(arr, value)` | 数组 + 值 | 最后出现位置 |
| `array.flat(arr, [depth])` | 数组 + 可选层级 | 扁平化 |
| `array.slice(arr, [start], [end])` | 数组 + 可选起止 | 返回子数组 |
| `array.splice(arr, start, [deleteCount], [...items])` | 数组 + 起点 + 删除数 + 可选插入项 | 返回被删除项 |
| `array.join(arr, [sep])` | 数组 + 可选分隔符 | 拼成字符串 |
| `array.sort(arr)` | 数组 | 按内建比较逻辑升序返回新数组 |
| `array.reverse(arr)` | 数组 | 反转返回新数组 |
| `array.shuffle(arr)` | 数组 | 返回打乱后的新数组 |
| `array.fill(arr, value)` | 数组 + 值 | 返回同长度填充值数组 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `concat/range/from/of/map/filter/flat/flatMap/slice/sort/reverse/shuffle/fill()` | 数组 |
| `isArray/includes/some/every()` | `bool` |
| `find()` | 命中项或 `nil` |
| `findIndex/indexOf/lastIndexOf()` | `number` |
| `forEach()` | `true` |
| `reduce()` | 聚合结果 |
| `splice()` | 被删除项数组 |
| `join()` | 字符串 |

## 最小示例

```javascript
import * as array from "array";

function main() {
  let nums = array.range(1, 5);
  print(array.join(nums, "-"));
  print(array.map(nums, (n) => n * 2));
  print(array.reduce(nums, (acc, n) => acc + n, 0));
}
```

## 代理/网关场景示例

```javascript
import * as array from "array";

let upstreams = array.of("a.internal", "b.internal", "b.internal", "c.internal");
let active = array.filter(upstreams, (name) => name != "c.internal");

print(array.join(active, ","));
print(array.findIndex(active, (name) => name == "b.internal"));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


# Sort

- 上级目录：[工具类库](tools)
- 导入：`import * as sort from "sort";`

## 模块定位

`sort` 用于排序和去重，适合整理路由优先级、候选节点或聚合结果。

## 接口说明

- 核心入口：`asc`、`desc`、`reverse`、`unique`
- 这些函数都会返回新数组，不会原地修改输入。
- `asc/desc` 目前只支持 `number[]` 或 `string[]`；混合类型会报错。
- 和 `array`、`set` 一起使用时，数据整形会更顺手。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `sort.asc(arr)` | 数字数组或字符串数组 | 升序排序 |
| `sort.desc(arr)` | 数字数组或字符串数组 | 降序排序 |
| `sort.reverse(arr)` | 任意数组 | 反转顺序 |
| `sort.unique(arr)` | 任意数组 | 去重并保留首次出现顺序 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `asc()` | 数组 |
| `desc()` | 数组 |
| `reverse()` | 数组 |
| `unique()` | 数组 |

## 最小示例

```javascript
import * as sort from "sort";

function main() {
  print(sort.asc([5, 2, 9, 2]));
  print(sort.unique([1, 1, 2, 3, 3]));
}
```

## 代理/网关场景示例

```javascript
import * as sort from "sort";

print(sort.asc([30, 10, 20]));
print(sort.unique(["a", "a", "b"]));
print(sort.desc([5, 2, 9]));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


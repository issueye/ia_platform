# Rand

- 上级目录：[工具类库](tools)
- 导入：`import * as rand from "rand";`

## 模块定位

`rand` 适合做简单采样、随机路由、测试流量标识和临时字符串生成。

## 接口说明

- 核心入口：`int`、`float`、`pick`、`string`
- 常用在测试、采样、轻量随机选择和临时标识生成场景。
- 如果代理需要从多个候选中随机挑一个上游，这个模块很直接。
- 安全相关随机值应先确认场景；这里是通用伪随机工具，不应替代密码学强随机。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `rand.int()` | 无 | 返回一个较大的非负整数 |
| `rand.int(max)` | 正整数 | 返回 `[0, max)` 区间随机整数 |
| `rand.int(min, max)` | 2 个整数 | 返回 `[min, max)` 区间随机整数 |
| `rand.float()` | 无 | 返回 `0-1` 间随机浮点数 |
| `rand.float(min, max)` | 2 个数字 | 返回 `[min, max)` 区间随机浮点数 |
| `rand.pick(arr)` | 非空数组 | 随机挑选一个元素 |
| `rand.string(length, [charset])` | 长度 + 可选字符集 | 生成随机字符串 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `int()` | `number` |
| `float()` | `number` |
| `pick()` | 数组中的任意元素 |
| `string()` | 字符串 |

## 最小示例

```javascript
import * as rand from "rand";

function main() {
  print(rand.int(100));
  print(rand.float(1, 10));
  print(rand.pick(["red", "green", "blue"]));
  print(rand.string(8));
  print(rand.string(6, "ABC123"));
}
```

## 代理/网关场景示例

```javascript
import * as rand from "rand";

let upstream = rand.pick(["a.internal", "b.internal", "c.internal"]);
print(upstream);
print(rand.string(12));
print(rand.int(10, 100));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


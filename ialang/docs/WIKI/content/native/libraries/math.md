# Math

- 上级目录：[工具类库](tools)
- 导入：`import * as math from "math";`

## 模块定位

`math` 提供数值计算能力，适合在代理脚本里做权重计算、超时回退、采样比例和简单统计。

## 接口说明

- 核心入口：`abs`、`ceil`、`floor`、`round`、`sqrt`、`pow`、`max`、`min`、`mod`、`random`、`log`、`log10`
- 这类函数通常是纯计算，不依赖外部状态，适合配置预处理和响应整形阶段。
- 如果代理需要做加权路由、随机抖动、限流窗口换算或归一化，`math` 会很常见。
- 模块同时暴露常量 `PI`、`E`、`sqrt2`。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `math.abs(x)` | 数字 | 绝对值 |
| `math.ceil(x)` | 数字 | 向上取整 |
| `math.floor(x)` | 数字 | 向下取整 |
| `math.round(x)` | 数字 | 四舍五入 |
| `math.sqrt(x)` | 数字 | 平方根 |
| `math.pow(base, exp)` | 2 个数字 | 幂运算 |
| `math.max(a, b)` | 2 个数字 | 较大值 |
| `math.min(a, b)` | 2 个数字 | 较小值 |
| `math.mod(a, b)` | 2 个数字 | 取模；`b` 不能为 `0` |
| `math.random()` | 无 | 返回 `0-1` 间随机浮点数 |
| `math.random(min, max)` | 2 个数字 | 返回 `[min, max)` 区间随机浮点数 |
| `math.log(x)` | 数字 | 自然对数 |
| `math.log10(x)` | 数字 | 以 10 为底对数 |
| `math.sin/cos/tan(x)` | 数字 | 三角函数 |

## 返回值

| 入口 | 返回值 |
|---|---|
| 所有函数 | `number` |
| `PI` / `E` / `sqrt2` | `number` 常量 |

## 最小示例

```javascript
import * as math from "math";

function main() {
  print(math.sqrt(16));
  print(math.pow(2, 8));
  print(math.PI);
  print(math.round(3.6));
}
```

## 代理/网关场景示例

```javascript
import { min, max } from "math";

let latency = 128;
let clamped = min(max(latency, 50), 300);
print(clamped);
print((clamped / 300) * 100);
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


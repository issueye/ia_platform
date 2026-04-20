# Regexp

- 上级目录：[工具类库](tools)
- 导入：`import * as regexp from "regexp";`

## 模块定位

`regexp` 适合路径重写、Header 提取、日志过滤和规则匹配，是代理路由逻辑里的高频模块。

## 接口说明

- 核心入口：`compile`、`test`、`find/findAll`、`replaceAll`、`split`、`findSubmatch/findAllSubmatch`
- `compile()` 适合重复使用的模式；一次性匹配可以直接调用模块函数。
- 当前实现支持 flags `i`、`m`、`s`，并带有编译缓存，适合高频路由匹配。
- 最常见场景是按路径提取资源 ID、版本号、租户前缀和头部中的标识片段。

## 参数要点

### 模块级函数

| 入口 | 参数 | 说明 |
|---|---|---|
| `regexp.compile(pattern, [flags])` | 模式 + 可选 flags | 编译后返回正则对象 |
| `regexp.test(pattern, text, [flags])` | 模式 + 文本 + 可选 flags | 是否匹配 |
| `regexp.find(pattern, text, [flags])` | 模式 + 文本 + 可选 flags | 返回第一个匹配字符串 |
| `regexp.findAll(pattern, text, [n], [flags])` | 模式 + 文本 + 可选数量 + 可选 flags | 返回所有匹配 |
| `regexp.replaceAll(pattern, text, replacement, [flags])` | 模式 + 文本 + 替换串 + 可选 flags | 全部替换 |
| `regexp.split(pattern, text, [n], [flags])` | 模式 + 文本 + 可选数量 + 可选 flags | 按模式切分 |
| `regexp.findSubmatch(pattern, text, [flags])` | 模式 + 文本 + 可选 flags | 返回首个完整匹配及捕获组 |
| `regexp.findAllSubmatch(pattern, text, [n], [flags])` | 模式 + 文本 + 可选数量 + 可选 flags | 返回所有捕获组结果 |
| `regexp.quoteMeta(text)` | 普通文本 | 转义元字符，生成安全模式 |

### 支持的 flags

| flag | 说明 |
|---|---|
| `i` | 忽略大小写 |
| `m` | 多行模式 |
| `s` | 让 `.` 匹配换行 |

### `regexp.compile()` 返回对象

编译后的对象字段和方法：

- `pattern`
- `flags`
- `test(text)`
- `find(text)`
- `findAll(text, [n])`
- `split(text, [n])`
- `replaceAll(text, replacement)`
- `findSubmatch(text)`
- `findAllSubmatch(text, [n])`

## 返回值

| 入口 | 返回值 |
|---|---|
| `test()` | `bool` |
| `find()` | 第一个匹配字符串，未匹配时为 `nil` |
| `findAll()` | 匹配字符串数组 |
| `replaceAll()` | 替换后的字符串 |
| `split()` | 字符串数组 |
| `findSubmatch()` | 数组，包含完整匹配和各捕获组；未匹配时为 `nil` |
| `findAllSubmatch()` | 二维数组，每一项为一次匹配的完整结果和捕获组 |
| `quoteMeta()` | 转义后的模式字符串 |

## 最小示例

```javascript
import * as regexp from "regexp";

function main() {
  let re = regexp.compile("(\\w+)-(\\d+)", "i");
  print(re.test("ITEM-42"));
  print(re.find("item-42"));
  print(regexp.findAllSubmatch("(\\w+)-(\\d+)", "item-42 task-7"));
}
```

## 代理/网关场景示例

```javascript
import * as regexp from "regexp";

let re = regexp.compile("^/api/v(\\d+)/(.*)$");
let matched = re.findSubmatch("/api/v1/orders");

print(matched[1]);
print(matched[2]);
print(re.replaceAll("/api/v1/orders", "/internal/$2"));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


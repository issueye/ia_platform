# ialang 语言概览

`ialang` 是一个用 Go 实现的、语法风格接近 TypeScript / JavaScript 的脚本语言原型。当前已经具备“可写业务脚本、可导入模块、可调用原生库、可打包分发”的完整闭环。

## 语言特性

当前实现已覆盖这些核心能力：

- 变量声明与赋值：`let name = expr;`
- 数组与对象：`[1, 2, 3]`、`{name: "demo"}`
- 解构：`let [a, b] = arr;`、`let {x, y: z} = obj;`
- 函数与闭包：`function`、`async function`、剩余参数、词法作用域
- 控制流：`if`、`while`、`for`、`break`、`continue`
- 运算符：算术、比较、逻辑、位运算、复合赋值、三元表达式
- 类与继承：`class`、`extends`、`super`、实例方法
- 错误处理：`throw`、`try/catch/finally`
- 异步：`await`、`Promise.all`、`Promise.race`、`Promise.allSettled`
- 模块系统：`import`、`export`、`export default`、动态 `import()`

## 模块导入

本地模块使用相对路径：

```javascript
import { greet } from "./modules/utils/index";
```

如果项目在 `pkg.toml` 中配置了根别名，也可以从项目根目录开始导入：

```toml
[imports]
root_alias = "@"
```

```javascript
import { greet } from "@/modules/utils/index";
```

原生库支持三种导入形式：

```javascript
import { readFile } from "fs";
import { readFile } from "@std/fs";
import { readFile } from "@stdlib/fs";
```

这三种写法在当前运行时中等价。

## 执行模型

`ialang` 当前的执行链路是：

1. 源码经过词法分析与语法分析生成 AST
2. AST 编译为字节码 chunk
3. 虚拟机执行字节码

这意味着它既支持直接运行源码，也支持先构建 `.iapkg` 包，再由 `run-pkg` 或独立二进制执行。

## 建议阅读顺序

- [快速开始](../getting-started)
- [命令与项目结构](usage)
- [原生库概览](../native)

如果你要确认某段语法是否已实现，优先查看：

- [`ialang/README.md`](/E:/code/issueye/ialang/ialang/README.md)
- [`ialang/docs/language-spec.md`](/E:/code/issueye/ialang/ialang/docs/language-spec.md)

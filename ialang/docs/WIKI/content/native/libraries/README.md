# 原生库目录

这一组页面把 `ialang` 原生库进一步细化到“按库查看”的粒度。与上层分类页不同，这里每个库都会单独列出：

- 推荐导入方式
- 常用 API 入口
- 文档内完整示例代码

## 导入规则

大多数原生库都支持三种等价写法：

```javascript
import * as fs from "fs";
import * as fs from "@std/fs";
import * as fs from "@stdlib/fs";
```

库级页面默认展示 plain 导入名，便于直接抄写最短形式。

## 别名与兼容名

- `db` 也可以写成 `database`
- `asset` 也可以写成 `bundle`
- `iax` 也可以写成 `interaction`
- `exec` 也可以写成 `os/exec`
- `timer` 额外保留 plain 兼容名 `setTimeout`
- `Promise` 的模块名区分大小写，示例统一使用 `Promise`

## 目录

- [网络类库](network)
- [系统类库](system)
- [数据类库](data)
- [工具类库](tools)

## 使用方式

单库页面现在直接内嵌了最小示例和代理/网关场景示例，阅读时不需要再跳到外部示例文件。

推荐阅读方式：

- 先从分组目录页进入
- 再打开单库页查看接口说明
- 直接参考页面中的代码块理解用法和组合方式

# Wiki Examples

这个目录存放和 `ialang_wiki` 文档对应的示例代码。

## 文件结构

- `http-and-network.ia`
- `system-and-runtime.ia`
- `data-and-utils.ia`
- `libraries/README.md`

前三个文件按文档分类聚合示例，`libraries/` 目录则按“每个原生库一个最小示例”继续拆分。

## 在仓库根目录运行

```bash
go run ./ialang/cmd/ialang check ./ialang/examples/wiki_examples/http-and-network.ia
go run ./ialang/cmd/ialang run ./ialang/examples/wiki_examples/http-and-network.ia

go run ./ialang/cmd/ialang check ./ialang/examples/wiki_examples/system-and-runtime.ia
go run ./ialang/cmd/ialang run ./ialang/examples/wiki_examples/system-and-runtime.ia

go run ./ialang/cmd/ialang check ./ialang/examples/wiki_examples/data-and-utils.ia
go run ./ialang/cmd/ialang run ./ialang/examples/wiki_examples/data-and-utils.ia
```

## 使用说明

- 每个聚合文件都对应 wiki 的一个分类页
- `main()` 默认只执行一部分更安全、可重复运行的演示
- 更容易阻塞、监听端口或产生额外副作用的示例，通常保留为文件里的独立函数
- 如果你想逐个看库，继续进入 `libraries/`

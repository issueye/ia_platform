# Asset

- 上级目录：[系统类库](system)
- 导入：`import * as asset from "asset";`
- 别名导入：`bundle`

## 模块定位

`asset` / `bundle` 用于处理静态资源打包、压缩、版本化和目录分析，适合给代理控制台、嵌入式前端和静态发布目录生成可分发资源。

## 接口说明

- 核心入口：`bundle`、`minify`、`hash`、`version`、`clean`、`analyze`
- `bundle(outputDir, options)` 读取文件列表，按需压缩、加哈希并写入目标目录
- `minify(content, type)` 直接压缩单段 `css/js/html`
- `hash(content, [length])` 计算内容哈希，默认返回 8 位短哈希和完整 SHA-256
- `version(filePath)` 基于现有文件内容生成版本文件名和查询串
- `clean(dir, [pattern])` 删除目录中匹配模式的文件；`analyze(dir)` 统计文件数、大小和扩展名分布

## 参数要点

### `asset.bundle(outputDir, options)`

| 字段 | 类型 | 说明 |
|---|---|---|
| `outputDir` | `string` | 输出目录，不存在会自动创建 |
| `options.files` | `array<string>` | 要处理的输入文件列表 |
| `options.minify` | `bool` | 是否压缩，默认 `true` |
| `options.hash` | `bool` | 是否生成哈希文件名，默认 `true` |
| `options.concat` | `bool` | 是否按首个文件扩展名合并同类型文件，默认 `false` |

### 其他入口

| 入口 | 参数 | 说明 |
|---|---|---|
| `asset.minify(content, type)` | `string`, `string` | `type` 仅支持 `css/js/javascript/html/htm` |
| `asset.hash(content, [length])` | `string`, `number` | `length` 默认 `8` |
| `asset.version(filePath)` | `string` | 读取现有文件生成版本名 |
| `asset.clean(dir, [pattern])` | `string`, `string` | `pattern` 默认 `*` |
| `asset.analyze(dir)` | `string` | 递归分析目录 |

## 返回值

### `asset.bundle()`

| 字段 | 说明 |
|---|---|
| `manifest` | 以原始文件路径为 key 的清单对象 |
| `files` | 已处理文件数组，每项含 `input/output/hash/size/minified` |
| `output` | 输出目录 |
| `count` | 输入文件数量 |

若启用 `concat`，`manifest["__bundle__"]` 还会包含：

- `output`
- `hash`
- `size`
- `files`

### 其他入口

| 入口 | 返回值 |
|---|---|
| `minify()` | `{ original, minified, content, ratio }` |
| `hash()` | `{ hash, full, length }` |
| `version()` | `{ path, hash, versioned, query }` |
| `clean()` | `{ removed, count }` |
| `analyze()` | `{ totalSize, fileCount, byExtension, files }` |

## 最小示例

```javascript
import * as asset from "asset";

function main() {
  let minified = asset.minify("body { color: red; }", "css");
  print(minified.content);
  print(minified.ratio);

  let hashed = asset.hash(minified.content, 12);
  print(hashed.hash);
  print(hashed.full);
}
```

## 代理/网关场景示例

```javascript
import * as asset from "asset";
import * as fs from "fs";

function main() {
  fs.mkdir("./tmp/wiki-assets");
  fs.writeFile("./tmp/wiki-assets/app.js", "const hello = 1; // keep");
  fs.writeFile("./tmp/wiki-assets/app.css", "body { color: #333; }");

  let build = asset.bundle("./tmp/wiki-assets/dist", {
    files: [
      "./tmp/wiki-assets/app.js",
      "./tmp/wiki-assets/app.css"
    ],
    minify: true,
    hash: true,
    concat: false
  });

  print(build.output);
  print(build.count);
  print(build.files[0].output);

  let report = asset.analyze("./tmp/wiki-assets/dist");
  print(report.fileCount);
  print(report.totalSize);
}
```

## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。

# MIME

- 上级目录：[数据类库](data)
- 导入：`import * as mime from "mime";`

## 模块定位

`mime` 用于在路径、扩展名和内容之间推断媒体类型，适合静态资源代理、上传校验、对象存储网关和下载服务。

## 接口说明

- 核心入口：`typeByExt`、`extByType`、`detectType`、`detectByPath`
- `typeByExt(ext)` 根据扩展名查 MIME；没有点号时会自动补 `.` 前缀
- `extByType(mimeType)` 返回该类型关联的扩展名数组
- `detectType(text)` 使用内容片段做探测，返回类似 `text/plain; charset=utf-8`
- `detectByPath(path)` 只依据文件扩展名推断，不会读取文件内容

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `mime.typeByExt(ext)` | `string` | 如 `.json`、`json`、`.html` |
| `mime.extByType(mimeType)` | `string` | 如 `text/html` |
| `mime.detectType(text)` | `string` | 按内容探测 MIME |
| `mime.detectByPath(path)` | `string` | 按路径扩展名探测 MIME |

## 返回值

| 入口 | 返回值 |
|---|---|
| `typeByExt()` | 字符串或 `null` |
| `extByType()` | 扩展名数组 |
| `detectType()` | 字符串 |
| `detectByPath()` | 字符串或 `null` |

说明：

- `typeByExt()` 与 `detectByPath()` 找不到已知类型时返回 `null`
- `extByType()` 可能返回多个扩展名，例如同一 MIME 的不同变体
- `detectType()` 基于传入文本内容，不适合替代二进制文件头完整探测

## 最小示例

```javascript
import * as mime from "mime";

function main() {
  print(mime.typeByExt("json"));
  print(mime.extByType("text/html"));
  print(mime.detectType("{\"ok\":true}"));
  print(mime.detectByPath("./public/app.css"));
}
```

## 代理/网关场景示例

```javascript
import * as mime from "mime";
import * as path from "path";

function main() {
  let assetPath = "/gateway/static/app.js";
  let ext = path.ext(assetPath);
  let contentType = mime.typeByExt(ext);

  if (contentType == null) {
    contentType = "application/octet-stream";
  }

  print(ext);
  print(contentType);
  print(mime.detectType("console.log('proxy panel');"));
}
```

## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。

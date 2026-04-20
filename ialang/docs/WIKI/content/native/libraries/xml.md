# XML

- 上级目录：[数据类库](data)
- 导入：`import * as xml from "xml";`

## 模块定位

`xml` 适合处理旧系统接口、配置交换格式或需要兼容 SOAP/XML 生态的代理脚本。

## 接口说明

- 核心入口：`parse`、`fromFile`、`stringify`、`saveToFile`、`valid`、`escape`
- `xml` 适合处理旧系统接口、配置交换格式或需要兼容 SOAP/XML 生态的代理脚本。
- 当前实现把 XML 文档解析为显式节点对象，而不是直接按标签名铺平成普通对象。
- 如果代理需要兼容 XML 上游，建议先验证结构，再按 `name/attrs/text/children` 读取。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `xml.parse(text)` | XML 字符串 | 解析为节点对象 |
| `xml.fromFile(path)` | 文件路径 | 读取并解析 XML |
| `xml.stringify(node, [pretty])` | 节点对象 + 可选布尔 | 序列化为 XML 字符串 |
| `xml.saveToFile(node, path, [pretty])` | 节点对象 + 路径 + 可选布尔 | 序列化并写入文件 |
| `xml.valid(text)` | XML 字符串 | 检查是否合法 |
| `xml.escape(text)` | 普通字符串 | 转义 XML 文本 |

### 节点对象结构

`xml.parse()` / `xml.fromFile()` 返回的节点对象字段：

- `name`
- `attrs`
- `text`
- `children`

## 返回值

| 入口 | 返回值 |
|---|---|
| `parse()/fromFile()` | 节点对象 |
| `stringify()` | XML 字符串 |
| `saveToFile()` | `true` |
| `valid()` | `bool` |
| `escape()` | 字符串 |

## 最小示例

```javascript
import * as xml from "xml";

function main() {
  let doc = xml.parse("<user><name>alice</name></user>");
  print(doc.name);
  print(doc.children[0].name);
  print(doc.children[0].text);
  print(xml.valid("<root></root>"));
}
```

## 代理/网关场景示例

```javascript
import * as xml from "xml";

let doc = xml.parse("<route><name>billing</name><port>9000</port></route>");
print(doc.name);
print(doc.children[0].text);
print(xml.valid("<proxy></proxy>"));
print(xml.stringify(doc, true));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 数据类模块通常最适合和 `http`、`fs`、`db`、`log` 这些基础模块组合。
- 如果脚本要处理外部输入，优先在解析、校验和编码阶段收敛数据格式。


# 数据类库

这一页收拢“格式、编码、安全、标识、数据库、压缩”相关原生库。它们更偏向数据表达、解析、转换、校验和持久化。

## 配置与结构化格式

### json

- 文档页：[JSON](json)
- 导入：`import * as json from "json";`
- 常用入口：`parse`、`fromFile`、`stringify`、`saveToFile`、`valid`

### yaml

- 文档页：[YAML](yaml)
- 导入：`import * as yaml from "yaml";`
- 常用入口：`parse`、`stringify`

### toml

- 文档页：[TOML](toml)
- 导入：`import * as toml from "toml";`
- 常用入口：`parse`、`stringify`

### xml

- 文档页：[XML](xml)
- 导入：`import * as xml from "xml";`
- 常用入口：`parse`、`stringify`、`valid`

### csv

- 文档页：[CSV](csv)
- 导入：`import * as csv from "csv";`
- 常用入口：`parse`、`stringify`

## 编码、字节与内容识别

### encoding

- 文档页：[Encoding](encoding)
- 导入：`import * as encoding from "encoding";`
- 常用入口：`base64Encode`、`base64Decode`、`urlEncode`、`urlDecode`

### hex

- 文档页：[Hex](hex)
- 导入：`import * as hex from "hex";`
- 常用入口：`encode`、`decode`

### bytes

- 文档页：[Bytes](bytes)
- 导入：`import * as bytes from "bytes";`
- 常用入口：`fromString`、`toString`、`toBase64`

### mime

- 文档页：[MIME](mime)
- 导入：`import * as mime from "mime";`
- 常用入口：`typeByExt`、`extByType`、`detectType`、`detectByPath`

### compress

- 文档页：[Compress](compress)
- 导入：`import * as compress from "compress";`
- 常用入口：`gzipCompress`、`gzipDecompress`、`zlibCompress`、`zlibDecompress`

## 安全、签名与标识

### crypto

- 文档页：[Crypto](crypto)
- 导入：`import * as crypto from "crypto";`
- 常用入口：`sha256`、`md5`

### hash

- 文档页：[Hash](hash)
- 导入：`import * as hash from "hash";`
- 常用入口：`sha1`、`sha512`、`crc32`

### hmac

- 文档页：[HMAC](hmac)
- 导入：`import * as hmac from "hmac";`
- 常用入口：`sha256`、`verifySha256`

### uuid

- 文档页：[UUID](uuid)
- 导入：`import * as uuid from "uuid";`
- 常用入口：`v4`、`isValid`

### url

- 文档页：[URL](url)
- 导入：`import * as url from "url";`
- 常用入口：`parse`、`escape`、`unescape`、`queryEncode`、`queryDecode`

## 数据存储

### db / database

- 文档页：[DB](db)
- 导入：`import { db } from "db";`
- 别名导入：`database`
- 常用入口：`connect`、`sqlite`、`mysql`、`postgres`、`sqlserver`

### orm

- 文档页：[ORM](orm)
- 导入：`import * as orm from "orm";`
- 常用入口：`init`、`defineModel`、`QueryBuilder`、`buildQuery`

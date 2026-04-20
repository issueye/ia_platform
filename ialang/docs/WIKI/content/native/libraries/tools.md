# 工具类库

这一页收拢“数值、集合、文本、随机、日志、异步编排和高层代理能力”相关原生库。它们更偏向脚本开发效率和流程编排。

## 数值、文本与集合

### math

- 文档页：[Math](math)
- 导入：`import * as math from "math";`
- 常用入口：`abs`、`sqrt`、`pow`、`min`、`max`

### string

- 文档页：[String](string)
- 导入：`import * as strmod from "string";`
- 常用入口：`split`、`join`、`trim`、`replace`

### array

- 文档页：[Array](array)
- 导入：`import * as array from "array";`
- 常用入口：`range`、`chunk`、`flatten`

### sort

- 文档页：[Sort](sort)
- 导入：`import * as sort from "sort";`
- 常用入口：`asc`、`desc`、`unique`

### set

- 文档页：[Set](set)
- 导入：`import * as setlib from "set";`
- 常用入口：`union`、`intersect`、`diff`

### strconv

- 文档页：[Strconv](strconv)
- 导入：`import * as strconv from "strconv";`
- 常用入口：`atoi`、`itoa`、`parseBool`、`formatFloat`

### rand

- 文档页：[Rand](rand)
- 导入：`import * as rand from "rand";`
- 常用入口：`int`、`float`、`pick`、`string`

### regexp

- 文档页：[Regexp](regexp)
- 导入：`import * as regexp from "regexp";`
- 常用入口：`compile`、`findAllSubmatch`

## 日志、异步与代理能力

### log

- 文档页：[Log](log)
- 导入：`import * as log from "log";`
- 常用入口：`info`、`warn`、`error`、`setJSON`

### Promise

- 文档页：[Promise](promise)
- 导入：`import { all, race, allSettled } from "Promise";`
- 常用入口：`all`、`race`、`allSettled`

### @agent/sdk

- 文档页：[@agent/sdk](agent-sdk)
- 导入：`import { llm, tool, memory } from "@agent/sdk";`
- 常用入口：`llm.chat`、`llm.chatAsync`、`tool.call`、`memory.get`

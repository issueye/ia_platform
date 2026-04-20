# 数据与工具

本文汇总常用的数据处理、编码、安全和工具型模块。

## 数值与集合

### math

常用函数：

- `abs`
- `ceil`
- `floor`
- `round`
- `sqrt`
- `pow`
- `max`
- `min`
- `mod`
- `random`
- `log`
- `log10`
- `sin` / `cos` / `tan`

### array

常用方法：

- `sort()`
- `reverse()`
- `includes(val)`
- `join([sep])`
- `slice(start, [end])`
- `flat([depth])`
- `concat(...items)`
- `push(...vals)`
- `pop()`
- `shift()`
- `unshift(...vals)`
- `at(idx)`
- `fill(val)`
- `shuffle()`

### set / sort

分别用于集合和排序辅助能力。

## 字符串与格式

### string

既支持原型方法，也支持模块函数。

### strconv

适合做更明确的字符串数值转换。

## 编码与序列化

- `json`
- `yaml`
- `toml`
- `xml`
- `csv`
- `encoding`
- `bytes`
- `hex`
- `compress`

## 安全与标识

- `crypto`
- `hash`
- `hmac`
- `uuid`
- `rand`
- `regexp`
- `url`
- `mime`

### regexp

已知支持：

- `compile(pattern, [flags])`
- `findSubmatch`
- `findAllSubmatch`

## 更高层模块

- `express`
- `db` / `database` / `orm`
- `asset` / `bundle`
- `Promise`
- `@agent/sdk`
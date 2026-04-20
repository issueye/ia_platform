# Net

- 上级目录：[网络类库](network)
- 导入：`import * as net from "net";`

## 模块定位

`net` 提供地址、DNS 和 CIDR 工具，适合在代理脚本里做路由拼装、白名单校验和上游解析。

## 接口说明

- 核心入口：`isIP`、`isIPv4`、`isIPv6`、`parseHostPort`、`joinHostPort`、`lookupIP`、`parseCIDR`、`containsCIDR`
- 这类函数偏纯工具函数，不维护连接状态，适合在真正发起请求前做地址解析和白名单判断。
- 如果代理有来源 IP 白名单、出口规则或服务发现逻辑，`net` 会经常出现。
- `joinHostPort()` 会校验端口范围，`parseCIDR()` 会返回网络与掩码信息。

## 参数要点

| 入口 | 参数 | 说明 |
|---|---|---|
| `net.isIP(text)` | 字符串 | 是否为合法 IP |
| `net.isIPv4(text)` | 字符串 | 是否为 IPv4 |
| `net.isIPv6(text)` | 字符串 | 是否为 IPv6 |
| `net.parseHostPort(addr)` | `host:port` 字符串 | 拆成 `host` 与数值 `port` |
| `net.joinHostPort(host, port)` | 主机 + 端口号 | 合成地址字符串 |
| `net.lookupIP(host)` | 主机名 | DNS 解析为 IP 数组 |
| `net.parseCIDR(cidr)` | CIDR 字符串 | 解析网络和掩码信息 |
| `net.containsCIDR(cidr, ip)` | CIDR + IP | 判断 IP 是否落在网段内 |

## 返回值

| 入口 | 返回值 |
|---|---|
| `isIP()/isIPv4()/isIPv6()/containsCIDR()` | `bool` |
| `joinHostPort()` | 字符串 |
| `lookupIP()` | 字符串数组 |

### `parseHostPort()` 返回值

| 字段 | 类型 | 说明 |
|---|---|---|
| `host` | `string` | 主机部分 |
| `port` | `number` | 数值端口 |

### `parseCIDR()` 返回值

| 字段 | 类型 | 说明 |
|---|---|---|
| `ip` | `string` | 输入 CIDR 中的 IP |
| `network` | `string` | 网络地址 |
| `maskOnes` | `number` | 掩码前缀位数 |
| `maskBits` | `number` | 总位数 |

## 最小示例

```javascript
import * as net from "net";

function main() {
  let hp = net.parseHostPort("127.0.0.1:8080");
  print(hp.host);
  print(hp.port);
  print(net.isIPv4(hp.host));
  print(net.containsCIDR("192.168.1.0/24", "192.168.1.10"));
}
```

## 代理/网关场景示例

```javascript
import * as net from "net";

let target = net.parseHostPort("10.0.0.12:9000");
let upstream = net.joinHostPort(target.host, target.port);

print(upstream);
print(net.containsCIDR("10.0.0.0/24", "10.0.0.12"));
print(net.lookupIP("localhost"));
```


## 使用建议

- 先阅读并理解文档内最小示例，再把它整合进你的脚本主流程。
- 如果模块涉及监听、连接、文件或子进程，优先在开发环境验证资源释放逻辑。
- 需要跨模块组合时，优先和 `fs`、`json`、`log`、`time` 这类基础模块一起使用。


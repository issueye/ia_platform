# ialang

`ialang` is a TS-like scripting language prototype implemented in Go.

**Language Completeness: 92+/100** ✅

## Documentation

- Usage guide: [docs/usage-guide.md](docs/usage-guide.md)
- Language spec: [docs/language-spec.md](docs/language-spec.md)
- Syntax checklist: [docs/syntax-checklist.md](docs/syntax-checklist.md)
- Unsupported features: [docs/unsupported-features.md](docs/unsupported-features.md)
- Built-in modules index: [docs/2026-04-08/README.md](docs/2026-04-08/README.md)

## Features

### Core Syntax

- **Variables**: `let name = expr;`, `name = expr;` (assignment)
- **Destructuring Declarations**: `let [a, b] = arr;`, `let {x, y: z} = obj;`
- **Destructuring Assignment (shallow)**: `[a, b] = arr;`, `({x, y: z} = obj);`
- **Functions**: `function name(a, b) { ... }`, `async function name(a, b) { ... }`, `function f(...args) { ... }`
- **Closures**: nested function + lexical closure capture
- **Return**: `return expr;`
- **Control Flow**:
  - `if (...) { ... } else { ... }`
  - `while (...) { ... }`
  - `for (init; cond; post) { ... }`
  - `break;` / `continue;`
- **Ternary**: `condition ? thenExpr : elseExpr`

### Operators

- **Arithmetic**: `+`, `-`, `*`, `/`, `%`
- **Comparison**: `==`, `!=`, `<`, `>`, `<=`, `>=`
- **Logical**: `&&`, `||` (short-circuit), `!`
- **Bitwise**: `&`, `|`, `^`, `<<`, `>>`
- **Compound Assignment**: `+=`, `-=`, `*=`, `/=`, `%=`
- **Unary**: `-x`, `!x`

### Data Structures

- **Arrays**: `[1, 2, 3]` with prototype methods
- **Objects**: `{name: "x", value: 42}`
- **Index Access**: `arr[0]`, `obj["name"]`, `obj[key]`
- **Literals**: string, number, bool (`true`/`false`), `null`
- **Template Strings**: `` `hello ${name}` ``

### Object-Oriented Programming

- **Class**: `class C { constructor(...) { ... } m(...) { ... } }`
- **Inheritance**: `class Child extends Parent { ... }`
- **Super**: `super()`, `super.method()`
- **Private Fields**: `#x` and `this.#x` (initializer and strict privacy checks are partial)
- **Instantiation**: `new C(...)`, `this.x = ...`, `obj.m()`
- **Method Override**: child classes can override parent methods

### Error Handling

- `throw expr;`
- `try { ... } catch (err) { ... }`
- `try { ... } finally { ... }`
- `try { ... } catch (err) { ... } finally { ... }`
- Structured error types: `IaError` with type hierarchy

### Async/Await

- `await expr` (for Promise values)
- `Promise.all([...])` - wait for all promises
- `Promise.race([...])` - first settled promise
- `Promise.allSettled([...])` - wait for all to settle

### Module System

- **ES6 Import**: `import { x } from "module";`
- **Namespace Import**: `import * as mod from "module";`
- **Local Import**: `import { x } from "./path/module";`
- **Project Root Import**: `import { x } from "@/path/module";` via `pkg.toml` `[imports].root_alias`
- **Export**: `export let x = ...;`, `export function f() { ... }`, `export class C { ... }`
- **Named Export List**: `export { a, b }`, `export { a as b }`
- **Default Export**: `export default expr;`, `export default class Named {}`, `export default function named() {}`
- **Export All**: `export * from "module";` (skips upstream `default`)
- **Dynamic Import**: `import("module")` (returns Promise, use with `await`)
- **Aliases**: `"@std/*"`, `"@stdlib/*"`

### Sandboxing

- Step counter limits (`MaxSteps`)
- Duration limits (`MaxDuration`)
- Module whitelist (`AllowedModules`)
- Feature toggles: `AllowFS`, `AllowNetwork`, `AllowProcess`

---

## Built-in Modules

### `@agent/sdk` (mock)

| Function | Returns | Description |
|---|---|---|
| `llm.chat(prompt)` | string | Mock LLM chat |
| `llm.chatAsync(prompt)` | Promise | Async LLM chat |
| `tool.call(name, ...)` | any | Call a tool |
| `memory.get(key)` | any | Get memory value |

### `math`

| Function | Returns | Description |
|---|---|---|
| `abs(x)` | number | Absolute value |
| `ceil(x)` | number | Round up |
| `floor(x)` | number | Round down |
| `round(x)` | number | Round to nearest |
| `sqrt(x)` | number | Square root |
| `pow(base, exp)` | number | Power |
| `max(a, b)` | number | Maximum |
| `min(a, b)` | number | Minimum |
| `mod(a, b)` | number | Floating-point modulo |
| `random([min, max])` | number | Random number |
| `log(x)` | number | Natural logarithm |
| `log10(x)` | number | Base-10 logarithm |
| `sin(x)`, `cos(x)`, `tan(x)` | number | Trig functions |

| Constant | Value |
|---|---|
| `PI` | 3.14159... |
| `E` | 2.71828... |
| `sqrt2` | 1.41421... |

### `string` (prototype methods)

String prototype methods — use as `"hello".toUpperCase()`:

| Method | Returns | Description |
|---|---|---|
| `split(sep)` | Array | Split string |
| `trim()` | string | Trim whitespace |
| `trimLeft()` | string | Trim left |
| `trimRight()` | string | Trim right |
| `replace(old, new)` | string | Replace all |
| `toLowerCase()` | string | Lowercase |
| `toUpperCase()` | string | Uppercase |
| `startsWith(prefix)` | bool | Check prefix |
| `endsWith(suffix)` | bool | Check suffix |
| `contains(substr)` | bool | Check contains |
| `indexOf(substr)` | number | Find index |
| `repeat(n)` | string | Repeat string |
| `padStart(len, pad)` | string | Pad start |
| `padEnd(len, pad)` | string | Pad end |
| `substring(start, [end])` | string | Extract substring |
| `charAt(idx)` | string | Character at index |
| `slice(start, [end])` | string | Slice string |
| `parseInt()` | number | Parse integer |
| `parseFloat()` | number | Parse float |
| `fromCodePoint(code)` | string | Unicode char |

**Utility functions** (via `string` module):

| Function | Returns | Description |
|---|---|---|
| `string.split(s, sep)` | Array | Split string |
| `string.join(arr, [sep])` | string | Join array |
| `string.parseInt(s)` | number | Parse integer |
| `string.parseFloat(s)` | number | Parse float |
| `string.fromCodePoint(code)` | string | Unicode char |

### `array` (prototype methods)

Array prototype methods — use as `arr.sort()`:

| Method | Returns | Description |
|---|---|---|
| `sort()` | Array | Sort array |
| `reverse()` | Array | Reverse array |
| `includes(val)` | bool | Check contains |
| `join([sep])` | string | Join to string |
| `indexOf(val)` | number | Find index |
| `lastIndexOf(val)` | number | Find last index |
| `slice(start, [end])` | Array | Slice array |
| `flat([depth])` | Array | Flatten nested |
| `concat(...items)` | Array | Concat arrays |
| `push(...vals)` | Array | Append values |
| `pop()` | Array | Remove last |
| `unshift(...vals)` | Array | Prepend values |
| `shift()` | Array | Remove first |
| `at(idx)` | any | Access by index (supports negative) |
| `fill(val)` | Array | Fill with value |
| `shuffle()` | Array | Random shuffle |
| `length` | number | Array length (property) |

**Utility functions** (via `array` module):

| Function | Returns | Description |
|---|---|---|
| `array.concat(...arrs)` | Array | Concatenate arrays |
| `array.range([start], end, [step])` | Array | Generate range |
| `array.from(val)` | Array | Convert to array |
| `array.isArray(val)` | bool | Type check |
| `array.of(...vals)` | Array | Create from args |

### `Promise`

| Function | Returns | Description |
|---|---|---|
| `Promise.all([...])` | Promise | Resolve all |
| `Promise.race([...])` | Promise | First settled |
| `Promise.allSettled([...])` | Promise | All settled |

### `http`

**Client**: `client.request(url, [options])`, `client.get(...)`, `client.post(...)`
**Async**: `client.requestAsync(...)`, `client.getAsync(...)`, `client.postAsync(...)`
**Server**: `server.serve([options])`, `server.serveAsync([options])`, `server.proxy([options])`, `server.forward([options])`

`server.proxy/forward` 支持对象或函数式改写:

- `requestMutations`: object 或 `fn(req) -> object`，可改 method/path/query/header/body
- `responseMutations`: object 或 `fn(resp, req) -> object`，可改 status/header/body

`regexp` 支持 `compile(pattern, [flags])`、`findSubmatch`、`findAllSubmatch`，并可在模块级 API 使用可选 `flags`（`i/m/s`）。

### `ipc`（本机进程间通信）

- `ipc.server.listen([options])` / `listenAsync`
- `ipc.client.connect(addr, [options])` / `connectAsync`
- 连接对象：`send/recv`、`call/reply`，均支持 `Async` 版本
- 请求封装：`ipc.buildRequest(...)`、`ipc.buildResponse(...)`

See `docs/2026-04-08/IPC_ENCAPSULATION.md` for details.

### `iax` (ialang 程序间交互协议)

- `iax.version()`：协议版本（当前 `iax/1`）
- `iax.buildRequest(service, action, payload, [options])`：构建交互请求包
- `iax.call(conn, service, action, payload, [options])`：通过 `ipc` 发起请求并等待响应
- `iax.callAsync(...)`：异步调用
- `iax.buildEvent(topic, payload, [options])`：构建事件
- `iax.publish(conn, topic, payload, [options])`：按主题发布
- `iax.publishAsync(...)`：异步发布
- `iax.subscribe(conn, [topics], [options])`：按主题订阅（`next/nextAsync` 拉取）
- `iax.configurePersistence({ enabled, path })`：配置事件持久化
- `iax.getPersistence()`：查看持久化配置
- `iax.loadEvents([options])`：加载持久化事件（支持 topic/sinceMs/limit）
- `iax.replay(conn, [options])`：将持久化事件回放到连接
- `iax.receive(conn, [options])`：接收请求并解析业务字段
- `iax.reply(conn, recvResult, ok, data, [error])`：回复请求结果

`iax` 传输层可切换：默认推荐 `ipc`，也支持仅提供 `send/recv` 的 websocket/socket 风格连接对象。

See `docs/2026-04-08/IAX_INTERACTION_PROTOCOL.md` for protocol details.

### `fs`

**Sync**: `readFile(path)`, `writeFile(path, content)`, `appendFile(path, content)`, `exists(path)`, `mkdir(path, [recursive])`, `readDir(path)`, `stat(path)`
**Async**: `readFileAsync(path)`, `writeFileAsync(path, content)`, `appendFileAsync(path, content)`

### `exec`

`run(command, [options])`, `runAsync(command, [options])`, `lookPath(name)` (`which(name)` alias)

`options`: `args`, `cwd`, `env`, `stdin`, `timeoutMs`, `shell`

### `log`

封装 Go `slog`：

- `debug/info/warn/error(message, [fields])`
- `log(level, message, [fields])`
- `with(fields)` 返回带默认字段的 logger
- `setLevel(level)`, `getLevel()`
- `setJSON(bool)`, `isJSON()`

### `os`, `process`, `path`, `json`, `time`, `encoding`, `crypto`, `exec`, `log`, `ipc`, `socket`, `iax`

`os` 新增目录原生能力：`userDir()`, `dataDir()`, `configDir()`, `cacheDir()`, `tempDir()`（`tmpDir()` 兼容保留）。
`socket` 提供 TCP/UDP 原生通信能力：`socket.server.listen` / `socket.client.connect` / `socket.udp.bind`（含 Async 版本与 `send/recv`、`sendTo/recvFrom`）。
`http` 新增原生代理与转发能力：`http.server.proxy` / `http.server.forward`，支持 `http.client.*` 的 `proxy` 选项；`proxy` 为手写转发链路（不依赖 `net/http/httputil.ReverseProxy`），并支持对象/函数两种 `requestMutations/responseMutations`。
`regexp` 新增原生编译对象：`regexp.compile(pattern, [flags])`，并支持 `findSubmatch` / `findAllSubmatch`。

See `docs/2026-04-07/NATIVE_MODULES.md` for full API reference.

---

## Execution Pipeline

1. Source → lexer/parser → AST
2. AST → compiler → bytecode chunk
3. Bytecode → VM interpreter

## Run

```bash
go run ./cmd/ialang run examples/hello.ia
```

## Init Project

```bash
go run ./cmd/ialang init myapp
cd myapp
go run ../cmd/ialang run main.ia
```

## Build Package

```bash
go run ./cmd/ialang build examples/package_demo/main.ia -o app.iapkg
go run ./cmd/ialang run-pkg app.iapkg
```

## Build Standalone Binary

```bash
go run ./cmd/ialang build-bin examples/package_demo/main.ia -o package_demo.exe
./package_demo.exe
```

## Check Project Syntax

```bash
# check current project (uses pkg.toml entry, fallback main.ia)
go run ./cmd/ialang check

# check a specific entry file
go run ./cmd/ialang check examples/hello.ia

# check a specific project directory
go run ./cmd/ialang check ./examples/package_demo
```

## Format Source File

```bash
# format a source file in place
go run ./cmd/ialang fmt examples/hello.ia
```

## Examples

| Example | Description | Command |
|---|---|---|
| `hello.ia` | Hello World | `go run ./cmd/ialang run examples/hello.ia` |
| `function.ia` | Functions | `go run ./cmd/ialang run examples/function.ia` |
| `closure.ia` | Closures | `go run ./cmd/ialang run examples/closure.ia` |
| `control.ia` | Control flow | `go run ./cmd/ialang run examples/control.ia` |
| `logic_for.ia` | For loops | `go run ./cmd/ialang run examples/logic_for.ia` |
| `data.ia` | Data structures | `go run ./cmd/ialang run examples/data.ia` |
| `class.ia` | Classes | `go run ./cmd/ialang run examples/class.ia` |
| `inheritance.ia` | Inheritance | `go run ./cmd/ialang run examples/inheritance.ia` |
| `try_catch.ia` | Error handling | `go run ./cmd/ialang run examples/try_catch.ia` |
| `async.ia` | Async/await | `go run ./cmd/ialang run examples/async.ia` |
| `async_loop.ia` | Async loops | `go run ./cmd/ialang run examples/async_loop.ia` |
| `module_main.ia` | Modules | `go run ./cmd/ialang run examples/module_main.ia` |
| `operators.ia` | Operators | `go run ./cmd/ialang run examples/operators.ia` |
| `compound_assign.ia` | Compound assignment | `go run ./cmd/ialang run examples/compound_assign.ia` |
| `comparison.ia` | Comparison | `go run ./cmd/ialang run examples/comparison.ia` |
| `bitwise.ia` | Bitwise ops | `go run ./cmd/ialang run examples/bitwise.ia` |
| `math_test.ia` | Math module | `go run ./cmd/ialang run examples/math_test.ia` |
| `string_proto_test.ia` | String prototype | `go run ./cmd/ialang run examples/string_proto_test.ia` |
| `array_proto_test.ia` | Array prototype | `go run ./cmd/ialang run examples/array_proto_test.ia` |
| `promise_test.ia` | Promise utils | `go run ./cmd/ialang run examples/promise_test.ia` |
| `ipc_demo.ia` | IPC request/reply demo | `go run ./cmd/ialang run examples/ipc_demo.ia` |
| `iax_demo.ia` | IAX interaction demo | `go run ./cmd/ialang run examples/iax_demo.ia` |
| `iax_pubsub_demo.ia` | IAX pub/sub demo | `go run ./cmd/ialang run examples/iax_pubsub_demo.ia` |
| `http_proxy_demo/main.ia` | HTTP proxy mutation demo | `go run ./cmd/ialang run examples/http_proxy_demo/main.ia` |
| `http_proxy_demo/route_rewrite.ia` | HTTP proxy route rewrite demo | `go run ./cmd/ialang run examples/http_proxy_demo/route_rewrite.ia` |
| `package_demo/main.ia` | Build + run-pkg demo | `go run ./cmd/ialang run examples/package_demo/main.ia` |

### Package Demo (`build` / `run-pkg`)

```bash
go run ./cmd/ialang build examples/package_demo/main.ia -o examples/package_demo/app.iapkg
go run ./cmd/ialang run-pkg examples/package_demo/app.iapkg
go run ./cmd/ialang build-bin examples/package_demo/main.ia -o examples/package_demo/package_demo.exe
```

## Async Runtime Config

Optional environment variables (milliseconds):

| Variable | Description | Default |
|---|---|---|
| `IALANG_ASYNC_TASK_TIMEOUT_MS` | Timeout for async tasks | disabled |
| `IALANG_ASYNC_AWAIT_TIMEOUT_MS` | Timeout for await | disabled |
| `IALANG_STRUCTURED_RUNTIME_ERRORS` | Enable structured errors | disabled |

Example:

```bash
IALANG_ASYNC_TASK_TIMEOUT_MS=1000 IALANG_ASYNC_AWAIT_TIMEOUT_MS=500 IALANG_STRUCTURED_RUNTIME_ERRORS=1 go run ./cmd/ialang run examples/async.ia
```

### Structured Errors

When `IALANG_STRUCTURED_RUNTIME_ERRORS=1`, errors include:

| Field | Type | Description |
|---|---|---|
| `e.name` | string | Error type name |
| `e.code` | string | Error code |
| `e.kind` | string | Error kind |
| `e.message` | string | Full message |
| `e.retryable` | bool | Whether retry helps |
| `e.runtime` | string | Runtime name |
| `e.module` | string | Module path |
| `e.ip` | number | Bytecode IP |
| `e.op` | number | Opcode number |
| `e.stack_depth` | number | VM stack depth |

### Sandbox Config

```go
policy := &SandboxPolicy{
    MaxSteps:     100000,
    MaxDuration:  5 * time.Second,
    AllowImport:  true,
    AllowFS:      false,
    AllowNetwork: false,
    AllowedModules: map[string]bool{"math": true, "string": true},
}
vm := runtime.NewVMWithOptions(chunk, modules, resolver, path, asyncRuntime, runtime.VMOptions{
    Sandbox: policy,
})
```

## Error Types

| Error Type | Code | Description |
|---|---|---|
| `RuntimeError` | `RUNTIME_ERROR` | General runtime error |
| `TimeoutError` | `TIMEOUT` | Timeout exceeded |
| `SandboxError` | `SANDBOX_VIOLATION` | Sandbox policy violation |
| `ImportError` | `IMPORT_ERROR` | Module import failure |
| `TypeError` | `TYPE_ERROR` | Type mismatch |
| `ReferenceError` | `REFERENCE_ERROR` | Undefined variable |

## Package Structure

```
pkg/lang/
├── token/          # Token definitions
├── ast/            # AST node types
├── frontend/       # Lexer + Parser
├── bytecode/       # Bytecode instructions
├── compiler/       # AST → Bytecode compiler
├── runtime/        # VM + async + builtins
│   ├── builtin/    # Native modules (http, fs, math, string, array, etc.)
│   └── types/      # Shared runtime types
└── lang/           # Facade + orchestration
```

## Tests

```bash
go test ./...
```

All tests pass ✅

For `iapm` release-gate checks, see `docs/2026-04-09/RELEASE_GATE.md`.

HTTP proxy/forward performance examples:

```bash
# pressure test (error-rate / status-code stability)
go test ./pkg/lang/runtime/builtin -run TestHTTPServerPipelinePressure -v

# benchmark (latency + allocs)
go test ./pkg/lang/runtime/builtin -run "^$" -bench "BenchmarkHTTPServer(Proxy|Forward)Pipeline" -benchmem -benchtime=1s

# cpu scaling benchmark matrix
go test ./pkg/lang/runtime/builtin -run "^$" -bench "BenchmarkHTTPServer(Proxy|Forward)Pipeline" -benchmem -benchtime=1s -cpu "1,4,8,12"
```

## Development Status

**Current Version**: MVP + Extensions  
**Language Completeness**: 92+/100  
**Test Coverage**: Core paths covered  
**Production Ready**: Suitable for prototyping and internal tools

### Completed Features

- ✅ Full operator set (arithmetic, comparison, logical, bitwise, compound assignment)
- ✅ Ternary operator
- ✅ Class inheritance (extends/super)
- ✅ String prototype methods
- ✅ Array prototype methods
- ✅ Math module
- ✅ Promise utilities (all/race/allSettled)
- ✅ Sandbox execution policy
- ✅ Structured error types

### Future Roadmap

- [ ] Debugger / step-through / trace
- [ ] Package manager (remote modules, versioning)
- [ ] Performance optimization & benchmarks
- [ ] Full `export default` parity (anonymous class declaration form, etc.)
- [ ] Observability (trace/profile/metrics)

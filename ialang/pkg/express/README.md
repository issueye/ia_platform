# Express Framework for IALang

一个类 Node.js Express 的 Go Web 框架，为 ialang 提供优雅的路由和中间件支持。

## 特性

- ✅ **Express 风格 API** - 熟悉的 Express.js 语法
- ✅ **路由系统** - 支持路径参数、查询参数、路由组
- ✅ **中间件系统** - 灵活的中间件链，支持全局和局部中间件
- ✅ **请求/响应封装** - 简化的请求处理和响应发送
- ✅ **数据验证** - 内置验证器支持
- ✅ **错误处理** - 统一的错误处理和自定义错误类型
- ✅ **CORS 支持** - 内置 CORS 中间件
- ✅ **Panic 恢复** - 自动捕获和处理 panic

## 安装

```go
import "ialang/pkg/express"
```

## 快速开始

### 基本示例

```go
package main

import (
    "ialang/pkg/express"
    "log"
)

func main() {
    app := express.New()
    
    // 基本路由
    app.GET("/", func(ctx *express.Context) {
        ctx.JSON(map[string]interface{}{
            "message": "Hello, Express!",
        })
    })
    
    // 路径参数
    app.GET("/users/:id", func(ctx *express.Context) {
        id := ctx.Param("id")
        ctx.JSON(map[string]interface{}{
            "user_id": id,
        })
    })
    
    // 启动服务器
    log.Fatal(app.Listen(":3000"))
}
```

## 路由

### HTTP 方法

```go
app.GET("/path", handler)
app.POST("/path", handler)
app.PUT("/path", handler)
app.DELETE("/path", handler)
app.PATCH("/path", handler)
app.OPTIONS("/path", handler)
app.ALL("/path", handler)  // 匹配所有方法
```

### 路径参数

```go
app.GET("/users/:id", func(ctx *express.Context) {
    id := ctx.Param("id")
    ctx.JSON(map[string]interface{}{"id": id})
})
```

### 查询参数

```go
app.GET("/search", func(ctx *express.Context) {
    query := ctx.QueryParam("q")
    page := ctx.QueryParam("page")
    ctx.JSON(map[string]interface{}{
        "query": query,
        "page":  page,
    })
})
```

### 路由组

```go
api := app.Group("/api/v1")
{
    api.GET("/users", getUsers)
    api.POST("/users", createUser)
    api.GET("/users/:id", getUser)
}

// 路由组中间件
api.Use(authMiddleware())
```

## 中间件

### 内置中间件

```go
// 日志记录
app.Use(express.Logger())

// Panic 恢复
app.Use(express.Recovery())

// CORS
app.Use(express.CORS())

// 自定义 CORS 配置
app.Use(express.CORS(map[string]string{
    "Access-Control-Allow-Origin": "https://example.com",
}))
```

### 自定义中间件

```go
// 函数式中间件
app.Use(func(ctx *express.Context) {
    // 请求前处理
    log.Printf("Request: %s %s", ctx.Method(), ctx.Path())
    
    ctx.Next()  // 执行下一个中间件/处理器
    
    // 响应后处理
    log.Printf("Response: %d", ctx.Writer.Status())
})

// 带配置的中间件
func requireAuth(roles ...string) express.MiddlewareFunc {
    return func(ctx *express.Context) {
        token := ctx.Header("Authorization")
        if token == "" {
            ctx.Status(401).JSON(map[string]interface{}{
                "error": "Unauthorized",
            })
            ctx.Abort()
            return
        }
        ctx.Next()
    }
}

app.GET("/admin", requireAuth("admin"), adminHandler)
```

## 请求处理

### 获取请求数据

```go
app.POST("/users", func(ctx *express.Context) {
    // 路径参数
    id := ctx.Param("id")
    
    // 查询参数
    query := ctx.QueryParam("q")
    
    // 请求头
    token := ctx.Header("Authorization")
    
    // JSON Body（自动解析）
    body := ctx.GetBody()
    name := ctx.BodyParam("name")
    
    // 客户端 IP
    ip := ctx.IP()
    
    // 表单值
    username := ctx.FormValue("username")
})
```

### 发送响应

```go
app.GET("/response", func(ctx *express.Context) {
    // JSON 响应
    ctx.JSON(map[string]interface{}{
        "message": "Hello",
    })
    
    // 纯文本响应
    ctx.Text("Hello World")
    
    // HTML 响应
    ctx.HTML("<h1>Hello</h1>")
    
    // 设置状态码
    ctx.Status(201).JSON(map[string]interface{}{
        "message": "Created",
    })
    
    // 设置响应头
    ctx.Set("X-Custom-Header", "value")
    
    // 重定向
    ctx.Redirect("/new-path")
    
    // 发送文件
    ctx.File("./path/to/file.pdf")
    
    // 文件下载
    ctx.Attachment("./file.pdf", "download.pdf")
    
    // Cookie
    ctx.Cookie("session", "token")
})
```

## 数据验证

```go
import "regexp"

rules := map[string][]express.Rule{
    "name": {
        {Required: true, MinLength: ptrInt(2), MaxLength: ptrInt(50)},
    },
    "email": {
        {Required: true, Pattern: regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)},
    },
    "age": {
        {Type: "number", Min: ptrFloat(0), Max: ptrFloat(150)},
    },
}

app.POST("/users", func(ctx *express.Context) {
    body := ctx.GetBody()
    
    errors := express.Validate(body, rules)
    if len(errors) > 0 {
        ctx.Status(400).JSON(map[string]interface{}{
            "error": "Validation failed",
            "details": errors,
        })
        return
    }
    
    // 验证通过，处理数据
    ctx.Status(201).JSON(body)
})
```

## 错误处理

### 内置错误类型

```go
app.GET("/users/:id", func(ctx *express.Context) {
    id, err := ctx.IntParam("id")
    if err != nil {
        express.HandleError(ctx, express.BadRequest("Invalid user ID"))
        return
    }
    
    if id > 1000 {
        express.HandleError(ctx, express.NotFound("User not found"))
        return
    }
    
    // 自定义错误
    express.HandleError(ctx, express.NewError(403, "FORBIDDEN", "Access denied"))
})
```

### 错误中间件

```go
// 默认错误处理
app.Use(express.ErrorMiddleware())

// 自定义错误处理
customHandler := func(err error, ctx *express.Context) {
    if httpErr, ok := err.(*express.HTTPError); ok {
        ctx.Status(httpErr.StatusCode)
        ctx.JSON(map[string]interface{}{
            "success": false,
            "error": map[string]interface{}{
                "code":    httpErr.Code,
                "message": httpErr.Message,
            },
        })
        return
    }
    
    ctx.Status(500)
    ctx.JSON(map[string]interface{}{
        "success": false,
        "error": "Internal Server Error",
    })
}

app.Use(express.ErrorMiddleware(customHandler))
```

## 完整示例

查看 `examples/express/` 目录获取更多示例：

- `basic.go` - 基本用法示例
- `middleware.go` - 中间件使用示例
- `validation.go` - 数据验证和错误处理示例

## API 参考

### App

| 方法 | 描述 |
|------|------|
| `New()` | 创建应用实例 |
| `Use(...MiddlewareFunc)` | 添加全局中间件 |
| `GET/POST/PUT/DELETE/PATCH(path, ...handler)` | 注册路由 |
| `Group(prefix, ...middleware)` | 创建路由组 |
| `Listen(addr)` | 启动服务器 |
| `ListenTLS(addr, cert, key)` | 启动 HTTPS 服务器 |
| `Shutdown(timeout)` | 优雅关闭服务器 |
| `Static(prefix, root)` | 静态文件服务 |

### Context

| 方法 | 描述 |
|------|------|
| `Param(name)` | 获取路径参数 |
| `QueryParam(name)` | 获取查询参数 |
| `Header(name)` | 获取请求头 |
| `GetBody()` | 获取请求体 |
| `Status(code)` | 设置状态码 |
| `JSON(data)` | 发送 JSON 响应 |
| `Text(text)` | 发送文本响应 |
| `HTML(html)` | 发送 HTML 响应 |
| `File(path)` | 发送文件 |
| `Redirect(url)` | 重定向 |
| `Cookie(name, value)` | 设置 Cookie |
| `Set(key, value)` | 设置响应头 |
| `IP()` | 获取客户端 IP |

## 许可证

MIT

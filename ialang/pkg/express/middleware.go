package express

// MiddlewareFunc 中间件函数类型
type MiddlewareFunc func(ctx *Context)

// Next 中间件链中的下一个函数
type Next func()

// Middleware 中间件接口（可选使用）
type Middleware interface {
	Handle(ctx *Context, next Next)
}

// AdaptMiddleware 将接口类型的中间件适配为函数类型
func AdaptMiddleware(m Middleware) MiddlewareFunc {
	return func(ctx *Context) {
		m.Handle(ctx, func() {
			ctx.Next()
		})
	}
}

// 常用中间件

// Logger 日志记录中间件
func Logger() MiddlewareFunc {
	return func(ctx *Context) {
		// 记录请求开始
		method := ctx.Req.Method
		path := ctx.Req.URL.Path
		
		// 执行下一个中间件
		ctx.Next()
		
		// 记录请求结束
		status := ctx.Writer.status
		_ = method  // 可以在这里添加实际的日志记录逻辑
		_ = path
		_ = status
	}
}

// Recovery 恢复中间件（捕获 panic）
func Recovery() MiddlewareFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx.Status(500)
				ctx.JSON(map[string]interface{}{
					"error": "Internal Server Error",
				})
				ctx.Abort()
			}
		}()
		ctx.Next()
	}
}

// CORS CORS 中间件
func CORS(options ...map[string]string) MiddlewareFunc {
	opts := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}
	
	if len(options) > 0 {
		for k, v := range options[0] {
			opts[k] = v
		}
	}
	
	return func(ctx *Context) {
		// 设置 CORS 头
		for key, value := range opts {
			ctx.Writer.Header().Set(key, value)
		}
		
		// 处理预检请求
		if ctx.Req.Method == "OPTIONS" {
			ctx.Status(204)
			ctx.Abort()
			return
		}
		
		ctx.Next()
	}
}

// Static 静态文件服务中间件
func Static(root string) MiddlewareFunc {
	return func(ctx *Context) {
		// 这里可以添加静态文件服务逻辑
		// 简单实现：交给默认的文件服务器处理
		ctx.Next()
	}
}

package express

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// App Express 应用实例
type App struct {
	*Router
	server     *http.Server
	middleware []MiddlewareFunc
	settings   map[string]interface{}
}

// New 创建新的应用实例
func New() *App {
	app := &App{
		Router:     NewRouter(),
		middleware: make([]MiddlewareFunc, 0),
		settings:   make(map[string]interface{}),
	}
	
	return app
}

// Use 添加全局中间件
func (app *App) Use(handlers ...MiddlewareFunc) {
	app.Router.Use(handlers...)
}

// Listen 启动 HTTP 服务器
func (app *App) Listen(addr string) error {
	app.server = &http.Server{
		Addr:    addr,
		Handler: app.Router,
	}
	
	fmt.Printf("🚀 Server listening on http://%s\n", addr)
	
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	
	return app.server.Serve(ln)
}

// ListenTLS 启动 HTTPS 服务器
func (app *App) ListenTLS(addr, certFile, keyFile string) error {
	app.server = &http.Server{
		Addr:    addr,
		Handler: app.Router,
	}
	
	fmt.Printf("🚀 HTTPS Server listening on https://%s\n", addr)
	return app.server.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown 优雅关闭服务器
func (app *App) Shutdown(timeout ...time.Duration) error {
	if app.server == nil {
		return fmt.Errorf("server not started")
	}
	
	t := 5 * time.Second
	if len(timeout) > 0 {
		t = timeout[0]
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()
	
	fmt.Println("👋 Shutting down server...")
	return app.server.Shutdown(ctx)
}

// Route 动态加载路由
func (app *App) Route(path string, register func(*App)) {
	register(app)
}

// Settings 获取设置
func (app *App) Setting(name string, value ...interface{}) interface{} {
	if len(value) > 0 {
		app.settings[name] = value[0]
		return nil
	}
	return app.settings[name]
}

// Enable 启用设置
func (app *App) Enable(name string) {
	app.settings[name] = true
}

// Disable 禁用设置
func (app *App) Disable(name string) {
	app.settings[name] = false
}

// Enabled 检查设置是否启用
func (app *App) Enabled(name string) bool {
	v, ok := app.settings[name]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

// Disabled 检查设置是否禁用
func (app *App) Disabled(name string) bool {
	return !app.Enabled(name)
}

// 便捷路由方法

// GET 注册 GET 路由
func (app *App) GET(path string, handlers ...MiddlewareFunc) {
	app.Router.GET(path, handlers...)
}

// POST 注册 POST 路由
func (app *App) POST(path string, handlers ...MiddlewareFunc) {
	app.Router.POST(path, handlers...)
}

// PUT 注册 PUT 路由
func (app *App) PUT(path string, handlers ...MiddlewareFunc) {
	app.Router.PUT(path, handlers...)
}

// DELETE 注册 DELETE 路由
func (app *App) DELETE(path string, handlers ...MiddlewareFunc) {
	app.Router.DELETE(path, handlers...)
}

// PATCH 注册 PATCH 路由
func (app *App) PATCH(path string, handlers ...MiddlewareFunc) {
	app.Router.PATCH(path, handlers...)
}

// OPTIONS 注册 OPTIONS 路由
func (app *App) OPTIONS(path string, handlers ...MiddlewareFunc) {
	app.Router.OPTIONS(path, handlers...)
}

// ALL 注册所有方法的路由
func (app *App) ALL(path string, handlers ...MiddlewareFunc) {
	app.Router.Handle("ANY", path, handlers...)
}

// Static 静态文件服务
func (app *App) Static(prefix, root string) {
	app.GET(prefix+"/*filepath", func(ctx *Context) {
		filepath := ctx.Param("filepath")
		fullPath := root + "/" + filepath
		ctx.File(fullPath)
	})
}

// StaticFile 静态文件
func (app *App) StaticFile(path, filepath string) {
	app.GET(path, func(ctx *Context) {
		ctx.File(filepath)
	})
}

// NotFound 404 处理
func (app *App) NotFound(handler MiddlewareFunc) {
	// 这里可以添加自定义 404 处理逻辑
	_ = handler
}

// Error 错误处理
func (app *App) Error(handler func(error, *Context)) {
	// 这里可以添加自定义错误处理逻辑
	_ = handler
}

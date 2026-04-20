package express

import (
	"net/http"
	"strings"
	"sync"
)

// Router 路由管理器
type Router struct {
	mu       sync.RWMutex
	routes   []*Route
	middleware []MiddlewareFunc
}

// Route 单个路由定义
type Route struct {
	Method     string
	Path       string
	Handlers   []MiddlewareFunc
	Router     *Router
	isGroup    bool
	prefix     string
}

// RouteGroup 路由组
type RouteGroup struct {
	prefix     string
	middleware []MiddlewareFunc
	router     *Router
}

// NewRouter 创建新的路由管理器
func NewRouter() *Router {
	return &Router{
		routes: make([]*Route, 0),
		middleware: make([]MiddlewareFunc, 0),
	}
}

// Use 添加中间件
func (r *Router) Use(handlers ...MiddlewareFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middleware = append(r.middleware, handlers...)
}

// Handle 注册路由
func (r *Router) Handle(method, path string, handlers ...MiddlewareFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	route := &Route{
		Method:   strings.ToUpper(method),
		Path:     path,
		Handlers: handlers,
		isGroup:  false,
	}
	
	r.routes = append(r.routes, route)
}

// GET 注册 GET 路由
func (r *Router) GET(path string, handlers ...MiddlewareFunc) {
	r.Handle("GET", path, handlers...)
}

// POST 注册 POST 路由
func (r *Router) POST(path string, handlers ...MiddlewareFunc) {
	r.Handle("POST", path, handlers...)
}

// PUT 注册 PUT 路由
func (r *Router) PUT(path string, handlers ...MiddlewareFunc) {
	r.Handle("PUT", path, handlers...)
}

// DELETE 注册 DELETE 路由
func (r *Router) DELETE(path string, handlers ...MiddlewareFunc) {
	r.Handle("DELETE", path, handlers...)
}

// PATCH 注册 PATCH 路由
func (r *Router) PATCH(path string, handlers ...MiddlewareFunc) {
	r.Handle("PATCH", path, handlers...)
}

// OPTIONS 注册 OPTIONS 路由
func (r *Router) OPTIONS(path string, handlers ...MiddlewareFunc) {
	r.Handle("OPTIONS", path, handlers...)
}

// HEAD 注册 HEAD 路由
func (r *Router) HEAD(path string, handlers ...MiddlewareFunc) {
	r.Handle("HEAD", path, handlers...)
}

// Group 创建路由组
func (r *Router) Group(prefix string, middleware ...MiddlewareFunc) *RouteGroup {
	return &RouteGroup{
		prefix:     prefix,
		middleware: middleware,
		router:     r,
	}
}

// RouteGroup 方法实现

// Use 为路由组添加中间件
func (g *RouteGroup) Use(handlers ...MiddlewareFunc) {
	g.middleware = append(g.middleware, handlers...)
}

// Handle 为路由组注册路由
func (g *RouteGroup) Handle(method, path string, handlers ...MiddlewareFunc) {
	fullPath := g.prefix + path
	allHandlers := append(g.middleware, handlers...)
	g.router.Handle(method, fullPath, allHandlers...)
}

// GET 为路由组注册 GET 路由
func (g *RouteGroup) GET(path string, handlers ...MiddlewareFunc) {
	g.Handle("GET", path, handlers...)
}

// POST 为路由组注册 POST 路由
func (g *RouteGroup) POST(path string, handlers ...MiddlewareFunc) {
	g.Handle("POST", path, handlers...)
}

// PUT 为路由组注册 PUT 路由
func (g *RouteGroup) PUT(path string, handlers ...MiddlewareFunc) {
	g.Handle("PUT", path, handlers...)
}

// DELETE 为路由组注册 DELETE 路由
func (g *RouteGroup) DELETE(path string, handlers ...MiddlewareFunc) {
	g.Handle("DELETE", path, handlers...)
}

// PATCH 为路由组注册 PATCH 路由
func (g *RouteGroup) PATCH(path string, handlers ...MiddlewareFunc) {
	g.Handle("PATCH", path, handlers...)
}

// Match 匹配路由
func (r *Router) Match(method, path string) (*Route, map[string]string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	method = strings.ToUpper(method)
	
	for _, route := range r.routes {
		if route.Method != method && route.Method != "ANY" {
			continue
		}
		
		params, matched := matchPath(route.Path, path)
		if matched {
			return route, params, true
		}
	}
	
	return nil, nil, false
}

// matchPath 匹配路径并提取参数
func matchPath(routePath, requestPath string) (map[string]string, bool) {
	routeParts := strings.Split(strings.Trim(routePath, "/"), "/")
	requestParts := strings.Split(strings.Trim(requestPath, "/"), "/")

	params := make(map[string]string)

	// 检查是否有通配符
	hasWildcard := false
	for _, part := range routeParts {
		if strings.HasPrefix(part, "*") {
			hasWildcard = true
			break
		}
	}

	// 如果有通配符，请求路径分段可以更多
	if hasWildcard {
		if len(requestParts) < len(routeParts) {
			return nil, false
		}
	} else if len(routeParts) != len(requestParts) {
		return nil, false
	}

	for i, part := range routeParts {
		if strings.HasPrefix(part, "*") {
			// 通配符参数 - 匹配剩余所有
			paramName := part[1:]
			if paramName == "" {
				paramName = "wildcard"
			}
			params[paramName] = strings.Join(requestParts[i:], "/")
			return params, true
		} else if strings.HasPrefix(part, ":") {
			// 路径参数
			paramName := part[1:]
			params[paramName] = requestParts[i]
		} else {
			if i >= len(requestParts) || part != requestParts[i] {
				return nil, false
			}
		}
	}

	return params, true
}

// ServeHTTP 实现 http.Handler 接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := NewContext(req, w)
	
	// 全局 panic 恢复
	defer func() {
		if err := recover(); err != nil {
			// 如果还没有响应，发送 500 错误
			if !ctx.Writer.written {
				ctx.Status(500)
				ctx.JSON(map[string]interface{}{
					"error": "Internal Server Error",
				})
			}
		}
	}()
	
	// 应用全局中间件
	for _, mw := range r.middleware {
		mw(ctx)
		if ctx.Aborted() {
			return
		}
	}
	
	// 匹配路由
	route, params, found := r.Match(req.Method, req.URL.Path)
	if !found {
		http.NotFound(w, req)
		return
	}
	
	// 设置路径参数
	ctx.Params = params
	
	// 执行路由处理器
	for _, handler := range route.Handlers {
		handler(ctx)
		if ctx.Aborted() {
			return
		}
	}
}

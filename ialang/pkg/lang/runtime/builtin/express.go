package builtin

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"ialang/pkg/express"
	rtvm "ialang/pkg/lang/runtime/vm"
)

// expressModule 创建 express 模块
func newExpressModule(asyncRuntime AsyncRuntime) Value {
	// express() - 创建应用
	expressFn := NativeFunction(func(args []Value) (Value, error) {
		app := express.New()
		return createExpressAppObject(app, asyncRuntime), nil
	})

	// 内置中间件
	loggerFn := NativeFunction(func(args []Value) (Value, error) {
		mw := express.Logger()
		return createMiddlewareObject(mw), nil
	})

	recoveryFn := NativeFunction(func(args []Value) (Value, error) {
		mw := express.Recovery()
		return createMiddlewareObject(mw), nil
	})

	corsFn := NativeFunction(func(args []Value) (Value, error) {
		var mw express.MiddlewareFunc
		if len(args) > 0 {
			if opts, ok := args[0].(Object); ok {
				corsOpts := make(map[string]string)
				for k, v := range opts {
					if s, ok := v.(string); ok {
						corsOpts[k] = s
					}
				}
				mw = express.CORS(corsOpts)
			} else {
				mw = express.CORS()
			}
		} else {
			mw = express.CORS()
		}
		return createMiddlewareObject(mw), nil
	})

	// Router 构造函数
	routerFn := NativeFunction(func(args []Value) (Value, error) {
		router := express.NewRouter()
		return createRouterObject(router, asyncRuntime), nil
	})

	// Create module with self-reference for namespace pattern
	module := Object{
		// express() 主函数
		"express": expressFn,

		// 内置中间件
		"logger":   loggerFn,
		"recovery": recoveryFn,
		"cors":     corsFn,
		"router":   routerFn,

		// 便捷方法
		"New": expressFn, // express.New() 的别名
	}
	// Note: we don't add self-reference here because "express" key already exists as function
	// Users can use: import { express } from "express"; express()

	return module
}

// createExpressAppObject 创建应用对象
func createExpressAppObject(app *express.App, asyncRuntime AsyncRuntime) Object {
	var serverMu sync.Mutex
	var server *http.Server

	// use 方法
	useFn := NativeFunction(func(args []Value) (Value, error) {
		for _, arg := range args {
			if mwObj, ok := arg.(Object); ok {
				if mwFunc, ok := mwObj["__middleware__"].(express.MiddlewareFunc); ok {
					app.Use(mwFunc)
				}
			}
		}
		return true, nil
	})

	// 路由方法
	getFn := createRouteMethod("GET", app)
	postFn := createRouteMethod("POST", app)
	putFn := createRouteMethod("PUT", app)
	deleteFn := createRouteMethod("DELETE", app)
	patchFn := createRouteMethod("PATCH", app)
	optionsFn := createRouteMethod("OPTIONS", app)
	allFn := createRouteMethod("ANY", app)

	// group 方法
	groupFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("express.app.group expects at least 1 arg, got %d", len(args))
		}

		prefix, err := asStringValue("express.app.group", args[0])
		if err != nil {
			return nil, err
		}

		var middlewares []express.MiddlewareFunc
		for i := 1; i < len(args); i++ {
			if mwObj, ok := args[i].(Object); ok {
				if mwFunc, ok := mwObj["__middleware__"].(express.MiddlewareFunc); ok {
					middlewares = append(middlewares, mwFunc)
				}
			}
		}

		group := app.Group(prefix, middlewares...)
		return createGroupObject(group, asyncRuntime), nil
	})

	// listen 方法
	listenFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("express.app.listen expects at least 1 arg, got %d", len(args))
		}

		addr, err := asStringValue("express.app.listen", args[0])
		if err != nil {
			return nil, err
		}

		handler := createHandler(app)
		server = &http.Server{
			Addr:    addr,
			Handler: handler,
		}

		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}

		actualAddr := ln.Addr().String()

		// 后台启动
		go func() {
			_ = server.Serve(ln)
		}()

		fmt.Printf("🚀 Express server listening on http://%s\n", actualAddr)

		return Object{
			"addr": actualAddr,
			"close": NativeFunction(func(args []Value) (Value, error) {
				serverMu.Lock()
				defer serverMu.Unlock()
				if server != nil {
					return server.Close(), nil
				}
				return false, nil
			}),
		}, nil
	})

	// static 方法
	staticFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("express.app.static expects 2 args, got %d", len(args))
		}

		prefix, err := asStringValue("express.app.static.prefix", args[0])
		if err != nil {
			return nil, err
		}

		root, err := asStringValue("express.app.static.root", args[1])
		if err != nil {
			return nil, err
		}

		app.Static(prefix, root)
		return true, nil
	})

	return Object{
		"use":     useFn,
		"get":     getFn,
		"post":    postFn,
		"put":     putFn,
		"delete":  deleteFn,
		"patch":   patchFn,
		"options": optionsFn,
		"all":     allFn,
		"group":   groupFn,
		"listen":  listenFn,
		"static":  staticFn,
	}
}

// createRouterObject 创建路由器对象
func createRouterObject(router *express.Router, asyncRuntime AsyncRuntime) Object {
	useFn := NativeFunction(func(args []Value) (Value, error) {
		for _, arg := range args {
			if mwObj, ok := arg.(Object); ok {
				if mwFunc, ok := mwObj["__middleware__"].(express.MiddlewareFunc); ok {
					router.Use(mwFunc)
				}
			}
		}
		return true, nil
	})

	getFn := createRouterRouteMethod("GET", router)
	postFn := createRouterRouteMethod("POST", router)
	putFn := createRouterRouteMethod("PUT", router)
	deleteFn := createRouterRouteMethod("DELETE", router)
	patchFn := createRouterRouteMethod("PATCH", router)

	return Object{
		"use":    useFn,
		"get":    getFn,
		"post":   postFn,
		"put":    putFn,
		"delete": deleteFn,
		"patch":  patchFn,
	}
}

// createGroupObject 创建路由组对象
func createGroupObject(group *express.RouteGroup, asyncRuntime AsyncRuntime) Object {
	useFn := NativeFunction(func(args []Value) (Value, error) {
		for _, arg := range args {
			if mwObj, ok := arg.(Object); ok {
				if mwFunc, ok := mwObj["__middleware__"].(express.MiddlewareFunc); ok {
					group.Use(mwFunc)
				}
			}
		}
		return true, nil
	})

	getFn := createGroupRouteMethod("GET", group)
	postFn := createGroupRouteMethod("POST", group)
	putFn := createGroupRouteMethod("PUT", group)
	deleteFn := createGroupRouteMethod("DELETE", group)
	patchFn := createGroupRouteMethod("PATCH", group)

	return Object{
		"use":    useFn,
		"get":    getFn,
		"post":   postFn,
		"put":    putFn,
		"delete": deleteFn,
		"patch":  patchFn,
	}
}

// createMiddlewareObject 创建中间件对象
func createMiddlewareObject(mw express.MiddlewareFunc) Object {
	return Object{
		"__middleware__": mw,
		"isMiddleware":   true,
	}
}

// createHandler 从 app 创建 HTTP handler
func createHandler(app *express.App) http.Handler {
	return app.Router
}

// createRouteMethod 创建路由方法
func createRouteMethod(method string, app *express.App) Value {
	return NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("express.app.%s expects at least 2 args, got %d", method, len(args))
		}

		path, err := asStringValue(fmt.Sprintf("express.app.%s", method), args[0])
		if err != nil {
			return nil, err
		}

		var handlers []express.MiddlewareFunc
		for i := 1; i < len(args); i++ {
			if handler, ok := createHandlerFromValue(args[i]); ok {
				handlers = append(handlers, handler)
			}
		}

		if method == "ANY" {
			app.ALL(path, handlers...)
		} else {
			app.Router.Handle(method, path, handlers...)
		}

		return true, nil
	})
}

// createRouterRouteMethod 创建路由器路由方法
func createRouterRouteMethod(method string, router *express.Router) Value {
	return NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("express.router.%s expects at least 2 args, got %d", method, len(args))
		}

		path, err := asStringValue(fmt.Sprintf("express.router.%s", method), args[0])
		if err != nil {
			return nil, err
		}

		var handlers []express.MiddlewareFunc
		for i := 1; i < len(args); i++ {
			if handler, ok := createHandlerFromValue(args[i]); ok {
				handlers = append(handlers, handler)
			}
		}

		router.Handle(method, path, handlers...)

		return true, nil
	})
}

// createGroupRouteMethod 创建路由组路由方法
func createGroupRouteMethod(method string, group *express.RouteGroup) Value {
	return NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("express.group.%s expects at least 2 args, got %d", method, len(args))
		}

		path, err := asStringValue(fmt.Sprintf("express.group.%s", method), args[0])
		if err != nil {
			return nil, err
		}

		var handlers []express.MiddlewareFunc
		for i := 1; i < len(args); i++ {
			if handler, ok := createHandlerFromValue(args[i]); ok {
				handlers = append(handlers, handler)
			}
		}

		group.Handle(method, path, handlers...)

		return true, nil
	})
}

// createHandlerFromValue 从值创建处理器
func createHandlerFromValue(val Value) (express.MiddlewareFunc, bool) {
	if fn, ok := val.(NativeFunction); ok {
		return func(ctx *express.Context) {
			// 创建 ialang 上下文
			ialangCtx := createIALangContext(ctx)

			// 调用函数
			result, err := fn([]Value{ialangCtx})
			if err != nil {
				ctx.Status(500).JSON(map[string]interface{}{
					"error": err.Error(),
				})
				ctx.Abort()
				return
			}

			// 如果返回 false，中止上下文
			if b, ok := result.(bool); ok && !b {
				ctx.Abort()
			}
		}, true
	}
	if fn, ok := val.(*UserFunction); ok {
		return func(ctx *express.Context) {
			ialangCtx := createIALangContext(ctx)
			result, err := rtvm.CallUserFunctionSync(fn, []Value{ialangCtx})
			if err != nil {
				ctx.Status(500).JSON(map[string]interface{}{
					"error": err.Error(),
				})
				ctx.Abort()
				return
			}
			if b, ok := result.(bool); ok && !b {
				ctx.Abort()
			}
		}, true
	}
	if fn, ok := val.(*BoundMethod); ok {
		return func(ctx *express.Context) {
			ialangCtx := createIALangContext(ctx)
			result, err := rtvm.CallBoundMethodSync(fn, []Value{ialangCtx})
			if err != nil {
				ctx.Status(500).JSON(map[string]interface{}{
					"error": err.Error(),
				})
				ctx.Abort()
				return
			}
			if b, ok := result.(bool); ok && !b {
				ctx.Abort()
			}
		}, true
	}

	return nil, false
}

// createIALangContext 创建 ialang 上下文
func createIALangContext(ctx *express.Context) Object {
	// 请求相关
	paramFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("context.param expects 1 arg, got %d", len(args))
		}
		name, err := asStringValue("context.param", args[0])
		if err != nil {
			return nil, err
		}
		return ctx.Param(name), nil
	})

	queryFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("context.query expects 1 arg, got %d", len(args))
		}
		name, err := asStringValue("context.query", args[0])
		if err != nil {
			return nil, err
		}
		return ctx.QueryParam(name), nil
	})

	bodyFn := NativeFunction(func(args []Value) (Value, error) {
		body := ctx.GetBody()
		if body == nil {
			return Object{}, nil
		}
		return mapToObject(body), nil
	})

	headerFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("context.header expects 1 arg, got %d", len(args))
		}
		name, err := asStringValue("context.header", args[0])
		if err != nil {
			return nil, err
		}
		return ctx.Header(name), nil
	})

	methodFn := NativeFunction(func(args []Value) (Value, error) {
		return ctx.Method(), nil
	})

	pathFn := NativeFunction(func(args []Value) (Value, error) {
		return ctx.Path(), nil
	})

	ipFn := NativeFunction(func(args []Value) (Value, error) {
		return ctx.IP(), nil
	})

	// 响应相关
	jsonFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("context.json expects at least 1 arg, got %d", len(args))
		}

		var data interface{}
		if obj, ok := args[0].(Object); ok {
			data = objectToMap(obj)
		} else {
			data = args[0]
		}

		if len(args) > 1 {
			if code, ok := args[1].(float64); ok {
				ctx.Status(int(code))
			}
		}

		return nil, ctx.JSON(data)
	})

	sendFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("context.send expects 1 arg, got %d", len(args))
		}

		if str, ok := args[0].(string); ok {
			return nil, ctx.Text(str)
		}

		if obj, ok := args[0].(Object); ok {
			return nil, ctx.JSON(objectToMap(obj))
		}

		return nil, ctx.Send(args[0])
	})

	statusFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("context.status expects 1 arg, got %d", len(args))
		}

		code, err := asIntValue("context.status", args[0])
		if err != nil {
			return nil, err
		}

		ctx.Status(code)
		return true, nil
	})

	setFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("context.set expects 2 args, got %d", len(args))
		}

		key, err := asStringValue("context.set.key", args[0])
		if err != nil {
			return nil, err
		}

		value, err := asStringValue("context.set.value", args[1])
		if err != nil {
			return nil, err
		}

		ctx.Set(key, value)
		return true, nil
	})

	redirectFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("context.redirect expects at least 1 arg, got %d", len(args))
		}

		url, err := asStringValue("context.redirect", args[0])
		if err != nil {
			return nil, err
		}

		if len(args) > 1 {
			if code, ok := args[1].(float64); ok {
				ctx.Redirect(url, int(code))
			} else {
				ctx.Redirect(url)
			}
		} else {
			ctx.Redirect(url)
		}

		return true, nil
	})

	cookieFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("context.cookie expects 2 args, got %d", len(args))
		}

		name, err := asStringValue("context.cookie.name", args[0])
		if err != nil {
			return nil, err
		}

		value, err := asStringValue("context.cookie.value", args[1])
		if err != nil {
			return nil, err
		}

		ctx.Cookie(name, value)
		return true, nil
	})

	nextFn := NativeFunction(func(args []Value) (Value, error) {
		ctx.Next()
		return true, nil
	})

	endFn := NativeFunction(func(args []Value) (Value, error) {
		ctx.Abort()
		return true, nil
	})

	return Object{
		// 请求
		"param":  paramFn,
		"query":  queryFn,
		"body":   bodyFn,
		"header": headerFn,
		"method": methodFn,
		"path":   pathFn,
		"ip":     ipFn,

		// 响应
		"json":     jsonFn,
		"send":     sendFn,
		"status":   statusFn,
		"set":      setFn,
		"redirect": redirectFn,
		"cookie":   cookieFn,

		// 控制流
		"next": nextFn,
		"end":  endFn,
	}
}

// mapToObject 将 map 转换为 Object
func mapToObject(m map[string]interface{}) Object {
	obj := make(Object)
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			obj[k] = mapToObject(val)
		case []interface{}:
			arr := make(Array, len(val))
			for i, item := range val {
				if m, ok := item.(map[string]interface{}); ok {
					arr[i] = mapToObject(m)
				} else {
					arr[i] = item
				}
			}
			obj[k] = arr
		default:
			obj[k] = v
		}
	}
	return obj
}

// objectToMap 将 Object 转换为 map
func objectToMap(obj Object) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range obj {
		switch val := v.(type) {
		case Object:
			m[k] = objectToMap(val)
		case Array:
			arr := make([]interface{}, len(val))
			for i, item := range val {
				if obj, ok := item.(Object); ok {
					arr[i] = objectToMap(obj)
				} else {
					arr[i] = item
				}
			}
			m[k] = arr
		default:
			m[k] = v
		}
	}
	return m
}

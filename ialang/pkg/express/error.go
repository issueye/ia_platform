package express

import (
	"fmt"
	"runtime/debug"
)

// HTTPError HTTP 错误
type HTTPError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    interface{} `json:"details,omitempty"`
}

// Error 实现 error 接口
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// NewError 创建新的 HTTP 错误
func NewError(statusCode int, code, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// NewErrorWithDetails 创建带详细信息的错误
func NewErrorWithDetails(statusCode int, code, message string, details interface{}) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}

// 常见错误工厂函数

// BadRequest 400 错误
func BadRequest(message ...string) *HTTPError {
	msg := "Bad Request"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewError(400, "BAD_REQUEST", msg)
}

// Unauthorized 401 错误
func Unauthorized(message ...string) *HTTPError {
	msg := "Unauthorized"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewError(401, "UNAUTHORIZED", msg)
}

// Forbidden 403 错误
func Forbidden(message ...string) *HTTPError {
	msg := "Forbidden"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewError(403, "FORBIDDEN", msg)
}

// NotFound 404 错误
func NotFound(message ...string) *HTTPError {
	msg := "Not Found"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewError(404, "NOT_FOUND", msg)
}

// MethodNotAllowed 405 错误
func MethodNotAllowed(message ...string) *HTTPError {
	msg := "Method Not Allowed"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewError(405, "METHOD_NOT_ALLOWED", msg)
}

// Conflict 409 错误
func Conflict(message ...string) *HTTPError {
	msg := "Conflict"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewError(409, "CONFLICT", msg)
}

// InternalServerError 500 错误
func InternalServerError(message ...string) *HTTPError {
	msg := "Internal Server Error"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewError(500, "INTERNAL_ERROR", msg)
}

// ErrorHandler 错误处理函数类型
type ErrorHandler func(err error, ctx *Context)

// DefaultErrorHandler 默认错误处理器
func DefaultErrorHandler() ErrorHandler {
	return func(err error, ctx *Context) {
		// 检查是否是 HTTPError
		if httpErr, ok := err.(*HTTPError); ok {
			ctx.Status(httpErr.StatusCode)
			ctx.JSON(map[string]interface{}{
				"error": map[string]interface{}{
					"code":    httpErr.Code,
					"message": httpErr.Message,
					"details": httpErr.Details,
				},
			})
			return
		}
		
		// 未知错误，返回 500
		ctx.Status(500)
		ctx.JSON(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "INTERNAL_ERROR",
				"message": "Internal Server Error",
			},
		})
	}
}

// ErrorMiddleware 错误处理中间件
func ErrorMiddleware(handler ...ErrorHandler) MiddlewareFunc {
	errHandler := DefaultErrorHandler()
	if len(handler) > 0 {
		errHandler = handler[0]
	}
	
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				// 捕获 panic
				var errObj error
				switch e := err.(type) {
				case error:
					errObj = e
				default:
					errObj = fmt.Errorf("%v", e)
				}
				
				// 记录堆栈信息
				stack := string(debug.Stack())
				_ = stack // 可以在这里添加日志记录
				
				errHandler(errObj, ctx)
				ctx.Abort()
			}
		}()
		
		ctx.Next()
	}
}

// HandleError 在路由处理器中处理错误
func HandleError(ctx *Context, err error) {
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok {
			ctx.Status(httpErr.StatusCode)
			ctx.JSON(map[string]interface{}{
				"error": map[string]interface{}{
					"code":    httpErr.Code,
					"message": httpErr.Message,
				},
			})
			ctx.Abort()
		} else {
			ctx.Status(500)
			ctx.JSON(map[string]interface{}{
				"error": map[string]interface{}{
					"code":    "INTERNAL_ERROR",
					"message": "Internal Server Error",
				},
			})
			ctx.Abort()
		}
	}
}

// Try 尝试执行函数并捕获错误
func Try(fn func() error) func(ctx *Context) {
	return func(ctx *Context) {
		if err := fn(); err != nil {
			HandleError(ctx, err)
			return
		}
	}
}

// AsyncTry 异步尝试执行函数
func AsyncTry(fn func(ctx *Context) error) MiddlewareFunc {
	return func(ctx *Context) {
		if err := fn(ctx); err != nil {
			HandleError(ctx, err)
			return
		}
	}
}

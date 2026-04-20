package express

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Context 请求上下文（核心对象）
type Context struct {
	Req     *http.Request
	Writer  *Response
	Params  map[string]string
	Query   map[string][]string
	Body    map[string]interface{}
	Headers map[string]string
	Store   map[string]interface{}
	
	index       int
	handlers    []MiddlewareFunc
	aborted     bool
}

// NewContext 创建新的上下文
func NewContext(req *http.Request, w http.ResponseWriter) *Context {
	ctx := &Context{
		Req:    req,
		Writer: NewResponse(w, req),
		Params: make(map[string]string),
		Query:  req.URL.Query(),
		Store:  make(map[string]interface{}),
		index:  -1,
	}
	
	// 解析请求头
	ctx.Headers = make(map[string]string)
	for key, values := range req.Header {
		if len(values) > 0 {
			ctx.Headers[key] = values[0]
		}
	}
	
	// 自动解析 JSON Body
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		contentType := req.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			body, err := io.ReadAll(req.Body)
			if err == nil {
				json.Unmarshal(body, &ctx.Body)
			}
		}
	}
	
	return ctx
}

// Next 执行下一个中间件
func (ctx *Context) Next() {
	ctx.index++
	for ctx.index < len(ctx.handlers) {
		handler := ctx.handlers[ctx.index]
		ctx.index++
		handler(ctx)
	}
}

// Abort 中止中间件链
func (ctx *Context) Abort() {
	ctx.aborted = true
}

// Aborted 检查是否已中止
func (ctx *Context) Aborted() bool {
	return ctx.aborted
}

// 请求相关方法

// Param 获取路径参数
func (ctx *Context) Param(name string) string {
	return ctx.Params[name]
}

// Query 获取查询参数
func (ctx *Context) QueryParam(name string) string {
	values := ctx.Query[name]
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

// QueryParams 获取所有查询参数
func (ctx *Context) QueryParams() map[string]string {
	result := make(map[string]string)
	for key, values := range ctx.Query {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// Header 获取请求头
func (ctx *Context) Header(name string) string {
	return ctx.Headers[name]
}

// BodyParam 获取 Body 参数
func (ctx *Context) BodyParam(name string) interface{} {
	if ctx.Body == nil {
		return nil
	}
	return ctx.Body[name]
}

// GetBody 获取完整 Body
func (ctx *Context) GetBody() map[string]interface{} {
	return ctx.Body
}

// 响应相关方法

// Status 设置状态码
func (ctx *Context) Status(code int) *Context {
	ctx.Writer.Status(code)
	return ctx
}

// Set 设置响应头
func (ctx *Context) Set(key, value string) *Context {
	ctx.Writer.Header().Set(key, value)
	return ctx
}

// Send 发送响应
func (ctx *Context) Send(data interface{}) error {
	switch v := data.(type) {
	case string:
		_, err := ctx.Writer.Text(v)
		return err
	case []byte:
		_, err := ctx.Writer.Write(v)
		return err
	case map[string]interface{}, []interface{}:
		return ctx.Writer.JSON(v)
	default:
		return ctx.Writer.JSON(data)
	}
}

// JSON 返回 JSON 响应
func (ctx *Context) JSON(data interface{}) error {
	return ctx.Writer.JSON(data)
}

// HTML 返回 HTML 响应
func (ctx *Context) HTML(html string) error {
	_, err := ctx.Writer.HTML(html)
	return err
}

// Text 返回纯文本响应
func (ctx *Context) Text(text string) error {
	_, err := ctx.Writer.Text(text)
	return err
}

// File 返回文件
func (ctx *Context) File(filepath string) {
	http.ServeFile(ctx.Writer, ctx.Req, filepath)
}

// Attachment 发送文件附件
func (ctx *Context) Attachment(filepath, filename string) {
	ctx.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.File(filepath)
}

// Redirect 重定向
func (ctx *Context) Redirect(url string, code ...int) {
	ctx.Writer.Redirect(url, code...)
}

// Cookie 设置 Cookie
func (ctx *Context) Cookie(name, value string, options ...CookieOptions) {
	ctx.Writer.Cookie(name, value, options...)
}

// ClearCookie 清除 Cookie
func (ctx *Context) ClearCookie(name string) {
	ctx.Writer.Cookie(name, "", CookieOptions{
		MaxAge: -1,
	})
}

// 便捷方法

// Get 获取存储的值
func (ctx *Context) Get(key string) interface{} {
	return ctx.Store[key]
}

// Set 设置存储值
func (ctx *Context) SetStore(key string, value interface{}) {
	ctx.Store[key] = value
}

// Path 获取请求路径
func (ctx *Context) Path() string {
	return ctx.Req.URL.Path
}

// Method 获取请求方法
func (ctx *Context) Method() string {
	return ctx.Req.Method
}

// IP 获取客户端 IP
func (ctx *Context) IP() string {
	// 优先从 X-Forwarded-For 获取
	forwarded := ctx.Header("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	
	// 从 X-Real-IP 获取
	realIP := ctx.Header("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	
	// 从 RemoteAddr 获取
	ip := ctx.Req.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// Is 检查 Content-Type
func (ctx *Context) Is(contentType string) bool {
	ct := ctx.Header("Content-Type")
	return strings.Contains(ct, contentType)
}

// Accepts 检查 Accept 头
func (ctx *Context) Accepts(types ...string) string {
	accept := ctx.Header("Accept")
	if accept == "" {
		return ""
	}
	
	for _, t := range types {
		if strings.Contains(accept, t) {
			return t
		}
	}
	return ""
}

// FormValue 获取表单值
func (ctx *Context) FormValue(name string) string {
	return ctx.Req.FormValue(name)
}

// FormFile 获取上传文件
func (ctx *Context) FormFile(name string) (*multipart.FileHeader, error) {
	_, file, err := ctx.Req.FormFile(name)
	return file, err
}

// SaveUploadedFile 保存上传的文件
func (ctx *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	
	_, err = io.Copy(out, src)
	return err
}

// IntParam 获取整数类型路径参数
func (ctx *Context) IntParam(name string) (int, error) {
	val := ctx.Param(name)
	return strconv.Atoi(val)
}

// IntQueryParam 获取整数类型查询参数
func (ctx *Context) IntQueryParam(name string) (int, error) {
	val := ctx.QueryParam(name)
	return strconv.Atoi(val)
}

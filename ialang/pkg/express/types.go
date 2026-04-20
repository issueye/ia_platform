package express

import (
	"encoding/json"
	"net/http"
)

// Request 请求封装
type Request struct {
	*http.Request
}

// Response 响应封装
type Response struct {
	http.ResponseWriter
	written bool
	status  int
	req     *http.Request
}

// NewResponse 创建响应对象
func NewResponse(w http.ResponseWriter, req *http.Request) *Response {
	return &Response{
		ResponseWriter: w,
		status:         200,
		req:            req,
	}
}

// Status 设置响应状态码
func (r *Response) Status(code int) int {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
	return code
}

// Write 写入响应数据
func (r *Response) Write(b []byte) (int, error) {
	r.written = true
	return r.ResponseWriter.Write(b)
}

// JSON 返回 JSON 响应
func (r *Response) JSON(data interface{}) error {
	r.Header().Set("Content-Type", "application/json")
	r.written = true
	encoder := json.NewEncoder(r)
	return encoder.Encode(data)
}

// HTML 返回 HTML 响应
func (r *Response) HTML(html string) (int, error) {
	r.Header().Set("Content-Type", "text/html; charset=utf-8")
	return r.Write([]byte(html))
}

// Text 返回纯文本响应
func (r *Response) Text(text string) (int, error) {
	r.Header().Set("Content-Type", "text/plain; charset=utf-8")
	return r.Write([]byte(text))
}

// Redirect 重定向
func (r *Response) Redirect(url string, code ...int) {
	status := 302
	if len(code) > 0 {
		status = code[0]
	}
	http.Redirect(r.ResponseWriter, r.req, url, status)
}

// Cookie 设置 Cookie
func (r *Response) Cookie(name, value string, options ...CookieOptions) {
	cookie := &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
	}
	
	if len(options) > 0 {
		opt := options[0]
		if opt.Path != "" {
			cookie.Path = opt.Path
		}
		if opt.Domain != "" {
			cookie.Domain = opt.Domain
		}
		cookie.MaxAge = opt.MaxAge
		cookie.Secure = opt.Secure
		cookie.HttpOnly = opt.HttpOnly
		cookie.SameSite = opt.SameSite
	}
	
	http.SetCookie(r.ResponseWriter, cookie)
}

// CookieOptions Cookie 选项
type CookieOptions struct {
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

package express

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouter_BasicRoutes(t *testing.T) {
	app := New()
	
	app.GET("/", func(ctx *Context) {
		ctx.JSON(map[string]interface{}{
			"message": "Hello",
		})
	})
	
	app.POST("/users", func(ctx *Context) {
		ctx.Status(201).JSON(map[string]interface{}{
			"message": "Created",
		})
	})
	
	// 测试 GET 请求
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["message"] != "Hello" {
		t.Errorf("Expected message 'Hello', got %v", response["message"])
	}
	
	// 测试 POST 请求
	req = httptest.NewRequest("POST", "/users", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	if w.Code != 201 {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestRouter_PathParams(t *testing.T) {
	app := New()
	
	app.GET("/users/:id", func(ctx *Context) {
		id := ctx.Param("id")
		ctx.JSON(map[string]interface{}{
			"id": id,
		})
	})
	
	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	if response["id"] != "123" {
		t.Errorf("Expected id '123', got %v", response["id"])
	}
}

func TestRouter_QueryParams(t *testing.T) {
	app := New()
	
	app.GET("/search", func(ctx *Context) {
		query := ctx.QueryParam("q")
		page := ctx.QueryParam("page")
		ctx.JSON(map[string]interface{}{
			"query": query,
			"page":  page,
		})
	})
	
	req := httptest.NewRequest("GET", "/search?q=golang&page=1", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	if response["query"] != "golang" {
		t.Errorf("Expected query 'golang', got %v", response["query"])
	}
	if response["page"] != "1" {
		t.Errorf("Expected page '1', got %v", response["page"])
	}
}

func TestRouter_RouteGroup(t *testing.T) {
	app := New()
	
	api := app.Group("/api/v1")
	{
		api.GET("/users", func(ctx *Context) {
			ctx.JSON(map[string]interface{}{
				"users": []string{"Alice", "Bob"},
			})
		})
		api.POST("/users", func(ctx *Context) {
			ctx.Status(201).JSON(map[string]interface{}{
				"message": "Created",
			})
		})
	}
	
	// 测试 GET 请求
	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// 测试 POST 请求
	req = httptest.NewRequest("POST", "/api/v1/users", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	if w.Code != 201 {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestMiddleware_Execution(t *testing.T) {
	app := New()
	
	executed := []string{}
	
	app.Use(func(ctx *Context) {
		executed = append(executed, "middleware1")
		ctx.Next()
	})
	
	app.Use(func(ctx *Context) {
		executed = append(executed, "middleware2")
		ctx.Next()
	})
	
	app.GET("/", func(ctx *Context) {
		executed = append(executed, "handler")
		ctx.JSON(map[string]interface{}{"status": "ok"})
	})
	
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	expected := []string{"middleware1", "middleware2", "handler"}
	if len(executed) != len(expected) {
		t.Errorf("Expected %d executions, got %d", len(expected), len(executed))
	}
	
	for i, v := range expected {
		if executed[i] != v {
			t.Errorf("Expected '%s' at position %d, got '%s'", v, i, executed[i])
		}
	}
}

func TestMiddleware_Abort(t *testing.T) {
	app := New()
	
	executed := []string{}
	
	app.Use(func(ctx *Context) {
		executed = append(executed, "middleware1")
		ctx.Next()
	})
	
	app.Use(func(ctx *Context) {
		executed = append(executed, "middleware2")
		ctx.Abort()  // 中止后续执行
	})
	
	app.GET("/", func(ctx *Context) {
		executed = append(executed, "handler")
		ctx.JSON(map[string]interface{}{"status": "ok"})
	})
	
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	// 应该只执行前两个中间件
	if len(executed) != 2 {
		t.Errorf("Expected 2 executions, got %d", len(executed))
	}
	
	if executed[0] != "middleware1" || executed[1] != "middleware2" {
		t.Errorf("Middleware execution order incorrect: %v", executed)
	}
}

func TestMiddleware_GroupMiddleware(t *testing.T) {
	app := New()
	
	executed := []string{}
	
	api := app.Group("/api")
	api.Use(func(ctx *Context) {
		executed = append(executed, "group_middleware")
		ctx.Next()
	})
	
	api.GET("/users", func(ctx *Context) {
		executed = append(executed, "handler")
		ctx.JSON(map[string]interface{}{"status": "ok"})
	})
	
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	if len(executed) != 2 {
		t.Errorf("Expected 2 executions, got %d", len(executed))
	}
	
	if executed[0] != "group_middleware" || executed[1] != "handler" {
		t.Errorf("Middleware execution order incorrect: %v", executed)
	}
}

func TestContext_ResponseMethods(t *testing.T) {
	app := New()
	
	app.GET("/text", func(ctx *Context) {
		ctx.Text("Hello World")
	})
	
	app.GET("/html", func(ctx *Context) {
		ctx.HTML("<h1>Hello</h1>")
	})
	
	app.GET("/redirect", func(ctx *Context) {
		ctx.Redirect("/new-path")
	})
	
	// 测试 Text 响应
	req := httptest.NewRequest("GET", "/text", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	if w.Body.String() != "Hello World" {
		t.Errorf("Expected 'Hello World', got %s", w.Body.String())
	}
	
	// 测试 HTML 响应
	req = httptest.NewRequest("GET", "/html", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	if !strings.Contains(w.Body.String(), "<h1>Hello</h1>") {
		t.Errorf("Expected HTML content not found")
	}
	
	// 测试重定向
	req = httptest.NewRequest("GET", "/redirect", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	if w.Code != 302 {
		t.Errorf("Expected redirect status 302, got %d", w.Code)
	}
}

func TestContext_RequestMethods(t *testing.T) {
	app := New()
	
	app.POST("/test", func(ctx *Context) {
		method := ctx.Method()
		path := ctx.Path()
		header := ctx.Header("X-Custom")
		
		ctx.JSON(map[string]interface{}{
			"method": method,
			"path":   path,
			"header": header,
		})
	})
	
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-Custom", "test-value")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	if response["method"] != "POST" {
		t.Errorf("Expected method 'POST', got %v", response["method"])
	}
	if response["path"] != "/test" {
		t.Errorf("Expected path '/test', got %v", response["path"])
	}
	if response["header"] != "test-value" {
		t.Errorf("Expected header 'test-value', got %v", response["header"])
	}
}

func TestError_Handling(t *testing.T) {
	app := New()
	
	app.Use(ErrorMiddleware())
	
	app.GET("/error", func(ctx *Context) {
		HandleError(ctx, NotFound("User not found"))
	})
	
	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	if w.Code != 404 {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	errMap, ok := response["error"].(map[string]interface{})
	if !ok {
		t.Errorf("Expected error object")
	}
	
	if errMap["code"] != "NOT_FOUND" {
		t.Errorf("Expected error code 'NOT_FOUND', got %v", errMap["code"])
	}
}

func TestError_PanicRecovery(t *testing.T) {
	app := New()
	
	app.Use(Recovery())
	
	app.GET("/panic", func(ctx *Context) {
		panic("test panic")
	})
	
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	
	// 不应该抛出 panic
	app.ServeHTTP(w, req)
	
	if w.Code != 500 {
		t.Errorf("Expected status 500 after panic, got %d", w.Code)
	}
}

func TestValidation_Basic(t *testing.T) {
	rules := map[string][]Rule{
		"name":  {
			{Required: true, MinLength: ptrInt(2), MaxLength: ptrInt(50)},
		},
		"email": {
			{Required: true},
		},
		"age":   {
			{Type: "number", Min: ptrFloat(0), Max: ptrFloat(150)},
		},
	}
	
	// 测试有效数据
	validData := map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
		"age":   25.0,
	}
	
	errors := Validate(validData, rules)
	if len(errors) != 0 {
		t.Errorf("Expected no errors, got %v", errors)
	}
	
	// 测试无效数据
	invalidData := map[string]interface{}{
		"name":  "A",  // 太短
		"email": "",   // 必填
		"age":   200.0, // 超出范围
	}
	
	errors = Validate(invalidData, rules)
	if len(errors) != 3 {
		t.Errorf("Expected 3 errors, got %d: %v", len(errors), errors)
	}
}

func TestRouter_Wildcard(t *testing.T) {
	app := New()
	
	app.GET("/files/*", func(ctx *Context) {
		wildcard := ctx.Param("wildcard")
		ctx.JSON(map[string]interface{}{
			"path": wildcard,
		})
	})
	
	req := httptest.NewRequest("GET", "/files/path/to/file.txt", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	if response["path"] != "path/to/file.txt" {
		t.Errorf("Expected path 'path/to/file.txt', got %v", response["path"])
	}
}

func TestApp_Settings(t *testing.T) {
	app := New()
	
	app.Enable("trust proxy")
	app.Setting("port", 3000)
	
	if !app.Enabled("trust proxy") {
		t.Errorf("Expected 'trust proxy' to be enabled")
	}
	
	if app.Setting("port") != 3000 {
		t.Errorf("Expected port 3000, got %v", app.Setting("port"))
	}
}

// 辅助函数
func ptrInt(i int) *int {
	return &i
}

func ptrFloat(f float64) *float64 {
	return &f
}

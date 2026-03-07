package http

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewGinAdapter(t *testing.T) {
	config := HTTPConfig{
		Framework:  GinFramework,
		Port:       8080,
		Host:       "localhost",
		Workers:    4,
		MultiNodes: []string{"192.168.1.10:8080", "192.168.1.11:8080"},
	}

	adapter := NewGinAdapter(config)

	if adapter == nil {
		t.Fatal("Expected NewGinAdapter to return non-nil adapter")
	}

	if adapter.Engine == nil {
		t.Error("Expected Engine to be initialized")
	}

	if adapter.Config.Port != 8080 {
		t.Errorf("Expected port to be 8080, got %d", adapter.Config.Port)
	}

	if adapter.Config.Host != "localhost" {
		t.Errorf("Expected host to be 'localhost', got '%s'", adapter.Config.Host)
	}

	if adapter.Config.Framework != GinFramework {
		t.Errorf("Expected framework to be 'gin', got '%s'", adapter.Config.Framework)
	}

	if adapter.Middlewares == nil {
		t.Error("Expected Middlewares slice to be initialized")
	}
}

func TestGinAdapter_AddRoute_GET(t *testing.T) {
	config := HTTPConfig{
		Framework: GinFramework,
		Port:      8081,
		Host:      "127.0.0.1",
	}

	adapter := NewGinAdapter(config)

	handlerCalled := false
	testHandler := func(ctx Context) {
		handlerCalled = true
		ctx.JSON(200, map[string]string{"message": "success"})
	}

	adapter.AddRoute("GET", "/test", testHandler)

	// 创建测试请求
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	adapter.Engine.ServeHTTP(recorder, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if recorder.Code != 200 {
		t.Errorf("Expected status code 200, got %d", recorder.Code)
	}
}

func TestGinAdapter_AddRoute_POST(t *testing.T) {
	config := HTTPConfig{
		Framework: GinFramework,
		Port:      8082,
		Host:      "127.0.0.1",
	}

	adapter := NewGinAdapter(config)

	handlerCalled := false
	testHandler := func(ctx Context) {
		handlerCalled = true
		ctx.JSON(201, map[string]string{"message": "created"})
	}

	adapter.AddRoute("POST", "/create", testHandler)

	req := httptest.NewRequest("POST", "/create", nil)
	recorder := httptest.NewRecorder()

	adapter.Engine.ServeHTTP(recorder, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if recorder.Code != 201 {
		t.Errorf("Expected status code 201, got %d", recorder.Code)
	}
}

func TestGinAdapter_AddRoute_PUT(t *testing.T) {
	config := HTTPConfig{
		Framework: GinFramework,
		Port:      8083,
		Host:      "127.0.0.1",
	}

	adapter := NewGinAdapter(config)

	handlerCalled := false
	testHandler := func(ctx Context) {
		handlerCalled = true
		ctx.JSON(200, map[string]string{"message": "updated"})
	}

	adapter.AddRoute("PUT", "/update", testHandler)

	req := httptest.NewRequest("PUT", "/update", nil)
	recorder := httptest.NewRecorder()

	adapter.Engine.ServeHTTP(recorder, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if recorder.Code != 200 {
		t.Errorf("Expected status code 200, got %d", recorder.Code)
	}
}

func TestGinAdapter_AddRoute_DELETE(t *testing.T) {
	config := HTTPConfig{
		Framework: GinFramework,
		Port:      8084,
		Host:      "127.0.0.1",
	}

	adapter := NewGinAdapter(config)

	handlerCalled := false
	testHandler := func(ctx Context) {
		handlerCalled = true
		ctx.JSON(200, map[string]string{"message": "deleted"})
	}

	adapter.AddRoute("DELETE", "/delete", testHandler)

	req := httptest.NewRequest("DELETE", "/delete", nil)
	recorder := httptest.NewRecorder()

	adapter.Engine.ServeHTTP(recorder, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if recorder.Code != 200 {
		t.Errorf("Expected status code 200, got %d", recorder.Code)
	}
}

func TestGinAdapter_UseMiddlewares(t *testing.T) {
	config := HTTPConfig{
		Framework: GinFramework,
		Port:      8085,
		Host:      "127.0.0.1",
	}

	adapter := NewGinAdapter(config)

	middlewareCalled := false
	middleware := MiddlewareFunc(func(ctx Context) bool {
		middlewareCalled = true
		return true
	})

	adapter.UseMiddlewares(middleware)

	if len(adapter.Middlewares) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(adapter.Middlewares))
	}

	// 测试中间件是否被调用
	testHandler := func(ctx Context) {
		ctx.JSON(200, map[string]string{"message": "success"})
	}

	adapter.AddRoute("GET", "/test", testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	adapter.Engine.ServeHTTP(recorder, req)

	if !middlewareCalled {
		t.Error("Expected middleware to be called")
	}
}

func TestGinAdapter_MiddlewareAborts(t *testing.T) {
	config := HTTPConfig{
		Framework: GinFramework,
		Port:      8086,
		Host:      "127.0.0.1",
	}

	adapter := NewGinAdapter(config)

	// 中间件返回 false，应该中断处理
	middleware := MiddlewareFunc(func(ctx Context) bool {
		ctx.JSON(403, map[string]string{"error": "forbidden"})
		return false
	})

	adapter.UseMiddlewares(middleware)

	handlerCalled := false
	testHandler := func(ctx Context) {
		handlerCalled = true
		ctx.JSON(200, map[string]string{"message": "success"})
	}

	adapter.AddRoute("GET", "/test", testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	adapter.Engine.ServeHTTP(recorder, req)

	// 处理器不应该被调用
	if handlerCalled {
		t.Error("Expected handler NOT to be called when middleware aborts")
	}

	// 应该返回中间件设置的 403 状态码
	if recorder.Code != 403 {
		t.Errorf("Expected status code 403, got %d", recorder.Code)
	}
}

func TestGinAdapter_AddGlobalMiddleware(t *testing.T) {
	config := HTTPConfig{
		Framework: GinFramework,
		Port:      8087,
		Host:      "127.0.0.1",
	}

	adapter := NewGinAdapter(config)

	globalMiddlewareCalled := false
	globalMiddleware := MiddlewareFunc(func(ctx Context) bool {
		globalMiddlewareCalled = true
		return true
	})

	adapter.AddGlobalMiddleware(globalMiddleware)

	if len(adapter.Middlewares) != 1 {
		t.Errorf("Expected 1 global middleware, got %d", len(adapter.Middlewares))
	}

	// 测试全局中间件是否被调用
	testHandler := func(ctx Context) {
		ctx.JSON(200, map[string]string{"message": "success"})
	}

	adapter.AddRoute("GET", "/test", testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	adapter.Engine.ServeHTTP(recorder, req)

	if !globalMiddlewareCalled {
		t.Error("Expected global middleware to be called")
	}
}

func TestGinContext_JSON(t *testing.T) {
	adapter := NewGinAdapter(HTTPConfig{Framework: GinFramework, Port: 8088})

	testHandler := func(ctx Context) {
		ctx.JSON(200, map[string]interface{}{
			"message": "hello",
			"count":   42,
		})
	}

	adapter.AddRoute("GET", "/json", testHandler)

	req := httptest.NewRequest("GET", "/json", nil)
	recorder := httptest.NewRecorder()

	adapter.Engine.ServeHTTP(recorder, req)

	if recorder.Code != 200 {
		t.Errorf("Expected status code 200, got %d", recorder.Code)
	}

	// 验证响应内容类型
	if recorder.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type to be application/json, got %s", recorder.Header().Get("Content-Type"))
	}

	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "hello" {
		t.Errorf("Expected message to be 'hello', got %v", response["message"])
	}
}

func TestGinContext_Param(t *testing.T) {
	adapter := NewGinAdapter(HTTPConfig{Framework: GinFramework, Port: 8089})

	var receivedID string
	testHandler := func(ctx Context) {
		receivedID = ctx.Param("id")
		ctx.JSON(200, map[string]string{"id": receivedID})
	}

	adapter.AddRoute("GET", "/users/:id", testHandler)

	req := httptest.NewRequest("GET", "/users/123", nil)
	recorder := httptest.NewRecorder()

	adapter.Engine.ServeHTTP(recorder, req)

	if receivedID != "123" {
		t.Errorf("Expected param 'id' to be '123', got '%s'", receivedID)
	}
}

func TestGinContext_Query(t *testing.T) {
	adapter := NewGinAdapter(HTTPConfig{Framework: GinFramework, Port: 8090})

	var receivedPage string
	testHandler := func(ctx Context) {
		receivedPage = ctx.Query("page")
		ctx.JSON(200, map[string]string{"page": receivedPage})
	}

	adapter.AddRoute("GET", "/search", testHandler)

	req := httptest.NewRequest("GET", "/search?page=2", nil)
	recorder := httptest.NewRecorder()

	adapter.Engine.ServeHTTP(recorder, req)

	if receivedPage != "2" {
		t.Errorf("Expected query param 'page' to be '2', got '%s'", receivedPage)
	}
}

func TestGinAdapter_Configuration(t *testing.T) {
	expectedConfig := HTTPConfig{
		Framework:  GinFramework,
		Port:       9090,
		Host:       "test.host",
		Workers:    8,
		MultiNodes: []string{"node1:8080", "node2:8080"},
	}

	adapter := NewGinAdapter(expectedConfig)

	if adapter.Config.Port != expectedConfig.Port {
		t.Errorf("Expected port %d, got %d", expectedConfig.Port, adapter.Config.Port)
	}

	if adapter.Config.Host != expectedConfig.Host {
		t.Errorf("Expected host %s, got %s", expectedConfig.Host, adapter.Config.Host)
	}

	if adapter.Config.Framework != expectedConfig.Framework {
		t.Errorf("Expected framework %s, got %s", expectedConfig.Framework, adapter.Config.Framework)
	}

	if adapter.Config.Workers != expectedConfig.Workers {
		t.Errorf("Expected workers %d, got %d", expectedConfig.Workers, adapter.Config.Workers)
	}

	if len(adapter.Config.MultiNodes) != len(expectedConfig.MultiNodes) {
		t.Errorf("Expected %d nodes, got %d", len(expectedConfig.MultiNodes), len(adapter.Config.MultiNodes))
	}
}

func TestGinAdapter_FullIntegration(t *testing.T) {
	config := HTTPConfig{
		Framework: GinFramework,
		Port:      8091,
		Host:      "127.0.0.1",
	}

	adapter := NewGinAdapter(config)

	// 添加中间件
	middlewareCalls := 0
	authMiddleware := MiddlewareFunc(func(ctx Context) bool {
		middlewareCalls++
		// 检查请求头中的 token
		token := ctx.Request().Header.Get("Authorization")
		if token != "valid-token" {
			ctx.JSON(401, map[string]string{"error": "unauthorized"})
			return false
		}
		return true
	})

	adapter.UseMiddlewares(authMiddleware)

	// 添加路由
	testHandler := func(ctx Context) {
		ctx.JSON(200, map[string]interface{}{
			"message": "success",
			"user_id": ctx.Param("id"),
		})
	}

	adapter.AddRoute("GET", "/api/users/:id", testHandler)

	// 测试授权成功的情况
	t.Run("Authorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/456", nil)
		req.Header.Set("Authorization", "valid-token")
		recorder := httptest.NewRecorder()

		adapter.Engine.ServeHTTP(recorder, req)

		if recorder.Code != 200 {
			t.Errorf("Expected status code 200, got %d", recorder.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &response)

		if response["user_id"] != "456" {
			t.Errorf("Expected user_id to be '456', got %v", response["user_id"])
		}
	})

	// 测试未授权的情况
	t.Run("Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/789", nil)
		// 不设置 Authorization header
		recorder := httptest.NewRecorder()

		adapter.Engine.ServeHTTP(recorder, req)

		if recorder.Code != 401 {
			t.Errorf("Expected status code 401, got %d", recorder.Code)
		}
	})

	// 验证中间件被调用了两次
	if middlewareCalls != 2 {
		t.Errorf("Expected middleware to be called 2 times, got %d", middlewareCalls)
	}
}

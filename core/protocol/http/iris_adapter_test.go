package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewIrisAdapter(t *testing.T) {
	config := HTTPConfig{
		Framework:  IrisFramework,
		Port:       8080,
		Host:       "localhost",
		Workers:    4,
		MultiNodes: []string{"192.168.1.10:8080", "192.168.1.11:8080"},
	}

	adapter := NewIrisAdapter(config)

	if adapter == nil {
		t.Fatal("Expected NewIrisAdapter to return non-nil adapter")
	}

	if adapter.app == nil {
		t.Error("Expected app to be initialized")
	}

	if adapter.config.Port != 8080 {
		t.Errorf("Expected port to be 8080, got %d", adapter.config.Port)
	}

	if adapter.config.Host != "localhost" {
		t.Errorf("Expected host to be 'localhost', got '%s'", adapter.config.Host)
	}

	if adapter.config.Framework != IrisFramework {
		t.Errorf("Expected framework to be 'iris', got '%s'", adapter.config.Framework)
	}
}

func TestIrisAdapter_AddRoute_GET(t *testing.T) {
	config := HTTPConfig{
		Framework: IrisFramework,
		Port:      8081,
		Host:      "127.0.0.1",
	}

	adapter := NewIrisAdapter(config)

	handlerCalled := false
	testHandler := func(ctx Context) {
		handlerCalled = true
		ctx.JSON(200, map[string]string{"message": "success"})
	}

	adapter.AddRoute("GET", "/test", testHandler)

	// 构建路由器
	adapter.app.Build()

	// 创建测试请求
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	adapter.app.ServeHTTP(recorder, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if recorder.Code != 200 {
		t.Errorf("Expected status code 200, got %d", recorder.Code)
	}
}

func TestIrisAdapter_AddRoute_POST(t *testing.T) {
	config := HTTPConfig{
		Framework: IrisFramework,
		Port:      8082,
		Host:      "127.0.0.1",
	}

	adapter := NewIrisAdapter(config)

	handlerCalled := false
	testHandler := func(ctx Context) {
		handlerCalled = true
		ctx.JSON(201, map[string]string{"message": "created"})
	}

	adapter.AddRoute("POST", "/create", testHandler)
	adapter.app.Build()

	req := httptest.NewRequest("POST", "/create", nil)
	recorder := httptest.NewRecorder()

	adapter.app.ServeHTTP(recorder, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if recorder.Code != 201 {
		t.Errorf("Expected status code 201, got %d", recorder.Code)
	}
}

func TestIrisAdapter_AddRoute_PUT(t *testing.T) {
	config := HTTPConfig{
		Framework: IrisFramework,
		Port:      8083,
		Host:      "127.0.0.1",
	}

	adapter := NewIrisAdapter(config)

	handlerCalled := false
	testHandler := func(ctx Context) {
		handlerCalled = true
		ctx.JSON(200, map[string]string{"message": "updated"})
	}

	adapter.AddRoute("PUT", "/update", testHandler)
	adapter.app.Build()

	req := httptest.NewRequest("PUT", "/update", nil)
	recorder := httptest.NewRecorder()

	adapter.app.ServeHTTP(recorder, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if recorder.Code != 200 {
		t.Errorf("Expected status code 200, got %d", recorder.Code)
	}
}

func TestIrisAdapter_AddRoute_DELETE(t *testing.T) {
	config := HTTPConfig{
		Framework: IrisFramework,
		Port:      8084,
		Host:      "127.0.0.1",
	}

	adapter := NewIrisAdapter(config)

	handlerCalled := false
	testHandler := func(ctx Context) {
		handlerCalled = true
		ctx.JSON(200, map[string]string{"message": "deleted"})
	}

	adapter.AddRoute("DELETE", "/delete", testHandler)
	adapter.app.Build()

	req := httptest.NewRequest("DELETE", "/delete", nil)
	recorder := httptest.NewRecorder()

	adapter.app.ServeHTTP(recorder, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if recorder.Code != 200 {
		t.Errorf("Expected status code 200, got %d", recorder.Code)
	}
}

func TestIrisAdapter_UseMiddlewares(t *testing.T) {
	config := HTTPConfig{
		Framework: IrisFramework,
		Port:      8085,
		Host:      "127.0.0.1",
	}

	adapter := NewIrisAdapter(config)

	middlewareCalled := false
	middleware := MiddlewareFunc(func(ctx Context) bool {
		middlewareCalled = true
		return true
	})

	adapter.UseMiddlewares(middleware)

	// 测试中间件是否被调用
	testHandler := func(ctx Context) {
		ctx.JSON(200, map[string]string{"message": "success"})
	}

	adapter.AddRoute("GET", "/test", testHandler)
	adapter.app.Build()

	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	adapter.app.ServeHTTP(recorder, req)

	if !middlewareCalled {
		t.Error("Expected middleware to be called")
	}
}

func TestIrisAdapter_MiddlewareAborts(t *testing.T) {
	config := HTTPConfig{
		Framework: IrisFramework,
		Port:      8086,
		Host:      "127.0.0.1",
	}

	adapter := NewIrisAdapter(config)

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
	adapter.app.Build()

	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	adapter.app.ServeHTTP(recorder, req)

	// 处理器不应该被调用
	if handlerCalled {
		t.Error("Expected handler NOT to be called when middleware aborts")
	}

	// 应该返回中间件设置的 403 状态码
	if recorder.Code != 403 {
		t.Errorf("Expected status code 403, got %d", recorder.Code)
	}
}

func TestIrisAdapter_AddGlobalMiddleware(t *testing.T) {
	config := HTTPConfig{
		Framework: IrisFramework,
		Port:      8087,
		Host:      "127.0.0.1",
	}

	adapter := NewIrisAdapter(config)

	globalMiddlewareCalled := false
	globalMiddleware := MiddlewareFunc(func(ctx Context) bool {
		globalMiddlewareCalled = true
		return true
	})

	adapter.AddGlobalMiddleware(globalMiddleware)

	// 测试全局中间件是否被调用
	testHandler := func(ctx Context) {
		ctx.JSON(200, map[string]string{"message": "success"})
	}

	adapter.AddRoute("GET", "/test", testHandler)
	adapter.app.Build()

	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	adapter.app.ServeHTTP(recorder, req)

	if !globalMiddlewareCalled {
		t.Error("Expected global middleware to be called")
	}
}

func TestIrisContext_JSON(t *testing.T) {
	adapter := NewIrisAdapter(HTTPConfig{Framework: IrisFramework, Port: 8088})

	testHandler := func(ctx Context) {
		ctx.JSON(200, map[string]interface{}{
			"message": "hello",
			"count":   42,
		})
	}

	adapter.AddRoute("GET", "/json", testHandler)
	adapter.app.Build()

	req := httptest.NewRequest("GET", "/json", nil)
	recorder := httptest.NewRecorder()

	adapter.app.ServeHTTP(recorder, req)

	if recorder.Code != 200 {
		t.Errorf("Expected status code 200, got %d", recorder.Code)
	}

	// 验证响应内容类型
	contentType := recorder.Header().Get("Content-Type")
	if contentType == "" {
		t.Log("Warning: Content-Type not set")
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

	if int(response["count"].(float64)) != 42 {
		t.Errorf("Expected count to be 42, got %v", response["count"])
	}
}

func TestIrisContext_Param(t *testing.T) {
	adapter := NewIrisAdapter(HTTPConfig{Framework: IrisFramework, Port: 8089})

	var receivedID string
	testHandler := func(ctx Context) {
		receivedID = ctx.Param("id")
		ctx.JSON(200, map[string]string{"id": receivedID})
	}

	adapter.AddRoute("GET", "/users/{id}", testHandler)
	adapter.app.Build()

	req := httptest.NewRequest("GET", "/users/123", nil)
	recorder := httptest.NewRecorder()

	adapter.app.ServeHTTP(recorder, req)

	if receivedID != "123" {
		t.Errorf("Expected param 'id' to be '123', got '%s'", receivedID)
	}
}

func TestIrisContext_Query(t *testing.T) {
	adapter := NewIrisAdapter(HTTPConfig{Framework: IrisFramework, Port: 8090})

	var receivedPage string
	testHandler := func(ctx Context) {
		receivedPage = ctx.Query("page")
		ctx.JSON(200, map[string]string{"page": receivedPage})
	}

	adapter.AddRoute("GET", "/search", testHandler)
	adapter.app.Build()

	req := httptest.NewRequest("GET", "/search?page=2", nil)
	recorder := httptest.NewRecorder()

	adapter.app.ServeHTTP(recorder, req)

	if receivedPage != "2" {
		t.Errorf("Expected query param 'page' to be '2', got '%s'", receivedPage)
	}
}

func TestIrisContext_Bind(t *testing.T) {
	adapter := NewIrisAdapter(HTTPConfig{Framework: IrisFramework, Port: 8091})

	var receivedData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	testHandler := func(ctx Context) {
		err := ctx.Bind(&receivedData)
		if err != nil {
			t.Errorf("Bind failed: %v", err)
		}
		ctx.JSON(200, map[string]string{"status": "ok"})
	}

	adapter.AddRoute("POST", "/bind", testHandler)
	adapter.app.Build()

	// 创建包含 JSON 数据的请求
	jsonData := strings.NewReader(`{"name":"John","age":30}`)
	req := httptest.NewRequest("POST", "/bind", jsonData)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	adapter.app.ServeHTTP(recorder, req)

	// 验证状态码
	if recorder.Code != 200 {
		t.Errorf("Expected status code 200, got %d", recorder.Code)
	}
}

func TestIrisAdapter_Configuration(t *testing.T) {
	expectedConfig := HTTPConfig{
		Framework:  IrisFramework,
		Port:       9090,
		Host:       "test.host",
		Workers:    8,
		MultiNodes: []string{"node1:8080", "node2:8080"},
	}

	adapter := NewIrisAdapter(expectedConfig)

	if adapter.config.Port != expectedConfig.Port {
		t.Errorf("Expected port %d, got %d", expectedConfig.Port, adapter.config.Port)
	}

	if adapter.config.Host != expectedConfig.Host {
		t.Errorf("Expected host %s, got %s", expectedConfig.Host, adapter.config.Host)
	}

	if adapter.config.Framework != expectedConfig.Framework {
		t.Errorf("Expected framework %s, got %s", expectedConfig.Framework, adapter.config.Framework)
	}

	if adapter.config.Workers != expectedConfig.Workers {
		t.Errorf("Expected workers %d, got %d", expectedConfig.Workers, adapter.config.Workers)
	}

	if len(adapter.config.MultiNodes) != len(expectedConfig.MultiNodes) {
		t.Errorf("Expected %d nodes, got %d", len(expectedConfig.MultiNodes), len(adapter.config.MultiNodes))
	}
}

func TestIrisAdapter_FullIntegration(t *testing.T) {
	config := HTTPConfig{
		Framework: IrisFramework,
		Port:      8092,
		Host:      "127.0.0.1",
	}

	adapter := NewIrisAdapter(config)

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

	adapter.AddRoute("GET", "/api/users/{id}", testHandler)
	adapter.app.Build()

	// 测试授权成功的情况
	t.Run("Authorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/456", nil)
		req.Header.Set("Authorization", "valid-token")
		recorder := httptest.NewRecorder()

		adapter.app.ServeHTTP(recorder, req)

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

		adapter.app.ServeHTTP(recorder, req)

		if recorder.Code != 401 {
			t.Errorf("Expected status code 401, got %d", recorder.Code)
		}
	})

	// 验证中间件被调用了两次
	if middlewareCalls != 2 {
		t.Errorf("Expected middleware to be called 2 times, got %d", middlewareCalls)
	}
}

func TestIrisContext_RequestResponse(t *testing.T) {
	adapter := NewIrisAdapter(HTTPConfig{Framework: IrisFramework, Port: 8093})

	var capturedRequest *http.Request
	var capturedResponse http.ResponseWriter

	testHandler := func(ctx Context) {
		capturedRequest = ctx.Request()
		capturedResponse = ctx.ResponseWriter()
		ctx.JSON(200, map[string]string{"status": "ok"})
	}

	adapter.AddRoute("GET", "/test", testHandler)
	adapter.app.Build()

	req := httptest.NewRequest("GET", "/test?test=value", nil)
	recorder := httptest.NewRecorder()

	adapter.app.ServeHTTP(recorder, req)

	if capturedRequest == nil {
		t.Error("Expected Request to be captured")
	} else {
		if capturedRequest.URL.Query().Get("test") != "value" {
			t.Errorf("Expected query param 'test' to be 'value', got '%s'", capturedRequest.URL.Query().Get("test"))
		}
	}

	if capturedResponse == nil {
		t.Error("Expected ResponseWriter to be captured")
	}
}

func TestIrisAdapter_MultipleMiddlewares(t *testing.T) {
	adapter := NewIrisAdapter(HTTPConfig{Framework: IrisFramework, Port: 8094})

	callOrder := []string{}

	middleware1 := MiddlewareFunc(func(ctx Context) bool {
		callOrder = append(callOrder, "middleware1")
		return true
	})

	middleware2 := MiddlewareFunc(func(ctx Context) bool {
		callOrder = append(callOrder, "middleware2")
		return true
	})

	adapter.UseMiddlewares(middleware1, middleware2)

	testHandler := func(ctx Context) {
		callOrder = append(callOrder, "handler")
		ctx.JSON(200, map[string]string{"status": "ok"})
	}

	adapter.AddRoute("GET", "/test", testHandler)
	adapter.app.Build()

	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	adapter.app.ServeHTTP(recorder, req)

	// 验证中间件和处理器按顺序被调用
	expected := []string{"middleware1", "middleware2", "handler"}
	if len(callOrder) != len(expected) {
		t.Errorf("Expected %d calls, got %d", len(expected), len(callOrder))
	}

	for i, expectedCall := range expected {
		if i < len(callOrder) && callOrder[i] != expectedCall {
			t.Errorf("Expected call %d to be '%s', got '%s'", i, expectedCall, callOrder[i])
		}
	}
}

func TestIrisAdapter_AnyMethod(t *testing.T) {
	adapter := NewIrisAdapter(HTTPConfig{Framework: IrisFramework, Port: 8095})

	handlerCalled := false
	testHandler := func(ctx Context) {
		handlerCalled = true
		ctx.JSON(200, map[string]string{"method": ctx.Request().Method})
	}

	// 测试默认路由（Any 方法）
	adapter.AddRoute("CUSTOM", "/custom", testHandler)
	adapter.app.Build()

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			handlerCalled = false
			req := httptest.NewRequest(method, "/custom", nil)
			recorder := httptest.NewRecorder()

			adapter.app.ServeHTTP(recorder, req)

			if !handlerCalled {
				t.Errorf("Expected handler to be called for method %s", method)
			}
		})
	}
}

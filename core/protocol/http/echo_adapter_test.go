package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// EchoMockContext Echo适配器测试的模拟上下文
type EchoMockContext struct {
	responseWriter http.ResponseWriter
	request        *http.Request
	jsonCode       int
	jsonData       interface{}
}

func (emc *EchoMockContext) JSON(code int, obj interface{}) {
	emc.jsonCode = code
	emc.jsonData = obj
}

func (emc *EchoMockContext) Param(key string) string {
	return emc.request.URL.Query().Get(key) // 简单实现
}

func (emc *EchoMockContext) Query(key string) string {
	return emc.request.URL.Query().Get(key)
}

func (emc *EchoMockContext) Bind(obj interface{}) error {
	// 简单实现：不处理绑定
	return nil
}

func (emc *EchoMockContext) Request() *http.Request {
	return emc.request
}

func (emc *EchoMockContext) ResponseWriter() http.ResponseWriter {
	return emc.responseWriter
}

func TestNewEchoAdapter(t *testing.T) {
	config := HTTPConfig{
		Framework:  EchoFramework,
		Port:       8080,
		Host:       "localhost",
		Workers:    4,
		MultiNodes: []string{"192.168.1.10:8080", "192.168.1.11:8080"},
	}

	adapter := NewEchoAdapter(config)

	if adapter == nil {
		t.Fatal("Expected NewEchoAdapter to return non-nil adapter")
	}

	if adapter.config.Port != 8080 {
		t.Errorf("Expected port to be 8080, got %d", adapter.config.Port)
	}

	if adapter.config.Host != "localhost" {
		t.Errorf("Expected host to be 'localhost', got '%s'", adapter.config.Host)
	}

	if adapter.config.Framework != EchoFramework {
		t.Errorf("Expected framework to be 'echo', got '%s'", adapter.config.Framework)
	}

	// 验证 handlers 映射初始化
	if adapter.handlers == nil {
		t.Error("Expected handlers map to be initialized")
	}

	if adapter.handlers["GET"] == nil {
		t.Error("Expected GET handler map to be initialized")
	}

	if adapter.handlers["POST"] == nil {
		t.Error("Expected POST handler map to be initialized")
	}

	if adapter.handlers["PUT"] == nil {
		t.Error("Expected PUT handler map to be initialized")
	}

	if adapter.handlers["DELETE"] == nil {
		t.Error("Expected DELETE handler map to be initialized")
	}
}

func TestEchoAdapter_AddRoute(t *testing.T) {
	config := HTTPConfig{
		Framework: EchoFramework,
		Port:      8081,
		Host:      "127.0.0.1",
	}

	adapter := NewEchoAdapter(config)

	// 创建测试处理器
	testHandler := func(ctx Context) {
		// 只是记录一下被调用
	}

	// 添加路由
	adapter.AddRoute("GET", "/test", testHandler)

	// 检查路由是否注册
	if adapter.handlers["GET"]["/test"] == nil {
		t.Error("Expected handler to be registered for GET /test")
	} else {
		// 验证处理器存在
		t.Logf("Handler is registered for GET /test")
	}

	// 额外测试：不同方法类型的路由注册
	adapter.AddRoute("POST", "/post-test", testHandler)
	if adapter.handlers["POST"]["/post-test"] == nil {
		t.Error("Expected handler to be registered for POST /post-test")
	} else {
		t.Logf("Handler is registered for POST /post-test")
	}
}

func TestEchoAdapter_UseMiddlewares(t *testing.T) {
	config := HTTPConfig{
		Framework: EchoFramework,
		Port:      8082,
		Host:      "127.0.0.1",
	}

	adapter := NewEchoAdapter(config)

	// 创建测试中间件 - 总是返回true，允许通过
	middleware1 := MiddlewareFunc(func(ctx Context) bool {
		return true
	})

	middleware2 := MiddlewareFunc(func(ctx Context) bool {
		return true
	})

	// 应用中间件 - 对于当前Echo适配器实现，UseMiddlewares只返回自身
	modifiedServer := adapter.UseMiddlewares(middleware1, middleware2)

	// 验证返回值是同一实例
	if modifiedServer != adapter {
		t.Error("Expected UseMiddlewares to return the same instance")
	}
}

func TestEchoAdapter_AddGlobalMiddleware(t *testing.T) {
	config := HTTPConfig{
		Framework: EchoFramework,
		Port:      8083,
		Host:      "127.0.0.1",
	}

	adapter := NewEchoAdapter(config)

	// 创建一个全局中间件
	globalMiddleware := MiddlewareFunc(func(ctx Context) bool {
		return true
	})

	// 添加全局中间件
	// 在当前Echo适配器实现中，这是空实现，但我们仍然测试其调用
	adapter.AddGlobalMiddleware(globalMiddleware)

	// 确保方法调用不会导致panic
	t.Log("Successfully called AddGlobalMiddleware - implementation doesn't store middleware yet")
}

// 更完善的Echo适配器集成测试
func TestEchoAdapter_FullIntegration(t *testing.T) {
	config := HTTPConfig{
		Framework: EchoFramework,
		Port:      8084,
		Host:      "127.0.0.1",
	}

	adapter := NewEchoAdapter(config)

	// 注册一个测试路由
	responseHandled := false
	testHandler := func(ctx Context) {
		ctx.JSON(200, map[string]string{"message": "success"})
		responseHandled = true
	}

	adapter.AddRoute("GET", "/integration-test", testHandler)

	// 验证路由是否已注册
	routeHandler := adapter.handlers["GET"]["/integration-test"]
	if routeHandler == nil {
		t.Fatal("Expected route handler to be registered for GET /integration-test")
	}

	// 创建模拟的上下文用于测试
	mockWriter := httptest.NewRecorder()
	mockReq := httptest.NewRequest("GET", "/integration-test", nil)

	// 直接调用注册的处理器
	mockCtx := &EchoContext{
		writer:  mockWriter,
		request: mockReq,
		params:  make(map[string]string),
	}

	routeHandler(mockCtx)

	// 验证是否调用了处理器
	if !responseHandled {
		t.Error("Expected handler to be called")
	} else {
		t.Log("Handler was successfully called in integration test")
	}
}

func TestEchoAdapter_Configuration(t *testing.T) {
	expectedConfig := HTTPConfig{
		Framework:  EchoFramework,
		Port:       9090,
		Host:       "test.host",
		Workers:    8,
		MultiNodes: []string{"node1:8080", "node2:8080"},
	}

	adapter := NewEchoAdapter(expectedConfig)

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

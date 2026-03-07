package http

import (
	"net/http/httptest"
	"testing"

	"gospacex/core/common"
)

// TestHTTPProtocol_Interface 测试 HTTPProtocol 接口类型别名
func TestHTTPProtocol_Interface(t *testing.T) {
	// 验证 HTTPProtocol 类型别名的正确性
	var _ HTTPProtocol = (*MockHTTPServer)(nil)
	t.Log("HTTPProtocol interface type alias is valid")
}

// TestHandlerFunc_Type 测试 HandlerFunc 类型别名
func TestHandlerFunc_Type(t *testing.T) {
	// 创建一个 HandlerFunc
	handler := func(ctx Context) {
		ctx.JSON(200, map[string]string{"message": "test"})
	}

	// 验证可以赋值给 HandlerFunc 类型
	var h HandlerFunc = handler
	if h == nil {
		t.Error("HandlerFunc should not be nil")
	}

	// 验证可以调用
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	mockCtx := &MockContext{
		responseWriter: recorder,
		request:        req,
	}

	h(mockCtx)

	if mockCtx.jsonCode != 200 {
		t.Errorf("Expected status code 200, got %d", mockCtx.jsonCode)
	}
}

// TestContext_Interface 测试 Context 接口
func TestContext_Interface(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	recorder := httptest.NewRecorder()

	// 使用 chi_adapter_test.go 中定义的 MockContext
	mockCtx := &MockContext{
		responseWriter: recorder,
		request:        req,
	}

	// 测试所有 Context 接口方法
	t.Run("JSON", func(t *testing.T) {
		mockCtx.JSON(200, map[string]string{"status": "ok"})
		if mockCtx.jsonCode != 200 {
			t.Errorf("Expected JSON code 200, got %d", mockCtx.jsonCode)
		}
	})

	t.Run("Query", func(t *testing.T) {
		value := mockCtx.Query("param")
		if value != "value" {
			t.Errorf("Expected query param 'value', got '%s'", value)
		}
	})

	t.Run("Request", func(t *testing.T) {
		r := mockCtx.Request()
		if r == nil {
			t.Error("Request should not be nil")
		}
	})

	t.Run("ResponseWriter", func(t *testing.T) {
		w := mockCtx.ResponseWriter()
		if w == nil {
			t.Error("ResponseWriter should not be nil")
		}
	})
}

// TestMiddleware_Interface 测试 Middleware 接口
func TestMiddleware_Interface(t *testing.T) {
	// 创建一个中间件
	middleware := MiddlewareFunc(func(ctx Context) bool {
		return true
	})

	// 验证 MiddlewareFunc 实现了 Middleware 接口
	var _ Middleware = middleware

	// 测试 Process 方法
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	mockCtx := &MockContext{
		responseWriter: recorder,
		request:        req,
	}

	result := middleware.Process(mockCtx)
	if !result {
		t.Error("Expected middleware to return true")
	}
}

// TestMiddlewareFunc_Process 测试 MiddlewareFunc 的 Process 方法
func TestMiddlewareFunc_Process(t *testing.T) {
	tests := []struct {
		name     string
		process  func(ctx Context) bool
		expected bool
	}{
		{
			name: "returns true",
			process: func(ctx Context) bool {
				return true
			},
			expected: true,
		},
		{
			name: "returns false",
			process: func(ctx Context) bool {
				return false
			},
			expected: false,
		},
		{
			name: "with context operations",
			process: func(ctx Context) bool {
				ctx.JSON(200, map[string]string{"status": "ok"})
				return true
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := MiddlewareFunc(tt.process)

			req := httptest.NewRequest("GET", "/test", nil)
			recorder := httptest.NewRecorder()
			mockCtx := &MockContext{
				responseWriter: recorder,
				request:        req,
			}

			result := middleware.Process(mockCtx)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestTypeAliases_Compatability 测试类型别名与 common 包的兼容性
func TestTypeAliases_Compatability(t *testing.T) {
	// 测试 HTTPProtocol 类型别名
	var protocol HTTPProtocol
	var commonProtocol common.HTTPProtocol = protocol
	_ = commonProtocol // 避免未使用变量错误
	t.Log("HTTPProtocol type alias is compatible with common.HTTPProtocol")

	// 测试 HandlerFunc 类型别名
	var handler HandlerFunc
	var commonHandler common.HTTPHandlerFunc = handler
	_ = commonHandler
	t.Log("HandlerFunc type alias is compatible with common.HTTPHandlerFunc")

	// 测试 Context 类型别名
	var ctx Context
	var commonCtx common.HTTPContext = ctx
	_ = commonCtx
	t.Log("Context type alias is compatible with common.HTTPContext")

	// 测试 Middleware 类型别名
	var mid Middleware
	var commonMid common.HTTPMiddleware = mid
	_ = commonMid
	t.Log("Middleware type alias is compatible with common.HTTPMiddleware")

	// 测试 MiddlewareFunc 类型别名
	var midFunc MiddlewareFunc
	var commonMidFunc common.HTTPMiddlewareFunc = midFunc
	_ = commonMidFunc
	t.Log("MiddlewareFunc type alias is compatible with common.HTTPMiddlewareFunc")
}

// TestMiddlewareFunc_ImplementsMiddleware 测试 MiddlewareFunc 实现 Middleware 接口
func TestMiddlewareFunc_ImplementsMiddleware(t *testing.T) {
	// 验证 MiddlewareFunc 实现了 Middleware 接口
	var _ Middleware = MiddlewareFunc(func(ctx Context) bool {
		return true
	})
	t.Log("MiddlewareFunc implements Middleware interface")
}

// TestMultipleMiddlewares 测试多个中间件的链式调用
func TestMultipleMiddlewares(t *testing.T) {
	callOrder := []string{}

	middleware1 := MiddlewareFunc(func(ctx Context) bool {
		callOrder = append(callOrder, "middleware1")
		return true
	})

	middleware2 := MiddlewareFunc(func(ctx Context) bool {
		callOrder = append(callOrder, "middleware2")
		return true
	})

	middleware3 := MiddlewareFunc(func(ctx Context) bool {
		callOrder = append(callOrder, "middleware3")
		return true
	})

	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	mockCtx := &MockContext{
		responseWriter: recorder,
		request:        req,
	}

	// 按顺序调用中间件
	if !middleware1.Process(mockCtx) {
		t.Error("Middleware1 should return true")
	}
	if !middleware2.Process(mockCtx) {
		t.Error("Middleware2 should return true")
	}
	if !middleware3.Process(mockCtx) {
		t.Error("Middleware3 should return true")
	}

	// 验证调用顺序
	expected := []string{"middleware1", "middleware2", "middleware3"}
	if len(callOrder) != len(expected) {
		t.Errorf("Expected %d calls, got %d", len(expected), len(callOrder))
		return
	}

	for i, expectedCall := range expected {
		if callOrder[i] != expectedCall {
			t.Errorf("Expected call %d to be '%s', got '%s'", i, expectedCall, callOrder[i])
		}
	}
}

// TestMiddlewareAbortsChain 测试中间件中断链
func TestMiddlewareAbortsChain(t *testing.T) {
	callOrder := []string{}

	middleware1 := MiddlewareFunc(func(ctx Context) bool {
		callOrder = append(callOrder, "middleware1")
		return true
	})

	middleware2 := MiddlewareFunc(func(ctx Context) bool {
		callOrder = append(callOrder, "middleware2")
		return false // 中断链
	})

	middleware3 := MiddlewareFunc(func(ctx Context) bool {
		callOrder = append(callOrder, "middleware3")
		return true
	})

	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	mockCtx := &MockContext{
		responseWriter: recorder,
		request:        req,
	}

	// 模拟中间件链式调用
	proceed := middleware1.Process(mockCtx)
	if !proceed {
		t.Error("Middleware1 should return true")
	}

	proceed = middleware2.Process(mockCtx)
	if proceed {
		t.Error("Middleware2 should return false")
	}

	// 如果 middleware2 返回 false，middleware3 不应该被调用
	if !proceed {
		// 不调用 middleware3
	} else {
		// 只有 proceed 为 true 时才调用 middleware3
		middleware3.Process(mockCtx)
	}

	// 验证调用顺序（middleware3 不应该被调用）
	expected := []string{"middleware1", "middleware2"}
	if len(callOrder) != len(expected) {
		t.Errorf("Expected %d calls, got %d: %v", len(expected), len(callOrder), callOrder)
		return
	}

	for i, expectedCall := range expected {
		if callOrder[i] != expectedCall {
			t.Errorf("Expected call %d to be '%s', got '%s'", i, expectedCall, callOrder[i])
		}
	}
}

// MockHTTPServer 用于测试的模拟 HTTP 服务器
type MockHTTPServer struct{}

func (m *MockHTTPServer) Start() error {
	return nil
}

func (m *MockHTTPServer) Stop() error {
	return nil
}

func (m *MockHTTPServer) AddRoute(method, path string, handler HandlerFunc) {
}

func (m *MockHTTPServer) UseMiddlewares(mids ...Middleware) HTTPProtocol {
	return m
}

func (m *MockHTTPServer) AddGlobalMiddleware(middleware Middleware) {
}

// BenchmarkMiddlewareFunc_Process 基准测试：MiddlewareFunc 的 Process 方法
func BenchmarkMiddlewareFunc_Process(b *testing.B) {
	middleware := MiddlewareFunc(func(ctx Context) bool {
		return true
	})

	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	mockCtx := &MockContext{
		responseWriter: recorder,
		request:        req,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware.Process(mockCtx)
	}
}

// BenchmarkContext_JSON 基准测试：Context 的 JSON 方法
func BenchmarkContext_JSON(b *testing.B) {
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	mockCtx := &MockContext{
		responseWriter: recorder,
		request:        req,
	}

	data := map[string]string{
		"message": "hello world",
		"status":  "ok",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockCtx.JSON(200, data)
	}
}

// BenchmarkContext_Query 基准测试：Context 的 Query 方法
func BenchmarkContext_Query(b *testing.B) {
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	recorder := httptest.NewRecorder()
	mockCtx := &MockContext{
		responseWriter: recorder,
		request:        req,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockCtx.Query("param")
	}
}

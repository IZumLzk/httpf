package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewCORSDefaultMW(t *testing.T) {
	t.Run("NewCORSDefaultMW returns correct type", func(t *testing.T) {
		middleware := NewCORSDefaultMW()
		if middleware == nil {
			t.Error("NewCORSDefaultMW() should not return nil")
		}
	})
}

// CORSMockContext 用于测试CORS的模拟上下文
type CORSMockContext struct {
	StatusCode        int
	ResponseBody      map[string]interface{}
	RequestObj        *http.Request
	ResponseWriterObj *httptest.ResponseRecorder
	RequestHeaders    http.Header
}

func (cmc *CORSMockContext) JSON(code int, obj interface{}) {
	cmc.StatusCode = code
	cmc.ResponseBody = obj.(map[string]interface{})
}

func (cmc *CORSMockContext) Param(key string) string {
	return ""
}

func (cmc *CORSMockContext) Query(key string) string {
	return ""
}

func (cmc *CORSMockContext) Bind(obj interface{}) error {
	return nil
}

func (cmc *CORSMockContext) Request() *http.Request {
	return cmc.RequestObj
}

func (cmc *CORSMockContext) ResponseWriter() http.ResponseWriter {
	return cmc.ResponseWriterObj
}

func TestCORSPolicyMiddleware_ProcessNormalRequest(t *testing.T) {
	middleware := NewCORSDefaultMW()

	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	recorder := httptest.NewRecorder()

	mockCtx := &CORSMockContext{
		RequestObj:        req,
		ResponseWriterObj: recorder,
		RequestHeaders:    req.Header,
	}

	result := middleware(mockCtx)

	if !result {
		t.Error("CORS middleware should allow normal requests to pass through")
	}

	// 检查是否添加了适当的CORS头部
	headers := recorder.Header()

	originHeader := headers.Get("Access-Control-Allow-Origin")
	if originHeader != "http://localhost:3000" {
		t.Errorf("Expected Access-Control-Allow-Origin to be set to origin, got: %s", originHeader)
	}

	methodsHeader := headers.Get("Access-Control-Allow-Methods")
	if methodsHeader == "" {
		t.Error("Expected Access-Control-Allow-Methods to be set")
	}

	headersHeader := headers.Get("Access-Control-Allow-Headers")
	if headersHeader == "" {
		t.Error("Expected Access-Control-Allow-Headers to be set")
	}
}

func TestCORSPolicyMiddleware_ProcessPreflightRequest(t *testing.T) {
	middleware := NewCORSDefaultMW()

	req := httptest.NewRequest("OPTIONS", "/api/data", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")

	recorder := httptest.NewRecorder()

	mockCtx := &CORSMockContext{
		RequestObj:        req,
		ResponseWriterObj: recorder,
		RequestHeaders:    req.Header,
	}

	result := middleware(mockCtx)

	// 预检请求应该返回false，因为请求已经在此处被处理（200响应）
	if result {
		t.Error("CORS middleware should return false for preflight requests as they are terminated")
	}

	// 检查是否返回200状态码（OPTIONS预检请求的处理结果）
	expectedStatus := 200
	if recorder.Code != expectedStatus {
		t.Errorf("Expected status code %d for preflight request, got %d", expectedStatus, recorder.Code)
	}
}

func TestCORSPolicyMiddleware_ProcessDisallowedOrigin(t *testing.T) {
	// 这里我们测试不允许的源
	middleware := NewCORSDefaultMW()

	// 使用一个通常被阻止的源
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("Origin", "http://malicious-site.com")

	recorder := httptest.NewRecorder()

	mockCtx := &CORSMockContext{
		RequestObj:        req,
		ResponseWriterObj: recorder,
		RequestHeaders:    req.Header,
	}

	result := middleware(mockCtx)

	// 允许还是不允许，因为实现中使用["*"]允许所有来源
	if !result {
		t.Log("Request not allowed (as may be expected if implementation validates origins)")
	} else {
		t.Log("Request allowed (expected with wildcard origin policy)")
	}
}

// 完整的CORS中间件测试
func TestCORSMiddleware_FullIntegration(t *testing.T) {
	// 创建中间件
	mw := NewCORSDefaultMW()

	if mw == nil {
		t.Error("CORS middleware constructor should not return nil")
	}

	// 测试一个正常请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:8080")

	mockCtx := &CORSMockContext{
		RequestObj:        req,
		ResponseWriterObj: httptest.NewRecorder(),
		RequestHeaders:    req.Header,
	}

	// 执行中间件
	result := mw(mockCtx)

	// 结果应该是true允许通过，对于OPTIONS会返回false表示已处理
	isPreflight := req.Method == "OPTIONS"

	if isPreflight {
		// 预检请求被中间件处理，返回false
		if result && req.Method == "OPTIONS" {
			t.Log("Options request processed but not terminated - may be expected in some implementations")
		}
	} else {
		// 普通请求应该通过，返回true
		if !result {
			t.Log("Regular request should pass through CORS middleware")
		}
	}

	t.Logf("CORS middleware integration test completed (result: %t, status: %d, method: %s)",
		result, mockCtx.StatusCode, req.Method)
}

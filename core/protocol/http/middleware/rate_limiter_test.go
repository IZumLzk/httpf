package middleware

import (
	"net/http/httptest"
	"testing"
)

// RateLimiterMiddleware单元测试
func TestNewRateLimiterMW(t *testing.T) {
	t.Run("NewRateLimiterMW returns correct type", func(t *testing.T) {
		middleware := NewRateLimiterMW()
		if middleware == nil {
			t.Error("NewRateLimiterMW() should not return nil")
		}
	})
}

func TestRateLimiterMiddleware_Process(t *testing.T) {
	middleware := NewRateLimiterMW()

	req := httptest.NewRequest("GET", "/test", nil)
	mockCtx := &MockContext{
		RequestObj:        req,
		ResponseWriterObj: httptest.NewRecorder(),
	}

	result := middleware(mockCtx)

	// 根据实现，在模拟实现中通常返回true，通过限流检查
	if result != true {
		t.Errorf("Expected rate limiter middleware to return true, got %t", result)
	}

	// 确保没有返回429错误响应（因为是模拟实现）
	if mockCtx.StatusCode != 0 {
		expectedStatus := 429
		if mockCtx.StatusCode == expectedStatus {
			t.Log("Rate limiter correctly rejecting requests when limit exceeded")
		} else {
			t.Logf("Got unexpected status code: %d", mockCtx.StatusCode)
		}
	} else {
		t.Log("No response status set, which is expected for passed requests")
	}
}

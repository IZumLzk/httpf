package http

import (
	"testing"

	"gospacex/core/common"
)

func TestNewFiberAdapter(t *testing.T) {
	config := HTTPConfig{
		Framework:  FiberFramework,
		Port:       8080,
		Host:       "localhost",
		Workers:    4,
		MultiNodes: []string{"192.168.1.10:8080", "192.168.1.11:8080"},
	}

	adapter := NewFiberAdapter(config)

	if adapter == nil {
		t.Fatal("Expected NewFiberAdapter to return non-nil adapter")
	}

	if adapter.config.Port != 8080 {
		t.Errorf("Expected port to be 8080, got %d", adapter.config.Port)
	}

	if adapter.config.Host != "localhost" {
		t.Errorf("Expected host to be 'localhost', got '%s'", adapter.config.Host)
	}

	if adapter.config.Framework != FiberFramework {
		t.Errorf("Expected framework to be 'fiber', got '%s'", adapter.config.Framework)
	}
}

func TestFiberAdapter_AddRoute(t *testing.T) {
	config := HTTPConfig{
		Framework: FiberFramework,
		Port:      8081,
		Host:      "127.0.0.1",
	}

	adapter := NewFiberAdapter(config)

	// 添加路由
	adapter.AddRoute("GET", "/test", func(ctx common.HTTPContext) {
		ctx.JSON(200, map[string]interface{}{"message": "Route executed"})
	})

	// 验证路由是否注册成功（此处主要是验证没有编译错误或panic）
	t.Log("Route was registered with Fiber adapter")
}

func TestFiberAdapter_UseMiddlewares(t *testing.T) {
	config := HTTPConfig{
		Framework: FiberFramework,
		Port:      8082,
		Host:      "127.0.0.1",
	}

	adapter := NewFiberAdapter(config)

	// 创建测试中间件 - 总是返回true，允许通过
	middleware1 := common.HTTPMiddlewareFunc(func(ctx common.HTTPContext) bool {
		return true
	})

	middleware2 := common.HTTPMiddlewareFunc(func(ctx common.HTTPContext) bool {
		return true
	})

	// 应用中间件
	modifiedServer := adapter.UseMiddlewares(middleware1, middleware2)

	// 检查中间件数量
	expectedCount := 2
	actualCount := len(adapter.middlewares)
	if actualCount != expectedCount {
		t.Errorf("Expected %d middlewares to be added, got %d", expectedCount, actualCount)
	}

	// 验证返回值是同一实例
	if modifiedServer != adapter {
		t.Error("Expected UseMiddlewares to return the same instance")
	}
}

func TestFiberAdapter_AddGlobalMiddleware(t *testing.T) {
	config := HTTPConfig{
		Framework: FiberFramework,
		Port:      8083,
		Host:      "127.0.0.1",
	}

	adapter := NewFiberAdapter(config)

	initialCount := len(adapter.middlewares)

	// 创建一个全局中间件
	globalMiddleware := common.HTTPMiddlewareFunc(func(ctx common.HTTPContext) bool {
		return true
	})

	// 添加全局中间件
	adapter.AddGlobalMiddleware(globalMiddleware)

	// 检查中间件数量是否增加
	actualCount := len(adapter.middlewares)
	if actualCount != initialCount+1 {
		t.Errorf("Expected 1 global middleware to be added, got difference of %d", actualCount-initialCount)
	} else {
		t.Logf("Successfully added global middleware, count increased from %d to %d", initialCount, actualCount)
	}
}

func TestFiberAdapter_FullIntegration(t *testing.T) {
	config := HTTPConfig{
		Framework: FiberFramework,
		Port:      8084,
		Host:      "127.0.0.1",
	}

	adapter := NewFiberAdapter(config)

	// 注册一个测试路由
	adapter.AddRoute("GET", "/integration-test", func(ctx common.HTTPContext) {
		ctx.JSON(200, map[string]string{"message": "success"})
	})

	// 添加一个中间件
	middleware := common.HTTPMiddlewareFunc(func(ctx common.HTTPContext) bool {
		return true // 允许继续处理
	})

	adapter.UseMiddlewares(middleware)

	// 验证中间件数量
	expectedMiddlewareCount := 1
	actualCount := len(adapter.middlewares)
	if actualCount != expectedMiddlewareCount {
		t.Errorf("Expected %d middleware after UseMiddlewares call, got %d", expectedMiddlewareCount, actualCount)
	}

	// 检查配置是否正确应用
	if adapter.config.Port != 8084 {
		t.Errorf("Expected adapter to maintain configuration, got port: %d", adapter.config.Port)
	}

	t.Logf("Fiber adapter integration test completed, port: %d", adapter.config.Port)
}

func TestFiberAdapter_Configuration(t *testing.T) {
	expectedConfig := HTTPConfig{
		Framework:  FiberFramework,
		Port:       9090,
		Host:       "test.host",
		Workers:    8,
		MultiNodes: []string{"node1:8080", "node2:8080"},
	}

	adapter := NewFiberAdapter(expectedConfig)

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

	t.Log("Fiber adapter configuration test passed")
}

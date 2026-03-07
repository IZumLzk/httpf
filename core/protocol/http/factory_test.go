package http

import (
	"testing"
)

func TestHTTPFrameworkTypes(t *testing.T) {
	// 验证所有框架常量定义
	tests := []struct {
		name     string
		expected HTTPFrameworkType
	}{
		{"Gin", GinFramework},
		{"Iris", IrisFramework},
		{"Echo", EchoFramework},
		{"Fiber", FiberFramework},
		{"Chi", ChiFramework},
		{"Hertz", HertzFramework},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Errorf("Expected framework type for %s, but got empty string", tt.name)
			}
		})
	}
}

func TestHTTPConfig_Structure(t *testing.T) {
	config := HTTPConfig{
		Framework:  GinFramework,
		Port:       8080,
		Host:       "localhost",
		Workers:    4,
		MultiNodes: []string{"node1:8080", "node2:8080"},
	}

	if config.Framework != GinFramework {
		t.Errorf("Expected framework: %s, got: %s", GinFramework, config.Framework)
	}
	if config.Port != 8080 {
		t.Errorf("Expected port: %d, got: %d", 8080, config.Port)
	}
	if config.Host != "localhost" {
		t.Errorf("Expected host: %s, got: %s", "localhost", config.Host)
	}
	if config.Workers != 4 {
		t.Errorf("Expected workers: %d, got: %d", 4, config.Workers)
	}
	if len(config.MultiNodes) != 2 {
		t.Errorf("Expected 2 multi-nodes, got: %d", len(config.MultiNodes))
	}
	if config.MultiNodes[0] != "node1:8080" && config.MultiNodes[1] != "node2:8080" {
		t.Errorf("Expected multi-node addresses to match")
	}
}

func TestGetHTTPFramework_GinFramework(t *testing.T) {
	factory := GetHTTPFramework(GinFramework)
	if factory == nil {
		t.Fatal("Expected a factory, but got nil")
	}

	// 验证工厂能够创建服务器
	config := HTTPConfig{Framework: GinFramework, Port: 8080, Host: "localhost"}
	server := factory.CreateHTTPServer(config)
	if server == nil {
		t.Error("Expected server to be created, but got nil")
	}
}

func TestGetHTTPFramework_IrisFramework(t *testing.T) {
	factory := GetHTTPFramework(IrisFramework)
	if factory == nil {
		t.Fatal("Expected a factory, but got nil")
	}

	// 验证工厂能够创建服务器
	config := HTTPConfig{Framework: IrisFramework, Port: 8081, Host: "localhost"}
	server := factory.CreateHTTPServer(config)
	if server == nil {
		t.Error("Expected server to be created, but got nil")
	}
}

func TestGetHTTPFramework_EchoFramework(t *testing.T) {
	factory := GetHTTPFramework(EchoFramework)
	if factory == nil {
		t.Fatal("Expected a factory, but got nil")
	}

	// 验证工厂能够创建服务器
	config := HTTPConfig{Framework: EchoFramework, Port: 8082, Host: "localhost"}
	server := factory.CreateHTTPServer(config)
	if server == nil {
		t.Error("Expected server to be created, but got nil")
	}
}

func TestGetHTTPFramework_FiberFramework(t *testing.T) {
	factory := GetHTTPFramework(FiberFramework)
	if factory == nil {
		t.Fatal("Expected a factory, but got nil")
	}

	// 验证工厂能够创建服务器
	config := HTTPConfig{Framework: FiberFramework, Port: 8083, Host: "localhost"}
	server := factory.CreateHTTPServer(config)
	if server == nil {
		t.Error("Expected server to be created, but got nil")
	}
}

func TestGetHTTPFramework_ChiFramework(t *testing.T) {
	factory := GetHTTPFramework(ChiFramework)
	if factory == nil {
		t.Fatal("Expected a factory, but got nil")
	}

	// 验证工厂能够创建服务器
	config := HTTPConfig{Framework: ChiFramework, Port: 8084, Host: "localhost"}
	server := factory.CreateHTTPServer(config)
	if server == nil {
		t.Error("Expected server to be created, but got nil")
	}
}

func TestGetHTTPFramework_HertzFramework(t *testing.T) {
	factory := GetHTTPFramework(HertzFramework)
	if factory == nil {
		t.Fatal("Expected a factory, but got nil")
	}

	// 验证工厂能够创建服务器
	config := HTTPConfig{Framework: HertzFramework, Port: 8085, Host: "localhost"}
	server := factory.CreateHTTPServer(config)
	if server == nil {
		t.Error("Expected server to be created, but got nil")
	}
}

func TestGetHTTPFramework_DefaultFramework(t *testing.T) {
	// 测试当传递未知/空框架类型时，默认返回Gin框架
	factory := GetHTTPFramework("")
	if factory == nil {
		t.Fatal("Expected a factory, but got nil")
	}

	// 这应该返回gin框架作为默认选项
	config := HTTPConfig{Framework: "", Port: 9000, Host: "localhost"}
	server := factory.CreateHTTPServer(config)
	if server == nil {
		t.Error("Expected server to be created with default framework, but got nil")
	}
}

func TestGetHTTPFramework_UnknownFramework(t *testing.T) {
	// 测试当传递未知框架类型时，默认返回Gin框架
	factory := GetHTTPFramework("unknown-framework")
	if factory == nil {
		t.Fatal("Expected a factory to be returned as default, but got nil")
	}

	// 这应该返回gin框架作为默认选项
	config := HTTPConfig{Framework: "unknown-framework", Port: 9001, Host: "localhost"}
	server := factory.CreateHTTPServer(config)
	if server == nil {
		t.Error("Expected server to be created with default framework, but got nil")
	}
}

func TestGetHTTPFramework_AllFrameworksReturnServers(t *testing.T) {
	// 验证所有的框架类型都能正确创建服务器实例
	frameworks := []HTTPFrameworkType{
		GinFramework,
		IrisFramework,
		EchoFramework,
		FiberFramework,
		ChiFramework,
		HertzFramework,
	}

	for _, framework := range frameworks {
		t.Run(string(framework), func(t *testing.T) {
			factory := GetHTTPFramework(framework)
			if factory == nil {
				t.Fatalf("Expected factory for framework: %s, but got nil", framework)
			}

			config := HTTPConfig{Framework: framework, Port: 8090, Host: "localhost"}
			server := factory.CreateHTTPServer(config)
			if server == nil {
				t.Errorf("Expected server to be created for framework: %s, but got nil", framework)
			}
		})
	}
}

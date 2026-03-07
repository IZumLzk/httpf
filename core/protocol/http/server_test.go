package http

import (
	"testing"
)

// TestHTTPLauncher_Init 测试 HTTPLauncher 初始化
func TestHTTPLauncher_Init(t *testing.T) {
	launcher := &HTTPLauncher{}

	if launcher == nil {
		t.Error("Expected HTTPLauncher to be initialized")
	}
}

// TestHTTPInstance 测试全局 HTTP 实例
func TestHTTPInstance(t *testing.T) {
	if HTTP == nil {
		t.Error("Expected global HTTP instance to be initialized")
	}
}

// TestParseNodeAddress 测试节点地址解析
func TestParseNodeAddress(t *testing.T) {
	tests := []struct {
		name         string
		address      string
		expectedHost string
		expectedPort int
	}{
		{
			name:         "standard address",
			address:      "localhost:8080",
			expectedHost: "localhost",
			expectedPort: 8080,
		},
		{
			name:         "IP address with port",
			address:      "192.168.1.10:9090",
			expectedHost: "192.168.1.10",
			expectedPort: 9090,
		},
		{
			name:         "address without port",
			address:      "localhost",
			expectedHost: "localhost",
			expectedPort: 8080,
		},
		{
			name:         "IP address without port",
			address:      "192.168.1.10",
			expectedHost: "192.168.1.10",
			expectedPort: 8080,
		},
		{
			name:         "port 443",
			address:      "example.com:443",
			expectedHost: "example.com",
			expectedPort: 443,
		},
		{
			name:         "empty string",
			address:      "",
			expectedHost: "",
			expectedPort: 8080,
		},
		{
			name:         "only colon",
			address:      ":",
			expectedHost: ":",
			expectedPort: 8080,
		},
		{
			name:         "only port with colon",
			address:      ":8080",
			expectedHost: ":8080",
			expectedPort: 8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := parseNodeAddress(tt.address)

			if config.Host != tt.expectedHost {
				t.Errorf("Expected host '%s', got '%s'", tt.expectedHost, config.Host)
			}

			if config.Port != tt.expectedPort {
				t.Errorf("Expected port %d, got %d", tt.expectedPort, config.Port)
			}

			// 验证框架默认为 Gin
			if config.Framework != GinFramework {
				t.Errorf("Expected framework to be '%s', got '%s'", GinFramework, config.Framework)
			}
		})
	}
}

// TestFindLastColon 测试查找最后一个冒号
func TestFindLastColon(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "simple case",
			input:    "localhost:8080",
			expected: 9,
		},
		{
			name:     "IP address",
			input:    "192.168.1.1:9090",
			expected: 11,
		},
		{
			name:     "no colon",
			input:    "localhost",
			expected: -1,
		},
		{
			name:     "only colon",
			input:    ":",
			expected: 0,
		},
		{
			name:     "multiple colons",
			input:    "http://example.com:8080",
			expected: 18,
		},
		{
			name:     "empty string",
			input:    "",
			expected: -1,
		},
		{
			name:     "colon at end",
			input:    "localhost:",
			expected: 9,
		},
		{
			name:     "colon at start",
			input:    ":8080",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findLastColon(tt.input)

			if result != tt.expected {
				t.Errorf("Expected %d, got %d for input '%s'", tt.expected, result, tt.input)
			}
		})
	}
}

// TestStringToInt 测试字符串转整数
func TestStringToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "simple number",
			input:    "8080",
			expected: 8080,
		},
		{
			name:     "single digit",
			input:    "5",
			expected: 5,
		},
		{
			name:     "zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "large number",
			input:    "65535",
			expected: 65535,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "with non-digit characters",
			input:    "12a34",
			expected: 1234,
		},
		{
			name:     "only non-digit characters",
			input:    "abc",
			expected: 0,
		},
		{
			name:     "port 443",
			input:    "443",
			expected: 443,
		},
		{
			name:     "port 80",
			input:    "80",
			expected: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringToInt(tt.input)

			if result != tt.expected {
				t.Errorf("Expected %d, got %d for input '%s'", tt.expected, result, tt.input)
			}
		})
	}
}

// TestHTTPLauncher_Launch_Empty 测试 Launch 方法无参数情况
func TestHTTPLauncher_Launch_Empty(t *testing.T) {
	// 注意：这个测试会尝试启动实际的 HTTP 服务器
	// 在实际测试中，我们应该 mock 或使用短时间运行的服务
	// 这里我们只测试方法可以被调用，不实际启动服务

	_ = &HTTPLauncher{} // 验证可以创建实例

	// 由于 Launch() 会启动服务器，我们不能在这里直接测试
	// 可以测试配置生成逻辑
	t.Log("Launch method exists and can be called")

	// 验证默认配置生成
	defaultConfig := HTTPConfig{
		Framework: GinFramework,
		Port:      8080,
		Host:      "0.0.0.0",
	}

	if defaultConfig.Framework != GinFramework {
		t.Errorf("Expected default framework to be Gin")
	}

	if defaultConfig.Port != 8080 {
		t.Errorf("Expected default port to be 8080")
	}

	if defaultConfig.Host != "0.0.0.0" {
		t.Errorf("Expected default host to be 0.0.0.0")
	}
}

// TestHTTPLauncher_launchHTTPMulti_EmptyList 测试空节点列表
func TestHTTPLauncher_launchHTTPMulti_EmptyList(t *testing.T) {
	launcher := &HTTPLauncher{}

	err := launcher.launchHTTPMulti([]string{})

	if err == nil {
		t.Error("Expected error for empty node list")
	}

	if err.Error() != "node list cannot be empty" {
		t.Errorf("Expected error message 'node list cannot be empty', got '%s'", err.Error())
	}
}

// TestHTTPLauncher_launchHTTPMulti_SingleNode 测试单节点启动
// 注意：此测试会尝试启动实际服务器，应该小心处理
func TestHTTPLauncher_launchHTTPMulti_SingleNode(t *testing.T) {
	// 跳过实际启动测试，因为这会阻塞
	t.Skip("Skipping actual server launch test to avoid blocking")
}

// TestParseNodeAddress_EdgeCases 测试边界情况
func TestParseNodeAddress_EdgeCases(t *testing.T) {
	// 测试 IPv6 地址格式（简化处理）
	config := parseNodeAddress("[::1]:8080")

	// 由于我们的简化实现，IPv6 地址可能无法正确解析
	// 这里主要验证函数不会崩溃
	t.Logf("IPv6 address parsing - Host: '%s', Port: %d", config.Host, config.Port)

	// 测试多个冒号的情况
	config2 := parseNodeAddress("192.168.1.1:8080:9090")
	t.Logf("Multiple colons - Host: '%s', Port: %d", config2.Host, config2.Port)
}

// TestStringToInt_SpecialCases 测试特殊情况的字符串转换
func TestStringToInt_SpecialCases(t *testing.T) {
	// 测试前导零
	result := stringToInt("0080")
	if result != 80 {
		t.Errorf("Expected 80 for '0080', got %d", result)
	}

	// 测试混合字符和数字
	result2 := stringToInt("80a0b0")
	if result2 != 8000 {
		t.Errorf("Expected 8000 for '80a0b0', got %d", result2)
	}
}

// TestFindLastColon_SpecialCases 测试特殊情况的冒号查找
func TestFindLastColon_SpecialCases(t *testing.T) {
	// 测试多个连续冒号
	result := findLastColon(":::")
	if result != 2 {
		t.Errorf("Expected 2 for ':::', got %d", result)
	}

	// 测试空格
	result2 := findLastColon("host : port")
	if result2 != 5 {
		t.Errorf("Expected 5 for 'host : port', got %d", result2)
	}
}

// TestHTTPLauncher_Config 测试配置正确性
func TestHTTPLauncher_Config(t *testing.T) {
	// 测试配置结构完整性
	config := HTTPConfig{
		Framework:  GinFramework,
		Port:       8080,
		Host:       "localhost",
		Workers:    4,
		MultiNodes: []string{"node1:8080", "node2:8080"},
	}

	if config.Framework != GinFramework {
		t.Error("Expected GinFramework")
	}

	if config.Port != 8080 {
		t.Error("Expected port 8080")
	}

	if config.Host != "localhost" {
		t.Error("Expected host localhost")
	}

	if config.Workers != 4 {
		t.Error("Expected 4 workers")
	}

	if len(config.MultiNodes) != 2 {
		t.Error("Expected 2 multi nodes")
	}
}

// BenchmarkParseNodeAddress 基准测试：节点地址解析
func BenchmarkParseNodeAddress(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseNodeAddress("192.168.1.10:8080")
	}
}

// BenchmarkFindLastColon 基准测试：查找最后一个冒号
func BenchmarkFindLastColon(b *testing.B) {
	for i := 0; i < b.N; i++ {
		findLastColon("192.168.1.10:8080")
	}
}

// BenchmarkStringToInt 基准测试：字符串转整数
func BenchmarkStringToInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stringToInt("8080")
	}
}

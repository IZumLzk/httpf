package http

import (
	"errors"
	"github.com/IZumLzk/httpf/core/protocol/http/middleware"
	"log"
	"sync"
)

// HTTPLauncher HTTP启动器
type HTTPLauncher struct{}

// Launch 启动单个HTTP服务 (默认Gin)
func (h *HTTPLauncher) Launch(addresses ...string) error {
	if len(addresses) == 0 {
		// 单节点启动：使用默认配置
		defaultConfig := HTTPConfig{
			Framework: GinFramework, // 从配置中心获取默认值
			Port:      8080,
			Host:      "0.0.0.0",
		}
		return h.LaunchHTTP(defaultConfig)
	} else {
		// 多节点启动：解析提供的地址
		return h.launchHTTPMulti(addresses)
	}
}

// LaunchMultiNodes 启动多节点集群
func (h *HTTPLauncher) launchHTTPMulti(nodeList []string) error {
	if len(nodeList) == 0 {
		return errors.New("node list cannot be empty")
	}

	var wg sync.WaitGroup
	var resultMu sync.Mutex

	errorsOccurred := false

	for _, nodeAddr := range nodeList {
		wg.Add(1)

		go func(addr string) {
			defer wg.Done()

			config := parseNodeAddress(addr)
			config.Framework = GinFramework // 为了示例保持Gin框架的一致性

			if err := h.LaunchHTTP(config); err != nil {
				resultMu.Lock()
				log.Printf("Failed to launch node %s: %v", addr, err)
				errorsOccurred = true
				resultMu.Unlock()
			}
		}(nodeAddr)
	}

	wg.Wait()

	if errorsOccurred {
		return errors.New("some nodes failed to start")
	}

	return nil
}

// LaunchHTTP 内部启动方法
func (h *HTTPLauncher) LaunchHTTP(config HTTPConfig) error {
	// 创建对应框架实例
	factory := GetHTTPFramework(config.Framework)
	httpServer := factory.CreateHTTPServer(config)

	// 使用中间件
	httpServer.UseMiddlewares(
		middleware.NewRateLimiterMW(),      // 限流中间件
		middleware.NewAuthJWTMW(),          // JWT认证中间件
		middleware.NewCORSDefaultMW(),      // CORS策略中间件
		middleware.NewBlackWhiteFilterMW(), // 黑白名单中间件
	)

	// 启动服务
	return httpServer.Start()
}

// parseNodeAddress 解析节点地址字符串 "host:port" -> HTTPConfig
func parseNodeAddress(addr string) HTTPConfig {
	config := HTTPConfig{
		Framework: GinFramework, // 从配置读取
	}

	// 简单的地址解析示例
	hostPort := addr
	colonIdx := findLastColon(hostPort)

	if colonIdx > 0 {
		config.Host = hostPort[:colonIdx]
		portStr := hostPort[colonIdx+1:]
		config.Port = stringToInt(portStr)
	} else {
		config.Host = hostPort
		config.Port = 8080 // 默认端口
	}

	return config
}

// 辅助函数
func findLastColon(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == ':' {
			return i
		}
	}
	return -1
}

func stringToInt(s string) int {
	// 简单的字符串转整数实现
	num := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			num = num*10 + int(r-'0')
		}
	}
	return num
}

// 启动器实例
var HTTP = &HTTPLauncher{}

package main

import (
	"fmt"
	"github.com/IZumLzk/httpf/core/protocol/http"
	"github.com/IZumLzk/httpf/core/protocol/http/middleware"
)

func main() {
	fmt.Println("Starting GoSpaceX HTTP Demo...")

	// 配置使用Gin框架
	config := http.HTTPConfig{
		Framework:  http.GinFramework,
		Port:       8080,
		Host:       "127.0.0.1",
		Workers:    0,
		MultiNodes: nil,
	}

	// 通过工厂获取HTTP服务器实例
	factory := http.GetHTTPFramework(config.Framework)
	httpServer := factory.CreateHTTPServer(config)

	// 注册路由 - 示例路由
	httpServer.AddRoute("GET", "/", func(ctx http.Context) {
		ctx.JSON(200, map[string]interface{}{
			"message":   "Welcome to GoSpaceX HTTP Demo!",
			"framework": string(config.Framework),
			"features": []string{
				"Multiple HTTP Frameworks Support",
				"Auto Switchable Frameworks",
				"Four Major Middleware Components",
				"Cluster Deployment Support",
			},
		})
	})

	// 添加四大核心中间件组件
	httpServer.UseMiddlewares(
		middleware.NewRateLimiterMW(),      // 限流中间件（基于Sentinel）
		middleware.NewAuthJWTMW(),          // 认证中间件（JWT）
		middleware.NewCORSDefaultMW(),      // CORS策略中间件
		middleware.NewBlackWhiteFilterMW(), // 黑白名单过滤中间件
	)

	// 启动HTTP服务
	fmt.Printf("Starting GoSpaceX HTTP Server on %s:%d with %s framework...\n",
		config.Host, config.Port, string(config.Framework))

	if err := httpServer.Start(); err != nil {
		fmt.Printf("Failed to start HTTP server: %v\n", err)
		return
	}
}

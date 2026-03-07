# HTTP协议模块实施详细方案文档

## 项目概述

GoSpaceX框架的HTTP协议模块实现。此模块支持多种HTTP框架，采用抽象工厂模式和策略模式，允许动态选择HTTP引擎。

## 架构设计

### 核心组件

1. **HTTPProtocol接口** - 统一的HTTP服务接口
2. **抽象工厂** - 根据框架类型创建对应的服务实例
3. **六种框架适配器** - 每种流行的Go HTTP框架的适配器
4. **四大核心中间件** - 包括限流、认证、CORS、黑白名单过滤
5. **启动器服务** - 包括单机和集群启动功能

### 目录结构

```
core/protocol/http/
├── interface.go          # HTTP协议接口定义
├── factory.go            # HTTP框架工厂
├── server.go             # HTTP服务启动器  
├── gin_adapter.go        # Gin框架适配器
├── iris_adapter.go       # Iris框架适配器
├── echo_adapter.go       # Echo框架适配器
├── fiber_adapter.go      # Fiber框架适配器
├── chi_adapter.go        # Chi框架适配器
├── hertz_adapter.go      # Hertz框架适配器
└── middleware/           # 中间件目录
    ├── rate_limiter.go     # 限流中间件（使用sentinel库）
    ├── authentication.go   # JWT身份认证中间件
    ├── cors.go             # CORS策略中间件
    └── black_white_list.go # 黑白名单过滤中间件
```

## 路由注册指南

### 1. 创建路由处理器函数

要添加路由，首先需要创建符合框架要求的处理器函数。处理器函数应接收GoSpaceX标准化的`Context`接口，实现跨框架兼容性：

```go
package handlers

import (
    "gospacex/core/protocol/http"
)

// 定义处理器 - 所有框架共享同一签名
func ExampleRouteHandler(ctx http.Context) {
    // 获取URL路径参数
    id := ctx.Param("id")
    
    // 获取查询参数  
    page := ctx.Query("page")
    
    // 绑定请求体到结构体
    var reqObj struct {
        Name string `json:"name"`
    }
    if err := ctx.Bind(&reqObj); err != nil {
        ctx.JSON(400, map[string]interface{}{
            "error": "invalid request body",
        })
        return
    }
    
    // 返回JSON响应
    ctx.JSON(200, map[string]interface{}{
        "id": id,
        "page": page,
        "message": "Hello from GoSpaceX",
        "received_name": reqObj.Name,
    })  
}

// GET处理函数
func GetUsersHandler(ctx http.Context) {
    page := ctx.Query("page")
    pageSize := ctx.Query("size")
    ctx.JSON(200, map[string]interface{}{
        "data": []string{"user1", "user2", "user3"},
        "page": page,
        "pageSize": pageSize,
    })
}
```

### 2. 动态注册路由

GoSpaceX提供了多种路由注册方式，以下是最推荐的几种方式：

#### 方法一：配置驱动的批量路由注册
通过配置文件或代码集中定义路由，使用GoSpaceX框架自动化注册：

```go
package main

import (
    "github.com/gospacex/demo/handlers"
    "gospacex/core/protocol/http"
)

// 定义路由结构
var AppRoutes = []http.RouteDefinition{
    {"GET", "/api/users", handlers.GetUsersHandler},
    {"GET", "/api/users/:id", handlers.ExampleRouteHandler},
    {"POST", "/api/users", handlers.ExampleRouteHandler},
    {"PUT", "/api/users/:id", handlers.ExampleRouteHandler},
    {"DELETE", "/api/users/:id", handlers.ExampleRouteHandler},
}

func main() {
    config := http.HTTPConfig{
        Framework: http.GinFramework,  // 可动态切换为 "iris", "echo", "fiber", "chi", "hertz"
        Port:      8080,
        Host:      "0.0.0.0",
    }
    
    // 利用工厂创建HTTP服务
    factory := http.GetHTTPFramework(config.Framework)
    server := factory.CreateHTTPServer(config)
    
    // 注册所有应用路由
    for _, route := range AppRoutes {
        server.AddRoute(route.Method, route.Path, route.Handler)
    }
    
    // 添加默认中间件
    server.UseMiddlewares(
        http.NewRateLimiterMW(),
        http.NewAuthJWTMW(), 
        http.NewCORSDefaultMW(),
        http.NewBlackWhiteFilterMW(),
    )
    
    // 启动服务
    err := server.Start()
    if err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}
```

#### 方法二：在启动器中注册路由
使用框架内置的启动器，通过在启动器中注册路由来实现：

```go
package main

import (
    "log"
    "gospacex/core/protocol/http"
    "github.com/gospacex/demo/handlers"
)

// 扩展启动器功能
func launchWithCustomRoutes() error {
    config := http.HTTPConfig{
        Framework: http.EchoFramework, // 切换到Echo
        Port:      8081,
        Host:      "0.0.0.0",
    }
    
    factory := http.GetHTTPFramework(config.Framework)
    server := factory.CreateHTTPServer(config)
    
    // 添加自定义路由
    server.AddRoute("GET", "/", func(ctx http.Context) {
        ctx.JSON(200, map[string]string{
            "message": "Welcome to GoSpaceX framework!",
        })
    })
    
    server.AddRoute("GET", "/api/hello/:name", func(ctx http.Context) {
        name := ctx.Param("name")
        ctx.JSON(200, map[string]string{
            "greeting": fmt.Sprintf("Hello, %s!", name),
        })
    })
    
    server.AddRoute("POST", "/api/users", handlers.ExampleRouteHandler)
    server.AddRoute("GET", "/api/users", handlers.GetUsersHandler)
    
    // 使用默认中间件
    server.UseMiddlewares(
        http.NewRateLimiterMW(),
        http.NewAuthJWTMW(),
        http.NewCORSDefaultMW(), 
        http.NewBlackWhiteFilterMW(),
    )
    
    return server.Start()
}

func main() {
    err := launchWithCustomRoutes()
    if err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

#### 方法三：框架内路由注册
在GoSpaceX适配器层实现的更深层次集成：

```go
// 示例 - gin_adapter.go 扩展
func (ga *GinAdapter) AddDefaultRouteHandlers() {
    // 注册健康检查路由
    ga.AddRoute("GET", "/health", func(ctx http.Context) {
        ctx.JSON(200, map[string]interface{}{
            "status": "healthy",
            "framework": "gin",
            "uptime": time.Now(),
        })
    })
    
    // 添加您的其他默认路由...
    ga.AddRoute("GET", "/version", func(ctx http.Context) {
        ctx.JSON(200, map[string]string{
            "version": "v1.0.0", 
            "platform": "GoSpaceX HTTP Framework",
        })
    })
}
```

### 3. 路由定义规范

确保遵循以下规范创建路由：

1. **Handler函数签名**: 必须符合 `func(http.Context)` 格式
2. **路由路径格式**: 遵循各框架的路径参数约定 `/:param` 或 `/*wildcard`
3. **请求处理**:
    - 使用 ctx.Bind() 解析请求体
    - 使用 ctx.Param() 获取路径参数
    - 使用 ctx.Query() 获取查询参数
    - 使用 ctx.JSON() 返回JSON响应

### 4. 参数与响应处理

GoSpaceX为所有HTTP框架提供了统一的上下文API：

```go  
func CompleteRouteExample(ctx http.Context) {
    // 1. 获取请求参数
    userId := ctx.Param("id")          // 路径参数
    page := ctx.Query("page")         // 查询参数
    sort := ctx.Query("sort")         // 查询参数
    
    // 2. 解析请求体
    var requestData UserUpdateData
    if err := ctx.Bind(&requestData); err != nil {
        ctx.JSON(400, map[string]string{
            "error": "Malformed request body",
        })
        return
    }
    
    // 3. 处理业务逻辑
    result, err := processUserUpdate(userId, requestData)
    if err != nil {
        ctx.JSON(500, map[string]string{
            "error": "Internal server error occurred",
        })
        return
    }
    
    // 4. 返回响应
    ctx.JSON(200, map[string]interface{}{
        "status": "success",
        "data":   result,
        "meta": map[string]interface{}{
            "page": page,
            "sort": sort,
        },
    })
}
```

通过上述方式，您可以在GoSpaceX框架中以标准化的方式注册路由，无论使用哪种底层HTTP框架，都可获得一致的使用体验和强大的中间件支持。

## 实现组件详情

### 1、HTTP协议接口和工厂模式
- **interface.go** - 定义通用 HTTP 服务接口
- **factory.go** - 多框架选择工厂，提供 `GetHTTPFramework()` 方法
- **config.go** - 服务配置定义

### 2、框架适配器实现
分别实现了 Gin, Iris, Echo, Fiber, Chi, Hertz 六个适配器，每个都实现统一的 `HTTPProtocol` 接口

### 3、服务启停系统
- **server.go** - HTTPLauncher，启动器实现 `Launch()`/`LaunchMultiNodes()` 方法

### 4、中间件系统
四大核心中间件提供企业级功能：
- 服务限流（基于Sentinal）
- JWT身份验证
- CORS跨域策略
- IP 黑白名单过滤

该架构实现了框架无关的统一API，让开发者可以专注业务逻辑，无需关心底层HTTP框架细节。

## 使用方式

### 单服务启动
```go  
gospacex.http.HTTP.Launch()
```

### 集群启动
```go
gospacex.http.HTTP.LaunchMultiNodes([]string{
    "192.168.0.1:8081",
    "192.168.0.1:8082"
})
```

框架支持从配置文件动态指定使用的HTTP框架类型，以及对应的中间件、路由、部署策略等。通过工厂模式和适配器模式解藕了框架间的差异，实现了可插拔的HTTP框架支持。
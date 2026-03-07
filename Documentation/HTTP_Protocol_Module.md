# GoSpaceX 微服务框架 HTTP 协议模块设计与实现文档

## 1. 项目概述

GoSpaceX 是一款先进的 Go 语言微服务框架，支持多种 HTTP 框架，采用统一接口抽象，实现了架构灵活性与一致性相结合的设计目标。

### 1.1 设计目标

- **统一接口**：为不同 HTTP 框架提供一致的编程体验
- **模块化设计**：组件间松耦合，便于扩展和维护
- **高性能适配**：在不同框架间提供平滑高效的适配机制
- **中间件系统**：支持限流、认证、CORS、过滤等功能

### 1.2 核心架构
```
core/
├── types/            # 共享类型定义
├── protocol/         
│   └── http/         # HTTP 协议实现层
│       ├── gin_adapter.go      # Gin 框架适配
│       ├── iris_adapter.go     # Iris 框架适配
│       ├── echo_adapter.go     # Echo 框架适配
│       ├── fiber_adapter.go    # Fiber 框架适配
│       ├── chi_adapter.go      # Chi 框架适配  
│       ├── hertz_adapter.go    # Hertz 框架适配
│       ├── interface.go        # 统一接口定义
│       ├── factory.go          # 工厂创建机制
│       ├── server.go           # 服务启动器
│       └── middleware/         # 全局中间件
│           ├── rate_limiter.go     # 限流器
│           ├── authentication.go   # JWT 认证
│           ├── cors.go             # CORS 跨域治理
│           └── black_white_list.go # 黑白名单过滤
```

## 2. 技术选型和架构模式

### 2.1 设计模式
- **适配器模式（Adapter Pattern）**：将不同 HTTP 框架抽象为统一接口
- **工厂模式（Factory Pattern）**：根据配置动态创建指定框架实例
- **策略模式（Strategy Pattern）**：根据不同框架选择其运行时行为
- **外观模式（Facade Pattern）**：HTTP 协议对外提供统一接口

### 2.2 核心类型定义
```go
// HTTP 协议统一接口
type HTTPProtocol interface {
	Start() error                      // 启动服务
	Stop() error                       // 停止服务
	AddRoute(method, path string, handler HTTPHandlerFunc) // 添加路由
	UseMiddlewares(mids ...HTTPMiddleware) HTTPProtocol   // 使用中间件
	AddGlobalMiddleware(HTTPMiddleware)                   // 添加中间件
}

// HTTP 上下文抽象
type HTTPContext interface {
	JSON(code int, obj interface{})
	Param(key string) string
	Query(key string) string
	Bind(obj interface{}) error
	Request() *http.Request
	ResponseWriter() http.ResponseWriter
}

// 中间件接口
type HTTPMiddleware interface {
	Process(ctx HTTPContext) bool
}
```

## 3. 核心组件实现

### 3.1 HTTP 适配器实现

#### Gin 框架适配器
```go
type GinAdapter struct {
	engine      *gin.Engine    // Gin 核心引擎
	config      HTTPConfig     // 服务配置
	middlewares []Middleware   // 适配器中间件集合
}

func NewGinAdapter(config HTTPConfig) *GinAdapter {
	gin.SetMode(gin.ReleaseMode)
	adapter := &GinAdapter {
		engine: gin.New(),
		config: config,
		middlewares: make([]Middleware, 0),
	}
	return adapter
}

func (ga *GinAdapter) Start() error {
	addr := ga.config.Host + ":" + fmt.Sprintf("%d", ga.config.Port)
	return ga.engine.Run(addr)
}

func (ga *GinAdapter) UseMiddlewares(mids ...Middleware) HTTPProtocol {
	for _, mid := range mids {
		ga.engine.Use(wrapGinMiddleware(mid))
	}
	return ga
}
```

#### Fiber 框架适配器
```go
// 解决策略：通过中间层实现 net/http.Request 到 Fiber 请求的映射
type FiberAdapter struct {
	app         *fiber.App
	config      HTTPConfig
	middlewares []HTTPMiddleware
}

func (fc *FiberContext) Request() *http.Request {
	req := &http.Request{
		Method: fc.Ctx.Method(),
		URL:    parseURLFromFiberContext(fc.Ctx), // 适配 Fiber 特定的 URL 解析
		Header: make(http.Header),
	}

	// 从 Fiber 上下文传输请求头
	fc.Ctx.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})
	
	return req
}

// FiberResponseWriter 的适配实现
type fiberResponseWriterAdaptor fibere.Ctx
func (fwra *fiberResponseWriterAdaptor) Header() http.Header { /* ... */ }
```

### 3.2 中间件系统

#### 核心思路
中间件系统基于 `Process(ctx HTTPContext) bool` 接口设计，遵循责任链模式，在请求处理过程中插入检查逻辑。

```go
// 限流中间件 - 基于 Sentinel 
type RateLimiterMiddleware struct {
	limit int    // QPS 限制
}

func (rm *RateLimiterMiddleware) Process(ctx HTTPContext) bool {
	if !isWithinLimit() {  // 调用 Sentinel 流控检查逻辑
		ctx.JSON(429, map[string]string{
			"error": "rate limited", 
			"msg": "请求频率超出限制"
		})
		return false  // 阻断请求
	}
	return true       // 继续请求链
}

// 认证中间件 - JWT 解析与校验
func (am *AuthJWTMiddleware) Process(ctx HTTPContext) bool {
	authHeader := ctx.Request().Header.Get("Authorization") 
	token := parseBearerToken(authHeader)
	
	if !isValidJWT(token) { 
		ctx.JSON(401, map[string]string{"error": "unauthorized"})
		return false
	}
	return true
}
```

### 3.3 服务启动器

```go
type HTTPLauncher struct{}

// 单服务启动
func (h *HTTPLauncher) Launch() error {
	config := loadHTTPConfig()      // 从配置中心加载配置
	factory := GetHTTPFramework(config.Framework) // 获取工厂
	server := factory.CreateHTTPServer(config)    // 创建服务实例
	
	// 装载中间件
	server.UseMiddlewares(
		NewRateLimiterMW(),
		NewAuthJWTMW(), 
		NewCORSDefaultMW(),
		NewBlackWhiteFilterMW(),
	)
	
	return server.Start()
}

// 多节点集群启动
func (h *HTTPLauncher) LaunchMultiNodes(addrs []string) error {
	var wg sync.WaitGroup
	for idx, addr := range addrs {
		wg.Add(1)
		config := parseNodeAddress(addrs[idx])
		go launchNode(config, &wg)  // 并发启动
	}
	wg.Wait()
	return nil
}
```

## 4. 配置与管理

### 4.1 配置优先级
框架支持动态配置，支持从本地(Viper)和远程(Nacos)来源加载，优先级为：
```
远程配置 > 本地配置 = 默认配置
```

### 4.2 配置结构
```go
type HTTPConfig struct {
    Framework  HTTPFrameworkType `mapstructure:"framework"`
    Port       int               `mapstructure:"port"`
    Host       string            `mapstructure:"host"`
    Workers    int               `mapstructure:"workers"`
    MultiNodes []string          `mapstructure:"multi_nodes"` // 集群节点
}
```

## 5. 服务启动方式

### 5.1 默认启动
```go
// 启动默认配置服务（使用 Gin）
gospacex.http.Launch()
```

### 5.2 集群启动
```go
// 启动多节点集群
gospacex.http.Launch(
    "192.168.0.1:8081", 
    "192.168.0.1:8082"
)
```

## 6. 依赖管理

项目使用 Go Modules 进行依赖管理，核心依赖包括：
- `github.com/gin-gonic/gin`: 轻量级 Web 框架
- `github.com/gofiber/fiber/v2`: 基于 FastHTTP 的高性能 Web 框架
- `github.com/kataras/iris/v12`: 全功能 Web 框架
- `github.com/alibaba/sentinel-golang`: 限流熔断组件
- `github.com/golang-jwt/jwt/v4`: JWT 认证组件

## 7. 测试策略

### 7.1 单元测试
- 每个组件提供单元测试验证其行为正确性
- 基于 interface 定义的 mock 系统验证组件功能

### 7.2 集成测试
- HTTP 协议适配器正确性、一致性验证
- 中间件链、请求链路完整测试
- 各种 HTTP 框架功能等价性测试

### 7.3 压力测试
- HTTP 框架的并发处理能力测试
- 限流组件的准确控制测试

## 8. 性能优化

### 8.1 框架适配优化
- 最小化跨框架类型转换开销
- 中间件链的懒加载与缓存机制

### 8.2 内存管理优化
- 优化 Context 数据结构，复用内存
- 防止中间件链中不必要的临时对象分配

## 9. 扩展性

### 9.1 新框架支持
通过遵循 `HTTPProtocol` 接口即可接入新的 Web 框架（如 Beego、Revel 等）

### 9.2 新中间件支持
实现 `HTTPMiddleware` 接口即可创建新的通用中间件组件

### 9.3 集群扩展
支持通过配置方式水平扩展服务节点
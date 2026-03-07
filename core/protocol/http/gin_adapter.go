package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GinAdapter Gin框架适配器
type GinAdapter struct {
	Engine      *gin.Engine
	Config      HTTPConfig
	Middlewares []Middleware
}

// NewGinAdapter 创建新的Gin适配器
func NewGinAdapter(config HTTPConfig) *GinAdapter {
	gin.SetMode(gin.ReleaseMode) // 生产模式
	engine := gin.New()

	adapter := &GinAdapter{
		Engine:      engine,
		Config:      config,
		Middlewares: make([]Middleware, 0),
	}

	return adapter
}

// AddRoute 添加路由
func (ga *GinAdapter) AddRoute(method, path string, handler HandlerFunc) {
	ginHandler := func(c *gin.Context) {
		wrappedCtx := &GinContext{c}
		handler(wrappedCtx)
	}

	switch method {
	case "GET":
		ga.Engine.GET(path, ginHandler)
	case "POST":
		ga.Engine.POST(path, ginHandler)
	case "PUT":
		ga.Engine.PUT(path, ginHandler)
	case "DELETE":
		ga.Engine.DELETE(path, ginHandler)
	default:
		ga.Engine.Any(path, ginHandler) // 支持任意HTTP方法
	}
}

// UseMiddlewares 使用中间件
func (ga *GinAdapter) UseMiddlewares(mids ...Middleware) HTTPProtocol {
	ga.Middlewares = append(ga.Middlewares, mids...)

	for _, mid := range mids {
		ga.Engine.Use(func(c *gin.Context) {
			ctx := &GinContext{c}
			if !mid.Process(ctx) {
				c.Abort() // 如果中间件处理返回false，则终止后续处理
				return
			}
			c.Next()
		})
	}
	return ga
}

// AddGlobalMiddleware 添加全局中间件
func (ga *GinAdapter) AddGlobalMiddleware(middleware Middleware) {
	ga.Middlewares = append(ga.Middlewares, middleware)

	ga.Engine.Use(func(c *gin.Context) {
		ctx := &GinContext{c}
		if !middleware.Process(ctx) {
			c.Abort()
			return
		}
		c.Next()
	})
}

// Start 启动服务器
func (ga *GinAdapter) Start() error {
	addr := ga.Config.Host + ":" + fmt.Sprintf("%d", ga.Config.Port)
	return ga.Engine.Run(addr)
}

// Stop 停止服务器
func (ga *GinAdapter) Stop() error {
	// Gin目前没有内置的优雅停止功能，可能需要使用http.Server封装
	return nil
}

// GinContext Gin框架的上下文适配器
type GinContext struct {
	*gin.Context
}

// Implement the Context interface
func (gc *GinContext) JSON(code int, obj interface{}) {
	gc.Context.JSON(code, obj)
}

func (gc *GinContext) Param(key string) string {
	return gc.Context.Param(key)
}

func (gc *GinContext) Query(key string) string {
	return gc.Context.Query(key)
}

func (gc *GinContext) Bind(obj interface{}) error {
	return gc.Context.ShouldBind(obj)
}

// Request 和 ResponseWriter 方法需要从gin.Context获取
func (gc *GinContext) Request() *http.Request {
	return gc.Context.Request
}

func (gc *GinContext) ResponseWriter() http.ResponseWriter {
	return gc.Context.Writer
}

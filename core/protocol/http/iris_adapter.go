package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kataras/iris/v12"
)

// IrisAdapter Iris框架适配器
type IrisAdapter struct {
	app    *iris.Application
	config HTTPConfig
}

// NewIrisAdapter 创建新的Iris适配器
func NewIrisAdapter(config HTTPConfig) *IrisAdapter {
	app := iris.New()

	adapter := &IrisAdapter{
		app:    app,
		config: config,
	}

	return adapter
}

// AddRoute 添加路由
func (ia *IrisAdapter) AddRoute(method, path string, handler HandlerFunc) {
	irisHandler := func(ctx iris.Context) {
		wrappedCtx := &IrisContext{ctx}
		handler(wrappedCtx)
	}

	switch method {
	case "GET":
		ia.app.Get(path, irisHandler)
	case "POST":
		ia.app.Post(path, irisHandler)
	case "PUT":
		ia.app.Put(path, irisHandler)
	case "DELETE":
		ia.app.Delete(path, irisHandler)
	default:
		ia.app.Any(path, irisHandler)
	}
}

// UseMiddlewares 使用中间件
func (ia *IrisAdapter) UseMiddlewares(mids ...Middleware) HTTPProtocol {
	for _, mid := range mids {
		ia.app.Use(func(ctx iris.Context) {
			wrappedCtx := &IrisContext{ctx}
			if !mid.Process(wrappedCtx) {
				return // 如果中间件返回false，则不调用next
			}
			ctx.Next()
		})
	}
	return ia
}

// AddGlobalMiddleware 添加全局中间件
func (ia *IrisAdapter) AddGlobalMiddleware(middleware Middleware) {
	ia.app.Use(func(ctx iris.Context) {
		wrappedCtx := &IrisContext{ctx}
		if !middleware.Process(wrappedCtx) {
			return
		}
		ctx.Next()
	})
}

// Start 启动服务器
func (ia *IrisAdapter) Start() error {
	addr := ia.config.Host + ":" + fmt.Sprintf("%d", ia.config.Port)
	return ia.app.Listen(addr)
}

// Stop 停止服务器
func (ia *IrisAdapter) Stop() error {
	return ia.app.Shutdown(context.TODO())
}

// IrisContext Iris框架的上下文适配器
type IrisContext struct {
	iris.Context
}

// Implement the Context interface
func (ic *IrisContext) JSON(code int, obj interface{}) {
	ic.Context.StatusCode(code)
	ic.Context.JSON(obj)
}

func (ic *IrisContext) Param(key string) string {
	return ic.Context.Params().Get(key)
}

func (ic *IrisContext) Query(key string) string {
	return ic.Context.URLParam(key)
}

func (ic *IrisContext) Bind(obj interface{}) error {
	return ic.Context.ReadJSON(obj)
}

func (ic *IrisContext) Request() *http.Request {
	return ic.Context.Request()
}

func (ic *IrisContext) ResponseWriter() http.ResponseWriter {
	return ic.Context.ResponseWriter().Naive()
}

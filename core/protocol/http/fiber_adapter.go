package http

import (
	"fmt"
	"github.com/IZumLzk/httpf/core/common"
	"github.com/gofiber/fiber/v2"

	"net/http"
	"net/url"
)

// FiberAdapter Fiber框架适配器
// 注：在真实环境中，这应当对接fiber框架的实际API
type FiberAdapter struct {
	app         *fiber.App
	config      HTTPConfig
	middlewares []common.HTTPMiddleware
}

// NewFiberAdapter 创建新的Fiber适配器
func NewFiberAdapter(config HTTPConfig) *FiberAdapter {
	app := fiber.New(fiber.Config{
		ServerHeader: "gospacex-fiber",
		AppName:      "gospacex",
	})

	adapter := &FiberAdapter{
		app:         app,
		config:      config,
		middlewares: make([]common.HTTPMiddleware, 0),
	}

	return adapter
}

// AddRoute 添加路由
func (fa *FiberAdapter) AddRoute(method, path string, handler common.HTTPHandlerFunc) {
	fiberHandler := func(c *fiber.Ctx) error {
		wrappedCtx := &FiberContext{c}
		handler(wrappedCtx)
		// 在实际使用中，handler可能有返回值需要处理错误
		return nil
	}

	switch method {
	case "GET":
		fa.app.Get(path, fiberHandler)
	case "POST":
		fa.app.Post(path, fiberHandler)
	case "PUT":
		fa.app.Put(path, fiberHandler)
	case "DELETE":
		fa.app.Delete(path, fiberHandler)
	default:
		fa.app.All(path, fiberHandler) // 支持任意HTTP方法
	}
}

// UseMiddlewares 使用中间件
func (fa *FiberAdapter) UseMiddlewares(mids ...common.HTTPMiddleware) common.HTTPProtocol {
	fa.middlewares = append(fa.middlewares, mids...)

	for _, mid := range mids {
		fa.app.Use(func(c *fiber.Ctx) error {
			fiberCtx := &FiberContext{c}
			if !mid.Process(fiberCtx) {
				// 如果中间件返回false，发送禁止访问响应
				return c.Status(403).JSON(fiber.Map{
					"error":   "forbidden",
					"message": "Access denied by middleware",
				})
			}
			return c.Next()
		})
	}

	return fa
}

// AddGlobalMiddleware 添加全局中间件
func (fa *FiberAdapter) AddGlobalMiddleware(middleware common.HTTPMiddleware) {
	fa.middlewares = append(fa.middlewares, middleware)

	fa.app.Use(func(c *fiber.Ctx) error {
		fiberCtx := &FiberContext{c}
		if !middleware.Process(fiberCtx) {
			return c.Status(403).JSON(fiber.Map{
				"error":   "forbidden",
				"message": "Access denied by global middleware",
			})
		}
		return c.Next()
	})
}

// Start 启动服务器
func (fa *FiberAdapter) Start() error {
	addr := fa.config.Host + ":" + fmt.Sprintf("%d", fa.config.Port)
	return fa.app.Listen(addr)
}

// Stop 停止服务器
func (fa *FiberAdapter) Stop() error {
	return fa.app.Shutdown()
}

// FiberContext Fiber框架的上下文适配器
type FiberContext struct {
	*fiber.Ctx
}

// Implement the Context interface
func (fc *FiberContext) JSON(code int, obj interface{}) {
	fc.Ctx.Status(code).JSON(obj)
}

func (fc *FiberContext) Param(key string) string {
	return fc.Ctx.Params(key)
}

func (fc *FiberContext) Query(key string) string {
	return fc.Ctx.Query(key)
}

func (fc *FiberContext) Bind(obj interface{}) error {
	return fc.Ctx.BodyParser(obj)
}

func (fc *FiberContext) Request() *http.Request {
	// 将fiber的请求适配为标准的net/http.Request
	// 创建一个新的HTTP请求对象
	req := &http.Request{
		Method: fc.Ctx.Method(),
		URL:    &url.URL{},
		Header: make(http.Header),
		Body:   nil, // Body在Fiber中是直接处理的
	}

	// 构建URL
	scheme := "http"
	// 检查是否TLS/HTTPS连接
	if fc.Ctx.Protocol() == "https" {
		scheme = "https"
	}
	req.URL.Scheme = scheme

	req.URL.Host = fc.Ctx.Hostname()
	req.URL.Path = fc.Ctx.Path()
	req.URL.RawQuery = string(fc.Ctx.Request().URI().QueryString())

	// 设置查询参数到URL中
	rawURL := fmt.Sprintf("%s://%s%s?%s", scheme, req.URL.Host, req.URL.Path, req.URL.RawQuery)

	// 重新解析完整URL
	parsedURL, err := url.Parse(rawURL)
	if err == nil {
		req.URL = parsedURL
	}

	// 设置请求头
	fc.Ctx.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	// 设置远端地址
	req.RemoteAddr = fc.Ctx.IP()

	return req
}

func (fc *FiberContext) ResponseWriter() http.ResponseWriter {
	return (*fiberResponseWriterAdaptor)(fc.Ctx)
}

// fiberResponseWriterAdaptor 将fiber.Ctx转为 http.ResponseWriter 接口
type fiberResponseWriterAdaptor fiber.Ctx

func (fwra *fiberResponseWriterAdaptor) Header() http.Header {
	headers := make(http.Header)
	(*fiber.Ctx)(fwra).Response().Header.VisitAll(func(key, value []byte) {
		headers.Set(string(key), string(value))
	})
	return headers
}

func (fwra *fiberResponseWriterAdaptor) Write(data []byte) (int, error) {
	return (*fiber.Ctx)(fwra).Write(data)
}

func (fwra *fiberResponseWriterAdaptor) WriteHeader(statusCode int) {
	(*fiber.Ctx)(fwra).Status(statusCode)
}

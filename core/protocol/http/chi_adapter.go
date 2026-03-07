package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ChiContext Chi框架的上下文适配器
type ChiContext struct {
	writer  http.ResponseWriter
	request *http.Request
}

func (cc *ChiContext) JSON(code int, obj interface{}) {
	cc.writer.Header().Set("Content-Type", "application/json")
	cc.writer.WriteHeader(code)

	jsonData, err := json.Marshal(obj)
	if err != nil {
		cc.writer.WriteHeader(500) // 内部服务器错误
		_, _ = cc.writer.Write([]byte(`{"error":"Failed to serialize JSON"}`))
		return
	}

	cc.writer.Write(jsonData)
}

func (cc *ChiContext) Param(key string) string {
	return chi.URLParam(cc.request, key)
}

func (cc *ChiContext) Query(key string) string {
	return cc.request.URL.Query().Get(key)
}

func (cc *ChiContext) Bind(obj interface{}) error {
	// 实现请求绑定逻辑 - 根据 Content-Type 进行不同处理
	contentType := cc.request.Header.Get("Content-Type")
	if contentType == "" {
		return nil // 或返回错误
	}

	if contentType == "application/json" {
		// 使用 JSON 解码
		decoder := json.NewDecoder(cc.request.Body)
		return decoder.Decode(obj)
	}

	return nil // 其他类型待实现
}

func (cc *ChiContext) Request() *http.Request {
	return cc.request
}

func (cc *ChiContext) ResponseWriter() http.ResponseWriter {
	return cc.writer
}

// ChiAdapter Chi框架适配器
type ChiAdapter struct {
	config      HTTPConfig
	httpMux     *chi.Mux
	middlewares []Middleware
	server      *http.Server
}

// NewChiAdapter 创建新的Chi适配器
func NewChiAdapter(config HTTPConfig) *ChiAdapter {
	mux := chi.NewMux()

	// 使用Chi的基本中间件
	mux.Use(middleware.Logger)    // 请求日志
	mux.Use(middleware.Recoverer) // Panic恢复
	mux.Use(middleware.RequestID) // 请求ID

	adapter := &ChiAdapter{
		config:      config,
		httpMux:     mux,
		middlewares: make([]Middleware, 0),
	}

	return adapter
}

// AddRoute 添加路由
func (ca *ChiAdapter) AddRoute(method, path string, handler HandlerFunc) {
	chiHandler := func(w http.ResponseWriter, r *http.Request) {
		wrappedCtx := &ChiContext{
			writer:  w,
			request: r,
		}

		// 应用中间件链
		for _, middleware := range ca.middlewares {
			if !middleware.Process(wrappedCtx) {
				// 如果中间件返回false，停止处理
				return
			}
		}

		// 执行主处理程序
		handler(wrappedCtx)
	}

	ca.httpMux.MethodFunc(method, path, chiHandler)
}

// UseMiddlewares 使用中间件
func (ca *ChiAdapter) UseMiddlewares(mids ...Middleware) HTTPProtocol {
	ca.middlewares = append(ca.middlewares, mids...)

	// 为Chi框架添加中间件包装器
	for _, mid := range mids {
		ca.httpMux.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				wrappedCtx := &ChiContext{
					writer:  w,
					request: r,
				}

				if mid.Process(wrappedCtx) {
					// 继续处理
					next.ServeHTTP(w, r)
				}
				// 如果返回false，不继续处理
			})
		})
	}

	return ca
}

// AddGlobalMiddleware 添加全局中间件
func (ca *ChiAdapter) AddGlobalMiddleware(middleware Middleware) {
	ca.middlewares = append(ca.middlewares, middleware)

	// 为Chi框架添加全局中间件
	ca.httpMux.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrappedCtx := &ChiContext{
				writer:  w,
				request: r,
			}

			if middleware.Process(wrappedCtx) {
				// 继续处理
				next.ServeHTTP(w, r)
			}
			// 如果返回false，不继续处理
		})
	})
}

// Start 启动服务器
func (ca *ChiAdapter) Start() error {
	addr := fmt.Sprintf("%s:%d", ca.config.Host, ca.config.Port)

	ca.server = &http.Server{
		Addr:    addr,
		Handler: ca.httpMux,
	}

	return ca.server.ListenAndServe()
}

// Stop 停止服务器
func (ca *ChiAdapter) Stop() error {
	if ca.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return ca.server.Shutdown(ctx)
	}
	return nil
}

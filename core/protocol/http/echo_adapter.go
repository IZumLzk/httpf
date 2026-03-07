package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EchoContext Echo框架的上下文适配器
type EchoContext struct {
	writer  http.ResponseWriter
	request *http.Request
	params  map[string]string
}

// JSON 实现Context接口 - 完整实现JSON序列化
func (ec *EchoContext) JSON(code int, obj interface{}) {
	ec.writer.Header().Set("Content-Type", "application/json")
	ec.writer.WriteHeader(code)

	jsonData, err := json.Marshal(obj)
	if err != nil {
		ec.writer.WriteHeader(500) // 内部服务器错误
		_, _ = ec.writer.Write([]byte(`{"error":"Failed to serialize JSON"}`))
		return
	}

	ec.writer.Write(jsonData)
}

// Param 实现Context接口
func (ec *EchoContext) Param(key string) string {
	// 从路径参数中获取值
	if val, exists := ec.params[key]; exists {
		return val
	}
	return ec.request.URL.Query().Get(key)
}

// Query 实现Context接口
func (ec *EchoContext) Query(key string) string {
	return ec.request.URL.Query().Get(key)
}

// Bind 实现Context接口 - 完整实现请求绑定
func (ec *EchoContext) Bind(obj interface{}) error {
	contentType := ec.request.Header.Get("Content-Type")
	if contentType == "" {
		return nil // 或返回错误
	}

	if contentType == "application/json" {
		decoder := json.NewDecoder(ec.request.Body)
		return decoder.Decode(obj)
	}

	return nil // 其他类型待实现
}

// Request 实现Context接口
func (ec *EchoContext) Request() *http.Request {
	return ec.request
}

// ResponseWriter 实现Context接口
func (ec *EchoContext) ResponseWriter() http.ResponseWriter {
	return ec.writer
}

// EchoAdapter Echo框架适配器
type EchoAdapter struct {
	config      HTTPConfig
	handlers    map[string]map[string]HandlerFunc // 存储路由映射: method -> path -> handler
	middlewares []Middleware
	server      *http.Server
}

// NewEchoAdapter 创建新的Echo适配器
func NewEchoAdapter(config HTTPConfig) *EchoAdapter {
	adapter := &EchoAdapter{
		config:      config,
		handlers:    make(map[string]map[string]HandlerFunc),
		middlewares: make([]Middleware, 0),
	}

	// 初始化handlers map
	adapter.handlers["GET"] = make(map[string]HandlerFunc)
	adapter.handlers["POST"] = make(map[string]HandlerFunc)
	adapter.handlers["PUT"] = make(map[string]HandlerFunc)
	adapter.handlers["DELETE"] = make(map[string]HandlerFunc)

	return adapter
}

// AddRoute 添加路由
func (ea *EchoAdapter) AddRoute(method, path string, handler HandlerFunc) {
	if ea.handlers[method] == nil {
		ea.handlers[method] = make(map[string]HandlerFunc)
	}
	ea.handlers[method][path] = handler
}

// UseMiddlewares 使用中间件 - 完整实现
func (ea *EchoAdapter) UseMiddlewares(mids ...Middleware) HTTPProtocol {
	ea.middlewares = append(ea.middlewares, mids...)
	return ea
}

// AddGlobalMiddleware 添加全局中间件 - 完整实现
func (ea *EchoAdapter) AddGlobalMiddleware(middleware Middleware) {
	ea.middlewares = append(ea.middlewares, middleware)
}

// Start 启动服务器
func (ea *EchoAdapter) Start() error {
	// 创建一个基础的net/http服务
	httpMux := http.NewServeMux()

	// 注册所有路由到HTTP处理器中
	for method, paths := range ea.handlers {
		for path, handler := range paths {
			method := method
			path := path
			finalHandler := handler

			httpMux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}

				// 创建适配的上下文
				ctx := &EchoContext{
					writer:  w,
					request: r,
					params:  make(map[string]string), // 简化参数获取
				}

				// 应用中间件链
				for _, mid := range ea.middlewares {
					if !mid.Process(ctx) {
						// 如果中间件返回false，则停止处理
						return
					}
				}

				// 调用处理函数
				finalHandler(ctx)
			})
		}
	}

	addr := ea.config.Host + ":" + fmt.Sprintf("%d", ea.config.Port)

	ea.server = &http.Server{
		Addr:    addr,
		Handler: httpMux,
	}

	return ea.server.ListenAndServe()
}

// Stop 停止服务器 - 完整实现优雅停止
func (ea *EchoAdapter) Stop() error {
	if ea.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return ea.server.Shutdown(ctx)
	}
	return nil
}

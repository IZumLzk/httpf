package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HertzAdapter Hertz框架适配器
// 为简化实现，此处模拟Hertz框架行为，真实实现应导入Hertz SDK
type HertzAdapter struct {
	config      HTTPConfig
	httpMux     *http.ServeMux
	middlewares []Middleware
	server      *http.Server
}

// NewHertzAdapter 创建Hertz适配器实例
func NewHertzAdapter(config HTTPConfig) *HertzAdapter {
	mux := http.NewServeMux()

	adapter := &HertzAdapter{
		config:      config,
		httpMux:     mux,
		middlewares: make([]Middleware, 0),
	}

	return adapter
}

// AddRoute 添加路由
func (ha *HertzAdapter) AddRoute(method, path string, handler HandlerFunc) {
	hertzHandler := func(w http.ResponseWriter, r *http.Request) {
		// 检查 HTTP 方法是否匹配
		if r.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := &HertzContext{
			request:  r,
			response: w,
			params:   make(map[string]string),
		}

		// 应用中间件链
		for _, mid := range ha.middlewares {
			if !mid.Process(ctx) {
				// 如果中间件返回false，则停止处理
				return
			}
		}

		handler(ctx)
	}

	ha.httpMux.HandleFunc(path, hertzHandler)
}

// UseMiddlewares 使用中间件 - 完整实现
func (ha *HertzAdapter) UseMiddlewares(mids ...Middleware) HTTPProtocol {
	ha.middlewares = append(ha.middlewares, mids...)
	return ha
}

// AddGlobalMiddleware 添加全局中间件 - 完整实现
func (ha *HertzAdapter) AddGlobalMiddleware(middleware Middleware) {
	ha.middlewares = append(ha.middlewares, middleware)
}

// Start 启动服务
func (ha *HertzAdapter) Start() error {
	addr := ha.config.Host + ":" + fmt.Sprintf("%d", ha.config.Port)

	ha.server = &http.Server{
		Addr:    addr,
		Handler: ha.httpMux,
	}

	return ha.server.ListenAndServe()
}

// Stop 停止服务 - 完整实现优雅停止
func (ha *HertzAdapter) Stop() error {
	if ha.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return ha.server.Shutdown(ctx)
	}
	return nil
}

// HertzContext Hertz上下文适配器
type HertzContext struct {
	request  *http.Request
	response http.ResponseWriter
	params   map[string]string
}

// 实现 Context 接口
func (hc *HertzContext) JSON(code int, obj interface{}) {
	hc.response.Header().Set("Content-Type", "application/json")
	hc.response.WriteHeader(code)

	jsonData, err := json.Marshal(obj)
	if err != nil {
		hc.response.WriteHeader(500) // 内部服务器错误
		_, _ = hc.response.Write([]byte(`{"error":"Failed to serialize JSON"}`))
		return
	}

	hc.response.Write(jsonData)
}

func (hc *HertzContext) Param(key string) string {
	// 从路径参数或查询参数获取值
	if val, exists := hc.params[key]; exists {
		return val
	}
	return hc.request.URL.Query().Get(key)
}

func (hc *HertzContext) Query(key string) string {
	return hc.request.URL.Query().Get(key)
}

func (hc *HertzContext) Bind(obj interface{}) error {
	contentType := hc.request.Header.Get("Content-Type")
	if contentType == "" {
		return nil
	}

	if contentType == "application/json" {
		decoder := json.NewDecoder(hc.request.Body)
		return decoder.Decode(obj)
	}

	return nil
}

func (hc *HertzContext) Request() *http.Request {
	return hc.request
}

func (hc *HertzContext) ResponseWriter() http.ResponseWriter {
	return hc.response
}

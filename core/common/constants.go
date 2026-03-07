package common

import "net/http"

// HTTPMethod HTTP方法枚举
type HTTPMethod string

const (
	GET     HTTPMethod = "GET"
	POST    HTTPMethod = "POST"
	PUT     HTTPMethod = "PUT"
	DELETE  HTTPMethod = "DELETE"
	PATCH   HTTPMethod = "PATCH"
	HEAD    HTTPMethod = "HEAD"
	OPTIONS HTTPMethod = "OPTIONS"
)

// HTTPContext HTTP上下文接口
type HTTPContext interface {
	JSON(code int, obj interface{})
	Param(key string) string
	Query(key string) string
	Bind(obj interface{}) error
	Request() *http.Request
	ResponseWriter() http.ResponseWriter
}

// HTTPMiddleware HTTP中间件接口
type HTTPMiddleware interface {
	Process(ctx HTTPContext) bool // 返回false则中断
}

// HTTPMiddlewareFunc HTTP中间件函数类型
type HTTPMiddlewareFunc func(ctx HTTPContext) bool

func (mw HTTPMiddlewareFunc) Process(ctx HTTPContext) bool {
	return mw(ctx)
}

// HTTPProtocol HTTP协议接口
type HTTPProtocol interface {
	Start() error
	Stop() error
	AddRoute(method, path string, handler HTTPHandlerFunc)
	UseMiddlewares(mids ...HTTPMiddleware) HTTPProtocol
	AddGlobalMiddleware(middleware HTTPMiddleware)
}

// HTTPHandlerFunc 统一请求处理器函数签名
type HTTPHandlerFunc func(HTTPContext)

// CommonResult 通用响应结构
type CommonResult struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// KeyValue 键值对
type KeyValue struct {
	Key   string
	Value interface{}
}

// ConfigMap 配置映射
type ConfigMap map[string]interface{}

// ServiceInfo 服务信息
type ServiceInfo struct {
	ServiceName string
	Version     string
	Address     string
	Metadata    map[string]string
}

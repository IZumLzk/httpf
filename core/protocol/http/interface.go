package http

import (
	"github.com/IZumLzk/httpf/core/common"
)

// HTTPProtocol 定义统一的HTTP协议抽象层
type HTTPProtocol = common.HTTPProtocol

// HandlerFunc 统一请求处理器函数签名
type HandlerFunc = common.HTTPHandlerFunc

// Context 上下文抽象，兼容不同HTTP框架
type Context = common.HTTPContext

// Middleware 中间件统一接口
type Middleware = common.HTTPMiddleware
type MiddlewareFunc = common.HTTPMiddlewareFunc

// 强制类型检查，确保类型一致性但不在interface.go中定义方法
var _ = func() struct{} {
	// 确保MiddlewareFunc可以用于Process方法
	_ = func(mf common.HTTPMiddlewareFunc) common.HTTPMiddleware { return mf }
	return struct{}{}
}()

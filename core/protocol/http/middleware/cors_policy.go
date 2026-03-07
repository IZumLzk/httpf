package middleware

import (
	"gospacex/core/common"
	"strings"
)

type CORSPolicyMiddleware struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string
	enabled        bool
}

func NewCORSDefaultMW() common.HTTPMiddlewareFunc {
	return (&CORSPolicyMiddleware{
		allowedOrigins: []string{"*"}, // 从配置获取允许的源
		allowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"},
		allowedHeaders: []string{"Origin", "Content-Type", "Authorization", "X-Requested-With", "Accept", "X-CSRF-Token"},
		enabled:        true, // 从配置读取启用标志
	}).Process
}

func (cm *CORSPolicyMiddleware) Process(ctx common.HTTPContext) bool {
	req := ctx.Request()
	resp := ctx.ResponseWriter()

	origin := req.Header.Get("Origin")
	if origin != "" && cm.isOriginAllowed(origin) {
		resp.Header().Set("Access-Control-Allow-Origin", origin)
	}

	resp.Header().Set("Access-Control-Allow-Methods", strings.Join(cm.allowedMethods, ","))
	resp.Header().Set("Access-Control-Allow-Headers", strings.Join(cm.allowedHeaders, ","))
	resp.Header().Set("Access-Control-Allow-Credentials", "true")
	resp.Header().Set("Access-Control-Max-Age", "3600") // 预检缓存时间

	// 预检请求处理
	if req.Method == "OPTIONS" {
		// 对预检请求直接返回200
		resp.WriteHeader(200)
		return false // 阻止后续处理器执行
	}

	return true
}

func (cm *CORSPolicyMiddleware) isOriginAllowed(origin string) bool {
	for _, allowedOrigin := range cm.allowedOrigins {
		if allowedOrigin == origin || allowedOrigin == "*" {
			return true
		}
	}
	return false
}

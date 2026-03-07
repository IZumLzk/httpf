package middleware

import (
	"gospacex/core/common"
	"net"
	"net/http"
	"strings"
)

type IPFilterType string

const (
	BlackFilter IPFilterType = "black"
	WhiteFilter IPFilterType = "white"
)

type BlackWhiteFilter struct {
	ipBlacklist map[string]bool
	ipWhitelist map[string]bool
	ipType      IPFilterType
	enabled     bool
}

func NewBlackWhiteFilterMW() common.HTTPMiddlewareFunc {
	blacklist := make(map[string]bool)
	whitelist := make(map[string]bool)

	// 模拟从配置加载IP列表
	// 实际应从全局配置中心获取
	for _, ip := range getDefaultBlackList() {
		blacklist[ip] = true
	}

	for _, ip := range getDefaultWhiteList() {
		whitelist[ip] = true
	}

	return (&BlackWhiteFilter{
		ipBlacklist: blacklist,
		ipWhitelist: whitelist,
		ipType:      BlackFilter, // 默认黑名单模式
		enabled:     true,
	}).Process
}

func (bf *BlackWhiteFilter) Process(ctx common.HTTPContext) bool {
	if !bf.enabled {
		return true // 如果禁用，则通过检查
	}

	clientIP := getClientIP(ctx.Request())

	switch bf.ipType {
	case BlackFilter:
		// 黑名单模式：如果在黑名单中，则拒绝访问
		if bf.ipBlacklist[clientIP] {
			ctx.JSON(403, map[string]interface{}{
				"error":   "forbidden",
				"message": "IP address is blocked",
			})
			return false
		}
	case WhiteFilter:
		// 白名单模式：只允许在白名单中的IP访问
		if !bf.ipWhitelist[clientIP] && clientIP != "127.0.0.1" {
			ctx.JSON(403, map[string]interface{}{
				"error":   "forbidden",
				"message": "Access restricted to allowed IPs only",
			})
			return false
		}
	}

	// 允许访问
	return true
}

// getClientIP 获取客户端IP地址
func getClientIP(req *http.Request) string {
	// 获取实际的客户端IP，考虑代理头部
	xForwardedFor := req.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		clientIP := strings.TrimSpace(ips[0])
		return clientIP
	}

	xRealIP := req.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// 从RemoteAddr提取IP
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	return host
}

// 模拟配置获取函数
func getDefaultBlackList() []string {
	// 实际实现中会从配置中心获取
	return []string{"192.168.100.1", "10.0.0.1"}
}

func getDefaultWhiteList() []string {
	// 实际实现中会从配置中心获取
	return []string{"127.0.0.1", "10.0.0.0", "192.168.1.0"}
}

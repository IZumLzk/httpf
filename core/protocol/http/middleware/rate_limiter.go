package middleware

import (
	"github.com/IZumLzk/httpf/core/common"
	"sync"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
)

var (
	// sentinelInitOnce 确保 Sentinel 只初始化一次
	sentinelInitOnce sync.Once
	// sentinelInitialized 标记 Sentinel 是否已初始化
	sentinelInitialized = false
)

// RateLimiterMiddleware 限流中间件
type RateLimiterMiddleware struct {
	resource string  // 资源名称
	qps      float64 // QPS 阈值
	statMs   uint32  // 统计时间窗口（毫秒）
}

// RateLimiterConfig 限流配置
type RateLimiterConfig struct {
	Resource         string  // 资源名称
	QPS              float64 // QPS 阈值
	StatIntervalInMs uint32  // 统计时间窗口（毫秒），默认 1000
}

// DefaultRateLimiterConfig 默认限流配置
var DefaultRateLimiterConfig = RateLimiterConfig{
	Resource:         "default",
	QPS:              100,
	StatIntervalInMs: 1000,
}

// NewRateLimiterMW 创建限流中间件（使用默认配置）
func NewRateLimiterMW() common.HTTPMiddlewareFunc {
	return NewRateLimiterMWWithConfig(DefaultRateLimiterConfig)
}

// NewRateLimiterMWWithConfig 创建限流中间件（使用自定义配置）
func NewRateLimiterMWWithConfig(config RateLimiterConfig) common.HTTPMiddlewareFunc {
	// 初始化 Sentinel
	initSentinel()

	// 设置统计窗口
	statMs := config.StatIntervalInMs
	if statMs == 0 {
		statMs = 1000
	}

	// 创建限流中间件实例
	rlm := &RateLimiterMiddleware{
		resource: config.Resource,
		qps:      config.QPS,
		statMs:   statMs,
	}

	// 加载限流规则
	if err := rlm.loadRules(); err != nil {
		// 加载规则失败，返回一个总是放行的中间件
		return func(ctx common.HTTPContext) bool {
			return true
		}
	}

	return rlm.Process
}

// initSentinel 初始化 Sentinel
func initSentinel() {
	sentinelInitOnce.Do(func() {
		err := sentinel.InitDefault()
		if err != nil {
			// Sentinel 初始化失败，记录日志但不影响服务启动
			sentinelInitialized = false
			return
		}
		sentinelInitialized = true
	})
}

// loadRules 加载限流规则
func (rm *RateLimiterMiddleware) loadRules() error {
	// 创建限流规则
	rules := []*flow.Rule{
		{
			Resource:               rm.resource,
			Threshold:              rm.qps,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       rm.statMs,
		},
	}

	// 加载规则
	_, err := flow.LoadRules(rules)
	if err != nil {
		return err
	}

	return nil
}

// Process 处理请求
func (rm *RateLimiterMiddleware) Process(ctx common.HTTPContext) bool {
	// 如果 Sentinel 未初始化，直接放行
	if !sentinelInitialized {
		return true
	}

	// 尝试进入资源
	entry, blockErr := sentinel.Entry(rm.resource, sentinel.WithTrafficType(base.Inbound))
	if blockErr != nil {
		// 请求被限流
		ctx.JSON(429, map[string]interface{}{
			"error":   "rate_limited",
			"message": "Too many requests, please try again later",
			"details": map[string]interface{}{
				"resource": rm.resource,
				"qps":      rm.qps,
			},
		})
		return false
	}

	// 请求允许通过，业务逻辑结束后必须调用 Exit
	defer entry.Exit()

	return true
}

// UpdateQPS 动态更新 QPS 阈值
func (rm *RateLimiterMiddleware) UpdateQPS(qps float64) error {
	rm.qps = qps
	return rm.loadRules()
}

// UpdateResource 动态更新资源名称
func (rm *RateLimiterMiddleware) UpdateResource(resource string) error {
	rm.resource = resource
	return rm.loadRules()
}

// GetConfig 获取当前配置
func (rm *RateLimiterMiddleware) GetConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Resource:         rm.resource,
		QPS:              rm.qps,
		StatIntervalInMs: rm.statMs,
	}
}

// LoadRulesFromConfig 从配置批量加载限流规则
func LoadRulesFromConfig(configs []RateLimiterConfig) error {
	rules := make([]*flow.Rule, 0, len(configs))

	for _, cfg := range configs {
		statMs := cfg.StatIntervalInMs
		if statMs == 0 {
			statMs = 1000
		}

		rules = append(rules, &flow.Rule{
			Resource:               cfg.Resource,
			Threshold:              cfg.QPS,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       statMs,
		})
	}

	_, err := flow.LoadRules(rules)
	return err
}

// NewRateLimiterMWByResource 根据资源名称创建限流中间件（使用默认 QPS）
func NewRateLimiterMWByResource(resource string) common.HTTPMiddlewareFunc {
	config := DefaultRateLimiterConfig
	config.Resource = resource
	return NewRateLimiterMWWithConfig(config)
}

// NewRateLimiterMWByQPS 根据 QPS 创建限流中间件（使用默认资源名）
func NewRateLimiterMWByQPS(qps float64) common.HTTPMiddlewareFunc {
	config := DefaultRateLimiterConfig
	config.QPS = qps
	return NewRateLimiterMWWithConfig(config)
}

// NewRateLimiterMWCustom 创建完全自定义的限流中间件
func NewRateLimiterMWCustom(resource string, qps float64, statIntervalMs uint32) common.HTTPMiddlewareFunc {
	config := RateLimiterConfig{
		Resource:         resource,
		QPS:              qps,
		StatIntervalInMs: statIntervalMs,
	}
	return NewRateLimiterMWWithConfig(config)
}

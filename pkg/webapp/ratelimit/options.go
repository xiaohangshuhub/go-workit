package ratelimit

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
)

// RateLimitOptions 限流配置选项
type Options struct {
	DefaultPolicy string // 默认限流策略名称
	// 限流策略映射
	policies map[string]RateLimiter

	routeRateLimitMap  map[router.RouteKey][]string // 路由 → 限流方案
	rateLimitRoutesMap map[string][]router.RouteKey // 限流方案 → 路由
	router             *httprouter.Router           // httprouter 实例
	patternMap         map[string]string            // 处理函数标识到模式字符串的映射
	mu                 sync.Mutex                   // 保护并发访问
}

func NewOptions() *Options {
	return &Options{
		policies:           make(map[string]RateLimiter),
		routeRateLimitMap:  make(map[router.RouteKey][]string),
		rateLimitRoutesMap: make(map[string][]router.RouteKey),
		router:             httprouter.New(),
		patternMap:         make(map[string]string),
	}
}

// AddFixedWindowLimiter 添加固定窗口限流器
func (opt *Options) AddFixedWindowLimiter(name string, configure func(*FixedWindowOptions)) {
	options := &FixedWindowOptions{
		QueueProcessingOrder: OldestFirst,
	}
	configure(options)
	opt.policies[name] = NewFixedWindowLimiter(options)
}

// AddSlidingWindowLimiter 添加滑动窗口限流器
func (opt *Options) AddSlidingWindowLimiter(name string, configure func(*SlidingWindowOptions)) {
	options := &SlidingWindowOptions{
		QueueProcessingOrder: OldestFirst,
	}
	configure(options)
	opt.policies[name] = NewSlidingWindowLimiter(options)
}

// AddTokenBucketLimiter 添加令牌桶限流器
func (opt *Options) AddTokenBucketLimiter(name string, configure func(*TokenBucketOptions)) {
	options := &TokenBucketOptions{
		QueueProcessingOrder: OldestFirst,
	}
	configure(options)
	opt.policies[name] = NewTokenBucketLimiter(options)
}

// AddConcurrencyLimiter 添加并发限流器
func (opt *Options) AddConcurrencyLimiter(name string, configure func(*ConcurrencyOptions)) {
	options := &ConcurrencyOptions{
		QueueProcessingOrder: OldestFirst,
	}
	configure(options)
	opt.policies[name] = NewConcurrencyLimiter(options)
}

// 注册路由（内部方法）
func (opt *Options) registerRoute(method, pattern string) {
	opt.mu.Lock()
	defer opt.mu.Unlock()

	// 生成唯一 handlerID
	handlerID := fmt.Sprintf("%s:%s", method, pattern)
	if _, exists := opt.patternMap[handlerID]; exists {
		return // 已注册
	}

	// 注册到 httprouter
	switch method {
	case "GET":
		opt.router.GET(pattern, opt.dummyHandler)
	case "POST":
		opt.router.POST(pattern, opt.dummyHandler)
	case "PUT":
		opt.router.PUT(pattern, opt.dummyHandler)
	case "DELETE":
		opt.router.DELETE(pattern, opt.dummyHandler)
	case "PATCH":
		opt.router.PATCH(pattern, opt.dummyHandler)
	case "HEAD":
		opt.router.HEAD(pattern, opt.dummyHandler)
	case "OPTIONS":
		opt.router.OPTIONS(pattern, opt.dummyHandler)
	}
	opt.patternMap[handlerID] = pattern
}

// dummyHandler 空处理函数，仅用于路由匹配
func (opt *Options) dummyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

// useRouteRateLimitPolicies 注册路由限流策略
func (opt *Options) UseRouteRateLimitPolicies(routePolicies ...RoutePolicies) {
	for _, rp := range routePolicies {
		if len(rp.Routes) == 0 {
			panic("routes is empty")
		}
		if len(rp.RateLimitPolicy) == 0 {
			panic("rate limit policies is empty")
		}

		for _, route := range rp.Routes {
			if route.Path == "" {
				panic("path is empty")
			}
			if len(route.Methods) == 0 {
				panic("methods is empty: " + route.Path)
			}

			for _, m := range route.Methods {
				key := router.RouteKey{Method: string(m), Path: route.Path}

				// 合并 routeRateLimitMap
				existing := opt.routeRateLimitMap[key]
				policySet := make(map[string]struct{})
				for _, p := range existing {
					policySet[p] = struct{}{}
				}
				for _, p := range rp.RateLimitPolicy {
					if _, ok := policySet[p]; !ok {
						existing = append(existing, p)
						policySet[p] = struct{}{}
					}
				}
				opt.routeRateLimitMap[key] = existing

				// 合并 rateLimitRoutesMap
				for _, policy := range rp.RateLimitPolicy {
					routes := opt.rateLimitRoutesMap[policy]
					found := false
					for _, r := range routes {
						if r == key {
							found = true
							break
						}
					}
					if !found {
						opt.rateLimitRoutesMap[policy] = append(opt.rateLimitRoutesMap[policy], key)
					}
				}

				// 注册路由
				opt.registerRoute(string(m), route.Path)
			}
		}
	}
}

// findMatchingRoute 查找匹配路由
func (opt *Options) findMatchingRoute(method, path string) (router.RouteKey, bool) {
	handler, params, _ := opt.router.Lookup(method, path)
	if handler == nil {
		return router.RouteKey{}, false
	}

	// 还原参数化路径
	pattern := path
	for _, param := range params {
		pattern = replaceFirst(pattern, param.Value, ":"+param.Key)
	}
	return router.RouteKey{Method: method, Path: pattern}, true
}

// getLimitersForRequest 获取请求对应的限流器（返回策略名 → 限流器）
func (opt *Options) getLimitersForRequest(method, path string) map[string]RateLimiter {
	routeKey, found := opt.findMatchingRoute(method, path)

	limiters := make(map[string]RateLimiter)

	// 路由限流策略
	if found {
		if policies, exists := opt.routeRateLimitMap[routeKey]; exists {
			for _, p := range policies {
				if limiter, ok := opt.policies[p]; ok {
					limiters[p] = limiter
				}
			}
		}
	}

	// 全局默认限流策略（无论是否匹配路由，都加上全局保护）
	if opt.DefaultPolicy != "" {
		if limiter, ok := opt.policies[opt.DefaultPolicy]; ok {
			limiters[opt.DefaultPolicy] = limiter
		}
	}

	return limiters
}

// replaceFirst 替换字符串中第一次出现的子串
// 注意：如果 old 为空字符串，则返回 s
// s 原字符串 old 要替换的子串 new 要替换的子串
//
// 返回替换后的字符串
func replaceFirst(s, old, new string) string {
	if old == "" {
		return s
	}

	idx := strings.Index(s, old)
	if idx == -1 {
		return s
	}

	return s[:idx] + new + s[idx+len(old):]
}

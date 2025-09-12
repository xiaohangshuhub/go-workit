package webapp

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
)

// RouteAuthorizePolicies 表示路由和授权策略的关联
type RouteRateLimitPolicies struct {
	Routes          []Route  // 路由列表
	RateLimitPolicy []string // 授权策略列表
}

// RateLimitOptions 限流配置选项
type RateLimitOptions struct {
	DefaultPolicy string // 默认限流策略名称
	// 限流策略映射
	policies map[string]RateLimiter

	routeRateLimitMap  map[RouteKey][]string // 路由 → 限流方案
	rateLimitRoutesMap map[string][]RouteKey // 限流方案 → 路由
	router             *httprouter.Router    // httprouter 实例
	patternMap         map[string]string     // 处理函数标识到模式字符串的映射
	mu                 sync.Mutex            // 保护并发访问
}

func newRateLimitOptions() *RateLimitOptions {
	return &RateLimitOptions{
		policies:           make(map[string]RateLimiter),
		routeRateLimitMap:  make(map[RouteKey][]string),
		rateLimitRoutesMap: make(map[string][]RouteKey),
		router:             httprouter.New(),
		patternMap:         make(map[string]string),
	}
}

// AddFixedWindowLimiter 添加固定窗口限流器
func (opt *RateLimitOptions) AddFixedWindowLimiter(name string, configure func(*FixedWindowOptions)) {
	options := &FixedWindowOptions{
		QueueProcessingOrder: OldestFirst,
	}
	configure(options)
	opt.policies[name] = NewFixedWindowLimiter(options)
}

// AddSlidingWindowLimiter 添加滑动窗口限流器
func (opt *RateLimitOptions) AddSlidingWindowLimiter(name string, configure func(*SlidingWindowOptions)) {
	options := &SlidingWindowOptions{
		QueueProcessingOrder: OldestFirst,
	}
	configure(options)
	opt.policies[name] = NewSlidingWindowLimiter(options)
}

// AddTokenBucketLimiter 添加令牌桶限流器
func (opt *RateLimitOptions) AddTokenBucketLimiter(name string, configure func(*TokenBucketOptions)) {
	options := &TokenBucketOptions{
		QueueProcessingOrder: OldestFirst,
	}
	configure(options)
	opt.policies[name] = NewTokenBucketLimiter(options)
}

// AddConcurrencyLimiter 添加并发限流器
func (opt *RateLimitOptions) AddConcurrencyLimiter(name string, configure func(*ConcurrencyOptions)) {
	options := &ConcurrencyOptions{
		QueueProcessingOrder: OldestFirst,
	}
	configure(options)
	opt.policies[name] = NewConcurrencyLimiter(options)
}

// 注册路由（内部方法）
func (opt *RateLimitOptions) registerRoute(method, pattern string) {
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
func (opt *RateLimitOptions) dummyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

// useRouteRateLimitPolicies 注册路由限流策略
func (opt *RateLimitOptions) useRouteRateLimitPolicies(routePolicies ...RouteRateLimitPolicies) {
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
				key := RouteKey{Method: string(m), Path: route.Path}

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
func (opt *RateLimitOptions) findMatchingRoute(method, path string) (RouteKey, bool) {
	handler, params, _ := opt.router.Lookup(method, path)
	if handler == nil {
		return RouteKey{}, false
	}

	// 还原参数化路径
	pattern := path
	for _, param := range params {
		pattern = replaceFirst(pattern, param.Value, ":"+param.Key)
	}
	return RouteKey{Method: method, Path: pattern}, true
}

// getLimitersForRequest 获取请求对应的限流器
func (opt *RateLimitOptions) getLimitersForRequest(method, path string) []RateLimiter {
	routeKey, found := opt.findMatchingRoute(method, path)
	if !found {
		return []RateLimiter{}
	}

	policies, exists := opt.routeRateLimitMap[routeKey]
	if !exists || len(policies) == 0 {
		return []RateLimiter{}
	}

	var limiters []RateLimiter
	for _, p := range policies {
		if limiter, ok := opt.policies[p]; ok {
			limiters = append(limiters, limiter)
		}
	}
	return limiters
}

package ratelimit

import (
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

// Options 限流配置选项
type Options struct {
	DefaultPolicy      string                          // 默认限流策略名称
	policies           map[string]web.RateLimitHandler // 限流策略配置
	routeRateLimitMap  map[router.RouteKey][]string    // 路由 → 限流方案
	rateLimitRoutesMap map[string][]router.RouteKey    // 限流方案 → 路由
	*Builder
}

func NewOptions() *Options {

	opts := &Options{
		policies:           make(map[string]web.RateLimitHandler),
		routeRateLimitMap:  make(map[router.RouteKey][]string),
		rateLimitRoutesMap: make(map[string][]router.RouteKey),
	}

	opts.Builder = NewBuilder(opts)

	return opts
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
			}
		}
	}
}

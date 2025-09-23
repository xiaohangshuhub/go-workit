package ratelimit

import (
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

// Options 限流配置选项
type Options struct {
	DefaultPolicy string                     // 默认限流策略名称
	policies      map[string]web.RateLimiter // 限流策略配置

}

func NewOptions() *Options {

	opts := &Options{
		policies: make(map[string]web.RateLimiter),
	}

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

func (opt *Options) Policies(policyName ...string) map[string]web.RateLimiter {
	if len(policyName) == 0 {
		return opt.policies
	}

	policies := make(map[string]web.RateLimiter)

	for _, n := range policyName {
		if policy, exists := opt.policies[n]; exists {
			policies[n] = policy
		} else {
			panic("policy with name " + n + " does not exist")
		}
	}

	return policies
}

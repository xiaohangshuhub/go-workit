package webapp

import "time"

// RateLimitPolicy 限流策略类型
type RateLimitPolicy string

const (
	// FixedWindow 固定窗口
	FixedWindow RateLimitPolicy = "fixed"
	// SlidingWindow 滑动窗口
	SlidingWindow RateLimitPolicy = "sliding"
	// TokenBucket 令牌桶
	TokenBucket RateLimitPolicy = "token"
	// Concurrency 并发数
	Concurrency RateLimitPolicy = "concurrent"
)

// QueueProcessingOrder 队列处理顺序
type QueueProcessingOrder int

const (
	// OldestFirst 先进先出
	OldestFirst QueueProcessingOrder = iota
	// NewestFirst 后进先出
	NewestFirst
)

// RateLimitOptions 限流配置选项
type RateLimitOptions struct {
	// 限流策略映射
	policies map[string]RateLimiter
}

// FixedWindowOptions 固定窗口选项
type FixedWindowOptions struct {
	PermitLimit          int
	Window               time.Duration
	QueueProcessingOrder QueueProcessingOrder
	QueueLimit           int
}

// SlidingWindowOptions 滑动窗口选项
type SlidingWindowOptions struct {
	PermitLimit          int
	Window               time.Duration
	SegmentsPerWindow    int
	QueueProcessingOrder QueueProcessingOrder
	QueueLimit           int
}

// TokenBucketOptions 令牌桶选项
type TokenBucketOptions struct {
	TokenLimit           int
	QueueProcessingOrder QueueProcessingOrder
	QueueLimit           int
	ReplenishmentPeriod  time.Duration
	TokensPerPeriod      int
	AutoReplenishment    bool
}

// ConcurrencyOptions 并发限制选项
type ConcurrencyOptions struct {
	PermitLimit          int
	QueueProcessingOrder QueueProcessingOrder
	QueueLimit           int
}

func newRateLimitOptions() *RateLimitOptions {
	return &RateLimitOptions{
		policies: make(map[string]RateLimiter),
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

package ratelimit

import (
	"container/list"
	"sync"
)

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

// 基础限流器结构
type baseLimiter struct {
	mu    sync.RWMutex
	queue map[string]*list.List
}

func (l *baseLimiter) getOrCreateQueue(key string) *list.List {
	if q, exists := l.queue[key]; exists {
		return q
	}
	q := list.New()
	l.queue[key] = q
	return q
}

func (l *baseLimiter) processQueue(key string, order QueueProcessingOrder) {
	q := l.queue[key]
	if q == nil || q.Len() == 0 {
		return
	}

	var elem *list.Element
	if order == OldestFirst {
		elem = q.Front()
	} else {
		elem = q.Back()
	}
	q.Remove(elem)
}

func (l *baseLimiter) Release(key string) {
	// 默认空实现
}

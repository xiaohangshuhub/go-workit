package ratelimit

import (
	"container/list"
	"time"
)

// FixedWindowOptions 固定窗口选项
type FixedWindowOptions struct {
	PermitLimit          int
	Window               time.Duration
	QueueProcessingOrder QueueProcessingOrder
	QueueLimit           int
}

// FixedWindowLimiter 固定窗口限流器
type FixedWindowLimiter struct {
	baseLimiter
	windows map[string]*struct {
		count     int
		startTime time.Time
	}
	options *FixedWindowOptions
}

func NewFixedWindowLimiter(options *FixedWindowOptions) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		baseLimiter: baseLimiter{
			queue: make(map[string]*list.List),
		},
		windows: make(map[string]*struct {
			count     int
			startTime time.Time
		}),
		options: options,
	}
}

func (l *FixedWindowLimiter) TryAcquire(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	window, exists := l.windows[key]

	if !exists || now.Sub(window.startTime) >= l.options.Window {
		l.windows[key] = &struct {
			count     int
			startTime time.Time
		}{
			count:     1,
			startTime: now,
		}
		return true, 0
	}

	if window.count >= l.options.PermitLimit {
		if l.options.QueueLimit > 0 {
			queue := l.getOrCreateQueue(key)
			if queue.Len() < l.options.QueueLimit {
				queue.PushBack(now)
				return false, l.options.Window - now.Sub(window.startTime)
			}
		}
		return false, l.options.Window - now.Sub(window.startTime)
	}

	window.count++
	return true, 0
}

// ... 其他限流器实现 (TokenBucketLimiter, SlidingWindowLimiter, ConcurrencyLimiter)

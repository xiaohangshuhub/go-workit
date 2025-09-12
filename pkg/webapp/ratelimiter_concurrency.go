package webapp

import (
	"container/list"
	"time"
)

// ConcurrencyOptions 并发限制选项
type ConcurrencyOptions struct {
	PermitLimit          int
	QueueProcessingOrder QueueProcessingOrder
	QueueLimit           int
}

// ConcurrencyLimiter 并发限流器
type ConcurrencyLimiter struct {
	baseLimiter
	counters map[string]int
	options  *ConcurrencyOptions
}

func NewConcurrencyLimiter(options *ConcurrencyOptions) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		baseLimiter: baseLimiter{
			queue: make(map[string]*list.List),
		},
		counters: make(map[string]int),
		options:  options,
	}
}

func (l *ConcurrencyLimiter) TryAcquire(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	count := l.counters[key]
	if count >= l.options.PermitLimit {
		if l.options.QueueLimit > 0 {
			queue := l.getOrCreateQueue(key)
			if queue.Len() < l.options.QueueLimit {
				queue.PushBack(time.Now())
				return false, time.Millisecond * 100
			}
		}
		return false, time.Millisecond * 100
	}

	l.counters[key]++
	return true, 0
}

func (l *ConcurrencyLimiter) Release(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.counters[key] > 0 {
		l.counters[key]--
		if queue := l.queue[key]; queue != nil && queue.Len() > 0 {
			l.processQueue(key, l.options.QueueProcessingOrder)
			l.counters[key]++
		}
	}
}

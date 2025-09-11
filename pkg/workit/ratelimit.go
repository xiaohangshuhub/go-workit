package workit

import (
	"container/list"
	"sync"
	"time"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	// TryAcquire 尝试获取访问权限
	TryAcquire(key string) (bool, time.Duration)
	// Release 释放资源(用于并发限流)
	Release(key string)
}

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

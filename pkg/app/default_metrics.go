package app

import (
	"sync"
	"time"
)

type DefaultMetrics struct {
	startTime time.Time
	counters  map[string]int64
	mu        sync.Mutex
}

// newDefaultMetrics 创建一个默认的简单计数器
func newDefaultMetrics() *DefaultMetrics {
	return &DefaultMetrics{
		startTime: time.Now(),
		counters:  make(map[string]int64),
	}
}

// Increment 计数器+1
func (m *DefaultMetrics) Increment(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[key]++
}

// GetCounter 获取当前计数器值
func (m *DefaultMetrics) GetCounter(key string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.counters[key]
}

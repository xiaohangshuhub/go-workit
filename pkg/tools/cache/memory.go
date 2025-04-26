package cache

import (
	"sync"
	"time"
)

type MemoryCache struct {
	data  map[string]interface{}
	exp   map[string]time.Time
	mutex sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]interface{}),
		exp:  make(map[string]time.Time),
	}
	go cache.cleanExpired()
	return cache
}

func (c *MemoryCache) Set(key string, value interface{}, expiration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = value
	if expiration > 0 {
		c.exp[key] = time.Now().Add(expiration)
	}
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if val, ok := c.data[key]; ok {
		if exp, exists := c.exp[key]; !exists || exp.After(time.Now()) {
			return val, true
		}
	}
	return nil, false
}

func (c *MemoryCache) cleanExpired() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		now := time.Now()
		c.mutex.Lock()
		for k, exp := range c.exp {
			if exp.Before(now) {
				delete(c.data, k)
				delete(c.exp, k)
			}
		}
		c.mutex.Unlock()
	}
}

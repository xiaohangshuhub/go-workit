package workit

import (
	"container/list"
	"math"
	"time"
)

// TokenBucketLimiter 令牌桶限流器
type TokenBucketLimiter struct {
	baseLimiter
	buckets map[string]*struct {
		tokens     float64
		lastRefill time.Time
	}
	options *TokenBucketOptions
}

func NewTokenBucketLimiter(options *TokenBucketOptions) *TokenBucketLimiter {
	limiter := &TokenBucketLimiter{
		baseLimiter: baseLimiter{
			queue: make(map[string]*list.List),
		},
		buckets: make(map[string]*struct {
			tokens     float64
			lastRefill time.Time
		}),
		options: options,
	}

	if options.AutoReplenishment {
		go limiter.autoRefill()
	}

	return limiter
}

func (l *TokenBucketLimiter) TryAcquire(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = &struct {
			tokens     float64
			lastRefill time.Time
		}{
			tokens:     float64(l.options.TokenLimit),
			lastRefill: now,
		}
		l.buckets[key] = bucket
	}

	l.refill(bucket, now)

	if bucket.tokens >= 1 {
		bucket.tokens--
		return true, 0
	}

	if l.options.QueueLimit > 0 {
		queue := l.getOrCreateQueue(key)
		if queue.Len() < l.options.QueueLimit {
			queue.PushBack(now)
			waitTime := time.Duration(float64(time.Second) / float64(l.options.TokensPerPeriod))
			return false, waitTime
		}
	}

	waitTime := time.Duration(float64(time.Second) / float64(l.options.TokensPerPeriod))
	return false, waitTime
}

func (l *TokenBucketLimiter) refill(bucket *struct {
	tokens     float64
	lastRefill time.Time
}, now time.Time) {
	elapsed := now.Sub(bucket.lastRefill)
	tokens := float64(elapsed) * float64(l.options.TokensPerPeriod) / float64(l.options.ReplenishmentPeriod)
	bucket.tokens = math.Min(float64(l.options.TokenLimit), bucket.tokens+tokens)
	bucket.lastRefill = now
}

func (l *TokenBucketLimiter) autoRefill() {
	ticker := time.NewTicker(l.options.ReplenishmentPeriod)
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for _, bucket := range l.buckets {
			l.refill(bucket, now)
		}
		l.mu.Unlock()
	}
}

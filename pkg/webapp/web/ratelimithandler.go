package web

import "time"

// RateLimitHandler 限流器接口
type RateLimitHandler interface {
	TryAcquire(key string) (bool, time.Duration) // TryAcquire 尝试获取访问权限
	Release(key string)                          // Release 释放资源(用于并发限流)
}

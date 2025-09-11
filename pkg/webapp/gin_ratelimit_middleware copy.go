package webapp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type GinRateLimitMiddleware struct {
	options *RateLimitOptions
}

func newGinRateLimitMiddleware(options *RateLimitOptions) GinMiddleware {
	return &GinRateLimitMiddleware{
		options: options,
	}
}

func (m *GinRateLimitMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		policyName := c.GetString("RateLimitPolicy")
		if policyName == "" {
			policyName = "default"
		}

		limiter, exists := m.options.policies[policyName]
		if !exists {
			c.Next()
			return
		}

		key := c.ClientIP() // 默认使用IP作为限流键

		allowed, retryAfter := limiter.TryAcquire(key)
		if !allowed {
			c.Header("Retry-After", retryAfter.String())
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":       429,
				"message":    "Too Many Requests",
				"retryAfter": retryAfter.Seconds(),
			})
			return
		}

		c.Next()

		// 对于并发限流，需要在请求结束后释放资源
		if cl, ok := limiter.(*ConcurrencyLimiter); ok {
			cl.Release(key)
		}
	}
}

func (m *GinRateLimitMiddleware) ShouldSkip(path, method string) bool {
	return false
}

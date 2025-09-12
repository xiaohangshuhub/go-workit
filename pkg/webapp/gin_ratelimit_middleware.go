package webapp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GinRateLimitMiddleware struct {
	options *RateLimitOptions
	logger  *zap.Logger
}

func newGinRateLimitMiddleware(options *RateLimitOptions, logger *zap.Logger) GinMiddleware {
	return &GinRateLimitMiddleware{
		options: options,
		logger:  logger,
	}
}

// Handle 限流处理函数
func (m *GinRateLimitMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {

		method := c.Request.Method
		path := c.Request.URL.Path

		// 获取路由对应的限流器
		limiters := m.options.getLimitersForRequest(method, path)
		if len(limiters) == 0 {
			c.Next()
			return
		}

		key := c.ClientIP() // 默认使用IP作为限流键

		for _, limiter := range limiters {
			allowed, retryAfter := limiter.TryAcquire(key)
			if !allowed {
				m.logger.Warn("rate limit exceeded",
					zap.String("path", path),
					zap.String("method", method),
					//zap.String("policy", limiter.Name()),
					zap.String("clientIP", key))
				c.Header("Retry-After", retryAfter.String())
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"code":       429,
					"message":    "Too Many Requests",
					"retryAfter": retryAfter.Seconds(),
				})
				return
			}
		}

		c.Next()

		// 并发限流在请求结束后释放资源
		for _, limiter := range limiters {
			if cl, ok := limiter.(*ConcurrencyLimiter); ok {
				cl.Release(key)
			}
		}
	}
}

func (m *GinRateLimitMiddleware) ShouldSkip(path, method string) bool {
	return false
}

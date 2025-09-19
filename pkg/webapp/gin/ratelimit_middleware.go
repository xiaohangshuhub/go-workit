package gin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/ratelimit"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/zap"
)

type GinRateLimitMiddleware struct {
	web.RatelimitProvider
	logger *zap.Logger
}

func newGinRateLimitMiddleware(provider web.RatelimitProvider, logger *zap.Logger) GinMiddleware {
	return &GinRateLimitMiddleware{
		RatelimitProvider: provider,
		logger:            logger,
	}
}

// Handle 限流处理函数
func (m *GinRateLimitMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {

		method := c.Request.Method
		path := c.Request.URL.Path

		// 获取路由对应的限流器
		limiters := m.RoutePolicies(router.RequestMethod(method), path)
		if len(limiters) == 0 {
			c.Next()
			return
		}

		key := c.ClientIP()
		var maxRetryAfter time.Duration

		for _, limiter := range limiters {
			handler, ok := m.Hadnler(limiter)
			if !ok {

			}
			allowed, retryAfter := handler.TryAcquire(key)
			if !allowed {
				if retryAfter > maxRetryAfter {
					maxRetryAfter = retryAfter
				}
				m.logger.Warn("rate limit exceeded",
					zap.String("path", path),
					zap.String("method", method),
					zap.String("policy", limiter), // 策略名
					zap.String("clientIP", key))
			}
		}

		if maxRetryAfter > 0 {
			c.Header("Retry-After", maxRetryAfter.String())
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":       429,
				"message":    "Too Many Requests",
				"retryAfter": maxRetryAfter.Seconds(),
			})
			return
		}

		c.Next()

		// 并发限流在请求结束后释放资源
		for _, limiter := range limiters {
			handler, ok := m.Hadnler(limiter)
			if !ok {

			}
			if cl, ok := handler.(*ratelimit.ConcurrencyLimiter); ok {
				cl.Release(key)
			}
		}
	}
}

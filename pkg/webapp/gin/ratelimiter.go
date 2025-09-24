package gin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/ratelimit"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/zap"
)

type RateLimitr struct {
	web.Router
	logger *zap.Logger
}

func newRateLimiter(router web.Router, logger *zap.Logger) Middleware {
	return &RateLimitr{
		Router: router,
		logger: logger,
	}
}

// Handle 限流处理函数
func (m *RateLimitr) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path

		// 获取路由对应的限流器
		limiters := m.RateLimits(web.RequestMethod(method), path)

		// 如果没有配置路由策略，才使用默认策略
		if len(limiters) == 0 && m.GlobalRatelimit() != "" {
			limiters = append(limiters, m.GlobalRatelimit())
		}

		if len(limiters) == 0 {
			c.Next()
			return
		}

		key := c.ClientIP()
		var maxRetryAfter time.Duration
		blocked := false

		for _, limiter := range limiters {
			handler, ok := m.RateLimiter(limiter)
			if !ok {
				m.logger.Error("rate limit handler not found",
					zap.String("path", path),
					zap.String("method", method),
					zap.String("policy", limiter),
					zap.String("clientIP", key))
				continue
			}

			allowed, retryAfter := handler.TryAcquire(key)
			if !allowed {
				blocked = true
				if retryAfter > maxRetryAfter {
					maxRetryAfter = retryAfter
				}
				m.logger.Info("rate limit exceeded",
					zap.String("path", path),
					zap.String("method", method),
					zap.String("policy", limiter),
					zap.String("clientIP", key),
					zap.Duration("retryAfter", retryAfter))
			}
		}

		if blocked {
			// Retry-After 按规范需要返回秒数整数
			c.Header("Retry-After", strconv.Itoa(int(maxRetryAfter.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":       429,
				"message":    "Too Many Requests",
				"retryAfter": int(maxRetryAfter.Seconds()),
			})
			return
		}

		// 正常执行下游 handler
		c.Next()

		// 并发限流需要在请求结束后释放资源
		for _, limiter := range limiters {
			if handler, ok := m.RateLimiter(limiter); ok && handler != nil {
				if cl, ok := handler.(*ratelimit.ConcurrencyLimiter); ok {
					cl.Release(key)
				}
			}
		}
	}
}

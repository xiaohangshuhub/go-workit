package webapp

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type EchoRateLimitMiddleware struct {
	options *RateLimitOptions
	logger  *zap.Logger
}

func newEchoRateLimitMiddleware(options *RateLimitOptions, logger *zap.Logger) EchoMiddleware {
	return &EchoRateLimitMiddleware{
		options: options,
		logger:  logger,
	}
}

// Handle 限流处理函数
func (m *EchoRateLimitMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method
			path := c.Request().URL.Path

			// 获取路由对应的限流器
			limiters := m.options.getLimitersForRequest(method, path)
			if len(limiters) == 0 {
				return next(c)
			}

			key := c.RealIP() // 使用客户端IP作为限流键

			for _, limiter := range limiters {
				allowed, retryAfter := limiter.TryAcquire(key)
				if !allowed {
					m.logger.Warn("rate limit exceeded",
						zap.String("path", path),
						zap.String("method", method),
						//zap.String("policy", limiter.Name()),
						zap.String("clientIP", key))
					c.Response().Header().Set("Retry-After", retryAfter.String())
					return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
						"code":       429,
						"message":    "Too Many Requests",
						"retryAfter": retryAfter.Seconds(),
					})
				}
			}

			// 执行后续处理
			err := next(c)

			// 并发限流释放资源
			for _, limiter := range limiters {
				if cl, ok := limiter.(*ConcurrencyLimiter); ok {
					cl.Release(key)
				}
			}

			return err
		}
	}
}

func (m *EchoRateLimitMiddleware) ShouldSkip(path, method string) bool {
	return false
}

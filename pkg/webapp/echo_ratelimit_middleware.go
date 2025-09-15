package webapp

import (
	"fmt"
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

			// 获取路由对应的限流器 (策略名 -> 限流器)
			limiters := m.options.getLimitersForRequest(method, path)
			if len(limiters) == 0 {
				return next(c)
			}

			key := c.RealIP() // 使用客户端 IP 作为限流键
			var maxRetryAfterSeconds float64

			for name, limiter := range limiters {
				allowed, retryAfter := limiter.TryAcquire(key)
				if !allowed {
					m.logger.Warn("rate limit exceeded",
						zap.String("path", path),
						zap.String("method", method),
						zap.String("policy", name), // 输出策略名
						zap.String("clientIP", key))

					// 取最大 retryAfter
					if retryAfter.Seconds() > maxRetryAfterSeconds {
						maxRetryAfterSeconds = retryAfter.Seconds()
					}
				}
			}

			if maxRetryAfterSeconds > 0 {
				c.Response().Header().Set("Retry-After", fmt.Sprintf("%.0f", maxRetryAfterSeconds))
				return c.JSON(http.StatusTooManyRequests, map[string]any{
					"code":       429,
					"message":    "Too Many Requests",
					"retryAfter": maxRetryAfterSeconds,
				})
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

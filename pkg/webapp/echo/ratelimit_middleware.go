package echo

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/ratelimit"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/zap"
)

type EchoRateLimitMiddleware struct {
	web.RatelimitProvider
	logger *zap.Logger
}

func newEchoRateLimitMiddleware(provider web.RatelimitProvider, logger *zap.Logger) EchoMiddleware {
	return &EchoRateLimitMiddleware{
		RatelimitProvider: provider,
		logger:            logger,
	}
}

// Handle 限流处理函数
func (m *EchoRateLimitMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			method := c.Request().Method
			path := c.Request().URL.Path

			// 获取路由对应的限流器 (策略名 -> 限流器)
			limiters := m.RoutePolicies(router.RequestMethod(method), path)

			if len(limiters) == 0 {

				limiters = append(limiters, m.DefaultPolicy())
			}

			if len(limiters) == 0 {
				return next(c)
			}

			key := c.RealIP() // 使用客户端 IP 作为限流键
			var maxRetryAfterSeconds float64

			for _, limiter := range limiters {
				handler, ok := m.Handler(limiter)
				if !ok {

				}
				allowed, retryAfter := handler.TryAcquire(key)
				if !allowed {
					m.logger.Warn("rate limit exceeded",
						zap.String("path", path),
						zap.String("method", method),
						zap.String("policy", limiter), // 输出策略名
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
				handler, ok := m.Handler(limiter)
				if !ok {

				}
				if cl, ok := handler.(*ratelimit.ConcurrencyLimiter); ok {
					cl.Release(key)
				}
			}

			return err
		}
	}
}

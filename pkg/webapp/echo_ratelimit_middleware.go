package webapp

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type EchoRateLimitMiddleware struct {
	options *RateLimitOptions
}

func newEchoRateLimitMiddleware(options *RateLimitOptions) EchoMiddleware {
	return &EchoRateLimitMiddleware{
		options: options,
	}
}

func (m *EchoRateLimitMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 获取限流策略名称
			policyName := c.Get("RateLimitPolicy")
			if policyName == nil {
				policyName = "default"
			}

			// 获取限流器
			limiter, exists := m.options.policies[policyName.(string)]
			if !exists {
				return next(c)
			}

			// 获取限流键
			key := c.RealIP() // Echo 框架使用 RealIP 获取客户端IP

			// 尝试获取访问权限
			allowed, retryAfter := limiter.TryAcquire(key)
			if !allowed {
				// 设置重试时间头
				c.Response().Header().Set("Retry-After", retryAfter.String())
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"code":       429,
					"message":    "Too Many Requests",
					"retryAfter": retryAfter.Seconds(),
				})
			}

			// 执行后续中间件和处理程序
			err := next(c)

			// 对于并发限流，在请求结束后释放资源
			if cl, ok := limiter.(*ConcurrencyLimiter); ok {
				cl.Release(key)
			}

			return err
		}
	}
}

func (m *EchoRateLimitMiddleware) ShouldSkip(path, method string) bool {
	return false
}

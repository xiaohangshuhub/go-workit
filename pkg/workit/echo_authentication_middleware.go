package workit

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// 授权中间件
type EchoAuthenticationMiddleware struct {
	skipPaths []string
	handlers  map[string]AuthenticationHandler
	logger    *zap.Logger
}

// 初始化授权中间件
func NewEchoAuthenticationMiddleware(options AuthenticateOptions, auth *AuthenticateProvider, logger *zap.Logger) *EchoAuthenticationMiddleware {

	return &EchoAuthenticationMiddleware{
		handlers:  auth.handlers,
		skipPaths: options.SkipPaths,
		logger:    logger,
	}
}

// 授权中间件处理逻辑
func (a *EchoAuthenticationMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for _, handler := range a.handlers {
				claims, err := handler.Authenticate(c.Request())
				if err == nil && claims != nil {
					c.Set("claims", claims)
					return next(c) // 认证成功，继续下一个中间件/handler
				}

				if err != nil {
					a.logger.Error("authentication failed",
						zap.String("scheme", handler.Scheme()),
						zap.Error(err),
						zap.String("path", c.Request().URL.Path),
						zap.String("method", c.Request().Method),
						zap.String("ip", c.RealIP()),
					)
				}
			}

			// 所有 handler 都认证失败
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
	}
}

// 跳过路径判断（支持通配符）
func (a *EchoAuthenticationMiddleware) ShouldSkip(path string) bool {
	path = strings.TrimRight(strings.TrimSpace(path), "/")

	for _, pattern := range a.skipPaths {
		pattern = strings.TrimRight(pattern, "/")
		if strings.HasSuffix(pattern, "/*") {
			base := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(path, base+"/") {
				return true
			}
		} else if pattern == path {
			return true
		}
	}
	return false
}

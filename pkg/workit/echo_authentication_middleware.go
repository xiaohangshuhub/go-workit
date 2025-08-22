package workit

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// 授权中间件
type EchoAuthenticationMiddleware struct {
	handlers map[string]AuthenticationHandler
	logger   *zap.Logger
	*AuthenticateOptions
}

// 初始化授权中间件
func newEchoAuthenticationMiddleware(options *AuthenticateOptions, auth *AuthenticateProvider, logger *zap.Logger) *EchoAuthenticationMiddleware {

	return &EchoAuthenticationMiddleware{
		handlers:            auth.handlers,
		logger:              logger,
		AuthenticateOptions: options,
	}
}

// 授权中间件处理逻辑
func (a *EchoAuthenticationMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			schemas := a.GetSchemesForRequest(c.Request().Method, c.Request().URL.Path)

			for _, scheme := range schemas {
				if handler, ok := a.handlers[scheme]; ok {

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
			}

			// 所有 handler 都认证失败
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
	}
}

// 跳过路径判断（支持通配符）
func (a *EchoAuthenticationMiddleware) ShouldSkip(path string, method string) bool {
	return a.AuthenticateOptions.ShouldSkip(path, method)
}

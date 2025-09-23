package echo

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/zap"
)

const contextClaimsKey = "claims"

// EchoAuthenticationMiddleware echo 授权中间件
type EchoAuthenticationMiddleware struct {
	web.RouterConfig
	logger *zap.Logger
}

// newEchoAuthenticationMiddleware 初始化授权中间件
func newEchoAuthenticationMiddleware(provider web.RouterConfig, logger *zap.Logger) *EchoAuthenticationMiddleware {
	return &EchoAuthenticationMiddleware{
		logger:       logger,
		RouterConfig: provider,
	}
}

// Handle 授权中间件处理逻辑
func (a *EchoAuthenticationMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			req := c.Request()
			method, path, ip := req.Method, req.URL.Path, c.RealIP()
			commonFields := []zap.Field{
				zap.String("path", path),
				zap.String("method", method),
				zap.String("ip", ip),
			}

			// 跳过不需要授权的路由
			if a.AllowAnonymous(web.RequestMethod(method), path) {
				return next(c)
			}

			// 获取鉴权方案
			schemes := a.Schemes(web.RequestMethod(method), path)
			if len(schemes) == 0 {
				if defaultScheme := a.GlobalScheme(); defaultScheme != "" {
					schemes = append(schemes, defaultScheme)
				} else {
					a.logger.Error("route not config scheme", commonFields...)
					return c.NoContent(http.StatusUnauthorized)
				}
			}

			// 执行鉴权方案
			for _, scheme := range schemes {
				handler, ok := a.Authenticate(scheme)
				if !ok {
					a.logger.Warn("authentication scheme not found",
						append(commonFields, zap.String("scheme", scheme))...,
					)
					continue
				}

				claims, err := handler.Authenticate(req)
				if err == nil && claims != nil {
					c.Set(contextClaimsKey, claims)
					return next(c) // 认证成功，继续下一个中间件/handler
				}

				a.logger.Error("authentication failed",
					append(commonFields,
						zap.String("scheme", scheme),
						zap.Error(err),
					)...,
				)
			}

			// 所有 handler 都认证失败
			return c.NoContent(http.StatusUnauthorized)
		}
	}
}

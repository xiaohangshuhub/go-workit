package echo

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/zap"
)

// EchoAuthorizationMiddleware echo 授权中间件
type EchoAuthorizationMiddleware struct {
	web.RouterConfig
	logger *zap.Logger
}

// newEchoAuthorizationMiddleware 初始化授权中间件
func newEchoAuthorizationMiddleware(provider web.RouterConfig, logger *zap.Logger) *EchoAuthorizationMiddleware {
	return &EchoAuthorizationMiddleware{
		logger:       logger,
		RouterConfig: provider,
	}
}

// Handle 授权中间件处理函数
func (a *EchoAuthorizationMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			method, path, ip := req.Method, req.URL.Path, c.RealIP()
			commonFields := []zap.Field{
				zap.String("method", method),
				zap.String("path", path),
				zap.String("ip", ip),
			}

			// OPTIONS 请求直接放行
			if method == http.MethodOptions {
				return next(c)
			}

			// 跳过不需要授权的路由
			if a.AllowAnonymous(web.RequestMethod(method), path) {
				return next(c)
			}

			claims := echoGetClaimsPrincipal(c)
			if claims == nil {
				a.logger.Error("authorization failed: claims is nil", commonFields...)
				return c.NoContent(http.StatusUnauthorized)
			}

			policyNames := a.Policies(web.RequestMethod(method), path)
			if len(policyNames) == 0 {
				// 未配置策略，默认放行
				return next(c)
			}

			for _, policyName := range policyNames {
				policyFunc, ok := a.Authorize(policyName)
				if !ok {
					a.logger.Error("authorization policy not found",
						append(commonFields, zap.String("policy", policyName))...,
					)
					continue
				}

				if !policyFunc(claims) {
					a.logger.Warn("authorization denied",
						append(commonFields, zap.String("policy", policyName))...,
					)
					return c.NoContent(http.StatusForbidden)
				}
			}

			// 所有策略都通过，继续执行
			return next(c)
		}
	}
}

// echoGetClaimsPrincipal 获取 ClaimsPrincipal
func echoGetClaimsPrincipal(c echo.Context) *web.ClaimsPrincipal {
	if claims, ok := c.Get("claims").(*web.ClaimsPrincipal); ok {
		return claims
	}
	return nil
}

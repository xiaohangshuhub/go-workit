package webapp

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// EchoAuthorizationMiddleware echo 授权中间件
type EchoAuthorizationMiddleware struct {
	policies map[string]func(claims *ClaimsPrincipal) bool
	logger   *zap.Logger
	*AuthenticationOptions
	*AuthorizationOptions
}

// newEchoAuthorizationMiddleware 初始化授权中间件
func newEchoAuthorizationMiddleware(auhtOptions *AuthenticationOptions, authorOptions *AuthorizationOptions, author *AuthorizationProvider, logger *zap.Logger) *EchoAuthorizationMiddleware {
	return &EchoAuthorizationMiddleware{
		policies:              author.policies,
		logger:                logger,
		AuthenticationOptions: auhtOptions,
		AuthorizationOptions:  authorOptions,
	}
}

// Handle 授权中间件处理函数
func (a *EchoAuthorizationMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method
			requestPath := c.Request().URL.Path

			// OPTIONS 请求直接放行
			if method == http.MethodOptions {
				return next(c)
			}

			claims := echoGetClaimsPrincipal(c)
			if claims == nil {
				a.logger.Error("authorization failed: ClaimsPrincipal is nil")
				return c.NoContent(http.StatusUnauthorized)
			}

			policyNames := a.AuthorizationOptions.getPoliciesForRequest(requestPath, method)

			for _, policyName := range policyNames {
				policyFunc, ok := a.policies[policyName]
				if !ok {
					a.logger.Warn("authorization failed: policy not found",
						zap.String("path", requestPath),
						zap.String("policy", policyName))
					continue
				}

				if !policyFunc(claims) {
					a.logger.Warn("authorization failed",
						zap.String("path", requestPath),
						zap.String("policy", policyName))
					return c.NoContent(http.StatusForbidden)
				}
			}

			// 继续执行后续中间件
			return next(c)
		}
	}
}

// echoGetClaimsPrincipal 获取ClaimsPrincipal
func echoGetClaimsPrincipal(c echo.Context) *ClaimsPrincipal {
	claims := c.Get("claims")
	if claims == nil {
		return nil
	}

	principal, ok := claims.(*ClaimsPrincipal)
	if !ok {
		return nil
	}

	return principal
}

// ShouldSkip 跳过逻辑
func (a *EchoAuthorizationMiddleware) ShouldSkip(path string, method string) bool {
	return a.shouldSkip(method, path)
}

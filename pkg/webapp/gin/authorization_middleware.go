package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/zap"
)

// GinAuthorizationMiddleware 授权中间件
type GinAuthorizationMiddleware struct {
	web.AuthorizationProvider
	logger *zap.Logger
}

// newGinAuthorizationMiddleware 初始化授权中间件
func newGinAuthorizationMiddleware(provider web.AuthorizationProvider, logger *zap.Logger) *GinAuthorizationMiddleware {
	return &GinAuthorizationMiddleware{
		AuthorizationProvider: provider,
		logger:                logger,
	}
}

// Handle 授权中间件处理函数
func (a *GinAuthorizationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path

		// OPTIONS 请求直接放行
		if method == http.MethodOptions {
			c.Next()
			return
		}

		// 跳过不需要授权的路由
		if a.AllowAnonymous(router.RequestMethod(method), path) {
			c.Next()
		}

		claims := ginGetClaimsPrincipal(c)
		if claims == nil {
			a.logger.Error("authorization failed: ClaimsPrincipal is nil")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		policyNames := a.RoutePolicies(router.RequestMethod(method), path)

		for _, policyName := range policyNames {
			policyFunc, ok := a.Handler(policyName)
			if !ok {
				a.logger.Warn("authorization failed: policy not found",
					zap.String("path", path),
					zap.String("policy", policyName))
				continue
			}

			if !policyFunc(claims) {
				a.logger.Warn("authorization failed",
					zap.String("path", path),
					zap.String("policy", policyName))
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		// 继续执行后续中间件
		c.Next()
	}
}

// ginGetClaimsPrincipal 从gin.Context中获取ClaimsPrincipal
func ginGetClaimsPrincipal(c *gin.Context) *web.ClaimsPrincipal {
	claims, exists := c.Get("claims")
	if !exists {
		return nil
	}

	principal, ok := claims.(*web.ClaimsPrincipal)
	if !ok {
		return nil
	}

	return principal

}

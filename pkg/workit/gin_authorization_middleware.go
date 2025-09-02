package workit

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinAuthorizationMiddleware 授权中间件
type GinAuthorizationMiddleware struct {
	policies map[string]func(claims *ClaimsPrincipal) bool
	logger   *zap.Logger
	*AuthenticateOptions
	*AuthorizeOptions
}

// newGinAuthorizationMiddleware 初始化授权中间件
func newGinAuthorizationMiddleware(authOptions *AuthenticateOptions, authorOptions *AuthorizeOptions, author *AuthorizationProvider, logger *zap.Logger) *GinAuthorizationMiddleware {
	return &GinAuthorizationMiddleware{
		policies:            author.policies,
		logger:              logger,
		AuthenticateOptions: authOptions,
		AuthorizeOptions:    authorOptions,
	}
}

// Handle 授权中间件处理函数
func (a *GinAuthorizationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		requestPath := c.Request.URL.Path

		// OPTIONS 请求直接放行
		if method == http.MethodOptions {
			c.Next()
			return
		}

		claims := ginGetClaimsPrincipal(c)
		if claims == nil {
			a.logger.Error("authorization failed: ClaimsPrincipal is nil")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		policyNames := a.AuthorizeOptions.getPoliciesForRequest(requestPath, method)

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
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		// 继续执行后续中间件
		c.Next()
	}
}

// ginGetClaimsPrincipal 从gin.Context中获取ClaimsPrincipal
func ginGetClaimsPrincipal(c *gin.Context) *ClaimsPrincipal {
	claims, exists := c.Get("claims")
	if !exists {
		return nil
	}

	principal, ok := claims.(*ClaimsPrincipal)
	if !ok {
		return nil
	}

	return principal

}

// ShouldSkip 跳过逻辑
func (a *GinAuthorizationMiddleware) ShouldSkip(path string, method string) bool {
	return a.shouldSkip(path, method)
}

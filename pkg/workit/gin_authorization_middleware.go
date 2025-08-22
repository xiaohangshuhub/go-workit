package workit

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 授权中间件
type GinAuthorizationMiddleware struct {
	policies map[string]func(claims *ClaimsPrincipal) bool
	logger   *zap.Logger
	*AuthenticateOptions
	*AuthorizeOptions
}

// 初始化授权中间件
func newGinAuthorizationMiddleware(authOptions *AuthenticateOptions, authorOptions *AuthorizeOptions, author *AuthorizationProvider, logger *zap.Logger) *GinAuthorizationMiddleware {
	return &GinAuthorizationMiddleware{
		policies:            author.policies,
		logger:              logger,
		AuthenticateOptions: authOptions,
		AuthorizeOptions:    authorOptions,
	}
}

// matchPathTemplate 简单支持 {var} 形式的路径变量匹配
func ginmatchPathTemplate(requestPath, template string) bool {
	reqParts := strings.Split(strings.Trim(requestPath, "/"), "/")
	tplParts := strings.Split(strings.Trim(template, "/"), "/")

	if len(reqParts) != len(tplParts) {
		return false
	}

	for i := range tplParts {
		if strings.HasPrefix(tplParts[i], "{") && strings.HasSuffix(tplParts[i], "}") {
			continue
		}
		if reqParts[i] != tplParts[i] {
			return false
		}
	}

	return true
}

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

		policyNames := a.AuthorizeOptions.GetPoliciesForRequest(requestPath, method)

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

// 跳过逻辑
func (a *GinAuthorizationMiddleware) ShouldSkip(path string, method string) bool {
	return a.AuthenticateOptions.ShouldSkip(path, method)
}

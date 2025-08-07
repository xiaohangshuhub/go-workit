package workit

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 授权中间件
type AuthorizationMiddleware struct {
	policies map[string]func(claims *ClaimsPrincipal) bool
	skipMap  map[string]struct{} // For faster skip lookups
}

// 初始化授权中间件
func NewAuthorizationMiddleware(options *AuthMiddlewareOptions, author *AuthorizationProvider) *AuthorizationMiddleware {
	// 鉴权跳过的接口授权则不需要执行
	skipMap := make(map[string]struct{}, len(options.SkipPaths))
	for _, path := range options.SkipPaths {
		skipMap[path] = struct{}{}
	}
	return &AuthorizationMiddleware{
		policies: author.policies,
		skipMap:  skipMap,
	}
}

func (a *AuthorizationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaimsPrincipal(c)
		if claims == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		requestPath := c.Request.URL.Path

		// 直接获取匹配的策略
		var matchedPolicy func(claims *ClaimsPrincipal) bool = a.policies[requestPath]

		// 如果没有直接匹配的策略，则尝试查找最长前缀匹配
		if matchedPolicy == nil {

			var longestMatch string

			for prefix, policy := range a.policies {
				// 确保是独立的路径段（/hello 不能匹配 /helloworld）
				if strings.HasPrefix(requestPath, prefix) && (len(requestPath) == len(prefix) || requestPath[len(prefix)] == '/') {

					if len(prefix) > len(longestMatch) {
						longestMatch = prefix
						matchedPolicy = policy
					}
				}
			}
		}

		// 执行策略
		if matchedPolicy != nil {
			if !matchedPolicy(claims) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		c.Next()
	}
}

// 跳过逻辑
func (a *AuthorizationMiddleware) ShouldSkip(path string) bool {
	path = strings.TrimSpace(path)
	_, exists := a.skipMap[path]
	return exists
}

func GetClaimsPrincipal(c *gin.Context) *ClaimsPrincipal {
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

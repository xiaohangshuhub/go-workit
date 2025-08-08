package workit

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 授权中间件
type AuthorizationMiddleware struct {
	policies map[string]func(claims *ClaimsPrincipal) bool
	skipMap  map[string]struct{} // For faster skip lookups
	logger   *zap.Logger
}

// 初始化授权中间件
func NewAuthorizationMiddleware(options *AuthMiddlewareOptions, author *AuthorizationProvider, logger *zap.Logger) *AuthorizationMiddleware {
	// 鉴权跳过的接口授权则不需要执行
	skipMap := make(map[string]struct{}, len(options.SkipPaths))
	for _, path := range options.SkipPaths {
		skipMap[path] = struct{}{}
	}
	return &AuthorizationMiddleware{
		policies: author.policies,
		skipMap:  skipMap,
		logger:   logger,
	}
}

func (a *AuthorizationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaimsPrincipal(c)
		requestPath := c.Request.URL.Path

		if claims == nil {
			a.logger.Error("authorization failed: ClaimsPrincipal is nil",
				zap.String("path", requestPath),
			)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var matchedPolicy func(claims *ClaimsPrincipal) bool
		var matchedPolicyKey string

		// 尝试直接匹配策略
		if directPolicy, ok := a.policies[requestPath]; ok {
			matchedPolicy = directPolicy
			matchedPolicyKey = requestPath
		} else {
			// 尝试最长前缀匹配
			var longestMatch string
			for prefix, policy := range a.policies {
				if strings.HasPrefix(requestPath, prefix) && (len(requestPath) == len(prefix) || requestPath[len(prefix)] == '/') {
					if len(prefix) > len(longestMatch) {
						longestMatch = prefix
						matchedPolicy = policy
						matchedPolicyKey = prefix
					}
				}
			}
		}

		// 执行策略
		if matchedPolicy != nil {
			if !matchedPolicy(claims) {
				a.logger.Warn("authorization failed",
					zap.String("path", requestPath),
					zap.String("matched_policy", matchedPolicyKey))

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

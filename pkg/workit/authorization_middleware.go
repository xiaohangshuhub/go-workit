package workit

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 授权中间件
type AuthorizationMiddleware struct {
	policies  map[string]func(claims *ClaimsPrincipal) bool
	authorize map[string][]string
	skipPaths []string
	logger    *zap.Logger
}

// 初始化授权中间件
func NewAuthorizationMiddleware(options AuthenticateOptions, author *AuthorizationProvider, logger *zap.Logger) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		policies:  author.policies,
		authorize: author.authorize,
		skipPaths: options.SkipPaths,
		logger:    logger,
	}
}

func (a *AuthorizationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 启用授权后,检查 claims 是否为空
		claims := GetClaimsPrincipal(c)
		if claims == nil {
			a.logger.Error("authorization failed: ClaimsPrincipal is nil")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 组装 route key
		requestPath := c.Request.URL.Path
		method := c.Request.Method
		routeKey := requestPath + ":" + method

		// 1. 先直接精确匹配
		policyNames, exists := a.authorize[routeKey]

		// 2. 如果没有精确匹配，尝试最长前缀匹配
		if !exists {
			var longestMatch string
			for k := range a.authorize {
				if strings.HasPrefix(routeKey, k) && len(k) > len(longestMatch) {
					// "/some/path:METHOD"
					parts := strings.SplitN(k, ":", 2)

					if len(parts) != 2 {
						continue // 跳过不规范的 key
					}

					km := parts[1]

					// 方法要完全一致才匹配
					if km != method {
						continue
					}

					longestMatch = k
					policyNames = a.authorize[k]
					exists = true
				}
			}
		}

		// 3. 如果找到对应策略名，逐一执行
		if exists {
			for _, policyName := range policyNames {

				if policyFunc, ok := a.policies[policyName]; ok {
					if !policyFunc(claims) {
						a.logger.Warn("authorization failed",
							zap.String("path", requestPath),
							zap.String("policy", policyName))
						c.AbortWithStatus(http.StatusForbidden)
						return
					}
				} else {
					a.logger.Warn("authorization failed: policy not found",
						zap.String("path", requestPath),
						zap.String("policy", policyName))
				}

			}
		}

		// 4. 继续后续处理
		c.Next()
	}
}

// 跳过逻辑
func (a *AuthorizationMiddleware) ShouldSkip(path string) bool {
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

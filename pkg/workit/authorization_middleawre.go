package workit

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 跳过配置:依赖注入
type AuthMiddlewareOptions struct {
	SkipPaths []string // Paths to skip authorization check
}

// 授权中间件
type AuthorizationMiddleware struct {
	options  *AuthMiddlewareOptions
	skipMap  map[string]struct{} // For faster skip lookups
	handlers map[string]AuthenticationHandler
}

// 初始化授权中间件
func NewAuthorizationMiddleware(options *AuthMiddlewareOptions, auth *AuthenticateProvider) *AuthorizationMiddleware {
	// Build a map for O(1) skip path lookups
	skipMap := make(map[string]struct{}, len(options.SkipPaths))
	for _, path := range options.SkipPaths {
		skipMap[path] = struct{}{}
	}

	return &AuthorizationMiddleware{
		handlers: auth.handlers,
		options:  options,
		skipMap:  skipMap,
	}
}

// 授权中间件处理逻辑
func (a *AuthorizationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, handler := range a.handlers {
			claims, err := handler.Authenticate(c.Request)
			if err == nil && claims != nil {
				c.Set("claims", claims)
				c.Next() // 认证成功，继续下一个中间件/handler
				return
			}
		}

		// 所有 handler 都认证失败
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

// 跳过逻辑
func (a *AuthorizationMiddleware) ShouldSkip(path string) bool {
	// Normalize path (optional: lowercase or trim slashes if needed)
	path = strings.TrimSpace(path)
	_, exists := a.skipMap[path]
	return exists
}

package workit

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddlewareOptions defines options for the AuthorizationMiddleware.
type AuthMiddlewareOptions struct {
	SkipPaths []string // Paths to skip authorization check
}

// AuthorizationMiddleware handles authorization logic.
type AuthorizationMiddleware struct {
	options  *AuthMiddlewareOptions
	skipMap  map[string]struct{} // For faster skip lookups
	handlers map[string]AuthenticationHandler
}

// NewAuthorizationMiddleware creates a new AuthorizationMiddleware.
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
		c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
	}
}

// ShouldSkip determines whether the middleware should skip a specific path.
func (a *AuthorizationMiddleware) ShouldSkip(path string) bool {
	// Normalize path (optional: lowercase or trim slashes if needed)
	path = strings.TrimSpace(path)
	_, exists := a.skipMap[path]
	return exists
}

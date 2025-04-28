package middleware

import (
	"github.com/gin-gonic/gin"
)

// AuthorizationMiddleware is a middleware for handling authorization.
type AuthorizationMiddleware struct {
	SkipPaths   map[string]struct{} // Paths to skip authorization check
	TokenHeader string              // Header field to read token from, e.g., "Authorization"
}

// NewAuthorizationMiddleware creates a new AuthorizationMiddleware.
// skipPaths: list of paths that don't require authorization
func NewAuthorizationMiddleware(skipPaths []string) *AuthorizationMiddleware {
	skipMap := make(map[string]struct{})
	for _, path := range skipPaths {
		skipMap[path] = struct{}{}
	}
	return &AuthorizationMiddleware{
		SkipPaths:   skipMap,
		TokenHeader: "Authorization", // 默认从 Authorization Header 取token
	}
}

// Handle implements the Middleware interface for AuthorizationMiddleware.
func (a *AuthorizationMiddleware) Handle() gin.HandlerFunc {
	//直接返回 gin.BasicAuth、gin JWT中间件，或者自定义
	return gin.BasicAuth(gin.Accounts{
		"admin": "password123",
	})
}

// ShouldSkip determines whether the middleware should skip a specific path.
func (a *AuthorizationMiddleware) ShouldSkip(path string) bool {
	_, ok := a.SkipPaths[path]
	return ok
}

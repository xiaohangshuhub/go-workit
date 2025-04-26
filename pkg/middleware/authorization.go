package middleware

import "github.com/gin-gonic/gin"

// AuthorizationMiddleware is a middleware for handling authorization.
type AuthorizationMiddleware struct{}

// Handle implements the Middleware interface for AuthorizationMiddleware.
func (a *AuthorizationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add your authorization logic here.
		c.Next()
	}
}

// ShouldSkip determines whether the middleware should skip a specific path.
func (a *AuthorizationMiddleware) ShouldSkip(path string) bool {
	// Add logic to skip certain paths if needed.
	return false
}
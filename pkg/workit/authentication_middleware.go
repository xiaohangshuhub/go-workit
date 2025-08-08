package workit

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 跳过配置:依赖注入
type AuthMiddlewareOptions struct {
	SkipPaths []string // Paths to skip authorization check
}

// 授权中间件
type AuthenticationMiddleware struct {
	skipMap  map[string]struct{} // For faster skip lookups
	handlers map[string]AuthenticationHandler
	logger   *zap.Logger
}

// 初始化授权中间件
func NewAuthenticationMiddleware(options *AuthMiddlewareOptions, auth *AuthenticateProvider, logger *zap.Logger) *AuthenticationMiddleware {
	// Build a map for O(1) skip path lookups
	skipMap := make(map[string]struct{}, len(options.SkipPaths))
	for _, path := range options.SkipPaths {
		skipMap[path] = struct{}{}
	}

	return &AuthenticationMiddleware{
		handlers: auth.handlers,
		skipMap:  skipMap,
		logger:   logger,
	}
}

// 授权中间件处理逻辑
func (a *AuthenticationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, handler := range a.handlers {

			claims, err := handler.Authenticate(c.Request)
			if err == nil && claims != nil {
				c.Set("claims", claims)
				c.Next() // 认证成功，继续下一个中间件/handler
				return
			}

			if err != nil {
				a.logger.Error("authentication failed",
					zap.String("scheme", handler.Scheme()),
					zap.Error(err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)
			}
		}

		// 所有 handler 都认证失败
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

// 跳过逻辑
func (a *AuthenticationMiddleware) ShouldSkip(path string) bool {
	// Normalize path (optional: lowercase or trim slashes if needed)
	path = strings.TrimSpace(path)
	_, exists := a.skipMap[path]
	return exists
}

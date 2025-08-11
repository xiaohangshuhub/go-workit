package workit

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 跳过配置:依赖注入

// 授权中间件
type GinAuthenticationMiddleware struct {
	skipPaths []string
	handlers  map[string]AuthenticationHandler
	logger    *zap.Logger
}

// 初始化授权中间件
func NewGinAuthenticationMiddleware(options AuthenticateOptions, auth *AuthenticateProvider, logger *zap.Logger) *GinAuthenticationMiddleware {

	return &GinAuthenticationMiddleware{
		handlers:  auth.handlers,
		skipPaths: options.SkipPaths,
		logger:    logger,
	}
}

// 授权中间件处理逻辑
func (a *GinAuthenticationMiddleware) Handle() gin.HandlerFunc {
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

// 跳过路径判断（支持通配符）
func (a *GinAuthenticationMiddleware) ShouldSkip(path string) bool {
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

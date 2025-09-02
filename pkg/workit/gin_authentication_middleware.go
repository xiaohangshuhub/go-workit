package workit

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinAuthenticationMiddleware 授权中间件
type GinAuthenticationMiddleware struct {
	handlers map[string]AuthenticationHandler
	logger   *zap.Logger
	*AuthenticateOptions
}

// newGinAuthenticationMiddleware 初始化授权中间件
func newGinAuthenticationMiddleware(options *AuthenticateOptions, auth *AuthenticateProvider, logger *zap.Logger) *GinAuthenticationMiddleware {

	return &GinAuthenticationMiddleware{
		handlers:            auth.handlers,
		logger:              logger,
		AuthenticateOptions: options,
	}
}

// Handle 授权中间件处理逻辑
func (a *GinAuthenticationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {

		schemas := a.getSchemesForRequest(c.Request.Method, c.Request.URL.Path)

		for _, scheme := range schemas {
			if handler, ok := a.handlers[scheme]; ok {

				claims, err := handler.Authenticate(c.Request)
				if err == nil && claims != nil {
					c.Set("claims", claims)
					c.Next() // 认证成功，继续下一个中间件/handler
					return
				}

				if err != nil {
					a.logger.Error("authentication failed",
						zap.String("scheme", scheme),
						zap.Error(err),
						zap.String("path", c.Request.URL.Path),
						zap.String("method", c.Request.Method),
						zap.String("ip", c.ClientIP()),
					)
				}
			} else {
				a.logger.Warn("authentication scheme not found",
					zap.String("scheme", scheme),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)
			}

		}

		// 所有 scheme 都认证失败
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

// ShouldSkip 跳过路径判断（支持通配符）
func (a *GinAuthenticationMiddleware) ShouldSkip(path string, method string) bool {

	return a.shouldSkip(path, method)
}

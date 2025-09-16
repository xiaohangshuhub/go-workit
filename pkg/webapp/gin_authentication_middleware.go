package webapp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinAuthenticationMiddleware 授权中间件
type GinAuthenticationMiddleware struct {
	handlers map[string]AuthenticationHandler
	logger   *zap.Logger
	*AuthenticationOptions
}

// newGinAuthenticationMiddleware 初始化授权中间件
func newGinAuthenticationMiddleware(options *AuthenticationOptions, auth *AuthenticateProvider, logger *zap.Logger) *GinAuthenticationMiddleware {

	return &GinAuthenticationMiddleware{
		handlers:              auth.handlers,
		logger:                logger,
		AuthenticationOptions: options,
	}
}

// Handle 授权中间件处理逻辑
func (a *GinAuthenticationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {

		method := c.Request.Method
		path := c.Request.URL.Path

		// 跳过不需要授权的路由
		if a.shouldSkip(method, path) {

			c.Next()
		}

		schemas := a.getSchemesForRequest(method, path)

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

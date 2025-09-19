package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/zap"
)

const contextClaimsKey = "claims"

// GinAuthenticationMiddleware 授权中间件
type GinAuthenticationMiddleware struct {
	web.AuthenticateProvider
	logger *zap.Logger
}

// newGinAuthenticationMiddleware 初始化授权中间件
func newGinAuthenticationMiddleware(provider web.AuthenticateProvider, logger *zap.Logger) *GinAuthenticationMiddleware {
	return &GinAuthenticationMiddleware{
		AuthenticateProvider: provider,
		logger:               logger,
	}
}

// Handle 授权中间件处理逻辑
func (a *GinAuthenticationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		method, path, ip := req.Method, req.URL.Path, c.ClientIP()
		commonFields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", ip),
		}

		// 跳过不需要授权的路由
		if a.AllowAnonymous(router.RequestMethod(method), path) {
			c.Next()
			return
		}

		schemes := a.RouteSchemes(router.RequestMethod(method), path)
		if len(schemes) == 0 {
			if defaultScheme := a.DefaultScheme(); defaultScheme != "" {
				schemes = append(schemes, defaultScheme)
			} else {
				a.logger.Error("route not configured with scheme", commonFields...)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}

		for _, scheme := range schemes {
			handler, ok := a.Handler(scheme)
			if !ok {
				a.logger.Warn("authentication scheme not found",
					append(commonFields, zap.String("scheme", scheme))...,
				)
				continue
			}

			claims, err := handler.Authenticate(req)
			if err == nil && claims != nil {
				c.Set(contextClaimsKey, claims)
				c.Next() // 认证成功，继续下一个中间件/handler
				return
			}

			a.logger.Error("authentication failed",
				append(commonFields,
					zap.String("scheme", scheme),
					zap.Error(err),
				)...,
			)
		}

		// 所有 scheme 都认证失败
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

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
	handlers map[string]AuthenticationHandler
	logger   *zap.Logger
	*AuthenticateOptions
}

// 初始化授权中间件
func newGinAuthenticationMiddleware(options *AuthenticateOptions, auth *AuthenticateProvider, logger *zap.Logger) *GinAuthenticationMiddleware {

	return &GinAuthenticationMiddleware{
		handlers:            auth.handlers,
		logger:              logger,
		AuthenticateOptions: options,
	}
}

// 授权中间件处理逻辑
func (a *GinAuthenticationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 根据请求路径寻找handler
		path := strings.TrimRight(c.Request.URL.Path, "/")
		routeKey := RouteKey{Method: c.Request.Method, Path: path}

		schemas := a.routeSchemesMap[routeKey]

		if len(schemas) == 0 {
			// 如果路由没有配置scheme，则执行默认方案
			schemas = []string{a.DefaultScheme}
		}

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
						zap.String("scheme", handler.Scheme()),
						zap.Error(err),
						zap.String("path", c.Request.URL.Path),
						zap.String("method", c.Request.Method),
						zap.String("ip", c.ClientIP()),
					)
				}
			} else {
				a.logger.Error("authentication scheme not found",
					zap.String("scheme", scheme),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)
			}

		}

		// 所有 scheme 都认证失败
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
	}
}

// 跳过路径判断（支持通配符）
func (a *GinAuthenticationMiddleware) ShouldSkip(path string, method string) bool {
	path = strings.TrimRight(strings.TrimSpace(path), "/")

	for k := range a.skipRoutesMap {
		// 先比对 HTTP 方法
		if !strings.EqualFold(k.Method, method) {
			continue
		}

		pattern := strings.TrimRight(k.Path, "/")

		// 模糊匹配：支持 /xxx/* 形式
		if strings.HasSuffix(pattern, "/*") {
			base := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(path, base+"/") {
				return true
			}
		} else if pattern == path {
			// 精确匹配
			return true
		}
	}
	return false
}

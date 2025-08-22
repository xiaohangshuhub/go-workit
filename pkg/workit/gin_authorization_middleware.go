package workit

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 授权中间件
type GinAuthorizationMiddleware struct {
	policies  map[string]func(claims *ClaimsPrincipal) bool
	authorize map[string][]string
	logger    *zap.Logger
	*AuthenticateOptions
}

// 初始化授权中间件
func newGinAuthorizationMiddleware(options *AuthenticateOptions, author *AuthorizationProvider, logger *zap.Logger) *GinAuthorizationMiddleware {
	return &GinAuthorizationMiddleware{
		policies:            author.policies,
		authorize:           author.authorize,
		logger:              logger,
		AuthenticateOptions: options,
	}
}

// matchPathTemplate 简单支持 {var} 形式的路径变量匹配
func ginmatchPathTemplate(requestPath, template string) bool {
	reqParts := strings.Split(strings.Trim(requestPath, "/"), "/")
	tplParts := strings.Split(strings.Trim(template, "/"), "/")

	if len(reqParts) != len(tplParts) {
		return false
	}

	for i := range tplParts {
		if strings.HasPrefix(tplParts[i], "{") && strings.HasSuffix(tplParts[i], "}") {
			continue
		}
		if reqParts[i] != tplParts[i] {
			return false
		}
	}

	return true
}

func (a *GinAuthorizationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		requestPath := c.Request.URL.Path

		// OPTIONS 请求直接放行
		if method == http.MethodOptions {
			c.Next()
			return
		}

		claims := ginGetClaimsPrincipal(c)
		if claims == nil {
			a.logger.Error("authorization failed: ClaimsPrincipal is nil")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 先尝试精确匹配
		routeKey := method + ":" + requestPath
		policyNames, exists := a.authorize[routeKey]

		// 如果没精确匹配，尝试模板匹配，最长匹配优先
		if !exists {
			longestMatchLen := -1
			for k := range a.authorize {
				parts := strings.SplitN(k, ":", 2)
				if len(parts) != 2 {
					continue
				}
				km, kp := parts[0], parts[1]
				if km != method {
					continue
				}

				if ginmatchPathTemplate(requestPath, kp) {
					if len(kp) > longestMatchLen {
						longestMatchLen = len(kp)
						policyNames = a.authorize[k]
						exists = true
					}
				}
			}
		}

		// 如果找到策略，执行
		if exists {
			for _, policyName := range policyNames {
				policyFunc, ok := a.policies[policyName]
				if !ok {
					a.logger.Warn("authorization failed: policy not found",
						zap.String("path", requestPath),
						zap.String("policy", policyName))
					continue
				}

				if !policyFunc(claims) {
					a.logger.Warn("authorization failed",
						zap.String("path", requestPath),
						zap.String("policy", policyName))
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
			}
		}

		// 继续执行后续中间件
		c.Next()
	}
}

// 跳过逻辑
func (a *GinAuthorizationMiddleware) ShouldSkip(path string, method string) bool {
	path = strings.TrimRight(strings.TrimSpace(path), "/")

	for k := range a.skipRoutesMap {
		// 先比对 HTTP 方法（忽略大小写）
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

func ginGetClaimsPrincipal(c *gin.Context) *ClaimsPrincipal {
	claims, exists := c.Get("claims")
	if !exists {
		return nil
	}

	principal, ok := claims.(*ClaimsPrincipal)
	if !ok {
		return nil
	}

	return principal

}

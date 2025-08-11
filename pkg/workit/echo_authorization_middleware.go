package workit

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// 授权中间件
type EchoAuthorizationMiddleware struct {
	policies  map[string]func(claims *ClaimsPrincipal) bool
	authorize map[string][]string
	skipPaths []string
	logger    *zap.Logger
}

// 初始化授权中间件
func NewEchoAuthorizationMiddleware(options AuthenticateOptions, author *AuthorizationProvider, logger *zap.Logger) *EchoAuthorizationMiddleware {
	return &EchoAuthorizationMiddleware{
		policies:  author.policies,
		authorize: author.authorize,
		skipPaths: options.SkipPaths,
		logger:    logger,
	}
}

// matchPathTemplate 简单支持 {var} 形式的路径变量匹配
func matchPathTemplate(requestPath, template string) bool {
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

func (a *EchoAuthorizationMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method
			requestPath := c.Request().URL.Path

			// OPTIONS 请求直接放行
			if method == http.MethodOptions {
				return next(c)
			}

			claims := echoGetClaimsPrincipal(c)
			if claims == nil {
				a.logger.Error("authorization failed: ClaimsPrincipal is nil")
				return c.NoContent(http.StatusUnauthorized)
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

					if matchPathTemplate(requestPath, kp) {
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
						return c.NoContent(http.StatusForbidden)
					}
				}
			}

			// 继续执行后续中间件
			return next(c)
		}
	}
}

// 跳过逻辑
func (a *EchoAuthorizationMiddleware) ShouldSkip(path string) bool {
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

func echoGetClaimsPrincipal(c echo.Context) *ClaimsPrincipal {
	claims := c.Get("claims")
	if claims == nil {
		return nil
	}

	principal, ok := claims.(*ClaimsPrincipal)
	if !ok {
		return nil
	}

	return principal
}

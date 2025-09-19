package web

import "github.com/xiaohangshuhub/go-workit/pkg/webapp/router"

// AuthenticateProvider 鉴权提供者接口
type AuthenticateProvider interface {
	DefaultScheme() string                                          // 返回默认的scheme
	RouteSchemes(method router.RequestMethod, path string) []string // 返回路由的鉴权方案
	Handler(scheme string) (AuthenticationHandler, bool)            // 返回scheme对应的鉴权处理器
	AllowAnonymous(method router.RequestMethod, path string) bool   // 匿名访问
}

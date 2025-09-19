package web

import "github.com/xiaohangshuhub/go-workit/pkg/webapp/router"

// AuthenticateProvider 鉴权提供者接口
type AuthorizationProvider interface {
	DefaultPolicy() string                                            // 默认策略
	RoutePolicies(method router.RequestMethod, path string) []string  // 鉴权策略列表
	Handler(policy string) (func(claims *ClaimsPrincipal) bool, bool) // 鉴权策略
	AllowAnonymous(method router.RequestMethod, path string) bool     // 允许匿名访问
}

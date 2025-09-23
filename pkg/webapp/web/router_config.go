package web

type RouterConfig interface {
	RouteConfig() []*RouteConfig                                        // 路由配置
	GroupRouteConfig() []*GroupRouteConfig                              // 分组路由配置
	GlobalScheme() string                                               // 全局鉴权方案
	GlobalPolicy() string                                               // 全局授权方案
	GlobalRatelimit() string                                            // 全局限流方案
	AllowAnonymous(method RequestMethod, path string) bool              // 允许匿名访问
	Schemes(method RequestMethod, path string) []string                 // 路由鉴权方案
	Authenticate(scheme string) (Authenticate, bool)                    // 鉴权处理
	Policies(method RequestMethod, path string) []string                // 路由授权策略
	Authorize(policy string) (func(claims *ClaimsPrincipal) bool, bool) // 授权处理
	RateLimits(method RequestMethod, path string) []string              // 限流策略
	RateLimiter(policy string) (RateLimiter, bool)                      // 限流处理器
}

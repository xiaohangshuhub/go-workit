package web

import "github.com/gobwas/glob"

// RequestMethod is a type for HTTP request method
type RequestMethod string

const (
	GET     RequestMethod = "GET"
	POST    RequestMethod = "POST"
	PUT     RequestMethod = "PUT"
	DELETE  RequestMethod = "DELETE"
	PATCH   RequestMethod = "PATCH"
	HEAD    RequestMethod = "HEAD"
	OPTIONS RequestMethod = "OPTIONS"
)

// ANY is a slice of all HTTP request methods
var ANY = []RequestMethod{GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS}

// RouteKey is a struct for a route key definition
type RouteKey struct {
	Path   string
	Method string
	Glob   glob.Glob // 预编译
}

type Router interface {
	Config() []*RouteConfig                                             // 路由配置
	GroupConfig() []*GroupRouteConfig                                   // 分组路由配置
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

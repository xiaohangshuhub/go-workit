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

// RouteKey is a struct for a route key definition
type RouteKey struct {
	Path   string
	Method string
	Glob   glob.Glob // 预编译
}

type Router interface {
	GlobalScheme() string                                               // 全局鉴权方案
	GlobalPolicy() string                                               // 全局授权方案
	GlobalRatelimit() string                                            // 全局限流方案
	Authenticate(scheme string) (Authenticate, bool)                    // 鉴权处理
	Authorize(policy string) (func(claims *ClaimsPrincipal) bool, bool) // 授权处理
	RateLimiter(policy string) (RateLimiter, bool)                      // 限流处理器
}

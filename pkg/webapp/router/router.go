package router

import (
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/auth"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/authz"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/ratelimit"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

// Router 提供路由相关配置
type Router struct {
	authenticate    map[string]web.Authenticate                       // 鉴权handler
	authorize       map[string]func(claims *web.ClaimsPrincipal) bool // 授权handler
	rateLimiters    map[string]web.RateLimiter                        // 限流 handler 注册表 (name -> handler)
	globalScheme    string                                            // 默认鉴权方案默认鉴权方案
	globalPolicy    string                                            // 默认鉴权方案默认鉴权方案
	globalRatelimit string                                            // 默认鉴权方案默认鉴权方案
}

func NewRouter(authOpts *auth.Options, authzOpts *authz.Options, ratelimitOpts *ratelimit.Options) *Router {
	p := &Router{
		authenticate:    authOpts.Schemes(),
		authorize:       authzOpts.Policies(),
		rateLimiters:    ratelimitOpts.Policies(),
		globalScheme:    authOpts.DefaultScheme,
		globalPolicy:    authzOpts.DefaultPolicy,
		globalRatelimit: ratelimitOpts.DefaultPolicy,
	}
	return p
}

// GlobalScheme 全局鉴权方案
func (p *Router) GlobalScheme() string {
	return p.globalScheme
}

// GlobalPolicy 全局授权方案
func (p *Router) GlobalPolicy() string {
	return p.globalPolicy
}

// GlobalRatelimit 全局限流方案
func (p *Router) GlobalRatelimit() string {
	return p.globalRatelimit
}

// Authenticate 鉴权处理
func (p *Router) Authenticate(scheme string) (web.Authenticate, bool) {
	if handler, ok := p.authenticate[scheme]; ok {
		return handler, true
	}

	return nil, false
}

// Authorize 授权处理
func (p *Router) Authorize(policy string) (func(claims *web.ClaimsPrincipal) bool, bool) {
	if policy, ok := p.authorize[policy]; ok {
		return policy, true
	}

	return nil, false
}

// RateLimiter 限流处理器
func (p *Router) RateLimiter(policy string) (web.RateLimiter, bool) {
	if policy, ok := p.rateLimiters[policy]; ok {
		return policy, true
	}

	return nil, false
}

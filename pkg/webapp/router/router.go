package router

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/auth"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/authz"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/ratelimit"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

// Provider 提供路由相关配置
type Router struct {
	routeConfigs    []*web.RouteConfig                                // 路由配置
	groupConfigs    []*web.GroupRouteConfig                           // 路由组配置
	patternMap      map[string]string                                 // 处理函数标识到模式字符串的映射
	schemesMap      map[web.RouteKey][]string                         // 鉴权路由 → 鉴权方案列表
	allowAnonymous  map[web.RouteKey]struct{}                         // 跳过路由 → 空结构（集合）
	authenticate    map[string]web.Authenticate                       // 鉴权handler
	policiesMap     map[web.RouteKey][]string                         // 路由 → 授权策略列表
	authorize       map[string]func(claims *web.ClaimsPrincipal) bool // 授权handler
	rateLimitsMap   map[web.RouteKey][]string                         // 路由 → 限流策略列表
	rateLimiters    map[string]web.RateLimiter                        // 限流 handler 注册表 (name -> handler)
	globalScheme    string                                            // 默认鉴权方案默认鉴权方案
	globalPolicy    string                                            // 默认鉴权方案默认鉴权方案
	globalRatelimit string                                            // 默认鉴权方案默认鉴权方案
	router          *httprouter.Router                                // httprouter 实例
}

func NewRouter(opts *Options, authOpts *auth.Options, authzOpts *authz.Options, ratelimitOpts *ratelimit.Options) *Router {
	p := &Router{
		routeConfigs:    opts.routeConfigs,
		groupConfigs:    opts.groupConfigs,
		patternMap:      make(map[string]string),
		schemesMap:      make(map[web.RouteKey][]string),
		allowAnonymous:  make(map[web.RouteKey]struct{}),
		authenticate:    authOpts.Schemes(),
		policiesMap:     make(map[web.RouteKey][]string),
		authorize:       authzOpts.Policies(),
		rateLimitsMap:   make(map[web.RouteKey][]string),
		rateLimiters:    ratelimitOpts.Policies(),
		globalScheme:    authOpts.DefaultScheme,
		globalPolicy:    authzOpts.DefaultPolicy,
		globalRatelimit: ratelimitOpts.DefaultPolicy,
		router:          httprouter.New(),
	}

	// 根据传入的配置构建路由映射（注册到 httprouter 并填充各种映射表）
	p.buildRoutes()

	return p
}

// buildRoutes 根据传入的 routeConfigs / groupConfigs 注册路由并填充 maps
func (p *Router) buildRoutes() {
	// 先处理 groupConfigs（支持 prefix + group defaults）
	for _, group := range p.groupConfigs {
		prefix := normalizePrefix(group.Prefix)
		for _, r := range group.Routes {
			pattern := joinPaths(prefix, r.Path)
			pattern = p.normalizePattern(pattern)

			methodStr := string(r.Method)
			p.registerRoute(methodStr, pattern)

			key := web.RouteKey{Method: methodStr, Path: pattern}

			// 合并 schemes: group.Schemes + route.Schemes（route 追加到 group）
			mergedSchemes := append([]string{}, group.Schemes...)
			mergedSchemes = append(mergedSchemes, r.Schemes...)
			if len(mergedSchemes) > 0 {
				p.schemesMap[key] = mergedSchemes
			}

			// 合并 policies
			mergedPolicies := append([]string{}, group.Policies...)
			mergedPolicies = append(mergedPolicies, r.Policies...)
			if len(mergedPolicies) > 0 {
				p.policiesMap[key] = mergedPolicies
			}

			// 合并 rate limiters
			mergedRateLimits := append([]string{}, group.RateLimiter...)
			mergedRateLimits = append(mergedRateLimits, r.RateLimiter...)
			if len(mergedRateLimits) > 0 {
				p.rateLimitsMap[key] = mergedRateLimits
			}

			// allow anonymous 如果 group 或 route 任一为 true 则允许
			if group.AllowAnonymous || r.AllowAnonymous {
				p.allowAnonymous[key] = struct{}{}
			}
		}
	}

	// 再处理顶级 routeConfigs
	for _, r := range p.routeConfigs {
		pattern := p.normalizePattern(r.Path)
		methodStr := string(r.Method)

		p.registerRoute(methodStr, pattern)

		key := web.RouteKey{Method: methodStr, Path: pattern}

		if len(r.Schemes) > 0 {
			p.schemesMap[key] = append([]string{}, r.Schemes...)
		}
		if len(r.Policies) > 0 {
			p.policiesMap[key] = append([]string{}, r.Policies...)
		}
		if len(r.RateLimiter) > 0 {
			p.rateLimitsMap[key] = append([]string{}, r.RateLimiter...)
		}
		if r.AllowAnonymous {
			p.allowAnonymous[key] = struct{}{}
		}
	}

}

// normalizePattern：
// - 确保以 / 开头
// - 去除末尾多余的 /（非根路径）
// - 将 {param} 转为 :param 以兼容 httprouter 的参数形式
func (p *Router) normalizePattern(path string) string {
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// 去掉末尾 /（除根路径）
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	// 把 {id} => :id
	re := regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)
	path = re.ReplaceAllString(path, ":$1")
	return path
}

// normalizePrefix：确保前缀规范（允许 "" 或 "/prefix"）
func normalizePrefix(prefix string) string {
	if prefix == "" {
		return ""
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	// 去掉末尾 /
	if len(prefix) > 1 && strings.HasSuffix(prefix, "/") {
		prefix = strings.TrimSuffix(prefix, "/")
	}
	return prefix
}

// joinPaths 合并 prefix 和 path，保证格式正确
func joinPaths(prefix, path string) string {
	if prefix == "" {
		return path
	}
	if path == "" || path == "/" {
		return prefix
	}
	return strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(path, "/")
}

// RouteConfig 路由配置
func (p *Router) Config() []*web.RouteConfig {
	return p.routeConfigs
}

// GroupRouteConfig 路由组配置
func (p *Router) GroupConfig() []*web.GroupRouteConfig {
	return p.groupConfigs
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

// AllowAnonymous 是否允许匿名访问
func (p *Router) AllowAnonymous(method web.RequestMethod, path string) bool {
	routeKey, found := p.findMatchingRoute(string(method), path)
	if !found {
		return false
	}

	// 检查是否在跳过列表中
	_, exist := p.allowAnonymous[routeKey]
	return exist
}

// Schemes 路由鉴权方案
func (p *Router) Schemes(method web.RequestMethod, path string) []string {
	routeKey, found := p.findMatchingRoute(string(method), path)
	if !found {
		return []string{}
	}

	schemes := p.schemesMap[routeKey]

	return schemes
}

// Authenticate 鉴权处理
func (p *Router) Authenticate(scheme string) (web.Authenticate, bool) {
	if handler, ok := p.authenticate[scheme]; ok {
		return handler, true
	}

	return nil, false
}

// Policies 路由授权策略
func (p *Router) Policies(method web.RequestMethod, path string) []string {
	routeKey, found := p.findMatchingRoute(string(method), path)
	if !found {
		return []string{}
	}

	policies, exists := p.policiesMap[routeKey]
	if !exists || len(policies) == 0 {

		return []string{}
	}

	// 返回找到的策略
	return policies
}

// Authorize 授权处理
func (p *Router) Authorize(policy string) (func(claims *web.ClaimsPrincipal) bool, bool) {
	if policy, ok := p.authorize[policy]; ok {
		return policy, true
	}

	return nil, false
}

// RateLimits 限流策略
func (p *Router) RateLimits(method web.RequestMethod, path string) []string {
	routeKey, found := p.findMatchingRoute(string(method), path)
	if !found {
		return []string{}
	}

	policies, exists := p.rateLimitsMap[routeKey]
	if !exists || len(policies) == 0 {

		return []string{}
	}

	// 返回找到的策略
	return policies
}

// RateLimiter 限流处理器
func (p *Router) RateLimiter(policy string) (web.RateLimiter, bool) {
	if policy, ok := p.rateLimiters[policy]; ok {
		return policy, true
	}

	return nil, false
}

// findMatchingRoute 使用 httprouter 查找匹配的路由
func (p *Router) findMatchingRoute(method, path string) (web.RouteKey, bool) {
	// 使用 httprouter 的 Lookup 方法查找匹配的路由
	handler, params, _ := p.router.Lookup(method, path)
	if handler == nil {
		return web.RouteKey{}, false
	}

	// 重构路由模式
	pattern := path
	for _, param := range params {
		pattern = p.replaceFirst(pattern, param.Value, ":"+param.Key)
	}

	return web.RouteKey{Method: method, Path: pattern}, true
}

// replaceFirst 替换字符串中第一次出现的子串,返回替换后的字符串
func (p *Router) replaceFirst(s, old, new string) string {
	if old == "" {
		return s
	}

	idx := strings.Index(s, old)
	if idx == -1 {
		return s
	}

	return s[:idx] + new + s[idx+len(old):]
}

// registerRoute 注册路由到 httprouter（内部方法）
func (p *Router) registerRoute(method, pattern string) {

	// 为每个路由创建唯一的处理函数标识
	handlerID := fmt.Sprintf("%s:%s", method, pattern)

	// 检查是否已注册
	if _, exists := p.patternMap[handlerID]; exists {
		return // 已经注册过，跳过
	}

	// 注册路由到 httprouter
	switch method {
	case "GET":
		p.router.GET(pattern, p.dummyHandler)
	case "POST":
		p.router.POST(pattern, p.dummyHandler)
	case "PUT":
		p.router.PUT(pattern, p.dummyHandler)
	case "DELETE":
		p.router.DELETE(pattern, p.dummyHandler)
	case "PATCH":
		p.router.PATCH(pattern, p.dummyHandler)
	case "HEAD":
		p.router.HEAD(pattern, p.dummyHandler)
	case "OPTIONS":
		p.router.OPTIONS(pattern, p.dummyHandler)
	}

	// 存储模式映射
	p.patternMap[handlerID] = pattern
}

// dummyHandler 空处理函数
func (p *Router) dummyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 空处理函数，仅用于路由匹配
}

package authorization

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

// Provider is the interface for authorization providers.
type Provider struct {
	defaultPolicy    string
	routePoliciesMap map[router.RouteKey][]string // 路由 → 授权策略列表
	router           *httprouter.Router           // httprouter 实例
	patternMap       map[string]string            // 处理函数标识到模式字符串的映射
	policyMap        map[string]func(claims *web.ClaimsPrincipal) bool
	allowAnonymous   map[router.RouteKey]struct{}
}

// NewAuthorizationProvider creates a new AuthorizationProvider with the given policies.
func newAuthorizationProvider(defaultPolicy string, routePoliciesMap map[router.RouteKey][]string, allowAnonymous map[router.RouteKey]struct{}, policies map[string]func(claims *web.ClaimsPrincipal) bool) *Provider {
	p := &Provider{
		defaultPolicy:    defaultPolicy,
		routePoliciesMap: routePoliciesMap,
		allowAnonymous:   allowAnonymous,
		policyMap:        policies,
		router:           httprouter.New(),
		patternMap:       make(map[string]string),
	}

	// 注册匿名路由
	for key := range allowAnonymous {
		p.registerRoute(key.Method, key.Path)
	}

	// 注册策略路由
	for key := range routePoliciesMap {
		p.registerRoute(key.Method, key.Path)
	}

	return p
}

func (p *Provider) DefaultPolicy() string {
	return p.defaultPolicy
}

func (p *Provider) RoutePolicies(method router.RequestMethod, path string) []string {

	routeKey, found := p.findMatchingRoute(string(method), path)
	if !found {
		return []string{}
	}

	policies, exists := p.routePoliciesMap[routeKey]
	if !exists || len(policies) == 0 {

		return []string{}
	}

	// 返回找到的策略
	return policies

}
func (p *Provider) Handler(policy string) (func(claims *web.ClaimsPrincipal) bool, bool) {

	if policy, ok := p.policyMap[policy]; ok {
		return policy, true
	}

	return nil, false
}

func (p *Provider) AllowAnonymous(method router.RequestMethod, path string) bool {

	routeKey, found := p.findMatchingRoute(string(method), path)
	if !found {
		return false
	}

	_, exist := p.allowAnonymous[routeKey]
	return exist
}

// FindMatchingRoute 使用 httprouter 查找匹配的路由
//
// method 请求方法
// path 请求路径
//
// 返回 RouteKey 类型，包含方法和路径,以及是否找到匹配的路由
func (a *Provider) findMatchingRoute(method, path string) (router.RouteKey, bool) {
	// 使用 httprouter 的 Lookup 方法查找匹配的路由
	handler, params, _ := a.router.Lookup(method, path)
	if handler == nil {
		return router.RouteKey{}, false
	}

	// 重构路由模式
	pattern := path
	for _, param := range params {
		pattern = replaceFirst(pattern, param.Value, ":"+param.Key)
	}

	return router.RouteKey{Method: method, Path: pattern}, true
}

// registerRoute 注册路由到 httprouter（内部方法）
//
// method 请求方法
// pattern 请求路径
func (a *Provider) registerRoute(method, pattern string) {

	// 为每个路由创建唯一的处理函数标识
	handlerID := fmt.Sprintf("%s:%s", method, pattern)

	// 检查是否已注册
	if _, exists := a.patternMap[handlerID]; exists {
		return // 已经注册过，跳过
	}

	// 注册路由到 httprouter
	switch method {
	case "GET":
		a.router.GET(pattern, a.dummyHandler)
	case "POST":
		a.router.POST(pattern, a.dummyHandler)
	case "PUT":
		a.router.PUT(pattern, a.dummyHandler)
	case "DELETE":
		a.router.DELETE(pattern, a.dummyHandler)
	case "PATCH":
		a.router.PATCH(pattern, a.dummyHandler)
	case "HEAD":
		a.router.HEAD(pattern, a.dummyHandler)
	case "OPTIONS":
		a.router.OPTIONS(pattern, a.dummyHandler)
	}

	// 存储模式映射
	a.patternMap[handlerID] = pattern
}

// dummyHandler 空处理函数
func (a *Provider) dummyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {}

// replaceFirst 替换字符串中第一次出现的子串
// 注意：如果 old 为空字符串，则返回 s
// s 原字符串 old 要替换的子串 new 要替换的子串
//
// 返回替换后的字符串
func replaceFirst(s, old, new string) string {
	if old == "" {
		return s
	}

	idx := strings.Index(s, old)
	if idx == -1 {
		return s
	}

	return s[:idx] + new + s[idx+len(old):]
}

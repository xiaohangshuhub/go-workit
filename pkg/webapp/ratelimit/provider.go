package ratelimit

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

type Provider struct {
	defaultPolicy     string                          // 默认限流策略名称
	policiesMap       map[string]web.RateLimitHandler //策略映射
	routeRateLimitMap map[router.RouteKey][]string    // 路由 → 限流方案
	router            *httprouter.Router              // httprouter 实例
	patternMap        map[string]string               // 处理函数标识到模式字符串的映射
}

func NewProvider(defaultPolicy string, routeRateLimitMap map[router.RouteKey][]string, policiesMap map[string]web.RateLimitHandler) *Provider {
	p := &Provider{
		defaultPolicy:     defaultPolicy,
		routeRateLimitMap: routeRateLimitMap,
		policiesMap:       policiesMap,
		router:            httprouter.New(),
		patternMap:        make(map[string]string),
	}

	// 注册所有鉴权路由
	for key := range routeRateLimitMap {
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

	policies, exists := p.routeRateLimitMap[routeKey]
	if !exists || len(policies) == 0 {

		return []string{}
	}

	// 返回找到的策略
	return policies

}

// Handler
func (p *Provider) Handler(policy string) (web.RateLimitHandler, bool) {

	if policy, ok := p.policiesMap[policy]; ok {
		return policy, true
	}

	return nil, false
}

// findMatchingRoute 查找匹配路由
func (p *Provider) findMatchingRoute(method, path string) (router.RouteKey, bool) {
	handler, params, _ := p.router.Lookup(method, path)
	if handler == nil {
		return router.RouteKey{}, false
	}

	// 还原参数化路径
	pattern := path
	for _, param := range params {
		pattern = replaceFirst(pattern, param.Value, ":"+param.Key)
	}
	return router.RouteKey{Method: method, Path: pattern}, true
}

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

// 注册路由（内部方法）
func (p *Provider) registerRoute(method, pattern string) {
	// 生成唯一 handlerID
	handlerID := fmt.Sprintf("%s:%s", method, pattern)
	if _, exists := p.patternMap[handlerID]; exists {
		return // 已注册
	}

	// 注册到 httprouter
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
	p.patternMap[handlerID] = pattern
}

// dummyHandler 空处理函数，仅用于路由匹配
func (p *Provider) dummyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

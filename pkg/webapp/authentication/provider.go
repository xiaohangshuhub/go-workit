package authentication

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

// AuthenticateProvider 鉴权提供者
type Provider struct {
	defaultScheme    string                               // 默认鉴权方案
	router           *httprouter.Router                   // httprouter 实例
	routeSchemesMap  map[router.RouteKey][]string         // 鉴权路由 → 鉴权方案列表
	allowAnonymous   map[router.RouteKey]struct{}         // 跳过路由 → 空结构（集合）
	patternMap       map[string]string                    // 处理函数标识到模式字符串的映射
	schemeHandlerMap map[string]web.AuthenticationHandler // 鉴权handler
}

// NewAuthenticateProvider creates a new AuthenticateProvider.
func newProvider(defaultScheme string, routeSchemMap map[router.RouteKey][]string, allowAnonymous map[router.RouteKey]struct{}, schemesHandlerMap map[string]web.AuthenticationHandler) (*Provider, error) {

	if defaultScheme == "" {
		return nil, fmt.Errorf("no default authentication scheme configured")
	}

	p := &Provider{
		defaultScheme:    defaultScheme,
		routeSchemesMap:  routeSchemMap,
		allowAnonymous:   allowAnonymous,
		schemeHandlerMap: schemesHandlerMap,
		patternMap:       make(map[string]string),
		router:           httprouter.New(),
	}

	// 注册所有匿名路由
	for key := range allowAnonymous {
		p.registerRoute(key.Method, key.Path)
	}

	// 注册所有鉴权路由
	for key := range routeSchemMap {
		p.registerRoute(key.Method, key.Path)
	}

	return p, nil
}

func (p *Provider) DefaultScheme() string {

	return p.defaultScheme
}

func (p *Provider) RouteSchemes(method router.RequestMethod, path string) []string {

	routeKey, found := p.findMatchingRoute(string(method), path)
	if !found {
		return []string{}
	}

	schemes, exists := p.routeSchemesMap[routeKey]

	if !exists || len(schemes) == 0 {
		return []string{p.defaultScheme}
	}

	return schemes

}
func (p *Provider) Handler(scheme string) (web.AuthenticationHandler, bool) {

	if handler, ok := p.schemeHandlerMap[scheme]; ok {
		return handler, true
	}

	return nil, false
}

func (p *Provider) AllowAnonymous(method router.RequestMethod, path string) bool {

	routeKey, found := p.findMatchingRoute(string(method), path)
	if !found {
		return false
	}

	// 检查是否在跳过列表中
	_, shouldSkip := p.allowAnonymous[routeKey]
	return shouldSkip

}

// FindMatchingRoute 使用 httprouter 查找匹配的路由
//
// method  请求方法
// path  请求路径
//
// 返回 RouteKey 和 bool 值，bool 值为 true 表示找到了匹配的路由，否则为 false
func (p *Provider) findMatchingRoute(method, path string) (router.RouteKey, bool) {
	// 使用 httprouter 的 Lookup 方法查找匹配的路由
	handler, params, _ := p.router.Lookup(method, path)
	if handler == nil {
		return router.RouteKey{}, false
	}

	// 重构路由模式
	pattern := path
	for _, param := range params {
		pattern = p.replaceFirst(pattern, param.Value, ":"+param.Key)
	}

	return router.RouteKey{Method: method, Path: pattern}, true
}

// replaceFirst 替换字符串中第一次出现的子串
// 注意：如果 old 为空字符串，则返回 s
// s 原字符串 old 要替换的子串 new 要替换的子串
//
// 返回替换后的字符串
func (p *Provider) replaceFirst(s, old, new string) string {
	if old == "" {
		return s
	}

	idx := strings.Index(s, old)
	if idx == -1 {
		return s
	}

	return s[:idx] + new + s[idx+len(old):]
}

// RegisterRoute 注册路由到 httprouter（内部方法）
func (p *Provider) registerRoute(method, pattern string) {

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
func (p *Provider) dummyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 空处理函数，仅用于路由匹配
}

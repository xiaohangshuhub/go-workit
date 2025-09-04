package workit

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/julienschmidt/httprouter"
)

// RouteAuthenticationSchemes 表示路由和认证方案的关联
type RouteAuthenticationSchemes struct {
	Routes  []Route  // 鉴权路由
	Schemes []string // 鉴权方案列表
}

// AuthenticateOptions 表示授权选项配置。
type AuthenticateOptions struct {
	DefaultScheme   string
	routeSchemesMap map[RouteKey][]string // 鉴权路由 → 鉴权方案列表
	schemeRoutesMap map[string][]RouteKey // 鉴权方案 → 鉴权路由列表
	skipRoutesMap   map[RouteKey]struct{} // 跳过路由 → 空结构（集合）
	router          *httprouter.Router    // httprouter 实例
	patternMap      map[string]string     // 处理函数标识到模式字符串的映射
	mu              sync.Mutex            // 保护并发访问
}

// newAuthenticateOptions 创建一个新的 AuthenticateOptions 实例。
func newAuthenticateOptions() *AuthenticateOptions {
	return &AuthenticateOptions{
		routeSchemesMap: make(map[RouteKey][]string),
		schemeRoutesMap: make(map[string][]RouteKey),
		skipRoutesMap:   make(map[RouteKey]struct{}),
		router:          httprouter.New(),
		patternMap:      make(map[string]string),
	}
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

// RegisterRoute 注册路由到 httprouter（内部方法）
func (a *AuthenticateOptions) registerRoute(method, pattern string) {
	a.mu.Lock()
	defer a.mu.Unlock()

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
func (a *AuthenticateOptions) dummyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 空处理函数，仅用于路由匹配
}

// UseSkipRoutes 将指定的路由集合添加到不需要认证授权的列表中。
// 注意：如果路由已经在鉴权路由中，则会 panic。
func (a *AuthenticateOptions) useAllowAnonymous(routes ...Route) {
	for _, route := range routes {
		if route.Path == "" {
			panic("path is empty")
		}
		if len(route.Methods) == 0 {
			panic("methods is empty")
		}

		for _, m := range route.Methods {
			key := RouteKey{Method: string(m), Path: route.Path}

			// 检查是否已经在鉴权路由中
			if _, exists := a.routeSchemesMap[key]; exists {
				panic("route already exists in authentication routes: " + string(m) + " " + route.Path)
			}

			if _, ok := a.skipRoutesMap[key]; ok {
				panic("route already exists in skip routes: " + string(m) + " " + route.Path)
			}

			a.skipRoutesMap[key] = struct{}{}
			// 注册到 httprouter
			a.registerRoute(string(m), route.Path)
		}
	}
}

// UseRouteSchemes 将指定的路由授权方案添加到列表中。
// 注意：如果路由已经在跳过列表中，则会 panic。
//
// routeAuthenticationSchemes 路由和鉴权方案的关联列表
func (a *AuthenticateOptions) useRouteSchemes(routeAuthenticationSchemes ...RouteAuthenticationSchemes) {
	for _, ras := range routeAuthenticationSchemes {
		if len(ras.Routes) == 0 {
			panic("routes is empty")
		}
		if len(ras.Schemes) == 0 {
			panic("schemes is empty")
		}

		for _, route := range ras.Routes {
			if route.Path == "" {
				panic("path is empty")
			}
			if len(route.Methods) == 0 {
				panic("methods is empty:" + route.Path)
			}

			for _, m := range route.Methods {
				key := RouteKey{Method: string(m), Path: route.Path}

				// 检查是否已经在跳过路由中
				if _, exists := a.skipRoutesMap[key]; exists {
					panic("route is in skip list: " + string(m) + " " + route.Path)
				}

				// 合并 routeSchemesMap
				existing := a.routeSchemesMap[key]
				schemeSet := make(map[string]struct{})
				for _, s := range existing {
					schemeSet[s] = struct{}{}
				}
				for _, s := range ras.Schemes {
					if _, ok := schemeSet[s]; !ok {
						existing = append(existing, s)
						schemeSet[s] = struct{}{}
					}
				}
				a.routeSchemesMap[key] = existing

				// 合并 schemeRoutesMap
				for _, scheme := range ras.Schemes {
					routes := a.schemeRoutesMap[scheme]
					found := false
					for _, r := range routes {
						if r == key {
							found = true
							break
						}
					}
					if !found {
						a.schemeRoutesMap[scheme] = append(a.schemeRoutesMap[scheme], key)
					}
				}

				// 注册到 httprouter
				a.registerRoute(string(m), route.Path)
			}
		}
	}
}

// FindMatchingRoute 使用 httprouter 查找匹配的路由
//
// method  请求方法
// path  请求路径
//
// 返回 RouteKey 和 bool 值，bool 值为 true 表示找到了匹配的路由，否则为 false
func (a *AuthenticateOptions) findMatchingRoute(method, path string) (RouteKey, bool) {
	// 使用 httprouter 的 Lookup 方法查找匹配的路由
	handler, params, _ := a.router.Lookup(method, path)
	if handler == nil {
		return RouteKey{}, false
	}

	// 重构路由模式
	pattern := path
	for _, param := range params {
		pattern = replaceFirst(pattern, param.Value, ":"+param.Key)
	}

	return RouteKey{Method: method, Path: pattern}, true
}

// ShouldSkip 检查路径是否应该跳过鉴权
//
// method  请求方法
// path  请求路径
func (a *AuthenticateOptions) shouldSkip(method, path string) bool {
	// 使用 httprouter 查找匹配的路由
	routeKey, found := a.findMatchingRoute(method, path)
	if !found {
		return false
	}

	// 检查是否在跳过列表中
	_, shouldSkip := a.skipRoutesMap[routeKey]
	return shouldSkip
}

// GetSchemesForRequest 获取请求对应的鉴权方案。
// 如果没有找到匹配的路由，则返回默认的鉴权方案。
//
// method  请求方法
// path  请求路径
func (a *AuthenticateOptions) getSchemesForRequest(method, path string) []string {
	// 使用 httprouter 查找匹配的路由
	routeKey, found := a.findMatchingRoute(method, path)
	if !found {
		return []string{a.DefaultScheme}
	}

	// 获取对应的鉴权方案
	schemes, exists := a.routeSchemesMap[routeKey]

	if !exists || len(schemes) == 0 {
		return []string{a.DefaultScheme}
	}

	return schemes
}

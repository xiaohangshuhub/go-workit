package workit

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
)

// RouteAuthorizePolicies 表示路由和授权策略的关联
type RouteAuthorizePolicies struct {
	Routes   []Route  // 路由列表
	Policies []string // 授权策略列表
}

// AuthorizeOptions 表示授权选项配置
type AuthorizeOptions struct {
	DefaultPolicy    string
	routePoliciesMap map[RouteKey][]string // 路由 → 授权策略列表
	policyRoutesMap  map[string][]RouteKey // 授权策略 → 路由列表
	router           *httprouter.Router    // httprouter 实例
	patternMap       map[string]string     // 处理函数标识到模式字符串的映射
	mu               sync.Mutex            // 保护并发访问
}

// newAuthorizeOptions 创建一个新的 AuthorizeOptions 实例
func newAuthorizeOptions() *AuthorizeOptions {
	return &AuthorizeOptions{
		DefaultPolicy:    "",
		routePoliciesMap: make(map[RouteKey][]string),
		policyRoutesMap:  make(map[string][]RouteKey),
		router:           httprouter.New(),
		patternMap:       make(map[string]string),
	}
}

// registerRoute 注册路由到 httprouter（内部方法）
//
// method 请求方法
// pattern 请求路径
func (a *AuthorizeOptions) registerRoute(method, pattern string) {
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
func (a *AuthorizeOptions) dummyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 空处理函数，仅用于路由匹配
}

// UseRoutePolicies 将指定的路由授权策略添加到列表中
//
// routeAuthorizePolicies 包含多个 RouteAuthorizePolicies 实例
func (a *AuthorizeOptions) UseRoutePolicies(routeAuthorizePolicies ...RouteAuthorizePolicies) {
	for _, rap := range routeAuthorizePolicies {
		if len(rap.Routes) == 0 {
			panic("routes is empty")
		}
		if len(rap.Policies) == 0 {
			panic("policies is empty")
		}

		for _, route := range rap.Routes {
			if route.Path == "" {
				panic("path is empty")
			}
			if len(route.Methods) == 0 {
				panic("methods is empty:" + route.Path)
			}

			for _, m := range route.Methods {
				key := RouteKey{Method: string(m), Path: route.Path}

				// 合并 routePoliciesMap
				existing := a.routePoliciesMap[key]
				policySet := make(map[string]struct{})
				for _, p := range existing {
					policySet[p] = struct{}{}
				}
				for _, p := range rap.Policies {
					if _, ok := policySet[p]; !ok {
						existing = append(existing, p)
						policySet[p] = struct{}{}
					}
				}
				a.routePoliciesMap[key] = existing

				// 合并 policyRoutesMap
				for _, policy := range rap.Policies {
					routes := a.policyRoutesMap[policy]
					found := false
					for _, r := range routes {
						if r == key {
							found = true
							break
						}
					}
					if !found {
						a.policyRoutesMap[policy] = append(a.policyRoutesMap[policy], key)
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
// method 请求方法
// path 请求路径
//
// 返回 RouteKey 类型，包含方法和路径,以及是否找到匹配的路由
func (a *AuthorizeOptions) findMatchingRoute(method, path string) (RouteKey, bool) {
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

// GetPoliciesForRequest 获取请求对应的授权策略
// 如果没有找到匹配的路由，并默认不为空,则返回默认策略
// 否则,返回空列表
//
// 返回 []string 类型，可以包含多个策略
func (a *AuthorizeOptions) getPoliciesForRequest(method, path string) []string {
	// 使用 httprouter 查找匹配的路由
	routeKey, found := a.findMatchingRoute(method, path)

	// 如果没有找到匹配的路由
	if !found {
		if a.DefaultPolicy != "" {
			return []string{a.DefaultPolicy} // 默认不为空，返回默认策略
		}
		return []string{} // 默认为空，返回空列表
	}

	// 获取对应的授权策略
	policies, exists := a.routePoliciesMap[routeKey]

	// 如果路由存在但没有策略或策略为空
	if !exists || len(policies) == 0 {
		if a.DefaultPolicy != "" {
			return []string{a.DefaultPolicy} // 默认不为空，返回默认策略
		}
		return []string{} // 默认为空，返回空列表
	}

	// 返回找到的策略
	return policies
}

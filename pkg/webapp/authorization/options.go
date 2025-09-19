package authorization

import (
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
)

// Options 表示授权选项配置
type Options struct {
	DefaultPolicy    string
	routePoliciesMap map[router.RouteKey][]string // 路由 → 授权策略列表
	policyRoutesMap  map[string][]router.RouteKey // 授权策略 → 路由列表
	allowAnonymous   map[router.RouteKey]struct{} // 跳过路由 → 空结构（集合）

	*Builder
}

// NewOptions 创建一个新的 Options 实例
func NewOptions() *Options {

	opts := &Options{
		DefaultPolicy:    "",
		routePoliciesMap: make(map[router.RouteKey][]string),
		policyRoutesMap:  make(map[string][]router.RouteKey),
		allowAnonymous:   make(map[router.RouteKey]struct{}),
	}

	opts.Builder = NewBuilder(opts)

	return opts
}

// UseRoutePolicies 将指定的路由授权策略添加到列表中
//
// routeAuthorizePolicies 包含多个 RouteAuthorizePolicies 实例
func (a *Options) UseRoutePolicies(routeAuthorizePolicies ...RoutePolicies) {
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
				key := router.RouteKey{Method: string(m), Path: route.Path}

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
			}
		}
	}
}

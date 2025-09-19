package authentication

import (
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
)

// AuthenticationOptions 表示授权选项配置。
type Options struct {
	DefaultScheme   string
	routeSchemesMap map[router.RouteKey][]string // 鉴权路由 → 鉴权方案列表
	schemeRoutesMap map[string][]router.RouteKey // 鉴权方案 → 鉴权路由列表
	skipRoutesMap   map[router.RouteKey]struct{} // 跳过路由 → 空结构（集合）
	*Builder
}

// NewOptions 创建一个新的 Options 实例
func NewOptions() *Options {
	opt := &Options{
		routeSchemesMap: make(map[router.RouteKey][]string),
		schemeRoutesMap: make(map[string][]router.RouteKey),
		skipRoutesMap:   make(map[router.RouteKey]struct{}),
	}
	opt.Builder = NewBuilder(opt)
	return opt
}

// UseSkipRoutes 将指定的路由集合添加到不需要认证授权的列表中。
// 注意：如果路由已经在鉴权路由中，则会 panic。
func (a *Options) UseAllowAnonymous(routes ...router.Route) {
	for _, route := range routes {
		if route.Path == "" {
			panic("path is empty")
		}
		if len(route.Methods) == 0 {
			panic("methods is empty")
		}

		for _, m := range route.Methods {
			key := router.RouteKey{Method: string(m), Path: route.Path}

			// 检查是否已经在鉴权路由中
			if _, exists := a.routeSchemesMap[key]; exists {
				panic("route already exists in authentication routes: " + string(m) + " " + route.Path)
			}

			if _, ok := a.skipRoutesMap[key]; ok {
				panic("route already exists in skip routes: " + string(m) + " " + route.Path)
			}

			a.skipRoutesMap[key] = struct{}{}
		}
	}
}

// UseRouteSchemes 将指定的路由授权方案添加到列表中。
// 注意：如果路由已经在跳过列表中，则会 panic。
//
// routeAuthenticationSchemes 路由和鉴权方案的关联列表
func (a *Options) UseRouteSchemes(routeAuthenticationSchemes ...RouteSchemes) {
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
				key := router.RouteKey{Method: string(m), Path: route.Path}

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
			}
		}
	}
}

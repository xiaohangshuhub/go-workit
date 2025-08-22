package workit

import "github.com/gobwas/glob"

// RouteAuthenticationSchemes 表示路由级别的授权方案。
type RouteAuthenticationSchemes struct {
	Routes  []Route  // 路由列表
	Schemes []string // 对应的授权方案列表
}

// AuthenticateOptions 表示授权选项配置。
type AuthenticateOptions struct {
	DefaultScheme   string
	routeSchemesMap map[RouteKey][]string // (Method, Path) → 鉴权方案列表
	schemeRoutesMap map[string][]RouteKey // 鉴权方案 → (Method, Path) 列表
	skipRoutesMap   map[RouteKey]struct{} // (Method, Path) → 是否跳过授权
}

// newAuthenticateOptions 创建一个新的 AuthenticateOptions 实例。
//
// 返回初始化完成的实例，包括空的 SkipRoutes、默认方案为空、空的 RouteAuthenticationSchemes。
func newAuthenticateOptions() *AuthenticateOptions {
	return &AuthenticateOptions{
		routeSchemesMap: map[RouteKey][]string{},
		schemeRoutesMap: map[string][]RouteKey{},
		skipRoutesMap:   map[RouteKey]struct{}{},
	}
}

// UseSkipRoutes 将指定的路由集合添加到不需要认证授权的列表中。
//
// 每个路由必须包含非空 Path，并且至少包含一个 HTTP 方法。
// 如果路由路径重复或参数无效，会直接 panic。
//
// routes 要添加的路由列表
func (a *AuthenticateOptions) UseSkipRoutes(routes ...Route) {
	for _, route := range routes {
		if route.Path == "" {
			panic("path is empty")
		}
		if len(route.Methods) == 0 {
			panic("methods is empty")
		}

		for _, m := range route.Methods {
			key := RouteKey{Method: string(m), Path: route.Path, Glob: glob.MustCompile(route.Path)}

			if _, ok := a.skipRoutesMap[key]; ok {
				panic("route already exists: " + string(m) + " " + route.Path)
			}

			a.skipRoutesMap[key] = struct{}{}
		}
	}
}

// UseRouteAuthenticationSchemes 将指定的路由授权方案添加到列表中。
//
// 每个路由授权方案必须包含至少一个路由，并且至少包含一个授权方案。
// 如果路由路径重复或参数无效，会直接 panic。
//
// routeAuthenticationSchemes 要添加的路由授权方案列表
func (a *AuthenticateOptions) UseRouteAuthenticationSchemes(routeAuthenticationSchemes ...RouteAuthenticationSchemes) {
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

			g := glob.MustCompile(route.Path)

			for _, m := range route.Methods {
				key := RouteKey{Method: string(m), Path: route.Path, Glob: g}

				if _, ok := a.skipRoutesMap[key]; ok {
					panic("route is in skip list: " + string(m) + " " + route.Path)
				}

				// 合并 routeSchemesMap
				existing := a.routeSchemesMap[key]
				schemeSet := map[string]struct{}{}
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

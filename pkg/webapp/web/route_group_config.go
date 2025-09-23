package web

// GroupRouteConfig 分组路由配置
type GroupRouteConfig struct {
	Prefix         string
	Schemes        []string
	Policies       []string
	RateLimiter    []string
	Routes         []*RouteConfig
	AllowAnonymous bool
}

// MapGet 注册GET请求路由
func (group *GroupRouteConfig) MapGet(path string, handler any) *RouteConfig {
	return group.mapRoute(path, GET, handler)
}

// MapPost 注册POST请求路由
func (group *GroupRouteConfig) MapPost(path string, handler any) *RouteConfig {
	return group.mapRoute(path, POST, handler)
}

// MapPut 注册PUT请求路由
func (group *GroupRouteConfig) MapPut(path string, handler any) *RouteConfig {
	return group.mapRoute(path, PUT, handler)
}

// MapDelete 注册DELETE请求路由
func (group *GroupRouteConfig) MapDelete(path string, handler any) *RouteConfig {
	return group.mapRoute(path, DELETE, handler)
}

// MapPatch 注册PATCH请求路由
func (group *GroupRouteConfig) MapPatch(path string, handler any) *RouteConfig {
	return group.mapRoute(path, PATCH, handler)
}

// registerPattern 注册路由模式到 httprouter
func (group *GroupRouteConfig) WithAuthenticationScheme(schemes ...string) *GroupRouteConfig {
	group.Schemes = append(group.Schemes, schemes...)
	return group
}

// WithAuthorizationPolicy 配置授权策略
func (group *GroupRouteConfig) WithAuthorizationPolicy(policies ...string) *GroupRouteConfig {
	group.Policies = append(group.Policies, policies...)
	return group
}

// WithRateLimiter 配置限流器
func (group *GroupRouteConfig) WithRateLimiter(limiters ...string) *GroupRouteConfig {
	group.RateLimiter = append(group.RateLimiter, limiters...)
	return group
}

// WithAllowAnonymous 配置允许匿名访问
func (group *GroupRouteConfig) WithAllowAnonymous() *GroupRouteConfig {
	group.AllowAnonymous = true
	return group
}

// mapRoute 注册路由
func (group *GroupRouteConfig) mapRoute(path string, method RequestMethod, handler any) *RouteConfig {
	config := &RouteConfig{
		Path:        path,
		Method:      method,
		Handler:     handler,
		Schemes:     group.Schemes,
		Policies:    group.Policies,
		RateLimiter: group.RateLimiter,
	}
	group.Routes = append(group.Routes, config)
	return config
}

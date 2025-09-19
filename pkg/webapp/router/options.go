package router

type Options struct {
	routeConfigs []*RouteConfig
	groupConfigs []*GroupRouteConfig
	*Builder
}

func NewOptions() *Options {

	opts := &Options{
		routeConfigs: make([]*RouteConfig, 0),
		groupConfigs: make([]*GroupRouteConfig, 0),
	}

	opts.Builder = NewBuilder(opts)

	return opts
}

// MapGet 注册GET请求路由
func (opts *Options) MapGet(path string, handler any) *RouteConfig {
	return opts.mapRoute(path, GET, handler)
}

// MapPost 注册POST请求路由
func (opts *Options) MapPost(path string, handler any) *RouteConfig {
	return opts.mapRoute(path, POST, handler)
}

// MapPut 注册PUT请求路由
func (opts *Options) MapPut(path string, handler any) *RouteConfig {
	return opts.mapRoute(path, PUT, handler)
}

// MapDelete 注册DELETE请求路由
func (opts *Options) MapDelete(path string, handler any) *RouteConfig {
	return opts.mapRoute(path, DELETE, handler)
}

// MapPatch 注册PATCH请求路由
func (opts *Options) MapPatch(path string, handler any) *RouteConfig {
	return opts.mapRoute(path, PATCH, handler)
}

// MapGroup 注册分组路由
func (opts *Options) MapGroup(prefix string) *GroupRouteConfig {
	group := &GroupRouteConfig{
		Prefix: prefix,
		Routes: make([]*RouteConfig, 0),
	}
	opts.groupConfigs = append(opts.groupConfigs, group)
	return group
}

// mapRoute 注册路由
func (opts *Options) mapRoute(path string, method RequestMethod, handler any) *RouteConfig {
	config := &RouteConfig{
		Path:    path,
		Method:  method,
		Handler: handler,
	}
	opts.routeConfigs = append(opts.routeConfigs, config)
	return config
}

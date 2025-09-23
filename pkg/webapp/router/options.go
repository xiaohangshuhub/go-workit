package router

import "github.com/xiaohangshuhub/go-workit/pkg/webapp/web"

type Options struct {
	routeConfigs []*web.RouteConfig
	groupConfigs []*web.GroupRouteConfig
}

func NewOptions() *Options {
	opts := &Options{
		routeConfigs: make([]*web.RouteConfig, 0),
		groupConfigs: make([]*web.GroupRouteConfig, 0),
	}
	return opts
}

// MapGet 注册GET请求路由
func (opts *Options) MapGet(path string, handler any) *web.RouteConfig {
	return opts.mapRoute(path, web.GET, handler)
}

// MapPost 注册POST请求路由
func (opts *Options) MapPost(path string, handler any) *web.RouteConfig {
	return opts.mapRoute(path, web.POST, handler)
}

// MapPut 注册PUT请求路由
func (opts *Options) MapPut(path string, handler any) *web.RouteConfig {
	return opts.mapRoute(path, web.PUT, handler)
}

// MapDelete 注册DELETE请求路由
func (opts *Options) MapDelete(path string, handler any) *web.RouteConfig {
	return opts.mapRoute(path, web.DELETE, handler)
}

// MapPatch 注册PATCH请求路由
func (opts *Options) MapPatch(path string, handler any) *web.RouteConfig {
	return opts.mapRoute(path, web.PATCH, handler)
}

// MapGroup 注册分组路由
func (opts *Options) MapGroup(prefix string) *web.GroupRouteConfig {
	group := &web.GroupRouteConfig{
		Prefix: prefix,
		Routes: make([]*web.RouteConfig, 0),
	}
	opts.groupConfigs = append(opts.groupConfigs, group)
	return group
}

// mapRoute 注册路由
func (opts *Options) mapRoute(path string, method web.RequestMethod, handler any) *web.RouteConfig {
	config := &web.RouteConfig{
		Path:    path,
		Method:  method,
		Handler: handler,
	}
	opts.routeConfigs = append(opts.routeConfigs, config)
	return config
}

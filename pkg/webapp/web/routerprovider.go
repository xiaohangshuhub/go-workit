package web

import "github.com/xiaohangshuhub/go-workit/pkg/webapp/router"

type RouterProvider interface {
	RouteConfig() []*router.RouteConfig
	GroupRouteConfig() []*router.GroupRouteConfig
}

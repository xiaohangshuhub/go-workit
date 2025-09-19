package authentication

import "github.com/xiaohangshuhub/go-workit/pkg/webapp/router"

// RouteAuthenticationSchemes 表示路由和认证方案的关联
type RouteSchemes struct {
	Routes  []router.Route // 鉴权路由
	Schemes []string       // 鉴权方案列表
}

package authorization

import "github.com/xiaohangshuhub/go-workit/pkg/webapp/router"

// RoutePolicies 表示路由和授权策略的关联
type RoutePolicies struct {
	Routes   []router.Route // 路由列表
	Policies []string       // 授权策略列表
}

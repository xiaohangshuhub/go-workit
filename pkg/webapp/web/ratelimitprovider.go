package web

import "github.com/xiaohangshuhub/go-workit/pkg/webapp/router"

type RatelimitProvider interface {
	DefaultPolicy() string                                           // 默认策略
	RoutePolicies(method router.RequestMethod, path string) []string // 限流策略
	Handler(policy string) (RateLimitHandler, bool)                  // 限流处理器
}

package config

import "github.com/xiaohangshuhub/go-workit/pkg/workit"

// 授权配置
var Routecfg = []workit.RouteConfig{
	{
		Routes: []workit.Route{
			{Path: "/hello", Methods: []workit.RequestMethod{workit.GET}},
			{Path: "/api/v1/user", Methods: workit.ANY},
		},
		Policies: []string{"admin_role_policy"},
	},
	{
		Routes: []workit.Route{
			{Path: "/api/v1/login", Methods: []workit.RequestMethod{workit.POST}},
		},
		AllowAnonymous: true,
	},
}

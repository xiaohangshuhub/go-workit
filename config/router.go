package config

import "github.com/xiaohangshuhub/go-workit/pkg/webapp"

// 授权配置
var RouteSecurityCfg = []webapp.RouteSecurityConfig{
	{
		Routes: []webapp.Route{
			{Path: "/hello", Methods: []webapp.RequestMethod{webapp.GET}},
			{Path: "/api/v1/user", Methods: webapp.ANY},
		},
		Policies: []string{"admin_role_policy"},
	},
	{
		Routes: []webapp.Route{
			{Path: "/api/v1/login", Methods: []webapp.RequestMethod{webapp.POST}},
		},
		AllowAnonymous: true,
	},
}

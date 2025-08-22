package config

import "github.com/xiaohangshuhub/go-workit/pkg/workit"

// 授权配置
var RouteAuthenticationSchemes = []workit.RouteAuthenticationSchemes{
	{
		Routes: []workit.Route{
			{Path: "/user", Methods: []workit.RequestMethod{workit.POST, workit.DELETE, workit.PUT}},
		},
		Schemes: []string{"local_jwt_bearer"},
	},
	{
		Routes: []workit.Route{
			{Path: "/hello", Methods: workit.ANY},
		},
		Schemes: []string{"oauth2_jwt_bearer"},
	},
}

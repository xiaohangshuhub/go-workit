package config

import "github.com/xiaohangshuhub/go-workit/pkg/workit"

// 授权配置
var Authorize = []workit.AuthorizeOptions{
	{
		Routes: []workit.Route{
			{Path: "/user", Methods: []workit.RequestMethod{workit.POST, workit.DELETE, workit.PUT}},
			{Path: "/auth", Methods: []workit.RequestMethod{workit.POST, workit.DELETE, workit.PUT}},
		},
		Policies: []string{"admin_role_policy"},
	},
}

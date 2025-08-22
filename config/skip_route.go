package config

import "github.com/xiaohangshuhub/go-workit/pkg/workit"

//
var SkipRoutes = []workit.Route{
	{Path: "/health", Methods: []workit.RequestMethod{workit.GET}},
	{Path: "/api/v1/users/login", Methods: []workit.RequestMethod{workit.POST}},
}

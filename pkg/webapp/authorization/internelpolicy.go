package authorization

import "github.com/xiaohangshuhub/go-workit/pkg/webapp/web"

// requireRole 声明角色策略
var requireRole = func(role ...string) func(claims *web.ClaimsPrincipal) bool {

	return func(claims *web.ClaimsPrincipal) bool {
		for _, r := range role {
			if !claims.IsInRole(r) {
				return false
			}
		}
		return true
	}
}

// requireClaim 声明 key value 在Claims中的策略
var requireClaim = func(k string, v any) func(principal *web.ClaimsPrincipal) bool {

	return func(principal *web.ClaimsPrincipal) bool {
		for _, c := range principal.Claims {
			if c.Type == k && c.Value == v {
				return true
			}
		}
		return false
	}
}

// requireHasChaims 声明 key 在Claims中的策略
var requireHasChaims = func(k string) func(claims *web.ClaimsPrincipal) bool {

	return func(claims *web.ClaimsPrincipal) bool {
		for _, c := range claims.Claims {
			if c.Type == k {
				return true
			}
		}
		return false
	}
}

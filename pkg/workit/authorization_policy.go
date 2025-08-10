package workit

// 角色策略
var inRolePolicy = func(role string) func(claims *ClaimsPrincipal) bool {

	return func(claims *ClaimsPrincipal) bool {
		return claims.IsInRole(role)
	}
}

// 声明 key value 在Claims中的策略
var inClaimsPolicy = func(k string, v interface{}) func(principal *ClaimsPrincipal) bool {

	return func(principal *ClaimsPrincipal) bool {
		for _, c := range principal.Claims {
			if c.Type == k && c.Value == v {
				return true
			}
		}
		return false
	}
}

// 声明 key 在Claims中的策略
var hasChaimsPolicy = func(k string) func(claims *ClaimsPrincipal) bool {

	return func(claims *ClaimsPrincipal) bool {
		for _, c := range claims.Claims {
			if c.Type == k {
				return true
			}
		}
		return false
	}
}

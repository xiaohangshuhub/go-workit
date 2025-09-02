package workit

//  requireRole 声明角色策略
var requireRole = func(role ...string) func(claims *ClaimsPrincipal) bool {

	return func(claims *ClaimsPrincipal) bool {
		for _, r := range role {
			if !claims.IsInRole(r) {
				return false
			}
		}
		return true
	}
}

//  requireClaim 声明 key value 在Claims中的策略
var requireClaim = func(k string, v interface{}) func(principal *ClaimsPrincipal) bool {

	return func(principal *ClaimsPrincipal) bool {
		for _, c := range principal.Claims {
			if c.Type == k && c.Value == v {
				return true
			}
		}
		return false
	}
}

// requireHasChaims 声明 key 在Claims中的策略
var requireHasChaims = func(k string) func(claims *ClaimsPrincipal) bool {

	return func(claims *ClaimsPrincipal) bool {
		for _, c := range claims.Claims {
			if c.Type == k {
				return true
			}
		}
		return false
	}
}

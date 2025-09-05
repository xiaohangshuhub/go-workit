package workit

// AuthorizationBuilder 用于构建授权策略
type AuthorizationBuilder struct {
	// 策略名称 -> 策略函数
	policys map[string]func(claims *ClaimsPrincipal) bool
}

// NewAuthorizationBuilder 创建一个新的授权构建器
func newAuthorizationBuilder() *AuthorizationBuilder {
	b := &AuthorizationBuilder{
		policys: make(map[string]func(claims *ClaimsPrincipal) bool),
	}

	return b
}

// AddPolicy 添加一个策略
func (ab *AuthorizationBuilder) AddPolicy(policyName string, policy func(claims *ClaimsPrincipal) bool) *AuthorizationBuilder {

	// 检查是否已存在同名策略
	if _, exists := ab.policys[policyName]; exists {
		panic("policy with name " + policyName + " already exists")
	}

	ab.policys[policyName] = policy
	return ab
}

// Policies 根据名称获取策略, 如果没有指定名称则返回全部策略
func (ab *AuthorizationBuilder) Policies(policyName ...string) map[string]func(claims *ClaimsPrincipal) bool {

	if len(policyName) == 0 {
		return ab.policys
	}

	policies := make(map[string]func(claims *ClaimsPrincipal) bool)
	for _, n := range policyName {
		if policy, exists := ab.policys[n]; exists {
			policies[n] = policy
		} else {
			panic("policy with name " + n + " does not exist")
		}
	}
	return policies

}

// Policy 根据名称返回单个策略
func (ab *AuthorizationBuilder) Policy(policyName string) func(claims *ClaimsPrincipal) bool {
	if policy, exists := ab.policys[policyName]; exists {
		return policy
	}
	panic("policy with name " + policyName + " does not exist")
}

// RequireRolePolicy 添加一个要求指定角色的策略
func (ab *AuthorizationBuilder) RequireRole(policyName string, role ...string) *AuthorizationBuilder {

	ab.AddPolicy(policyName, requireRole(role...))

	return ab
}

// RequireClaimPolicy 添加一个要求指定 claim 的策略
func (ab *AuthorizationBuilder) RequireClaim(policyName, k string, v interface{}) *AuthorizationBuilder {

	ab.AddPolicy(policyName, requireClaim(k, v))

	return ab
}

// RequireHasChaimsPolicy 添加一个要求指定 claim 的策略
func (ab *AuthorizationBuilder) RequireHasChaims(policyName, k string) *AuthorizationBuilder {

	ab.AddPolicy(policyName, requireHasChaims(k))

	return ab
}

// Build 构建授权提供者
func (ab *AuthorizationBuilder) Build() *AuthorizationProvider {
	return newAuthorizationProvider(ab.policys)
}

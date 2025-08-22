package workit

type AuthorizationBuilder struct {
	// 策略名称 -> 策略函数
	policys map[string]func(claims *ClaimsPrincipal) bool
}

func newAuthorizationBuilder() *AuthorizationBuilder {
	b := &AuthorizationBuilder{
		policys: make(map[string]func(claims *ClaimsPrincipal) bool),
	}

	return b
}

func (ab *AuthorizationBuilder) AddPolicy(policyName string, policy func(claims *ClaimsPrincipal) bool) *AuthorizationBuilder {

	// 检查是否已存在同名策略
	if _, exists := ab.policys[policyName]; exists {
		panic("policy with name " + policyName + " already exists")
	}

	ab.policys[policyName] = policy
	return ab
}

// 根据名称获取策略, 如果没有指定名称则返回全部策略
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

// 根据名称返回单个策略
func (ab *AuthorizationBuilder) Policy(policyName string) func(claims *ClaimsPrincipal) bool {
	if policy, exists := ab.policys[policyName]; exists {
		return policy
	}
	panic("policy with name " + policyName + " does not exist")
}

func (ab *AuthorizationBuilder) RequireRolePolicy(policyName string, role ...string) *AuthorizationBuilder {

	ab.AddPolicy(policyName, requireRole(role...))

	return ab
}

func (ab *AuthorizationBuilder) RequireClaimPolicy(policyName, k string, v interface{}) *AuthorizationBuilder {

	ab.AddPolicy(policyName, requireClaim(k, v))

	return ab
}

func (ab *AuthorizationBuilder) RequireHasChaimsPolicy(policyName, k string) *AuthorizationBuilder {

	ab.AddPolicy(policyName, requireHasChaims(k))

	return ab
}

func (ab *AuthorizationBuilder) Build() *AuthorizationProvider {
	return newAuthorizationProvider(ab.policys)
}

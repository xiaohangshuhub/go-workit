package authorization

import "github.com/xiaohangshuhub/go-workit/pkg/webapp/web"

// AuthorizationBuilder 用于构建授权策略
type Builder struct {
	// 策略名称 -> 策略函数
	policys map[string]func(claims *web.ClaimsPrincipal) bool
	*Options
}

// NewAuthorizationBuilder 创建一个新的授权构建器
func NewBuilder(options *Options) *Builder {
	b := &Builder{
		policys: make(map[string]func(claims *web.ClaimsPrincipal) bool),
		Options: options,
	}

	return b
}

// AddPolicy 添加一个策略
func (a *Builder) AddPolicy(policyName string, policy func(claims *web.ClaimsPrincipal) bool) *Builder {

	// 检查是否已存在同名策略
	if _, exists := a.policys[policyName]; exists {
		panic("policy with name " + policyName + " already exists")
	}

	a.policys[policyName] = policy
	return a
}

// Policies 根据名称获取策略, 如果没有指定名称则返回全部策略
func (a *Builder) Policies(policyName ...string) map[string]func(claims *web.ClaimsPrincipal) bool {

	if len(policyName) == 0 {
		return a.policys
	}

	policies := make(map[string]func(claims *web.ClaimsPrincipal) bool)
	for _, n := range policyName {
		if policy, exists := a.policys[n]; exists {
			policies[n] = policy
		} else {
			panic("policy with name " + n + " does not exist")
		}
	}
	return policies

}

// Policy 根据名称返回单个策略
func (a *Builder) Policy(policyName string) func(claims *web.ClaimsPrincipal) bool {
	if policy, exists := a.policys[policyName]; exists {
		return policy
	}
	panic("policy with name " + policyName + " does not exist")
}

// RequireRolePolicy 添加一个要求指定角色的策略
func (a *Builder) RequireRole(policyName string, role ...string) *Builder {

	a.AddPolicy(policyName, requireRole(role...))

	return a
}

// RequireClaimPolicy 添加一个要求指定 claim 的策略
func (a *Builder) RequireClaim(policyName, k string, v interface{}) *Builder {

	a.AddPolicy(policyName, requireClaim(k, v))

	return a
}

// RequireHasChaimsPolicy 添加一个要求指定 claim 的策略
func (a *Builder) RequireHasChaims(policyName, k string) *Builder {

	a.AddPolicy(policyName, requireHasChaims(k))

	return a
}

// Build 构建授权提供者
func (a *Builder) Build() (*Provider, error) {
	return newAuthorizationProvider(a.DefaultPolicy, a.routePoliciesMap, a.allowAnonymous, a.policys), nil
}

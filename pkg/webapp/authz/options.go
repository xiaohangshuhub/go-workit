package authz

import (
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

// Options 表示授权选项配置
type Options struct {
	DefaultPolicy string
	policys       map[string]func(claims *web.ClaimsPrincipal) bool
}

// NewOptions 创建一个新的 Options 实例
func NewOptions() *Options {

	opts := &Options{
		DefaultPolicy: "",
		policys:       make(map[string]func(claims *web.ClaimsPrincipal) bool),
	}

	return opts
}

// AddPolicy 添加一个策略
func (o *Options) AddPolicy(policyName string, policy func(claims *web.ClaimsPrincipal) bool) *Options {

	// 检查是否已存在同名策略
	if _, exists := o.policys[policyName]; exists {
		panic("policy with name " + policyName + " already exists")
	}

	o.policys[policyName] = policy
	return o
}

// Policies 根据名称获取策略, 如果没有指定名称则返回全部策略
func (o *Options) Policies(policyName ...string) map[string]func(claims *web.ClaimsPrincipal) bool {

	if len(policyName) == 0 {
		return o.policys
	}

	policies := make(map[string]func(claims *web.ClaimsPrincipal) bool)
	for _, n := range policyName {
		if policy, exists := o.policys[n]; exists {
			policies[n] = policy
		} else {
			panic("policy with name " + n + " does not exist")
		}
	}
	return policies

}

// Policy 根据名称返回单个策略
func (o *Options) Policy(policyName string) func(claims *web.ClaimsPrincipal) bool {
	if policy, exists := o.policys[policyName]; exists {
		return policy
	}
	panic("policy with name " + policyName + " does not exist")
}

// RequireRolePolicy 添加一个要求指定角色的策略
func (o *Options) RequireRole(policyName string, role ...string) *Options {

	o.AddPolicy(policyName, requireRole(role...))

	return o
}

// RequireClaimPolicy 添加一个要求指定 claim 的策略
func (o *Options) RequireClaim(policyName, k string, v interface{}) *Options {

	o.AddPolicy(policyName, requireClaim(k, v))

	return o
}

// RequireHasChaimsPolicy 添加一个要求指定 claim 的策略
func (o *Options) RequireHasChaims(policyName, k string) *Options {

	o.AddPolicy(policyName, requireHasChaims(k))

	return o
}

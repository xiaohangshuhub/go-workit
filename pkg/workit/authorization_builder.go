package workit

type AuthorizationBuilder struct {
	// 策略名称 -> 策略函数
	policys map[string]func(claims *ClaimsPrincipal) bool

	// 配置策略映射关系
	authorize map[string][]string
}

func newAuthorizationBuilder(authorize ...AuthorizeOptions) *AuthorizationBuilder {
	b := &AuthorizationBuilder{
		policys:   make(map[string]func(claims *ClaimsPrincipal) bool),
		authorize: make(map[string][]string),
	}

	for _, mapping := range authorize {
		for _, route := range mapping.Routes {
			// 用 path + method 作为唯一 key
			for _, method := range route.Methods {
				key := string(method) + ":" + route.Path

				// 如果不存在，初始化 slice
				if _, exists := b.authorize[key]; !exists {
					b.authorize[key] = []string{}
				}

				// 添加策略（去重）
				for _, policy := range mapping.Policies {
					if !contains(b.authorize[key], policy) {
						b.authorize[key] = append(b.authorize[key], policy)
					}
				}
			}
		}
	}

	return b
}

// 工具函数：判断 slice 是否包含某元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (ab *AuthorizationBuilder) AddPolicy(name string, policy func(claims *ClaimsPrincipal) bool) *AuthorizationBuilder {

	// 检查是否已存在同名策略
	if _, exists := ab.policys[name]; exists {
		panic("policy with name " + name + " already exists")
	}

	ab.policys[name] = policy
	return ab
}

// 根据名称获取策略, 如果没有指定名称则返回全部策略
func (ab *AuthorizationBuilder) Policies(name ...string) map[string]func(claims *ClaimsPrincipal) bool {

	if len(name) == 0 {
		return ab.policys
	}

	policies := make(map[string]func(claims *ClaimsPrincipal) bool)
	for _, n := range name {
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

func (ab *AuthorizationBuilder) RequireRole(policyName string, role ...string) *AuthorizationBuilder {

	ab.AddPolicy(policyName, requireRole(role...))

	return ab
}

func (ab *AuthorizationBuilder) RequireClaim(policyName, k string, v interface{}) *AuthorizationBuilder {

	ab.AddPolicy(policyName, requireClaim(k, v))

	return ab
}

func (ab *AuthorizationBuilder) RequireHasChaims(policyName, k string) *AuthorizationBuilder {

	ab.AddPolicy(policyName, requireHasChaims(k))

	return ab
}

func (ab *AuthorizationBuilder) Build() *AuthorizationProvider {
	return newAuthorizationProvider(ab.policys, ab.authorize)
}

package workit

type AuthorizationBuilder struct {
	policys map[string]func(claims *ClaimsPrincipal) bool
}

func newAuthorizationBuilder() *AuthorizationBuilder {
	return &AuthorizationBuilder{
		policys: make(map[string]func(claims *ClaimsPrincipal) bool),
	}
}

func (ab *AuthorizationBuilder) AddPolicy(policy func(claims *ClaimsPrincipal) bool, path ...string) {

	for _, p := range path {
		ab.policys[p] = policy
	}
}

func (ab *AuthorizationBuilder) Policies() map[string]func(claims *ClaimsPrincipal) bool {
	return ab.policys
}

func (ab *AuthorizationBuilder) Build() *AuthorizationProvider {
	return newAuthorizationProvider(ab.policys)
}

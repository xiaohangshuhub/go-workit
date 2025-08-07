package workit

type AuthorizationProvider struct {
	policies map[string]func(claims *ClaimsPrincipal) bool
}

func newAuthorizationProvider(policies map[string]func(claims *ClaimsPrincipal) bool) *AuthorizationProvider {
	return &AuthorizationProvider{
		policies: policies,
	}
}

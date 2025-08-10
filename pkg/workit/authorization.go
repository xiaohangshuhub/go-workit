package workit

type AuthorizationProvider struct {
	policies map[string]func(claims *ClaimsPrincipal) bool

	authorize map[string][]string
}

func newAuthorizationProvider(policies map[string]func(claims *ClaimsPrincipal) bool, authorize map[string][]string) *AuthorizationProvider {
	return &AuthorizationProvider{
		policies:  policies,
		authorize: authorize,
	}
}

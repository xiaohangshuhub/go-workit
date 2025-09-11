package webapp

// AuthorizationProvider is the interface for authorization providers.
type AuthorizationProvider struct {
	policies map[string]func(claims *ClaimsPrincipal) bool
}

// NewAuthorizationProvider creates a new AuthorizationProvider with the given policies.
func newAuthorizationProvider(policies map[string]func(claims *ClaimsPrincipal) bool) *AuthorizationProvider {
	return &AuthorizationProvider{
		policies: policies,
	}
}

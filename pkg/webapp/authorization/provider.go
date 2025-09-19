package authorization

import "github.com/xiaohangshuhub/go-workit/pkg/webapp/web"

// AuthorizationProvider is the interface for authorization providers.
type Provider struct {
	policies map[string]func(claims *web.ClaimsPrincipal) bool
}

// NewAuthorizationProvider creates a new AuthorizationProvider with the given policies.
func newAuthorizationProvider(policies map[string]func(claims *web.ClaimsPrincipal) bool) *Provider {
	return &Provider{
		policies: policies,
	}
}

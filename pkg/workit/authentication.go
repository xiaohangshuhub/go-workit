package workit

import "net/http"

// AuthenticationHandler is an interface for handling authentication.
type AuthenticationHandler interface {
	Scheme() string
	Authenticate(r *http.Request) (*ClaimsPrincipal, error)
}

// 鉴权提供者
type AuthenticateProvider struct {
	handlers map[string]AuthenticationHandler
}

func newAuthenticateApplication(handlers map[string]AuthenticationHandler) *AuthenticateProvider {

	return &AuthenticateProvider{
		handlers: handlers,
	}
}

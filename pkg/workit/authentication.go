package workit

import "net/http"

// AuthenticationHandler is an interface for handling authentication.
type AuthenticationHandler interface {
	Scheme() string
	Authenticate(r *http.Request) (*ClaimsPrincipal, error)
}

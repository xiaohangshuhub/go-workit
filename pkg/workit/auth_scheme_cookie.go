package workit

import (
	"errors"
	"net/http"
)

type CookieSchem struct {
	CookieName string
}

func (h *CookieSchem) Scheme() string {
	return "Cookie"
}

func (h *CookieSchem) Authenticate(r *http.Request) (*ClaimsPrincipal, error) {
	cookie, err := r.Cookie(h.CookieName)
	if err != nil {
		return nil, errors.New("cookie not found")
	}

	// 这里假设 cookie 就是用户名
	return &ClaimsPrincipal{Name: cookie.Value, Roles: []string{"user"}}, nil
}

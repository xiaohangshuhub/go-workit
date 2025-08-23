package workit

import (
	"errors"
	"net/http"
)

type CookieSchem struct {
	options *CookieOptions
}

func newCookieHandler(options *CookieOptions) *CookieSchem {
	return &CookieSchem{
		options: options,
	}
}

func (h *CookieSchem) Scheme() string {
	return "Cookie"
}

func (h *CookieSchem) Authenticate(r *http.Request) (*ClaimsPrincipal, error) {
	cookie, err := r.Cookie(h.options.Name)
	if err != nil {
		return nil, errors.New("cookie not found")
	}

	// 这里假设 cookie 就是用户名
	return &ClaimsPrincipal{Name: cookie.Value, Roles: []string{"user"}}, nil
}

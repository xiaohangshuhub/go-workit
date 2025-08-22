package workit

type AuthorizeOptions struct {
	Routes   []Route
	Policies []string
}

func newAuthorizeOptions() *AuthorizeOptions {
	return &AuthorizeOptions{
		Routes:   make([]Route, 0),
		Policies: make([]string, 0),
	}
}

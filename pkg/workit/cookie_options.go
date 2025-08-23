package workit

type CookieOptions struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

func newCookieOptions() *CookieOptions {
	return &CookieOptions{}
}

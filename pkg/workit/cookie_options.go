package workit

type CookieOptions struct {
	Name              string
	Path              string
	Domain            string
	MaxAge            int
	Secure            bool
	HttpOnly          bool
	DataProtectionKey string // 数据保护密钥
}

func newCookieOptions() *CookieOptions {
	return &CookieOptions{}
}

package cookie

// CookieOptions 定义cookie的相关选项
type Options struct {
	Name              string
	Path              string
	Domain            string
	MaxAge            int
	Secure            bool
	HttpOnly          bool
	DataProtectionKey string // 数据保护密钥
}

// NewCookieOptions 创建一个新的CookieOptions实例
func NewCookieOptions() *Options {
	return &Options{}
}

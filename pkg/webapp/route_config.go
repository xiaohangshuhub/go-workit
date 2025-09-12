package webapp

// 路由配置
type RouteConfig struct {
	Path           string
	Method         RequestMethod
	Handler        any
	Schemes        []string
	Policies       []string
	RateLimiter    []string
	AllowAnonymous bool
}

// WithAuthenticationScheme 配置认证方案
func (config *RouteConfig) WithAuthenticationScheme(schemes ...string) *RouteConfig {
	config.Schemes = append(config.Schemes, schemes...)
	return config
}

// WithAuthorizationPolicy 配置授权策略
func (config *RouteConfig) WithAuthorizationPolicy(policies ...string) *RouteConfig {
	config.Policies = append(config.Policies, policies...)
	return config
}

// WithRateLimiter 配置限流器
func (config *RouteConfig) WithRateLimiter(limiters ...string) *RouteConfig {
	config.RateLimiter = append(config.RateLimiter, limiters...)
	return config
}

// WithAllowAnonymous 配置允许匿名访问
func (config *RouteConfig) WithAllowAnonymous() *RouteConfig {
	config.AllowAnonymous = true
	return config
}

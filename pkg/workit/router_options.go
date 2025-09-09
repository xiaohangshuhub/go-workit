package workit

// RouterConfigOptions 路由配置
type RouterOptions struct {
	authopts   *AuthenticationOptions
	authoropts *AuthorizationOptions
}

// newRouterConfigOptions 创建一个新的 RouterCofnigOptions 实例
func newRouterOptions(auth *AuthenticationOptions, author *AuthorizationOptions) *RouterOptions {
	return &RouterOptions{
		authopts:   auth,
		authoropts: author,
	}
}

// UseRouteSecurity 配置路由安全。包括鉴权方案和授权策略及匿名访问。
func (r *RouterOptions) UseRouteSecurity(cfg ...RouteSecurityConfig) {

	var schemes []RouteAuthenticationSchemes
	var policies []RouteAuthorizePolicies
	var allowAnonymous []Route

	// 遍历
	for _, cfg := range cfg {

		if len(cfg.Routes) == 0 {
			panic("Routes is empty")
		}

		if len(cfg.Schemes) != 0 {

			schemes = append(schemes, RouteAuthenticationSchemes{Routes: cfg.Routes, Schemes: cfg.Schemes})
		}

		if len(cfg.Policies) != 0 {
			policies = append(policies, RouteAuthorizePolicies{Routes: cfg.Routes, Policies: cfg.Policies})
		}

		if cfg.AllowAnonymous {
			allowAnonymous = append(allowAnonymous, cfg.Routes...)
		}
	}

	r.authopts.useRouteSchemes(schemes...)
	r.authopts.useAllowAnonymous(allowAnonymous...)
	r.authoropts.useRoutePolicies(policies...)
}

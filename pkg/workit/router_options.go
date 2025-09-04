package workit

// RouterConfigOptions 路由配置
type RouterOptions struct {
	auth   *AuthenticateOptions
	author *AuthorizeOptions
}

// newRouterConfigOptions 创建一个新的 RouterCofnigOptions 实例
func newRouterOptions() *RouterOptions {

	authopt := newAuthenticateOptions()
	authoropt := newAuthorizeOptions()

	return &RouterOptions{
		auth:   authopt,
		author: authoropt,
	}

}

func (r *RouterOptions) UseSettings(cfgs ...RouteConfigOptions) {
	r.auth.useRouteSchemes()
	r.auth.useAllowAnonymous()
	r.author.useRoutePolicies()
}

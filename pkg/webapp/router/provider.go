package router

type Provider struct {
	routeConfigs []*RouteConfig
	groupConfigs []*GroupRouteConfig
}

func NewProvider() *Provider {

	return &Provider{}
}

func (p *Provider) RouteConfig() []*RouteConfig {

	return p.routeConfigs

}
func (p *Provider) GroupRouteConfig() []*GroupRouteConfig {
	return p.GroupRouteConfig()
}

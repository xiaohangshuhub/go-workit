package workit

type RouteConfig struct {
	Routes         []Route
	Schemes        []string
	Policies       []string
	AllowAnonymous bool
}

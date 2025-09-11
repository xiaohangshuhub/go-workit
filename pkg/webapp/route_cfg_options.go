package webapp

type RouteSecurityConfig struct {
	Routes         []Route
	Schemes        []string
	Policies       []string
	AllowAnonymous bool
}

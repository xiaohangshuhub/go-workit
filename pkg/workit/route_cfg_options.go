package workit

type RouteConfigOptions struct {
	Routes         []Route
	Schemes        []string
	Policies       []string
	AllowAnonymous bool
}

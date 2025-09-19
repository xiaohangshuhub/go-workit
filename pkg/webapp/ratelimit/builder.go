package ratelimit

type Builder struct {
	*Options
}

func NewBuilder(options *Options) *Builder {
	return &Builder{
		Options: options,
	}
}

func (b *Builder) Build() (*Provider, error) {
	return NewProvider(b.DefaultPolicy, b.routeRateLimitMap, b.policies), nil
}

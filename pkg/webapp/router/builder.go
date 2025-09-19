package router

type Builder struct {
	*Options
}

func NewBuilder(options *Options) *Builder {
	return &Builder{
		Options: options,
	}
}

func (b *Builder) build() *Provider {
	return NewProvider()
}

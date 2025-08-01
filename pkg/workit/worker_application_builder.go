package workit

type WorkerApplicationBuilder struct {
	*ApplicationBuilder
	startFunc func() error
	stopFunc  func() error
}

func NewWorkerAppBuilder() *WorkerApplicationBuilder {
	return &WorkerApplicationBuilder{
		ApplicationBuilder: NewAppBuilder(),
	}
}

func (b *WorkerApplicationBuilder) OnStart(fn func() error) *WorkerApplicationBuilder {
	b.startFunc = fn
	return b
}

func (b *WorkerApplicationBuilder) OnStop(fn func() error) *WorkerApplicationBuilder {
	b.stopFunc = fn
	return b
}

func (b *WorkerApplicationBuilder) Build() (*WorkerApplication, error) {
	host, err := b.ApplicationBuilder.Build()
	if err != nil {
		return nil, err
	}
	return newWorkerApplication(host, b.startFunc, b.stopFunc), nil
}

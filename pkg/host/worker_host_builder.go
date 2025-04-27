package host

type WorkerHostBuilder struct {
	*ApplicationHostBuilder
	startFunc func() error
	stopFunc  func() error
}

func NewWorkerHostBuilder() *WorkerHostBuilder {
	return &WorkerHostBuilder{
		ApplicationHostBuilder: NewApplicationHostBuilder(),
	}
}

func (b *WorkerHostBuilder) OnStart(fn func() error) *WorkerHostBuilder {
	b.startFunc = fn
	return b
}

func (b *WorkerHostBuilder) OnStop(fn func() error) *WorkerHostBuilder {
	b.stopFunc = fn
	return b
}

func (b *WorkerHostBuilder) Build() (*WorkerApplication, error) {
	host, err := b.BuildHost()
	if err != nil {
		return nil, err
	}
	return newWorkerApplication(host, b.startFunc, b.stopFunc), nil
}

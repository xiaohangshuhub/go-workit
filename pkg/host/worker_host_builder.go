package host

import (
	"go.uber.org/fx"
)

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

func (b *WorkerHostBuilder) ConfigureServices(opts ...fx.Option) *WorkerHostBuilder {
	b.ApplicationHostBuilder.ConfigureServices(opts...)
	return b
}

func (b *WorkerHostBuilder) ConfigureAppConfiguration(fn func(builder ConfigBuilder)) *WorkerHostBuilder {
	b.ApplicationHostBuilder.ConfigureAppConfiguration(fn)
	return b
}

func (b *WorkerHostBuilder) AddBackgroundService(ctor interface{}) *WorkerHostBuilder {
	b.ApplicationHostBuilder.AddBackgroundService(ctor)
	return b
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

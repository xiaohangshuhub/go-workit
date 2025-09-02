package workit

// WorkerApplicationBuilder is a builder for creating a WorkerApplication.
type WorkerApplicationBuilder struct {
	*ApplicationBuilder
	startFunc func() error
	stopFunc  func() error
}

// NewWorkerAppBuilder creates a new WorkerApplicationBuilder.
func NewWorkerAppBuilder() *WorkerApplicationBuilder {
	return &WorkerApplicationBuilder{
		ApplicationBuilder: NewAppBuilder(),
	}
}

// OnStart sets the function to be called when the application starts.
func (b *WorkerApplicationBuilder) OnStart(fn func() error) *WorkerApplicationBuilder {
	b.startFunc = fn
	return b
}

// OnStop sets the function to be called when the application stops.
func (b *WorkerApplicationBuilder) OnStop(fn func() error) *WorkerApplicationBuilder {
	b.stopFunc = fn
	return b
}

// Build creates a new WorkerApplication.
func (b *WorkerApplicationBuilder) Build() *WorkerApplication {
	host := b.ApplicationBuilder.Build()
	return newWorkerApplication(host, b.startFunc, b.stopFunc)
}

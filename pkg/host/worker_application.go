package host

import "context"

type WorkerApplication struct {
	host      *ApplicationHost
	startFunc func() error
	stopFunc  func() error
}

var _ Application = (*WorkerApplication)(nil)

func newWorkerApplication(host *ApplicationHost, startFunc, stopFunc func() error) *WorkerApplication {
	return &WorkerApplication{
		host:      host,
		startFunc: startFunc,
		stopFunc:  stopFunc,
	}
}

func (app *WorkerApplication) Run(ctx context.Context) error {
	if err := app.host.Start(ctx); err != nil {
		return err
	}

	if app.startFunc != nil {
		if err := app.startFunc(); err != nil {
			return err
		}
	}

	<-ctx.Done()

	if app.stopFunc != nil {
		if err := app.stopFunc(); err != nil {
			return err
		}
	}

	return app.host.Stop(ctx)

}

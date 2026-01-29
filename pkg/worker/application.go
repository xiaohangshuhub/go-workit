package worker

import (
	"context"

	"github.com/xiaohangshu-dev/go-workit/pkg/app"
	"go.uber.org/fx"
)

// WorkerApplication 工作者应用
type WorkerApplication struct {
	*app.Application
	startFunc func() error
	stopFunc  func() error
}

// NewWorkerApplication 创建工作者应用
func newWorkerApplication(host *app.Application, startFunc, stopFunc func() error) *WorkerApplication {
	return &WorkerApplication{
		Application: host,
		startFunc:   startFunc,
		stopFunc:    stopFunc,
	}
}

// Run 运行应用
func (app *WorkerApplication) Run(ctx ...context.Context) {

	fxapp := app.FxApp(fx.New(app.Container()...))

	fxapp.Run()
}

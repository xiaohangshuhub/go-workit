package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/xiaohangshuhub/go-workit/pkg/app"
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
func (app *WorkerApplication) Run(ctx ...context.Context) error {
	var appCtx context.Context
	var cancel context.CancelFunc

	// 如果调用者未传递上下文，则创建默认上下文
	if len(ctx) == 0 || ctx[0] == nil {
		appCtx, cancel = context.WithCancel(context.Background())
		defer cancel()

		// 捕获系统信号，优雅关闭
		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			<-sigChan
			cancel()
		}()
	} else {
		// 使用调用者传递的上下文
		appCtx = ctx[0]
	}
	app.App = fx.New(app.Container()...)
	// 启动应用
	if err := app.Start(appCtx); err != nil {
		return err
	}

	// 执行自定义启动逻辑
	if app.startFunc != nil {
		if err := app.startFunc(); err != nil {
			return err
		}
	}

	// 等待上下文被取消
	<-appCtx.Done()

	// 执行自定义停止逻辑
	if app.stopFunc != nil {
		if err := app.stopFunc(); err != nil {
			return err
		}
	}

	// 停止应用
	return app.Stop(appCtx)
}

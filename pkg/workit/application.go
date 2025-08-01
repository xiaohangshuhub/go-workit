package workit

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Application struct {
	app        *fx.App
	config     *viper.Viper
	metrics    Metrics
	logger     *zap.Logger // 直接持有
	appoptions []fx.Option
}

func newApplication(options []fx.Option, config *viper.Viper, log *zap.Logger) *Application {
	metrics := newDefaultMetrics()

	opts := append(
		options,
		fx.Supply(config),
		fx.Supply(log),
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					metrics.Increment("application start")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					metrics.Increment("application stop")
					return nil
				},
			})
		}),
		//新增一个托管后台服务的调用：
		fx.Invoke(
			fx.Annotate(
				func(lc fx.Lifecycle, services []BackgroundService) {
					for _, svc := range services {
						svc := svc
						lc.Append(fx.Hook{
							OnStart: func(ctx context.Context) error {
								return svc.Start(ctx)
							},
							OnStop: func(ctx context.Context) error {
								return svc.Stop(ctx)
							},
						})
					}
				},
				fx.ParamTags(``, `optional:"true"`), // 声明可选参数 
			),
		),
	)

	return &Application{
		appoptions: opts,
		config:     config,
		metrics:    metrics,
		logger:     log,
	}
}

func (a *Application) Start(ctx context.Context) error {
	return a.app.Start(ctx)
}

func (a *Application) Stop(ctx context.Context) error {
	return a.app.Stop(ctx)
}

func (a *Application) Config() *viper.Viper {
	return a.config
}

func (a *Application) Metrics() Metrics {
	return a.metrics
}
func (a *Application) Logger() *zap.Logger {
	return a.logger
}

func (a *Application) Run(ctx ...context.Context) error {
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

	// 启动应用
	if err := a.Start(appCtx); err != nil {
		a.logger.Error("Failed to start application", zap.Error(err))
		return err
	}

	// 等待上下文被取消
	<-appCtx.Done()

	// 停止应用
	if err := a.Stop(appCtx); err != nil {
		a.logger.Error("Failed to stop application", zap.Error(err))
		return err
	}

	return nil
}

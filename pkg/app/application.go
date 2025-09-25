package app

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Application 应用
type Application struct {
	fxapp     *fx.App
	metrics   Metrics
	config    *viper.Viper
	logger    *zap.Logger // 直接持有
	container []fx.Option
}

// NewApplication 创建一个应用
func NewApplication(options []fx.Option, config *viper.Viper, log *zap.Logger) *Application {
	metrics := newDefaultMetrics()

	container := append(
		options,           // 容器选项
		fx.Supply(config), // config 实例
		fx.Supply(log),    // 日志实例
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
		container: container,
		config:    config,
		metrics:   metrics,
		logger:    log,
	}
}

// Config 获取配置实例
func (a *Application) Config() *viper.Viper {
	return a.config
}

// Metrics 获取指标实例
func (a *Application) Metrics() Metrics {
	return a.metrics
}

// Logger 获取日志实例
func (a *Application) Logger() *zap.Logger {
	return a.logger
}

// Run 运行应用
func (a *Application) Run(params ...string) {

	a.fxapp = fx.New(a.container...)
	a.fxapp.Run()

}

func (a *Application) AppendContainer(opts ...fx.Option) {
	a.container = append(a.container, opts...)
}

func (a *Application) Container() []fx.Option {
	return a.container
}

func (a *Application) FxApp(app *fx.App) *fx.App {

	a.fxapp = app

	return a.fxapp
}

package app

import (
	"context"
	"fmt"

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

	// 创建一个贯穿整个后台服务生命周期的根 Context 和它的取消函数
	bgCtx, bgCancel := context.WithCancel(context.Background())

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
		// 托管后台服务的调用
		fx.Invoke(
			func(lc fx.Lifecycle, log *zap.Logger, p struct {
				fx.In
				// 从组中安全读取服务切片
				Services []BackgroundService `group:"background_services"`
			}) {
				if len(p.Services) == 0 {
					return
				}

				for _, svc := range p.Services {
					svc := svc
					lc.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							log.Info("[BackgroundService] 正在启动后台服务...", zap.String("service", fmt.Sprintf("%T", svc)))
							go func() {
								defer func() {
									if r := recover(); r != nil {
										log.Error("[BackgroundService] 严重错误: 服务发生 Panic 崩溃",
											zap.String("service", fmt.Sprintf("%T", svc)),
											zap.Any("panic", r),
										)
									}
								}()
								// 使用我们在外面创建的长生命周期 bgCtx
								if err := svc.Start(bgCtx); err != nil {
									log.Error("[BackgroundService] 服务运行期间异常退出",
										zap.String("service", fmt.Sprintf("%T", svc)),
										zap.Error(err),
									)
								}
							}()
							return nil
						},
						OnStop: func(ctx context.Context) error {
							log.Info("[BackgroundService] 正在停止后台服务...", zap.String("service", fmt.Sprintf("%T", svc)))
							return svc.Stop(ctx)
						},
					})
				}

				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						log.Info("[BackgroundService] 正在广播关闭信号给所有后台服务...")
						bgCancel()
						return nil
					},
				})
			},
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

// AppendContainer 向容器中添加选项
func (a *Application) AppendContainer(opts ...fx.Option) {
	a.container = append(a.container, opts...)
}

// Container 获取容器选项
func (a *Application) Container() []fx.Option {
	return a.container
}

// FxApp 设置并返回 fx.App 实例
func (a *Application) FxApp(app *fx.App) *fx.App {
	a.fxapp = app
	return a.fxapp
}

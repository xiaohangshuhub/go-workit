package host

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ApplicationHost struct {
	app     *fx.App
	config  *viper.Viper
	metrics Metrics
	logger  *zap.Logger // 直接持有
}

func newApplicationHost(options []fx.Option, config *viper.Viper, log *zap.Logger) *ApplicationHost {
	metrics := newDefaultMetrics()

	opts := append(
		options,
		fx.Supply(config),
		fx.Supply(log),
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					metrics.Increment("host.start")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					metrics.Increment("host.stop")
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
				fx.ParamTags(``, `optional:"true"`), // <- 就加了这一行，解决问题！
			),
		),
	)

	return &ApplicationHost{
		app:     fx.New(opts...),
		config:  config,
		metrics: metrics,
		logger:  log,
	}
}

func (h *ApplicationHost) Start(ctx context.Context) error {
	return h.app.Start(ctx)
}

func (h *ApplicationHost) Stop(ctx context.Context) error {
	return h.app.Stop(ctx)
}

func (h *ApplicationHost) Config() *viper.Viper {
	return h.config
}

func (h *ApplicationHost) Metrics() Metrics {
	return h.metrics
}
func (h *ApplicationHost) Logger() *zap.Logger {
	return h.logger
}

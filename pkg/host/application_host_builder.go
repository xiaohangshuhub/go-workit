package host

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ApplicationHostBuilder struct {
	config        *viper.Viper
	options       []fx.Option
	configActions []func(ConfigBuilder)
	logger        *zap.Logger
}

func NewApplicationHostBuilder() *ApplicationHostBuilder {
	return &ApplicationHostBuilder{
		options: make([]fx.Option, 0),
	}
}

func (b *ApplicationHostBuilder) ConfigureAppConfiguration(fn func(builder ConfigBuilder)) *ApplicationHostBuilder {
	b.configActions = append(b.configActions, fn)
	return b
}

func (b *ApplicationHostBuilder) ConfigureServices(opts ...fx.Option) *ApplicationHostBuilder {
	b.options = append(b.options, opts...)
	return b
}

func (b *ApplicationHostBuilder) BuildHost() (*Application, error) {
	// 创建一个VIPER实例
	b.config = viper.New()

	// 创建配置构建器
	builder := newConfigBuilder(b.config)

	// 加载配置
	for _, action := range b.configActions {
		action(builder)
	}

	if err := b.config.ReadInConfig(); err != nil {
		// 配置文件不存在时跳过，不是错误
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// 配置 logger
	b.logger = newLogger(&Config{
		Level:      b.config.GetString("log.level"),    // 从配置文件/env/命令行拿
		Filename:   b.config.GetString("log.filename"), // 如果有配置文件路径则输出到文件
		MaxSize:    b.config.GetInt("log.max_size"),    // 单文件最大多少 MB，默认100
		MaxBackups: b.config.GetInt("log.max_backups"), // 保留几份备份，默认3
		MaxAge:     b.config.GetInt("log.max_age"),     // 最老的日志保留多少天，默认7
		Compress:   b.config.GetBool("log.compress"),   // 旧日志是否压缩，默认不开
		Console:    b.config.GetBool("log.console"),    // 是否同时输出到控制台，开发环境一般要 true
	})

	b.config.WatchConfig()

	b.config.OnConfigChange(func(e fsnotify.Event) {
		b.logger.Info("Config file changed", zap.String("file", e.Name))
	})

	return newApplicationHost(b.options, b.config, b.logger), nil
}

func (b *ApplicationHostBuilder) AddBackgroundService(ctor interface{}) *ApplicationHostBuilder {
	b.options = append(b.options, fx.Provide(ctor))
	return b
}

func (b *ApplicationHostBuilder) ConfigureOptions(provider interface{}) *ApplicationHostBuilder {
	b.options = append(b.options, fx.Provide(provider))
	return b
}

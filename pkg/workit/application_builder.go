package workit

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ApplicationBuilder struct {
	config        *viper.Viper
	options       []fx.Option
	logger        *zap.Logger
	configBuilder ConfigBuilder
}

func NewAppBuilder() *ApplicationBuilder {

	// 创建一个新的 viper 实例
	viper := viper.New()

	// 设置默认值
	viper.SetDefault("log.level", "info")              // 默认日志级别为 info
	viper.SetDefault("log.filename", "./logs/app.log") // 默认不输出到文件
	viper.SetDefault("log.max_size", 100)              // 默认单文件最大 100 MB
	viper.SetDefault("log.max_backups", 3)             // 默认保留 3 个备份
	viper.SetDefault("log.max_age", 7)                 // 默认日志保留 7 天
	viper.SetDefault("log.compress", true)             // 默认不压缩旧日志
	viper.SetDefault("log.console", true)              // 默认输出到控制台

	// 创建配置构建器
	configBuilder := newConfigBuilder(viper)

	return &ApplicationBuilder{
		config:        viper,
		options:       make([]fx.Option, 0),
		configBuilder: configBuilder,
	}
}

func (b *ApplicationBuilder) AddConfig(fn func(builder ConfigBuilder)) *ApplicationBuilder {

	// 先加载文件配置
	fn(b.configBuilder)

	// 加载环境变量
	b.configBuilder.addEnvironmentVariables()

	// 加载命令行参数
	b.configBuilder.addCommandLine()

	if err := b.config.ReadInConfig(); err != nil {
		// 配置文件不存在时跳过，不是错误
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}

	return b
}

func (b *ApplicationBuilder) AddServices(opts ...fx.Option) *ApplicationBuilder {
	b.options = append(b.options, opts...)
	return b
}

// Build 构建应用实例
//
// 参数:
//   - 无
//
// 返回:
//   - (*Application, error): 应用实例和错误信息
func (b *ApplicationBuilder) Build() (*Application, error) {

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

	// 监听配置文件变化(暂未实现)
	b.config.WatchConfig()

	b.config.OnConfigChange(func(e fsnotify.Event) {
		b.logger.Info("Config file changed", zap.String("file", e.Name))
	})

	return newApplication(b.options, b.config, b.logger), nil
}

// AddBackgroundService 添加后台服务
//
// 参数:
//   - ctor: BackgroundService 实现,启动阶段执行
func (b *ApplicationBuilder) AddBackgroundService(ctor interface{}) *ApplicationBuilder {
	b.options = append(b.options, fx.Provide(ctor))
	return b
}

// ConfigureOptions 配置选项(暂未实现)
func (b *ApplicationBuilder) ConfigureOptions(provider interface{}) *ApplicationBuilder {
	b.options = append(b.options, fx.Provide(provider))
	return b
}

// Config 返回配置实例,配置阶段也可读取配置
//
// 参数:
//   - 无
//
// 返回:
//   - *viper.Viper: 配置实例
func (b *ApplicationBuilder) Config() *viper.Viper {
	return b.config
}

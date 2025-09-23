package web

import (
	"github.com/spf13/viper"
	"github.com/xiaohangshuhub/go-workit/pkg/app"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// web config
type InstanceConfig struct {
	Applicaton   *app.Application // 应用实例
	Logger       *zap.Logger      // 日志管理器
	Config       *viper.Viper     // 配置管理器
	Container    []fx.Option      // 依赖注入容器
	RouterConfig RouterConfig     // 路由
}

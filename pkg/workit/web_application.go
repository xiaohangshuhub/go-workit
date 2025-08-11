package workit

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ServerOptions struct {
	HttpPort    string `mapstructure:"http_port"`
	GrpcPort    string `mapstructure:"grpc_port"`
	Environment string `mapstructure:"environment"`
}

type WebApplicationOptions struct {
	Config    *viper.Viper
	Logger    *zap.Logger
	Container []fx.Option
}

type EnvironmentOptions struct {
	IsDevelopment bool   // 是否开发环境
	Env           string // 环境名称

}

type WebApplication interface {
	Run(ctx ...context.Context) error
	MapRoutes(registerFunc interface{}) WebApplication
	UseSwagger() WebApplication
	UseCORS(interface{}) WebApplication
	UseStaticFiles(urlPath, root string) WebApplication
	UseHealthCheck() WebApplication
	MapGrpcServices(constructors ...interface{}) WebApplication
	UseMiddleware(constructors ...interface{}) WebApplication
	UseAuthentication() WebApplication
	UseAuthorization() WebApplication
	Logger() *zap.Logger
	Config() *viper.Viper
	Env() *EnvironmentOptions
}

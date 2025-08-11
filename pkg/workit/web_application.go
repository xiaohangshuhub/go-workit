package workit

import (
	"context"

	"github.com/gin-contrib/cors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type EnvironmentOptions struct {
	IsDevelopment bool   // 是否开发环境
	Env           string // 环境名称

}

type WebApplication interface {
	Run(ctx ...context.Context) error
	MapRoutes(registerFunc interface{}) WebApplication
	UseSwagger() WebApplication
	UseCORS(fn func(*cors.Config)) WebApplication
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

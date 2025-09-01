package workit

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type WebApplication interface {
	Run()
	MapRoutes(interface{}) WebApplication
	UseSwagger() WebApplication
	UseCORS(interface{}) WebApplication
	UseStaticFiles(urlPath, root string) WebApplication
	UseHealthCheck() WebApplication
	MapGrpcServices(...interface{}) WebApplication
	UseMiddleware(...interface{}) WebApplication
	UseAuthentication() WebApplication
	UseAuthorization() WebApplication
	Environment() *Environment
	Logger() *zap.Logger
	Config() *viper.Viper
}

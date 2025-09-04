package workit

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// WebApplication is the interface that defines the behavior of a web application.
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
	Logger() *zap.Logger
	Config() *viper.Viper
	Env() *Environment
	UseRecovery() WebApplication
	UseLogger() WebApplication
}

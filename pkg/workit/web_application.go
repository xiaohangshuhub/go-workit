package workit

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// WebApplication is the interface that defines the behavior of a web application.
type WebApplication interface {
	Run()
	MapRouter(interface{}) WebApplication
	MapGrpcServices(...interface{}) WebApplication
	Use(...interface{}) WebApplication
	UseSwagger() WebApplication
	UseCORS(interface{}) WebApplication
	UseStaticFiles(urlPath, root string) WebApplication
	UseHealthCheck() WebApplication
	UseAuthentication() WebApplication
	UseAuthorization() WebApplication
	UseRecovery() WebApplication
	UseLogger() WebApplication
	Logger() *zap.Logger
	Config() *viper.Viper
	Env() *Environment
}
